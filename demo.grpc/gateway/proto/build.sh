#!/bin/bash
set -eu

project_dir="${HOME}/Workspaces/zj_repos/zj_go2_project"

function clear_all_pb_go_files() {
  local pbs=$(find . -name "*pb.go" -type f)
  for pb in ${pbs}; do
    echo "rm file ${pb}"
    rm ${pb}
  done

  local gws=$(find . -name "*pb.gw.go" -type f)
  for gw in ${gws}; do
    echo "rm file ${gw}"
    rm ${gw}
  done
}

function build_pb_files() {
  echo "build pb files."
  local raw_cmd="protoc -I . --go_out=plugins=grpc:${project_dir}"
  ${raw_cmd} google/api/*.proto
  ${raw_cmd} demo/hello/*.proto
}

function build_gw_pb_files() {
  echo "build gw pb files."
  local raw_cmd="protoc --grpc-gateway_out=logtostderr=true:${project_dir}"
  ${raw_cmd} demo/hello/*.proto
}

function build_swagger_api_files() {
  echo "build swagger api files."
  local raw_cmd="protoc -I . --swagger_out=logtostderr=true:../api"
  ${raw_cmd} demo/hello/*.proto
}

if [[ $1 == "clear" ]]; then
  clear_all_pb_go_files
elif [[ $1 == "pb" ]]; then
  build_pb_files
elif [[ $1 == "gw" ]]; then
  build_gw_pb_files
elif [[ $1 == "api" ]]; then
  build_swagger_api_files
fi

echo "Done"