#!/bin/bash
set -e

# desc: grpc client aop by ioc.

function proto_gen {
    cd api; protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./hello.proto
}

function run_server {
    cd server; go run main.go
}

function run_client {
    cd cmd; go run .
}

function list {
    iocli list
}

function watch {
    iocli watch go1_1711_demo/ioc.demo/autowire_grpc_client/cmd/service2.Impl2 Hello
}

# proto_gen
# run_server
run_client

# list
# watch

echo "done"
