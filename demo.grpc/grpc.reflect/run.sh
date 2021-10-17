#!/bin/bash
set -eu

run_cmd="go run main.go"
# run_cmd="./grpc.reflect"

function run_help() {
    ${run_cmd} -help
}

function get_grpc_hello_meta() {
    ${run_cmd} -addr=127.0.0.1:50051 -debug
}

function invoke_grpc_hello() {
    local body='{"metadata":[],"data":[{"name":"grpc"}]}'
    ${run_cmd} -addr=127.0.0.1:50051 -method=proto.Greeter.SayHello -body="${body}"
}

function get_grpc_echo_meta() {
    ${run_cmd} -addr=127.0.0.1:9090 -debug
}

function invoke_grpc_echo() {
    local body='{"metadata":[],"data":[{"value":"go grpc"}]}'
    ${run_cmd} -addr=127.0.0.1:9090 -method=demo.hello.Service1.Echo -body="${body}" -debug
}

function invoke_grpc_echo_hello() {
    local body='{"metadata":[{"name":"x-user-id", "value":"tester"}],"data":[{"name":"golang grpc"}]}'
    ${run_cmd} -addr=127.0.0.1:9090 -method=demo.hello.Service2.SayHello -body="${body}"
}

function get_grpc_deposit_meta() {
    ${run_cmd} -addr=127.0.0.1:50051 -debug
}

function invoke_grpc_deposit() {
    local body='{"metadata":[],"data":[{"amount":-10.01}]}'
    ${run_cmd} -addr=127.0.0.1:50051 -method=account.DepositService.Deposit -body="${body}"
}

# main

# if reflection is disabled, then get:
# panic: server does not support the reflection API

run_help

# get_grpc_hello_meta
# invoke_grpc_hello

# get_grpc_echo_meta
# invoke_grpc_echo
# invoke_grpc_echo_hello

# get_grpc_deposit_meta
# invoke_grpc_deposit

echo "grpc reflect demo done."
