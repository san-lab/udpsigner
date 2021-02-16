package state

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"go.dedis.ch/kyber/v3"

	"go.dedis.ch/kyber/v3/pairing"
)

type State struct {
	ThisName            string
	ThisId              AgentID
	ThisPassword        string
	ThisEvaluationPoint kyber.Scalar
	ThisSecretValue     kyber.Scalar
	ThisPublicKey       kyber.Point
	DisableBroadcast    bool
	suite               pairing.Suite
	PendingJobs         map[string]*Job
	DoneJobs            map[string]*Job
	Results             map[ResultID]*JobResult
	JobBroadcast        map[string]int
	ResultBroadcast     map[ResultID]int
}

type ResultID struct {
	JobID   string
	AgentID AgentID
}

func (st *State) ComposeMessage() []byte {
	f := Frame{}
	if len(st.ThisName) > 0 {
		f.SenderName = st.ThisName
	} else {
		f.SenderName = "not set"
	}
	f.SenderID = st.ThisId
	f.JobRequests = []*Job{}
	f.JobResults = []*JobResult{}
	for k, v := range st.JobBroadcast {
		j, isajob := st.PendingJobs[k]
		if !isajob {
			fmt.Println("Removing nonexistent Job " + k + " from the BroadcastQueue")
			delete(st.JobBroadcast, k)
			continue
		} else {
			if v < 1 {
				//fmt.Println("Broadcast repeat < 1, deleting ", k)
				delete(st.JobBroadcast, k)
				continue
			}
			st.JobBroadcast[k] = v - 1
			f.JobRequests = append(f.JobRequests, j)
		}

	}
	for k, v := range st.ResultBroadcast {
		result, is := st.Results[k]
		if !is {
			fmt.Println("Removing nonexistent Result " + k.JobID + " from the ResultBroadcast")
			delete(st.ResultBroadcast, k)
			continue
		} else {
			if v < 1 {
				//fmt.Println("Broadcast repeat < 1, deleting ", k)
				delete(st.ResultBroadcast, k)
				continue
			}
			st.ResultBroadcast[k] = v - 1
			f.JobResults = append(f.JobResults, result)
		}

	}
	b, e := json.Marshal(f)
	if e != nil {
		fmt.Println(e)
		return []byte("Error State")
	}

	return b
}

var src = rand.NewSource(time.Now().UnixNano())
var CurrentState = State{
	suite:               pairing.NewSuiteBn256(),
	PendingJobs:         map[string]*Job{},
	Results:             map[ResultID]*JobResult{},
	JobBroadcast:        map[string]int{},
	ResultBroadcast:     map[ResultID]int{},
	ThisId:              AgentID(fmt.Sprint(src.Int63())),
	DoneJobs:            map[string]*Job{},
	ThisEvaluationPoint: pairing.NewSuiteBn256().G1().Scalar().One(),
}

func (st *State) AddJobRequest(jr *Job) {

	if st.PendingJobs == nil {
		st.PendingJobs = map[string]*Job{}
	}
	if _, known := st.PendingJobs[jr.JobID]; !known {

		st.PendingJobs[jr.JobID] = jr
	}

}

func (st *State) AddJobResult(r *JobResult) {
	r.Arrived = time.Now()
	st.Results[r.ResultID] = r
	fmt.Println("arrived", r.ResultID)
	if j, is := st.PendingJobs[r.ResultID.JobID]; is {
		fmt.Println("found job", j)
		j.AddPartialResult(r)
	} else {
		fmt.Println("no job for", r.ResultID)
	}
}

type JobType string

const TestJT = JobType("TestJob")
const MPPublicKeyJT = JobType("MPPublicKey")
const MPSignature = JobType("MPSignature")
const MPPrivateKey = JobType("MPPrivateKey")

type Job struct {
	JobID                string
	AgentID              AgentID
	Type                 JobType
	Accepted             time.Time `json:"-"`
	Finished             bool
	FinishedAt           time.Time `json:"-"`
	success              bool
	Error                string
	partialResults       map[AgentID]*JobResult `json:"-"`
	finalResult          string
	Payload              string
	partialResultArrival func(*Job) `json:"-"`
}

func SetRandomKey() {
	G2 := CurrentState.suite.G2()
	CurrentState.ThisSecretValue = G2.Scalar().Pick(CurrentState.suite.RandomStream())
	CurrentState.ThisPublicKey = G2.Point().Mul(CurrentState.ThisSecretValue, nil)
}

func publicFromPrivate(data []byte) (kyber.Point, error) {
	var G kyber.Group
	var S kyber.Scalar
	suite := pairing.NewSuiteBn256()

	S = suite.Scalar()
	e := S.UnmarshalBinary(data)
	if e != nil {
		return nil, e
	}
	G = suite.G2()
	Pb := G.Point().Mul(S, nil)
	return Pb, nil
}

//The return value marks if the Job should be marked as "Done" locally
func (st *State) ProcessJob(jb *Job) {

	switch jb.Type {
	case TestJT:
		if st.ProcessTestJob(jb) {
			st.markAsDone(jb)
		}
	case MPPublicKeyJT:
		if st.ProcessMPPubJob(jb) {
			st.markAsDone(jb)
		}
	default:
		fmt.Println("Unknown job type:", jb.Type)

	}

}

func (jb *Job) AddPartialResult(r *JobResult) bool {
	if jb.partialResults == nil {
		jb.partialResults = map[AgentID]*JobResult{}
	}
	jb.partialResults[r.ResultID.AgentID] = r
	if jb.partialResultArrival != nil {
		jb.partialResultArrival(jb)
	}
	return false
}

func (st *State) ResultToBroadcastQueue(jres *JobResult, retry int) {
	st.ResultBroadcast[jres.ResultID] = retry
}

func (st *State) markAsDone(jb *Job) {
	if _, pending := st.PendingJobs[jb.JobID]; !pending {
		return
	}
	CurrentState.DoneJobs[jb.JobID] = jb
	delete(CurrentState.PendingJobs, jb.JobID)
	jb.Finished = true
	jb.FinishedAt = time.Now()
}

func JobToBroadcastQueue(jb *Job, retry int) {
	CurrentState.PendingJobs[jb.JobID] = jb
	CurrentState.JobBroadcast[jb.JobID] = retry
}
