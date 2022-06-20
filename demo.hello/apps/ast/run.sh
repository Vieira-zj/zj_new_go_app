#!/bin/bash
set -eu

function run_go_test() {
    local case="TestFindAllCase"
    go test -timeout 30s -run ^${case}$ demo.hello/apps/ast -v -count=1
}

run_go_test
echo "done"
