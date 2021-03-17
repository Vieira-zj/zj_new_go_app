#!/bin/bash
set -e

cur_dir=$(pwd)
report_dir="${cur_dir}/reports"

function run_ut() {
    local go_pkgs="demo.hello/apps/cover/test"
    # -covermode default "set"
    go test -v -timeout 10s ${go_pkgs} -cover -covermode count \
      -coverprofile ${report_dir}/cov_results.txt > ${report_dir}/ut_results.txt
}

function cover_summary() {
    go tool cover -func=${report_dir}/cov_results.txt -o ${report_dir}/cover_summary.txt
}

function open_cover_html() {
    go tool cover -html=${report_dir}/cov_results.txt
}

# run_ut
# cover_summary
# open_cover_html

echo "done."
