package main

import (
	"github.com/san-lab/udpsigner/cli"
	"github.com/san-lab/udpsigner/peers"
)

func main() {

	go peers.Initialize()
	cli.Top()

}
