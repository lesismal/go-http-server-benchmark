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

var port = flag.Int("p", 8000, "server addr")
var rpcPort = flag.Int("r", 9000, "rpc server addr")

func main() {
	flag.Parse()

	alog.SetLevel(alog.LevelNone)

	go func() {
		router := gin.New()
		router.POST("/echo", onEcho)
		log.Fatalf("gin server exit: %v", router.Run(fmt.Sprintf(":%v", *port)))
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

func onEcho(c *gin.Context) {
	if c.Request.Body != nil {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Fatalf("read body failed: %v", err)
		}
		c.Data(http.StatusOK, c.ContentType(), body)
	}
}
