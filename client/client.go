package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/cloudwego/kitex-benchmark/runner"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
)

var (
	port          = flag.Int("p", 8000, "server addr")
	rpcPort       = flag.Int("r", 9000, "rpc server addr")
	framework     = flag.String("f", "none", "framework name")
	connectionNum = flag.Int("c", 100, "connection num")
	total         = flag.Int64("n", 10000000, "total test time")
	bufsize       = flag.Int("b", 1024, "buffer size")
)

func main() {
	flag.Parse()

	alog.SetLevel(alog.LevelNone)

	client, err := arpc.NewClient(func() (net.Conn, error) {
		return net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%v", *rpcPort), time.Second*3)
	})
	if err != nil {
		log.Fatalf("NewClient failed: %v", err)
	}
	defer client.Stop()

	chTask := make(chan chan error, *connectionNum)

	conns := make([]*http.Client, *connectionNum)
	for i := 0; i < *connectionNum; i++ {
		client := &http.Client{}
		conns = append(conns, client)
		go func() {
			url := fmt.Sprintf("http://127.0.0.1:%v/echo", *port)
			request := make([]byte, *bufsize)
			for waitting := range chTask {
				rand.Read(request)
				req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(request))
				if err != nil {
					panic(err)
				}
				res, err := client.Do(req)
				if err != nil {
					log.Fatalf("Do failed: %v", err)
				}
				defer res.Body.Close()
				response, err := ioutil.ReadAll(res.Body)
				if !bytes.Equal(response, request) {
					log.Fatal("not equal")
				}
				waitting <- err
			}
		}()
	}

	r := runner.NewRunner()

	handler := func() error {
		waitting := make(chan error, 1)
		chTask <- waitting
		return <-waitting
	}

	r.Warmup(handler, *connectionNum, 100*1000)

	err = client.Call("action", "begin", nil, time.Second)
	if err != nil {
		log.Fatalf("call begain failed: %v", err)
	}

	recorder := perf.NewRecorder(fmt.Sprintf("client@%v", *framework))
	recorder.Begin()

	r.Run(*framework, handler, *connectionNum, *total, *bufsize, 0)

	recorder.End()

	serverReport := ""
	err = client.Call("action", "end", &serverReport, time.Second)
	if err != nil {
		log.Fatalf("call begain failed: %v", err)
	}
	fmt.Print(serverReport)

	recorder.Report()
	fmt.Printf("\n\n")
}
