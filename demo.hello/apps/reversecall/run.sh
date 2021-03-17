#!/bin/bash
set -e

function help() {
    go run main.go -help
}

function run_analysis_for_example() {
    local go_project="${HOME}/Workspaces/zj_repos/zj_go2_project/demo.hello/apps/reversecall/pkg/test"
    local filter_opts="-nointer -include=github.com/labstack,golang/org/x"
    go build .
    mv reversecall ${go_project}
    cd ${go_project}
    ./reversecall -fullpackage=demo.hello/apps/reversecall/pkg/test \
      -gofile=${go_project}/example/test3.go -package=example -func=test3b ${filter_opts}
    rm ${go_project}/reversecall
}

function run_analysis_for_echo() {
    local go_project="${HOME}/Workspaces/zj_repos/zj_go2_project/demo.hello/echoserver"
    go build .
    mv reversecall ${go_project}
    cd ${go_project}
    ./reversecall -fullpackage=demo.hello/echoserver -gofile=${go_project}/handlers/cover.go -package=handlers -func=getCondition
    rm ${go_project}/reversecall
}

#help

run_analysis_for_example
# 反向调用链:example.test3b<-example.Test3<-main.main
# if no -nointer opt set, "test3b" will not be show.
# 反向调用链:example.test3b<-null

# run_analysis_for_echo
# 反向调用链:handlers.getCondition<-handlers.CoverHandler<-handlers.Deco<-main.runApp<-main.main

echo "reverse call demo done."
