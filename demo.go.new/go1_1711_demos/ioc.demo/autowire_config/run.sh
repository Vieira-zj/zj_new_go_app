#!/bin/bash
set -e

# desc: yaml config ioc.

function gen {
    iocli gen
}

function run {
    cd cmd; go run .
}

# gen
run

echo "done"
