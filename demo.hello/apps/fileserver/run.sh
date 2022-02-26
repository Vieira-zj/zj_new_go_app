#!/bin/bash

function build_linux_bin() {
    GOOS=linux GOARCH=amd64 go build .
}

build_linux_bin
echo "done"
