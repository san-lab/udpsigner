package state

import "time"

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
	return mnd
}
