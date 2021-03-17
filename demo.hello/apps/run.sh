#!/bin/bash
set -eu

function build_bin() {
  app_name=$1
  app_path="${app_name}/main.go"
  type=$2 # linux, mac

  if [[ $type == "linux" ]]; then
    GOOS=linux GOARCH=amd64 go build -o "${app_name}_linux" ${app_path}
  else
    go build -o "${app_name}_mac" ${app_path}
  fi
}

function start_etcd() {
  docker network create app-tier --driver bridge

  docker run -d --rm --name etcd-server \
    --network app-tier \
    --publish 2379:2379 \
    --publish 2380:2380 \
    --env ALLOW_NONE_AUTHENTICATION=yes \
    --env ETCD_ADVERTISE_CLIENT_URLS=http://etcd-server:2379 \
    bitnami/etcd:latest
}

function stop_etcd() {
  docker stop etcd-server
  docker network rm app-tier
}

# start_etcd
# stop_etcd

# build_bin etcd linux

echo "App Done."