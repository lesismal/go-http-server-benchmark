package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/julienschmidt/httprouter"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
)

var port = flag.Int("p", 8000, "server addr")
var rpcPort = flag.Int("r", 9000, "rpc server addr")

func main() {
	flag.Parse()

	alog.SetLevel(alog.LevelNone)

	go func() {
		router := httprouter.New()
		router.POST("/echo", onEcho)
		server := http.Server{
			Addr:    fmt.Sprintf(":%v", *port),
			Handler: router,
		}
		log.Fatalf("http server exit: %v", server.ListenAndServe())
	}()

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
