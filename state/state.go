package state

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/san-lab/secretsplitcli/goethkey"

	"go.dedis.ch/kyber/v3"

	"go.dedis.ch/kyber/v3/pairing"
)

type State struct {
	StateSerializable
	ThisSecretValue   kyber.Scalar
	ThisPublicKey     kyber.Point
	suite             pairing.Suite             `json:"-"`
	Results           map[ResultID]*JobResult   `json:"-"`
	ResultBroadcast   map[ResultID]int          `json:"-"`
	KnownScalarShares map[string][]*ScalarShare //First grouped by suite id
	Nodes             map[string]Plate
	LocalIP           string
}

type StateSerializable struct {
	ThisName         string
	ThisId           AgentID
	ThisPassword     string
	DisableBroadcast bool
	PendingJobs      map[string]*Job
	DoneJobs         map[string]*Job
	JobBroadcast     map[string]int
	HTTPPort         string
}

type ResultID struct {
	JobID   string
	AgentID AgentID
}

func (rid *ResultID) String() string {
	return "{\"ID\":\"" + rid.JobID + "\",\"AgentID\":\"" + string(rid.AgentID) + "\"}"
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
	f.KnownKeyShares = [][]string{}
	for suid, su := range st.KnownScalarShares {

		for _, ss := range su {
			h, _ := ss.E.MarshalBinary()
			kns := []string{suid, "TBLS", hex.EncodeToString(h)}
			f.KnownKeyShares = append(f.KnownKeyShares, kns)
		}

	}

	f.Timestamp = time.Now()

	//Broadcast local queues
	f.MyPendingJobs = st.PendingJobsList()
	f.MyDoneJobs = st.DoneJobsList()

	st.SignFrame(&f)
	b, e := json.Marshal(f)
	if e != nil {
		fmt.Println(e)
		return []byte("Error State")
	}

	return b
}

var src = rand.NewSource(time.Now().UnixNano())
var CurrentState = NewState()

func NewState() *State {
	return &State{
		StateSerializable: StateSerializable{
			PendingJobs:  map[string]*Job{},
			JobBroadcast: map[string]int{},
			ThisId:       AgentID(fmt.Sprint(src.Int63())),
			DoneJobs:     map[string]*Job{},
		},
		suite:             pairing.NewSuiteBn256(),
		Results:           map[ResultID]*JobResult{},
		ResultBroadcast:   map[ResultID]int{},
		KnownScalarShares: map[string][]*ScalarShare{},
		Nodes:             map[string]Plate{},
		LocalIP:           GetOutboundIP(),
	}
}

func (s *State) MarshalJSONb() ([]byte, error) {
	jsn := []byte{'{'}

	jsn = append(jsn, '}')
	return jsn, nil
}

func (st *State) AddJobRequest(jr *Job) {

	if st.PendingJobs == nil {
		st.PendingJobs = map[string]*Job{}
	}
	if _, known := st.PendingJobs[jr.ID]; !known {

		st.PendingJobs[jr.ID] = jr
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
	ID                   string
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
	//st.ThisEvaluationPoint = st.suite.G2().Scalar().SetInt64(int64(ps.I + 1))
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
	if _, pending := st.PendingJobs[jb.ID]; !pending {
		return
	}
	CurrentState.DoneJobs[jb.ID] = jb
	delete(CurrentState.PendingJobs, jb.ID)
	jb.Finished = true
	jb.FinishedAt = time.Now()
}

func JobToBroadcastQueue(jb *Job, retry int) {
	CurrentState.PendingJobs[jb.ID] = jb
	CurrentState.JobBroadcast[jb.ID] = retry
}

func (st *State) SetPrivKeyBytes(b []byte) {
	G2 := st.suite.G2()
	st.ThisSecretValue = G2.Scalar().SetBytes(b)
	st.ThisPublicKey = G2.Point().Mul(st.ThisSecretValue, nil)
}

type PresentationObject struct {
	ThisName            string
	ThisId              AgentID
	ThisPassword        string
	ThisEvaluationPoint kyber.Scalar
	ThisSecretValue     kyber.Scalar
	ThisPublicKey       kyber.Point
	DisableBroadcast    bool
	suite               pairing.Suite
	PendingJobs         []*Job
	DoneJobs            []*Job
	Results             map[ResultID]*JobResult `json:"-"`
	JobBroadcast        map[string]int
	ResultBroadcast     map[ResultID]int          `json:"-"`
	KnownScalarShares   map[string][]*ScalarShare //First grouped by suite id
	Nodes               []Plate
	LocalIP             string
	HTTPPort            string
}

func (st *State) StateToPresentation() PresentationObject {
	pendingJobs := []*Job{}
	for _, value := range st.PendingJobs {
		pendingJobs = append(pendingJobs, value)
	}

	doneJobs := []*Job{}
	for _, value := range st.DoneJobs {
		doneJobs = append(doneJobs, value)
	}

	pendingJobLabels := []JobLabel{}
	for _, s := range pendingJobs {
		pendingJobLabels = append(pendingJobLabels, JobLabel{ID: s.ID, Type: s.Type})
	}
	doneJobLabels := []JobLabel{}
	for _, s := range doneJobs {
		doneJobLabels = append(doneJobLabels, JobLabel{ID: s.ID, Type: s.Type})
	}
	knownKeyShares := [][]string{}
	for _, s := range st.KnownScalarShares {
		for _, t := range s {
			byteSlice, _ := t.E.MarshalBinary()
			knownKeyShares = append(knownKeyShares, []string{t.SuiteID, "TBLS", hex.EncodeToString(byteSlice)})
		}
	}

	var name string
	if st.ThisName == "" {
		name = "not set"
	} else {
		name = st.ThisName
	}
	actualNodePlate := Plate{Name: name, ID: st.ThisId, Address: st.LocalIP, LastSeen: time.Now(), PendingJobs: pendingJobLabels, DoneJobs: doneJobLabels, KnownKeyShares: knownKeyShares}

	nodesArray := []Plate{}
	nodesArray = append(nodesArray, actualNodePlate)
	for _, value := range st.Nodes {
		nodesArray = append(nodesArray, value)
	}

	PresentationObject := PresentationObject{ThisName: st.ThisName, ThisId: st.ThisId, ThisPassword: st.ThisPassword, ThisPublicKey: st.ThisPublicKey, DisableBroadcast: st.DisableBroadcast, suite: st.suite, PendingJobs: pendingJobs, DoneJobs: doneJobs, Results: st.Results, JobBroadcast: st.JobBroadcast, ResultBroadcast: st.ResultBroadcast, KnownScalarShares: st.KnownScalarShares, Nodes: nodesArray, LocalIP: st.LocalIP, HTTPPort: st.HTTPPort}

	return PresentationObject
}

func (st *State) PresentationObject() []byte {

	presentObject := st.StateToPresentation()

	b, _ := json.MarshalIndent(presentObject, " ", " ")
	return b
}

type Plate struct {
	Name           string
	ID             AgentID
	Address        string
	LastSeen       time.Time
	PendingJobs    []JobLabel
	DoneJobs       []JobLabel
	KnownKeyShares [][]string //suiteID, Type, EvalPoint
}

func (st *State) PruneDoneJobs(senderid AgentID, doneJobs []JobLabel) {
	for _, jl := range doneJobs {
		j, is := st.PendingJobs[jl.ID]

		if is && j.AgentID == senderid {
			delete(st.PendingJobs, j.ID)
		}
	}
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func (st *State) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(st.StateSerializable)
	return buf.Bytes(), err
}

func (st *State) Deserialize(b []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	sts := new(StateSerializable)
	err := dec.Decode(sts)
	st.StateSerializable = *sts
	return err
}
