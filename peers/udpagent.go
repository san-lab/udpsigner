package peers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/schollz/peerdiscovery"

	"github.com/san-lab/udpsigner/state"
)

var FrameSamples = map[string][]state.Frame{}
var Sample = false

var Repeat = 3

var S peerdiscovery.Settings

func Initialize(state *state.State, ctx context.Context) (err error) {
	S = peerdiscovery.Settings{Limit: -1,
		PayloadFunc:      state.ComposeMessage,
		TimeLimit:        -1,
		Delay:            1000 * time.Millisecond,
		Notify:           Incoming,
		DisableBroadcast: state.DisableBroadcast,
	}
	_, err = peerdiscovery.Discover(S)

	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func Incoming(d peerdiscovery.Discovered) {
	if len(d.Payload) == 0 {
		return
	}
	f := new(state.Frame)

	e := json.Unmarshal(d.Payload, &f)
	pl := state.Plate{}
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
	pl.PendingJobs = f.MyPendingJobs
	pl.DoneJobs = f.MyDoneJobs
	state.CurrentState.Nodes[d.Address] = pl
	if Sample {
		FrameSamples[d.Address] = append(FrameSamples[d.Address], *f)
	}

	jrq := f.JobRequests
	for _, v := range jrq {
		if _, known := state.CurrentState.PendingJobs[v.ID]; !known {
			fmt.Println("adding new job:", v.ID)
			v.Accepted = time.Now()
			state.CurrentState.AddJobRequest(v)
		} else {
			fmt.Println("Already seen ", v.ID)
		}
	}

	for _, r := range f.JobResults {
		state.CurrentState.AddJobResult(r)
	}

}

func DoSample(dur time.Duration) {
	FrameSamples = map[string][]state.Frame{}
	go func() {
		Sample = true
		time.Sleep(dur)
		Sample = false
	}()
}
