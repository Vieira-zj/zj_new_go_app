#!/bin/bash
set -eu

function compile_account_proto {
  set +e
  rm -r pb/account
  sleep 1
  set -e

  local pb_dir="proto/account"
  protoc --proto_path=${pb_dir} --go_out=. --go-grpc_out=. ${pb_dir}/*.proto
}

function compile_greeter_proto {
  set +e
  rm -r pb/greeter
  sleep 1
  set -e

  local pb_dir="proto/greeter"
  # --go_out generate pb struct
  # --go-grpc_out generate grpc invoke
  # --proto_path set root path for import proto file
  protoc --proto_path=${pb_dir} --go_out=. --go-grpc_out=. ${pb_dir}/helloworld.proto ${pb_dir}/msg/*.proto
}

# compile_account_proto
# compile_greeter_proto

echo "protoc done"
