package rpc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/san-lab/udpsigner/state"
	"github.com/san-lab/udpsigner/templates"
)

func StartRPC(httpPort string, ctx context.Context, cancel context.CancelFunc, interruptChan chan os.Signal) {
	fmt.Println("Starting http on port", httpPort)
	//Beware! This config means that all the static images - also the ones called from the templates -
	// have to be addressed as "/static/*", regardless of the location of the template
	fs := http.FileServer(http.Dir("static"))
	renderer = templates.NewRenderer()
	http.HandleFunc("/rpc", handleHttp)
	http.HandleFunc("/react", serveHTML)
	http.HandleFunc("/", fs.ServeHTTP)
	srv := http.Server{Addr: "0.0.0.0:" + httpPort}
	//This is to graciously serve the ^C signal - allow all registered routines to clean up
	go func() {
		select {
		case <-interruptChan:
			cancel()
			srv.Shutdown(context.TODO())
			return
		}
	}()

	go srv.ListenAndServe()
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(state.CurrentState.DumpState())

}

var renderer = templates.NewRenderer()

func serveHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	isSlash := func(c rune) bool { return c == '/' }
	f := strings.FieldsFunc(r.URL.Path, isSlash)
	if len(f) < 3 {
		fmt.Frintln(w, "Please, specify the template")
		return

	}
	tempname := f[2]

	w.Write([]byte("Templates here"))
}
