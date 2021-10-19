#!/bin/bash
set -eu

function build_error_exit_image() {
    docker build -t zhengjin/error-exit:v1.0 -f image/ErrorExit.DockerFile .
}

function run_debug_pod() {
    kubectl run test-pod -it --rm --restart=Never -n k8s-test --image=busybox:1.30 sh
}

function build_for_linux() {
    GOOS=linux GOARCH=amd64 go build .
    mv server /tmp/test/pod_monitor_linux
}

# build_error_exit_image
# build_for_linux
echo "build done."
