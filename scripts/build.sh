#!/bin/bash

# clean
rm -rf output/ && mkdir -p output/bin/ && mkdir -p output/log/

# build servers
go build -v -o output/bin/nbio_reciever ./frameworks/nbio
go build -v -o output/bin/net_reciever ./frameworks/net
go build -v -o output/bin/gin_reciever ./frameworks/gin
go build -v -o output/bin/fasthttp_reciever ./frameworks/fasthttp
go build -v -o output/bin/client_bencher ./client

