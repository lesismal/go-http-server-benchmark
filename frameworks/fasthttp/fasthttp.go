package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
	"github.com/valyala/fasthttp"
)

var port = flag.Int("p", 8400, "server addr")
var rpcPort = flag.Int("r", 9004, "rpc server addr")

func main() {
	flag.Parse()

	alog.SetLevel(alog.LevelNone)

	addrs := make([]string, 50)
	for i := 0; i < 50; i++ {
		addrs[i] = fmt.Sprintf(":%v", *port+i)
	}

	log.Printf("fasthttp server running on: %v", addrs)
	for idx, addr := range addrs {
		log.Printf("fasthttp server[%v] running on: %v", idx, addr)
		go log.Fatalf("fasthttp server[%v] exit: %v", idx, fasthttp.ListenAndServe(addr, onEcho))
	}

	recorder := perf.NewRecorder("server@fasthttp")

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

func onEcho(ctx *fasthttp.RequestCtx) {
	body := ctx.Request.Body()
	if len(body) > 0 {
		ctx.Write(body)
	}
}
