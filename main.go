// This is a client exposes a HTTP streaming interface to NSQ channels

package main

import (
	_ "net/http/pprof"
	//	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"strings"

	"app"
)

var (
	httpAddress      = flag.String("http-address", "0.0.0.0:8080", "<addr>:<port> to listen on for HTTP clients")
	maxInFlight      = flag.Int("max-in-flight", 100, "max number of messages to allow in flight")
	lookupdHTTPAddrs = app.StringArray{}
	timeout          = flag.Int("timeout", 10, "return within N seconds if maxMessages not reached")
	maxMessages      = flag.Int("max-messages", 1, "return if got N messages in a single poll")
)

func init() {
	flag.Var(&lookupdHTTPAddrs, "lookupd-http-address", "lookupd HTTP address (may be given multiple times)")
}

func (s *StreamServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/stats":
		StatsHandler(w, req)
	case "/sub":
		SubHandler(w, req)
	case "/pub":
		PubHandler(w, req)
	default:
		OpHandler(w, req)
	}
}

func main() {
	flag.Parse()

	if *maxInFlight <= 0 {
		log.Fatalf("--max-in-flight must be > 0")
	}

	if len(lookupdHTTPAddrs) == 0 {
		log.Fatalf("--lookupd-http-address required.")
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	httpAddr, err := net.ResolveTCPAddr("tcp", *httpAddress)
	if err != nil {
		log.Fatal(err)
	}

	httpListener, err := net.Listen("tcp", httpAddr.String())
	if err != nil {
		log.Fatalf("FATAL: listen (%s) failed - %s", httpAddr.String(), err.Error())
	}
	log.Printf("listening on %s", httpAddr.String())

	streamServer = &StreamServer{}
	streamServer.clients = make(map[uint]*StreamReader)

	server := &http.Server{Handler: streamServer}
	err = server.Serve(httpListener)

	// theres no direct way to detect this error because it is not exposed
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.Printf("ERROR: http.Serve() - %s", err.Error())
	}
}
