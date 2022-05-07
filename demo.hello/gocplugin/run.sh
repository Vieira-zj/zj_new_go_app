#!/bin/bash
set -eu

function run_golint {
    local root_dir=$(pwd)
    local enable_rules="asciicheck,rowserrcheck,exportloopref,unconvert,unparam,prealloc,gocyclo,gocognit,gosec,bodyclose,sqlclosecheck,dupl"
    local disable_rules="lll"
    local timeout="3m0s"
    local output="/tmp/test/golint_output.json"
    local rules="-E=${enable_rules} -D=${disable_rules}"
    golangci-lint run ${root_dir} --enable-all --sort-results=true --tests=false --timeout=${timeout} --out-format json | tee ${output} | jq .Issues
}

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

# run_golint
run_goc_report

echo "done"
