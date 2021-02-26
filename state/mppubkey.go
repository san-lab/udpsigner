package state

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/manifoldco/promptui"
)

func (st *State) ProcessMPPubJob(jb *Job) bool {
	if jb.AgentID == st.ThisId {
		return false
	}
	mppldreq := new(PointShare)

	err := mppldreq.UnmarshalJSON([]byte(jb.Payload))
	if err != nil {
		fmt.Println(err)
		return false
	}

	shares := st.KnownScalarShares[mppldreq.SuiteID]
	if len(shares) == 0 {
		fmt.Println("No such shares")
		fmt.Println(mppldreq.SuiteID)
		return true
	}
	for _, lps := range shares {
		locRes := ScalarShareToPointShare(lps)

		resp := new(JobResult)
		resp.ResultID = ResultID{jb.JobID, st.ThisId}
		b, _ := locRes.MarshalJSON()
		resp.Result = string(b)
		jb.AddPartialResult(resp)
		st.Results[resp.ResultID] = resp
		if jb.AgentID != st.ThisId {
			st.ResultToBroadcastQueue(resp, 1)
		}
	}

	return true

}

func NewMPPubJobStart() {
	for {
		suites := make([]string, len(CurrentState.KnownScalarShares))
		labels := make([]string, len(suites))
		i := 0
		for su, l := range CurrentState.KnownScalarShares {
			labels[i] = fmt.Sprintf("%s [%v of %v]", su, len(l), l[0].T)
			suites[i] = su
			i++
		}
		prpt := promptui.Select{
			Label: "Select the share to start an MP computation",
			Size:  len(suites) + 1,
			Items: append(labels, "Back"),
		}

		i, res, _ := prpt.Run()
		if res == "Back" {
			return
		}
		NewMPPubJob(suites[i])
		return

	}
}

func NewMPPubJob(suiteID string) error {
	st := CurrentState
	shares := st.KnownScalarShares[suiteID]
	if len(shares) == 0 {
		return fmt.Errorf("No shares known for suiteID %s", suiteID)
	}

	sh := shares[0]

	j := Job{
		Type:                 MPPublicKeyJT,
		JobID:                "ID" + strconv.Itoa(test) + "f" + string(st.ThisId),
		AgentID:              st.ThisId,
		PartialResults:       map[AgentID]*JobResult{},
		partialResultArrival: MPPubKeyResultArrival,
	}
	test++

	payload := ScalarShareToPointShare(sh)

	b, err := payload.MarshalJSON()
	if err != nil {
		return err
	}
	j.Payload = string(b)
	//Adding local share as a partilal result
	pres := new(JobResult)
	pres.ResultID = ResultID{j.JobID, st.ThisId}
	pres.Arrived = time.Now()
	pres.Result = string(b)
	j.PartialResults[st.ThisId] = pres

	JobToBroadcastQueue(&j, 1)
	return nil
}

func MPPubKeyResultArrival(jbr *JobResult) {
	jb, local := CurrentState.PendingJobs[jbr.ResultID.JobID]
	if !local {
		return
	}

	r := new(PointShare)
	r.UnmarshalJSON([]byte(jb.Payload))

	if len(jb.PartialResults) < r.T {
		return
	}

	//Deserialize all shares
	//Elimintate duplicates of same eval point
	shares := []*PointShare{}
	//pss[r.E] = r
	for _, parr := range jb.PartialResults {

		ps := new(PointShare)
		ps.UnmarshalJSON([]byte(parr.Result))
		shares = append(shares, ps)
	}

	unique := PruneDupShares(shares)

	if len(unique) < r.T {
		return
	}

	PBL, err := RecoverSecretPoint(CurrentState.suite.G2(), unique, r.T)
	if err != nil {
		fmt.Println(err)
		return
	}
	binform, _ := PBL.MarshalBinary()
	jb.FinalResult = hex.EncodeToString(binform)
	CurrentState.markAsDone(jb)

}
