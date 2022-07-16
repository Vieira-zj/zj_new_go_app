#!/bin/sh
set -ue

# desc: func aop by ioc.

function gen {
    iocli gen
}

function list {
    iocli list
}

function watch {
    iocli watch main.ServiceImpl1 GetHelloString
}

# gen
# list
# watch

echo "done"
