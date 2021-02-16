package peers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/schollz/peerdiscovery"

	"github.com/san-lab/udpsigner/state"
)

var Nodes = map[string]Plate{}

var Repeat = 3

type Plate struct {
	Name     string
	ID       state.AgentID
	Address  string
	LastSeen time.Time
}

func Initialize() (err error) {
	s := peerdiscovery.Settings{Limit: -1,
		PayloadFunc:      state.CurrentState.ComposeMessage,
		TimeLimit:        -1,
		Delay:            500 * time.Millisecond,
		Notify:           Incoming,
		DisableBroadcast: false}
	_, err = peerdiscovery.Discover(s)

	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func Incoming(d peerdiscovery.Discovered) {
	f := new(state.Frame)
	e := json.Unmarshal(d.Payload, &f)
	pl := Plate{}
	pl.Address = d.Address
	if e != nil {
		pl.Name = fmt.Sprint(e)
	} else {
		pl.ID = f.SenderID
		if len(f.SenderName) > 0 {
			pl.Name = f.SenderName
		} else {
			pl.Name = "Not set"
		}
		pl.LastSeen = time.Now()
	}
	Nodes[d.Address] = pl

	jrq := f.JobRequests
	for _, v := range jrq {
		if _, known := state.CurrentState.PendingJobs[v.JobID]; !known {
			fmt.Println("adding new job:", v.JobID)
			v.Accepted = time.Now()
			state.CurrentState.AddJobRequest(v)
		} else {
			fmt.Println("Already seen ", v.JobID)
		}
	}

	for _, r := range f.JobResults {
		state.CurrentState.AddJobResult(r)
	}

}
