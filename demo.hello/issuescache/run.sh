#!/bin/bash
set -ue

project_root_path="${HOME}/Workspaces/zj_repos/zj_go2_project/demo.hello"

function build_bin() {
    go build -o ${HOME}/Downloads/bins/cicd .
}

function run_golint() {
    local root_dir=$(pwd)
    local enable_rules="asciicheck,rowserrcheck,exportloopref,unconvert,unparam,prealloc,gocyclo,gocognit,gosec,bodyclose,sqlclosecheck,dupl"
    local disable_rules="lll"
    local timeout="3m0s"
    local output="/tmp/test/golint_output.json"
    golangci-lint run ${root_dir} -E="${enable_rules}" -D="${disable_rules}" --sort-results=true --tests=false --timeout=${timeout} --out-format json | tee ${output} | jq .Issues
}

function run_test() {
    case=$1
    go test -timeout 30s -run ^${case}$ demo.hello/cicd/pkg -v -count=1
}

function print_tickets() {
    go run main.go -cli -p 10 "fixVersion = apa_v1.0.26.20210607"
}

function run_server() {
    go run main.go -svc
}

# build_bin
# run_golint

# source ${project_root_path}/env.sh
# run_test TestIssueKeySort

# print_tickets
# run_server

echo "done"
