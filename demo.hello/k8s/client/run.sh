#!/bin/bash
set -e

root_dir="/tmp/test"
client_cli="${root_dir}/k8sclient"

function build_for_linux() {
    GOOS=linux GOARCH=amd64 go build .
    mv client /tmp/test/k8sclient_linux
}

function list_namespace_pods() {
    ${client_cli} -listnspods -n mini-test-ns
}

function monitor_pods() {
    # go run main.go -monitorpods
    ${client_cli} -monitorpods -n mini-test-ns
}

function list_service_pods() {
    ${client_cli} -getservicepods -n mini-test-ns -s hello-minikube
}

build_for_linux
# list_namespace_pods
# list_service_pods
# monitor_pods

echo "k8s client demo done."
