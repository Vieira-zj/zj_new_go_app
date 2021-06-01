#!/bin/bash
set -ue

project_root_path="${HOME}/Workspaces/zj_repos/zj_go2_project/demo.hello"

function print_releasecycle_tickets() {
    go run main.go -releaseCycle='2021.04.v3 - AirPay'
}

function run_server() {
    go run main.go -server
}

function run_test() {
    case=$1
    go test -timeout 30s -run ^${case}$ demo.hello/cicd/pkg -v -count=1
}

# source ${project_root_path}/env.sh
# print_releasecycle_tickets
run_test TestSingleTicketV2
# run_server

echo "done."
