#!/bin/bash
set -ue

project_root_path="${HOME}/Workspaces/zj_repos/zj_go2_project/demo.hello"

function print_releasecycle_tickets() {
    go run main.go -releaseCycle='2021.04.v3 - AirPay'
}

function run_server() {
    go run main.go -server
}

source ${project_root_path}/env.sh
# print_releasecycle_tickets
run_server

echo "done."
