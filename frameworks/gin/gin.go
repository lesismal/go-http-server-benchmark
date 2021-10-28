package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/gin-gonic/gin"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
)

var port = flag.Int("p", 8300, "server addr")
var rpcPort = flag.Int("r", 9003, "rpc server addr")

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

	router := gin.New()
	router.POST("/echo", onEcho)
	for idx, ln := range listeners {
		server := http.Server{
			Handler: router,
		}
		log.Printf("gin server[%v] running on: %v", idx, ln.Addr().String())
		go log.Fatalf("gin server [%v] exit: %v", idx, server.Serve(ln))
	}

	recorder := perf.NewRecorder("server@gin")

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

func onEcho(c *gin.Context) {
	if c.Request.Body != nil {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Fatalf("read body failed: %v", err)
		}
		c.Data(http.StatusOK, c.ContentType(), body)
	}
}
