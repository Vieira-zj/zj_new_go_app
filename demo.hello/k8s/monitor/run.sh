#!/bin/bash
set -eu

function build_for_linux() {
    local bin_name="pod_monitor"
    GOOS=linux GOARCH=amd64 go build -o ${bin_name}
    mv ${bin_name} /tmp/test/pod_monitor_linux
}

function build_error_exit_image() {
    docker build -t zhengjin/error-exit:v1.1 -f image/ErrorExit.DockerFile .
}

function build_pod_monitor_image() {
    docker build -t zhengjin/pod-monitor:v1.0 -f image/PodMonitor.DockerFile .
}

function run_debug_pod() {
    kubectl run test-pod -it --rm --restart=Never -n k8s-test --image=busybox:1.30 sh
}

function run_gotest() {
    local case="TestChannelCap"
    go test -timeout 10s -run ^${case}$ demo.hello/k8s/monitor/internal -v -count=1
}

# build_error_exit_image
# build_for_linux
# build_pod_monitor_image
# run_gotest
echo "done"
