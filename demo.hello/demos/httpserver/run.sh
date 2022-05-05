#!/bin/bash
set -e

# api test
curl http://localhost:8080/
curl http://localhost:8080/ping
curl http://localhost:8080/exec
curl http://localhost:8080/debug/vars | jq .

echo "done"
