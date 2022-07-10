#!/bin/sh
set -ue

function list {
    iocli list
}

function watch {
    iocli watch main.ServiceImpl1 GetHelloString
}

list
# watch

echo "done"
