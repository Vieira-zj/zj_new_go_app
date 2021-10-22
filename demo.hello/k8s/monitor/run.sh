#!/bin/bash
set -eu

function build_error_exit_image() {
    docker build -t zhengjin/error-exit:v1.1 -f image/ErrorExit.DockerFile .
}

function build_for_linux() {
    GOOS=linux GOARCH=amd64 go build .
    mv server /tmp/test/pod_monitor_linux
}

function run_debug_pod() {
    kubectl run test-pod -it --rm --restart=Never -n k8s-test --image=busybox:1.30 sh
}

function run_gotest() {
    local case="TestGetAllPodInfosByMultiNamespaces"
    go test -timeout 10s -run ^${case}$ demo.hello/k8s/monitor/internal -v -count=1
}

# build_error_exit_image
# build_for_linux
run_gotest
echo "build done."
