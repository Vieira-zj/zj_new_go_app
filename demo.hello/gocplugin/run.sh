#!/bin/bash
set -eu

function init {
    cp config/* /tmp/test
}

function run_goc_report {
    init
    go run main.go
}

function run_goc_watch_dog {
    init
    go run main.go -mode=watcher
}

run_goc_report
echo "done"
