#!/bin/bash
set -eu

#
# proto
#

function compile_proto {
  local pbs=$(find grpc.greeter/proto -name "*.pb.go" -type f)
  for pb in ${pbs}; do
    echo "rm ${pb}"
    rm ${pb}
  done

  # must be in parent dir of project to exec "protoc"
  cd ../
  # -IPATH relate to "import" of proto
  # --go_out=plugins=grpc:PATH relate to "option go_package" of proto
  protoc -I . --go_out=plugins=grpc:. demo.grpc/grpc.greeter/proto/message/*.proto
  protoc -I . --go_out=plugins=grpc:. demo.grpc/grpc.greeter/proto/helloworld.proto
}

function compile_account_proto {
  cd protoc/
  set +e
  rm -r pb/account
  sleep 1
  set -e

  local pb_dir="proto/account"
  protoc --proto_path=${pb_dir} --go_out=. --go-grpc_out=. ${pb_dir}/*.proto
}

function compile_greeter_proto {
  cd protoc/
  set +e
  rm -r pb/greeter
  sleep 1
  set -e

  local pb_dir="proto/greeter"
  # --go_out generate pb struct
  # --go-grpc_out generate grpc invoke
  # --proto_path set root path for import proto file
  protoc --proto_path=${pb_dir} --go_out=. --go-grpc_out=. proto/greeter/helloworld.proto ${pb_dir}/message/*.proto
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
# compile_greeter_proto
# compile_account_proto

# build_grpcserver_img
# build_grpcui_img
# build_goccenter_img

# run_grpc_unittest

echo "Done"