#!/bin/bash
set -e

# desc: dubbo rpc ioc, and watch.

function run_server {
    cd server; go run .
}

function run_client {
    cd client; go run .
}

function gen {
    iocli gen
}

function watch {
    iocli watch go1_1711_demo/ioc.demo/autowire_rpc/server/pkg/service.ServiceStruct GetUser
}

echo "done"
