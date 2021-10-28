package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/cloudwego/kitex-benchmark/runner"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
	"github.com/lesismal/nbio/mempool"
	"github.com/lesismal/nbio/nbhttp"
)

var (
	port        = flag.Int("p", 8000, "server addr")
	rpcPort     = flag.Int("r", 9000, "rpc server addr")
	framework   = flag.String("f", "none", "framework name")
	connections = flag.Int("connections", 100, "connection num")
	concurrency = flag.Int("concurrency", 100, "connection num")
	total       = flag.Int64("n", 10000000, "total test time")
	bufsize     = flag.Int("b", 1024, "buffer size")
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

	engine := nbhttp.NewEngine(nbhttp.Config{})

	err = engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer engine.Stop()

	httpClients := []*nbhttp.Client{}
	println("connections:", *connections)
	println("concurrency:", *concurrency)
	for i := 0; i < *connections; i++ {
		httpClients = append(httpClients, &nbhttp.Client{
			Engine:          engine,
			Timeout:         time.Second * 5,
			MaxConnsPerHost: 1,
		})
	}

	r := runner.NewRunner()

	var reqIndex uint32
	handler := func() error {
		idx := atomic.AddUint32(&reqIndex, 1)
		url := fmt.Sprintf("http://127.0.0.1:%v/echo", uint32(*port)+(idx%50))
		httpClient := httpClients[idx%uint32(len(httpClients))]
		waitting := make(chan error, 1)
		request := mempool.Malloc(*bufsize)
		response := mempool.Malloc(*bufsize)
		rand.Read(request)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(request))
		if err != nil {
			panic(err)
		}
		httpClient.Do(req, func(res *http.Response, conn net.Conn, err error) {
			defer mempool.Free(request)
			defer mempool.Free(response)
			if err != nil {
				log.Fatalf("Do failed: %v", err)
			}
			defer res.Body.Close()
			_, err = res.Body.Read(response)
			if err != nil {
				log.Fatalf("Do failed: %v", err)
			}
			if !bytes.Equal(response, request) {
				log.Fatal("not equal")
			}

			waitting <- err
		})
		return <-waitting
	}

	r.Warmup(handler, *concurrency, 100*1000)

	err = client.Call("action", "begin", nil, time.Second)
	if err != nil {
		log.Fatalf("call begain failed: %v", err)
	}

	recorder := perf.NewRecorder("client@nbio")
	recorder.Begin()

	r.Run(*framework, handler, *concurrency, *total, *bufsize, 0)

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
