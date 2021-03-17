#!/bin/bash
set -eu

function echo_api_perf_test() {
    # default run 30s
    wrk -t1 -c1 -d30s --script=post.lua --latency http://localhost:8081/v1/example/echo
}

echo_api_perf_test

echo "Test Done"
