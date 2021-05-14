package state

import (
	"encoding/hex"
	"fmt"
	"time"
)

func (st *State) Nodelist() []Plate {
	res := make([]Plate, 1, len(st.Nodes)+1)
	res[0] = Plate{Address: st.LocalIP, ID: st.ThisId, Name: st.ThisName, LastSeen: time.Now()}
	res[0].PendingJobs = st.PendingJobsList()
	res[0].DoneJobs = st.DoneJobsList()
	for k := range st.Nodes {
		res = append(res, st.Nodes[k])
	}
	return res
}

func (st *State) PendingJobsList() []JobLabel {
	mpd := []JobLabel{}
	for jid := range st.PendingJobs {
		mpd = append(mpd, JobLabel{ID: jid, Type: st.PendingJobs[jid].Type})
	}

	return mpd
}

func (st *State) DoneJobsList() []JobLabel {
	mdn := []JobLabel{}
	for jid := range st.DoneJobs {
		mdn = append(mdn, JobLabel{ID: jid, Type: st.DoneJobs[jid].Type})
	}
	return mdn
}

func (jb *Job) JobDetailsString() string {
	s := fmt.Sprintln("Job ID", jb.ID)
	s += fmt.Sprintln("Finished:", jb.Finished)
	switch jb.Type {

	case MPSignature:
		psig, _ := PartSigFromHEX(jb.Payload)
		s += fmt.Sprintf("Request to sign message: >>%s<<\n", psig.Plaintext)
		s += fmt.Sprintln("Requested by:", jb.AgentID)
		s += fmt.Sprintln("Suite:", psig.SuiteID)
		s += fmt.Sprintln("Responses from:")
		for i, r := range jb.PartialResults {
			psig2, _ := PartSigFromHEX(r.Result)
			s += fmt.Sprintln(i, ":")
			s += shortString(hex.EncodeToString(psig2.Signature), 50) + "\n"
		}
	case MPPublicKeyJT:
		ss := new(PointShare)
		ss.UnmarshalJSON([]byte(jb.Payload))
		s += fmt.Sprintln("Request to reassemble Public Key")
		s += fmt.Sprintln("Requested by:", jb.AgentID)
		s += fmt.Sprintln("Suite:", ss.SuiteID)
		s += fmt.Sprintln("Responses from")
		for i, r := range jb.PartialResults {
			ss.UnmarshalJSON([]byte(r.Result))
			hexpoint, _ := ss.P.MarshalBinary()
			s += fmt.Sprintf("%v: @%v %T %s\n", i, ss.E, ss.P, shortString(hex.EncodeToString(hexpoint), 50))
		}
	default:
		s += fmt.Sprintln("Request:", jb.Payload)
		s += fmt.Sprintln("Responses from")
		for i, r := range jb.PartialResults {
			lim := 32
			if lim > len(r.Result) {
				lim = len(r.Result)
			}
			s += fmt.Sprintln(i, r.Result[:lim], "...")
		}
	}
	s += fmt.Sprintln("Final result:", shortString(jb.FinalResult, 50))
	return s
}

func shortString(str string, limlen int) string {
	if limlen < 9 {
		return "No kidding!"
	}
	lng := len(str)
	if lng < limlen {
		return str
	}
	ofst := lng - limlen + 4
	return str[:(lng-ofst)/2] + "..." + str[(lng+ofst)/2:]
}
