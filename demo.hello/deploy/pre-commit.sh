#!/bin/bash
set -e

# Filter Golang files match Added (A), Copied (C), Modified (M) conditions.
gofiles=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [[ -n "$gofiles" ]]; then
    gofmt -s -w $gofiles
    goimports -w $gofiles
    git add $gofiles
fi
golangci-lint run
