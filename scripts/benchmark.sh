#!/bin/bash

. ./scripts/env.sh

repo=("nbio" "net" "gin" "fasthttp")
ports=(8100 8200 8300 8400)
rpcports=(9001 9002 9003 9004)

. ./scripts/build.sh

connections=$1
concurrency=$2

# benchmark
for b in ${body[@]}; do
  for ((i = 0; i < ${#repo[@]}; i++)); do
    rp=${repo[i]}
    port=${ports[i]}
    rpcport=${rpcports[i]}
    # server start
    nohup $taskset_server ./output/bin/${rp}_reciever -p=${port} -r=${rpcport} >> output/log/nohup.log 2>&1 &
    sleep 2
    echo "server $rp running with $taskset_server"

    # run client
    echo "client $rp running with $taskset_client"
    $taskset_client ./output/bin/client_bencher -p=${port} -r=${rpcport} -f=${rp} -b=$b -connections=$connections -concurrency=$concurrency -n=$n

    # stop server
    pid=$(ps -ef | grep ${rp}_reciever | grep -v grep | awk '{print $2}')
    disown $pid
    kill -9 $pid
    sleep 1
  done
done
