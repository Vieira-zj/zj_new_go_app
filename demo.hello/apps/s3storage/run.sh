#!/bin/bash
set -eu

function make_file() {
    dd if=/dev/zero of=/tmp/test/storagefile.test bs=1m count=100
}

function run_test() {
    case=$1
    go test -timeout 30s -run ^${case}$ demo.hello/apps/s3storage/pkg -v -count=1
}

# make_file
run_test TestDelete

echo "done."
