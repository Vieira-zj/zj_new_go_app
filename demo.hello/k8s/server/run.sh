#!/bin/bash
set -e

function build_for_linux() {
    GOOS=linux GOARCH=amd64 go build .
    mv server /tmp/test/k8s_monitor_linux
}

build_for_linux
echo "build done."
