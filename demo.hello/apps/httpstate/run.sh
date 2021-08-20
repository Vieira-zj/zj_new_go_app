#!/bin/bash
set -eu

function build() {
    go build -ldflags '-X=main.version=v1.0.0' -o '/tmp/test/httpstate' main.go
}

function run() {
    go run main.go http://www.baidu.com/
    # ./httpstate http://www.baidu.com/
    # ./httpstate -I http://www.baidu.com/
}

# build
run
echo "done"
