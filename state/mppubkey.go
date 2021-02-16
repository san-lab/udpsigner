package state

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
)

type MPPubKeyPayload struct {
	PartialKey string
	EvalPoint  string
	Threshold  int
}

func (st *State) ProcessMPPubJob(jb *Job) bool {
	byt1 := []byte(jb.Payload)
	mppldreq := new(MPPubKeyPayload)
	mppldres := new(MPPubKeyPayload)

	err := json.Unmarshal(byt1, mppldreq)
	if err != nil {
		fmt.Println(err)
		return false
	}

	resp := new(JobResult)
	resp.ResultID = ResultID{jb.JobID, st.ThisId}

	mppldres.Threshold = mppldreq.Threshold
	buf, _ := st.ThisEvaluationPoint.MarshalBinary()
	mppldres.EvalPoint = hex.EncodeToString(buf)
	buf, _ = st.ThisPublicKey.MarshalBinary()
	mppldres.PartialKey = hex.EncodeToString(buf)

	byt2, err := json.Marshal(mppldres)
	if err != nil {
		fmt.Println(err)
		return false
	}
	resp.Result = string(byt2)
	jb.AddPartialResult(resp)
	st.Results[resp.ResultID] = resp
	if jb.AgentID != st.ThisId {
		st.ResultToBroadcastQueue(resp, 1)
		return true
	}
	return false

}

func NewMPPubJob() {
	st := CurrentState
	if st.ThisSecretValue == nil {
		fmt.Println("Keys not initialized")
		return
	}
	j := Job{
		Type:                 MPPublicKeyJT,
		JobID:                "ID" + strconv.Itoa(test) + "f" + string(st.ThisId),
		AgentID:              st.ThisId,
		partialResults:       map[AgentID]*JobResult{},
		partialResultArrival: MPPubKeyResultArrival,
	}
	test++

	mppld := new(MPPubKeyPayload)
	mppld.Threshold = 3
	buf, _ := st.ThisEvaluationPoint.MarshalBinary()
	mppld.EvalPoint = hex.EncodeToString(buf)
	buf, _ = st.ThisPublicKey.MarshalBinary()
	mppld.PartialKey = hex.EncodeToString(buf)
	b, err := json.Marshal(mppld)
	if err != nil {
		fmt.Println(err)
		return
	}
	j.Payload = string(b)
	fmt.Println(j)

	st.ProcessTestJob(&j)
	JobToBroadcastQueue(&j, 1)
}

func MPPubKeyResultArrival(jb *Job) {
	for _, r := range jb.partialResults {
		fmt.Println(r.Result)
	}
	return
}
