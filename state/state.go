package state

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/san-lab/secretsplitcli/goethkey"
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
	KnownScalarShares   map[string][]*ScalarShare //First grouped by suite id

}

type ResultID struct {
	JobID   string
	AgentID AgentID
}

func (st *State) ComposeMessage() []byte {
	if st.DisableBroadcast {
		return nil
	}
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
	f.Timestamp = time.Now()

	st.SignFrame(&f)
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
	KnownScalarShares:   map[string][]*ScalarShare{},
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
	PartialResults       map[AgentID]*JobResult `json:"-"`
	FinalResult          string                 `json:"-"`
	Payload              string
	partialResultArrival func(*JobResult) `json:"-"`
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
	case MPSignature:
		if st.ProcessMPSignJob(jb) {
			st.markAsDone(jb)
		}
	default:
		fmt.Println("Unknown job type:", jb.Type)

	}

}

func (jb *Job) AddPartialResult(r *JobResult) bool {
	if jb.PartialResults == nil {
		jb.PartialResults = map[AgentID]*JobResult{}
	}
	jb.PartialResults[r.ResultID.AgentID] = r
	if jb.partialResultArrival != nil {
		jb.partialResultArrival(r)
	}
	return false
}

func (st *State) ImportKeyFile(filename string) (err error) {
	kf, err := goethkey.ReadAndProcessKeyfile(filename)
	if err != nil {
		return
	}
	ps, err := goethkey.Deserialize(kf.Plaintext)
	if err != nil {
		return
	}
	st.ThisEvaluationPoint = st.suite.G2().Scalar().SetInt64(int64(ps.I + 1))
	st.ThisSecretValue = ps.V
	st.ThisPublicKey = st.suite.G2().Point().Mul(ps.V, nil)

	return
}

func (st *State) ImportShareFile(filename string) (err error) {
	kf, err := goethkey.ReadAndProcessKeyfile(filename)
	if err != nil {
		return
	}
	if kf.ID[:8] != goethkey.SplitHeader {
		err = fmt.Errorf("Not a Sharefile ID")
		return
	}
	T, err := strconv.Atoi(kf.ID[8:10])
	if err != nil {
		return
	}
	sID := kf.ID[13:]
	ps, err := goethkey.Deserialize(kf.Plaintext)
	if err != nil {
		return
	}
	scs := new(ScalarShare)
	scs.T = T
	scs.SuiteID = sID
	scs.V = ps.V
	scs.E = st.suite.G2().Scalar().SetInt64(int64(ps.I + 1))
	//TODO Check and prevent submitting duplicates
	st.KnownScalarShares[sID] = append(st.KnownScalarShares[sID], scs)

	return
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

func (st *State) SetPrivKeyBytes(b []byte) {
	G2 := st.suite.G2()
	st.ThisSecretValue = G2.Scalar().SetBytes(b)
	st.ThisPublicKey = G2.Point().Mul(st.ThisSecretValue, nil)
}
