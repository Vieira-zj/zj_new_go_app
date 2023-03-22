#!/bin/bash
set -eu

function load_test() {
    echo "start load test"
    for i in $(seq 100); do
        echo $i
        curl http://localhost:8080/render --data-binary @data/README.md
    done
    echo "load test done"
}

filename=$1

function fetch_profile() {
    echo "fetch profile"
    curl -o bin/${filename} "http://localhost:8080/debug/pprof/profile?seconds=5"
}

fetch_profile &
load_test
