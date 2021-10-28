package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/julienschmidt/httprouter"
	"github.com/lesismal/arpc"
	"github.com/lesismal/nbio/nbhttp"
)

var port = flag.Int("p", 8100, "server addr")
var rpcPort = flag.Int("r", 9000, "rpc server addr")

func main() {
	flag.Parse()

	addrs := make([]string, 50)
	for i := 0; i < 50; i++ {
		addrs[i] = fmt.Sprintf(":%v", *port+i)
	}

	router := httprouter.New()
	router.POST("/echo", onEcho)
	engine := nbhttp.NewEngine(nbhttp.Config{
		Network: "tcp",
		Addrs:   addrs,
		Handler: router,
	})

	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer engine.Stop()

	recorder := perf.NewRecorder("server@nbio")

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
	data := r.Body.(*nbhttp.BodyReader).RawBody()
	if len(data) > 0 {
		w.Write(data)
	}
}
