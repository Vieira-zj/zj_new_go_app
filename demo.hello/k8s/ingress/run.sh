#!/bin/bash
set -eu

root_dir=$(pwd)

function build_image() {
    exec_file="k8s-simple-ingress-controller"
    cd $root_dir
    eval $(minikube -p minikube docker-env)
    GOOS=linux GOARCH=amd64 go build -o ${exec_file} .
    docker build -t zhengjin/k8s-simple-ingress-controller:v0.1 .
    rm ${exec_file}
}

build_image
