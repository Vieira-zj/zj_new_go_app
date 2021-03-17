#!/bin/bash
set -ue

function build_pb() {
    protoc -I . --go_out=plugins=grpc:. pb/random.proto
}

build_pb

echo "done"