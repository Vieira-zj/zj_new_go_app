#!/bin/bash
set -ue

project_root_path="${HOME}/Workspaces/zj_repos/zj_go2_project/demo.hello"

function build_bin() {
    go build -o ${HOME}/Downloads/bins/cicd .
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

# source ${project_root_path}/env.sh
# build_bin
# run_test TestTicketsTreeV2

# print_tickets
# run_server

echo "done."
