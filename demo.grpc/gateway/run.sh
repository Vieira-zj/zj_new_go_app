#!/bin/bash
set -eu

function build_bin() {
  local type=$1
  local src_main_path="./server/grpc_server/main.go"
  local bin_name="http_proxy"

  if [[ $type == "linux" ]]; then
    GOOS=linux GOARCH=amd64 go build -o "${bin_name}_linux" ${src_main_path}
  else
    go build -o "${bin_name}_mac" ${src_main_path}
  fi
}

function swaggerui() {
  local op=$1
  if [[ ${op} == "start" ]]; then
    docker run --name swaggerui --rm -d -p 8080:8080 metz/swaggerui
    sleep 1
    # fix issue: Error: request entity too large
    # app.use(bodyParser.json({limit: 10 * 1024 * 1024}))
    docker cp api/server.js swaggerui:/usr/src/app
    docker restart swaggerui
  else
    docker stop swaggerui
  fi
}

function upload_swagger_json() {
  local name=$1
  local target_file=$(find api -name "${name}.swagger.json" -type f)
  echo "upload swagger api file: ${target_file}"
  curl -v "http://localhost:8080/publish" -XPOST -H "Content-Type: application/json" -d "@${target_file}"
}

if [[ $1 == "build" ]]; then
  build_bin $2 
elif [[ $1 == "ui" ]]; then
  swaggerui $2
elif [[ $1 == "upload" ]]; then
  upload_swagger_json service1
else
  echo "no valid op specified!"
fi

echo "Done"