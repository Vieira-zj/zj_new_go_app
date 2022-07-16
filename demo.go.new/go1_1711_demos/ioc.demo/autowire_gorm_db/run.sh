#!/bin/bash
set -e

# desc: gorm sdk ioc, and watch.

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
    iocli watch go1_1711_demo/ioc.demo/autowire_gorm_db/sdk.GORMDB Find
}

# gen
run

# list
# watch

echo "done"
