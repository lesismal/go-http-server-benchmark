package main

import (
	"flag"
	"fmt"
	"io"
	"log"
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

	addrs := make([]string, 50)
	for i := 0; i < 50; i++ {
		addrs[i] = fmt.Sprintf(":%v", *port+i)
	}

	log.Printf("gin server running on: %v", addrs)
	for idx, addr := range addrs {
		router := gin.New()
		router.POST("/echo", onEcho)
		log.Printf("gin server[%v] running on: %v", idx, addr)
		go log.Fatalf("gin server[%v] exit: %v", idx, router.Run(addr))
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
