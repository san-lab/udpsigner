package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"

	"github.com/san-lab/udpsigner/cli"
	"github.com/san-lab/udpsigner/peers"
	"github.com/san-lab/udpsigner/rpc"
	"github.com/san-lab/udpsigner/state"
)

func main() {
	httpPort := flag.String("httpPort", "8100", "http port")
	rpcf := flag.Bool("withHttp", false, "enable service RPC endpoint")

	//wauth := flag.Bool("withAuth", true, "should Basic Authentication be enabled")
	flag.Parse()

	//This is to graciously serve the ^C signal - allow all registered routines to clean up
	interruptChan := make(chan os.Signal)
	wg := &sync.WaitGroup{}
	signal.Notify(interruptChan, os.Interrupt)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, "WaitGroup", wg)
	if *rpcf {
		rpc.StartRPC(*httpPort, ctx, cancel, interruptChan)
	}

	go peers.Initialize(state.CurrentState, ctx)
	cli.Top()
	wg.Wait()
}
