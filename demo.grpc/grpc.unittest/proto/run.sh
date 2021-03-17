#!/bin/bash
set -e

protoc -I account/ account/deposit.proto --go_out=plugins=grpc:account/

echo "Done"