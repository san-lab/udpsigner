package state

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/dedis/kyber/sign/bls"
	"github.com/manifoldco/promptui"
	"go.dedis.ch/kyber/v3"
)

func (st *State) ProcessMPSignJob(jb *Job) bool {

	signreq := new(PartSignedString)

	signreq, err := PartSigFromHEX(jb.Payload)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println(*signreq)
	shares := st.KnownScalarShares[signreq.SuiteID]
	if len(shares) == 0 {
		fmt.Println("No such shares")
		fmt.Println(signreq.SuiteID)
		return jb.AgentID != st.ThisId
	}
	for _, lps := range shares {
		locRes := new(PartSignedString)
		locRes.Signature, err = bls.Sign(st.suite, lps.V, []byte(signreq.Plaintext))
		if err != nil {
			fmt.Println(err)
			return false
		}
		locRes.E = lps.E
		locRes.T = lps.T
		locRes.SuiteID = lps.SuiteID

		resp := new(JobResult)
		resp.ResultID = ResultID{jb.JobID, st.ThisId}
		b := PartSigToHEX(locRes)
		resp.Result = b
		jb.AddPartialResult(resp)
		st.Results[resp.ResultID] = resp
		if jb.AgentID != st.ThisId {
			st.ResultToBroadcastQueue(resp, 1)
		}
	}

	return jb.AgentID != st.ThisId // Manual action finished job for a responder

}

func NewMPSignJobStart() {

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

		prpt2 := promptui.Prompt{
			Label: "Enter text to be signed",
		}
		res, _ = prpt2.Run()

		err := NewMPSignJob(suites[i], res)
		if err != nil {
			fmt.Println(err)
		}
		return

	}
}

type PartSignedString struct {
	Plaintext string
	SuiteID   string
	E         kyber.Scalar
	T         int
	Signature []byte
}

func PartSigToHEX(psig *PartSignedString) string {
	//First T 1
	buf := []byte{byte(psig.T)}
	//Then E 32
	be, _ := psig.E.MarshalBinary()
	buf = append(buf, be...)
	//Then Signature 64
	buf = append(buf, psig.Signature...)
	//Then len of SuiteID 1
	bid := []byte(psig.SuiteID)
	buf = append(buf, byte(len(bid)))
	//Then SuiteID
	buf = append(buf, bid...)
	//Them Plaintext
	buf = append(buf, []byte(psig.Plaintext)...)

	return hex.EncodeToString(buf)
}

func PartSigFromHEX(hxs string) (*PartSignedString, error) {
	hx, err := hex.DecodeString(hxs)
	if err != nil {
		return nil, err
	}
	if len(hx) < 99 {
		return nil, fmt.Errorf("Too few bytes: %v", len(hx))
	}
	psig := new(PartSignedString)
	psig.T = int(hx[0])
	psig.E = CurrentState.suite.G1().Scalar()
	err = psig.E.UnmarshalBinary(hx[1:33])
	if err != nil {
		return nil, err
	}
	psig.Signature = hx[33:97]
	idlen := int(hx[97])
	psig.SuiteID = string(hx[98 : 98+idlen])
	psig.Plaintext = string(hx[98+idlen:])
	return psig, nil
}

func NewMPSignJob(suiteID string, plaintext string) error {
	st := CurrentState
	shares := st.KnownScalarShares[suiteID]
	if len(shares) == 0 {
		return fmt.Errorf("No shares known for suiteID %s", suiteID)
	}

	sh := shares[0]

	j := Job{
		Type:                 MPSignature,
		JobID:                "ID" + strconv.Itoa(test) + "f" + string(st.ThisId),
		AgentID:              st.ThisId,
		PartialResults:       map[AgentID]*JobResult{},
		partialResultArrival: MPSignKeyResultArrival,
	}
	test++

	sigb, err := bls.Sign(CurrentState.suite, sh.V, []byte(plaintext))
	if err != nil {
		return err
	}

	psst := new(PartSignedString)
	psst.Plaintext = plaintext
	psst.Signature = sigb
	psst.SuiteID = sh.SuiteID
	psst.E = sh.E
	psst.T = sh.T

	j.Payload = PartSigToHEX(psst)

	//Adding local share as a partilal result
	pres := new(JobResult)
	pres.ResultID = ResultID{j.JobID, st.ThisId}
	pres.Arrived = time.Now()
	pres.Result = j.Payload
	j.PartialResults[st.ThisId] = pres

	JobToBroadcastQueue(&j, 1)
	return nil
}

func MPSignKeyResultArrival(jbr *JobResult) {
	jb, local := CurrentState.PendingJobs[jbr.ResultID.JobID]
	if !local {
		return
	}
	r, err := PartSigFromHEX(string(jbr.Result))
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(jb.PartialResults) < r.T {
		return
	}

	//Deserialize all shares
	//Elimintate duplicates of same eval point
	shares := []*PointShare{}
	//pss[r.E] = r
	for _, parr := range jb.PartialResults {

		ps := new(PointShare)
		psig, _ := PartSigFromHEX(parr.Result)
		ps.T = psig.T
		ps.E = psig.E
		ps.SuiteID = psig.SuiteID
		ps.P = CurrentState.suite.G1().Point()
		ps.P.UnmarshalBinary(psig.Signature)

		shares = append(shares, ps)
	}

	unique := PruneDupShares(shares)

	if len(unique) < r.T {
		return
	}

	PBL, err := RecoverSecretPoint(CurrentState.suite.G1(), unique, r.T)
	if err != nil {
		fmt.Println(err)
		return
	}
	binform, _ := PBL.MarshalBinary()
	jb.FinalResult = hex.EncodeToString(binform)
	CurrentState.markAsDone(jb)

}
