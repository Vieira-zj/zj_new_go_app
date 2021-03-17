#!/bin/bash
set -eu

DOCKER_USER="zhengjin.k8s.test"
BIN_FILE="bin/admission-webhook-app"

function build_bin() {
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ${BIN_FILE}
}

function build_image() {
  if [[ ! -f ${BIN_FILE} ]]; then
    echo "go bin file ${BIN_FILE} not exist."
    exit 99
  fi

  : ${DOCKER_USER:? required}
  docker build --no-cache -t ${DOCKER_USER}/admission-webhook-app:v1.0.1 .
  #docker push ${DOCKER_USER}/admission-webhook-app:v1

#   rm -rf ${BIN_FILE}
}

if [[ $1 == 'build' ]]; then
    build_bin
elif [[ $1 == 'push' ]]; then
    build_image
elif [[ $1 == 'clear' ]]; then
    docker ps -a | awk '/Exited/ {print $1}' | xargs docker rm
else
    echo "invalid arg: $1"
fi

echo "k8s admission webhook done."