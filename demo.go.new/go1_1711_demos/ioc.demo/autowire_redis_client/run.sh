#!/bin/bash
set -e

# desc: redis sdk ioc, and watch.

function gen {
    iocli gen
}

function run {
    cd cmd; go run .
}

function list {
    iocli list
}

function watch {
    iocli watch go1_1711_demo/ioc.demo/autowire_redis_client/sdk.Redis Get
}

# gen
run

# list
# watch

echo "done"
