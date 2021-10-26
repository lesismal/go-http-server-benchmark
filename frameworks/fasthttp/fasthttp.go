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

var port = flag.Int("p", 8000, "server addr")
var rpcPort = flag.Int("r", 9000, "rpc server addr")

func main() {
	flag.Parse()

	alog.SetLevel(alog.LevelNone)

	go func() {
		log.Fatalf("gin server exit: %v", fasthttp.ListenAndServe(fmt.Sprintf(":%v", *port), onEcho))
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

func onEcho(ctx *fasthttp.RequestCtx) {
	body := ctx.Request.Body()
	if len(body) > 0 {
		ctx.Write(body)
	}
}
