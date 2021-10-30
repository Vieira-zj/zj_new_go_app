#!/bin/bash
set -eu

function build_for_linux() {
    local bin_name="pod_monitor"
    GOOS=linux GOARCH=amd64 go build -o bin/${bin_name}
}

function build_error_exit_image() {
    docker build -t zhengjin/error-exit:v1.1 -f image/ErrorExit.dockerfile .
}

function build_pod_monitor_image() {
    docker build -t zhengjin/pod-monitor:v1.0 -f image/Monitor.dockerfile .
}

function run_debug_pod() {
    kubectl run test-pod -it --rm --restart=Never -n k8s-test --image=busybox:1.30 sh
}

function deploy_pod_monitor() {
    ./bin/envexpand -i deploy/monitor_deploy.yaml -p | kubectl create -f -
}

function run_gotest() {
    local case="TestRateLimiter"
    go test -timeout 60s -run ^${case}$ demo.hello/k8s/monitor/internal -v -count=1
}

# build_for_linux
# build_pod_monitor_image
# build_error_exit_image

# deploy_pod_monitor
run_gotest
echo "done"
