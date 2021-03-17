#!/bin/bash
set -eu

#
# proto
#

function compile_proto {
  local pbs=$(find grpc/proto -name "*.pb.go" -type f)
  for pb in ${pbs}; do
    echo "rm ${pb}"
    rm ${pb}
  done

  # must be in parent dir of project to exec "protoc"
  # -IPATH relate to "import" of proto
  # --go_out=plugins=grpc:PATH relate to "option go_package" of proto
  cd ../
  protoc -I . --go_out=plugins=grpc:. demo.grpc/grpc/proto/message/*.proto
  protoc -I . --go_out=plugins=grpc:. demo.grpc/grpc/proto/helloworld.proto
}

#
# image
#

function build_grpcserver_img {
  docker build -t grpcserver-test:v1.0 -f grpc/deploy/grpcserver.dockerfile .
}

function build_grpcui_img {
  docker build -t grpcui-test:v1.0 -f grpc/deploy/grpcui.dockerfile .
}

function build_goccenter_img {
  docker build -t goccenter-test:v1.0 -f grpc/deploy/goccenter.dockerfile .
}

#
# test
#

function run_grpc_unittest {
  local pkg="demo.grpc/grpc.unittest/internal/account"
  go test -timeout 10s ${pkg} -run ^TestDeposit -v -count=1
}

#
# run
#

# compile_proto

# build_grpcserver_img
# build_grpcui_img
# build_goccenter_img

run_grpc_unittest

echo "Done"