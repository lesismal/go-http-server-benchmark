package main

import (
	"flag"
	"fmt"
	"log"
	"net"

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

	listeners := make([]net.Listener, 50)
	for i := 0; i < 50; i++ {
		addr := fmt.Sprintf(":%v", *port+i)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("listen failed: %v", err)
		}
		listeners[i] = ln
	}

	for idx, ln := range listeners {
		log.Printf("fasthttp server[%v] running on: %v", idx, ln.Addr().String())
		go log.Fatalf("fasthttp server[%v] exit: %v", idx, fasthttp.Serve(ln, onEcho))
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
