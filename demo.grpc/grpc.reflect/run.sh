#!/bin/bash
set -eu

run_cmd="go run main.go"

function run_help() {
    ${run_cmd} -help
}

function show_grpc_hello_meta() {
    ${run_cmd} -addr=127.0.0.1:50051 -debug
}

function invoke_grpc_hello() {
    ${run_cmd} -addr=127.0.0.1:50051 -method=proto.Greeter.SayHello -body='{"metadata":[],"data":[{"name":"grpc"}]}'
}

function show_grpc_echo_meta() {
    ${run_cmd} -addr=127.0.0.1:9090 -debug
}

function invoke_grpc_echo() {
    ${run_cmd} -debug -addr=127.0.0.1:9090 -method=demo.hello.Service1.Echo -body='{"metadata":[],"data":[{"value":"golang grpc"}]}'
}

function invoke_grpc_deposit() {
    ${run_cmd} -debug -addr=127.0.0.1:50051 
}


# run_help

# show_grpc_hello_meta
invoke_grpc_hello

# show_grpc_echo_meta
# invoke_grpc_echo

# panic: server does not support the reflection API
# invoke_grpc_deposit

echo "grpc reflect demo done."
