#!/bin/bash
set -eu

function allocate_mem {
    for i in $(seq 100); do
        echo "iterator at ${i}"
        curl http://localhost:8082/mem/profile
        sleep 0.2
    done
}

function clear_mem {
    curl http://localhost:8082/mem/clear
}

allocate_mem &
allocate_mem &
allocate_mem
sleep 3
clear_mem

echo "test done"
