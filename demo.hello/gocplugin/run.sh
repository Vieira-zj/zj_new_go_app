#!/bin/bash
set -eu

# local test
# root="/tmp/test/goc_test_space" 
# staging test
root="${HOME}/Downloads/data/goc_staging_space"

function run_golint {
    local root_dir=$(pwd)
    local enable_rules="asciicheck,rowserrcheck,exportloopref,unconvert,unparam,prealloc,gocyclo,gocognit,gosec,bodyclose,sqlclosecheck,dupl"
    local disable_rules="lll"
    local timeout="3m0s"
    local output="${root}/golint_output.json"
    local rules="-E=${enable_rules} -D=${disable_rules}"
    golangci-lint run ${root_dir} --enable-all --sort-results=true --tests=false --timeout=${timeout} --out-format json | tee ${output} | jq .Issues
}

function init {
    cp config/* ${root}
}

function run_goc_plugin {
    go run main.go -root=${root}
}

function clearup {
    set +e
    rm ${root}/sqlite.db
    rm -r ${root}/public
    rm ${root}/apa_echoserver/cover_data/*
    set -e
}

# run_golint

# clearup
# init
run_goc_plugin

echo "done"
