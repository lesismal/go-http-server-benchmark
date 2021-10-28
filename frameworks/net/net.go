package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/julienschmidt/httprouter"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
)

var port = flag.Int("p", 8200, "server addr")
var rpcPort = flag.Int("r", 9002, "rpc server addr")

func main() {
	flag.Parse()

	alog.SetLevel(alog.LevelNone)

	listeners := make([]net.Listener, 50)
	for i := 0; i < 50; i++ {
		addr := fmt.Sprintf(":%v", *port+i)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("listen failed: %v", err)
		}
		listeners[i] = ln
	}

	router := httprouter.New()
	router.POST("/echo", onEcho)
	for idx, ln := range listeners {
		server := http.Server{
			Handler: router,
		}
		log.Printf("net server[%v] running on: %v", idx, ln.Addr().String())
		go log.Fatalf("net server [%v] exit: %v", idx, server.Serve(ln))
	}

	recorder := perf.NewRecorder("server@net")

	rpcSvr := arpc.NewServer()
	rpcSvr.Handler.Handle("action", func(ctx *arpc.Context) {
		cmd := ""
		ctx.Bind(&cmd)
		switch cmd {
		case "begin":
			recorder.Begin()
			ctx.Write(nil)
		case "end":
			recorder.End()
			ctx.Write(recorder.ReportString())
		}
	})
	defer rpcSvr.Stop()

	log.Fatal(rpcSvr.Run(fmt.Sprintf(":%v", *rpcPort)))
}

func onEcho(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Body != nil {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatalf("read body failed: %v", err)
		}
		w.Write(body)
	}
}
