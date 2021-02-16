package state

import (
	"fmt"
	"strconv"
)

func (st *State) ProcessTestJob(jb *Job) bool {
	resp := new(JobResult)
	resp.ResultID = ResultID{jb.JobID, st.ThisId}
	resp.Result = "OK fine"
	jb.AddPartialResult(resp)
	st.Results[resp.ResultID] = resp
	if jb.AgentID != st.ThisId {
		st.ResultToBroadcastQueue(resp, 1)
		return true
	}
	return false
}

func TestJobResultArrival(jb *Job) {
	if len(jb.partialResults) > 0 {

		for _, r := range jb.partialResults {
			if jb.AgentID != r.ResultID.AgentID {
				jb.finalResult = fmt.Sprintf("Response by %v: %v", r.ResultID.AgentID, r.Result)
				CurrentState.markAsDone(jb)
				break
			}

		}

	}
}

var test int

func (st *State) NewTestJob(payload string) *Job {

	j := Job{
		Type:                 TestJT,
		JobID:                "ID" + strconv.Itoa(test) + "f" + string(st.ThisId),
		Payload:              payload,
		AgentID:              st.ThisId,
		partialResults:       map[AgentID]*JobResult{},
		partialResultArrival: TestJobResultArrival,
	}
	test++
	return &j

}
