package main

import (
	"github.com/san-lab/udpsigner/cli"
	"github.com/san-lab/udpsigner/peers"
	"github.com/san-lab/udpsigner/state"
)

func main() {

	go peers.Initialize(&state.CurrentState)
	cli.Top()

}
