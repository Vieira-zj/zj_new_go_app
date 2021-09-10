#!/bin/bash
set -u

root_dir="results"

function inject_counter {
    local dst="${root_dir}/mock_cover.go"
    rm $dst
    go tool cover -mode=count mock.go >$dst
}

dst_profile="${root_dir}/mock_ut_cover.cov"

function cover_test {
    rm ${dst_profile}
    go test -v -timeout 5s -cover -covermode=count -coverprofile=${dst_profile}
}

function create_func_cover_report {
    local dst="${root_dir}/mock_func_report.txt"
    rm $dst
    go tool cover -func=${dst_profile} -o $dst
}

function create_html_cover_report {
    local dst="${root_dir}/mock_cover_report.html"
    rm $dst
    go tool cover -html=${dst_profile} -o $dst
}

# inject_counter
# cover_test

# create_func_cover_report
# create_html_cover_report

echo "Done"
