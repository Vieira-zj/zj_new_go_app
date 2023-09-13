#!/bin/bash
set -eu

tmp_dir="/tmp/test"
cover_file="${tmp_dir}/results.cov"

source ./common.sh

function echo_test {
    echo_info "this is info msg"
    echo_warn "this is warn msg"
    echo_error "this is error msg"
}

function update_pkg_to_latest {
    local pkg=$1
    go get ${pkg}@latest
}

function go_cover_test {
    if [[ -f $cover_file ]]; then
        rm $cover_file
    fi
    local dst_pkg="demo.apps/go.test/cover"
    ENV=dev go test -v -timeout=10s -cover -covermode=count -coverpkg="${dst_pkg}" -coverprofile="${cover_file}" ${dst_pkg}
}

function build_test_bin {
    local dst_pkg="demo.apps/go.test/cover"
    local dst_bin_file="${tmp_dir}/httpserve-test"

    if [[ -f $dst_bin_file ]]; then
        rm $dst_bin_file
    fi
    ENV=test go test -c -o=${dst_bin_file} ${dst_pkg}
}

function build_httpserve_cover_bin {
    # local cover_pkg="demo.apps/..."  # load all packages
    local cover_pkg="demo.apps/go.test/cover/..."
    local dst_pkg="demo.apps/go.test/cover/e2etest"
    local dst_bin_file="${tmp_dir}/httpserve-cover"

    if [[ -f $dst_bin_file ]]; then
        rm $dst_bin_file
    fi
    go test -c -cover -covermode=count -coverpkg="${cover_pkg}" -o=${dst_bin_file} ${dst_pkg}
}

function create_func_cover_report {
    local dst_file="${tmp_dir}/results.func"
    if [[ -f $dst_file ]]; then
        rm $dst_file
    fi
    go tool cover -func=${cover_file} -o ${dst_file}
}

function create_html_cover_report {
    local dst_file="${tmp_dir}/results.html"
    if [[ -f $dst_file ]]; then
        rm $dst_file
    fi
    go tool cover -html=${cover_file} -o ${dst_file}
}

if [[ $1 == "test" ]]; then
    echo "project path: ${project_dir}"
    echo_test
    exit 0
fi

if [[ $1 == "cover-test" ]]; then
    go_cover_test
    echo "go cover test done"
    exit 0
fi

if [[ $1 == "build-test" ]]; then
    build_test_bin
    echo "build go test bin done"
    exit 0
fi

if [[ $1 == "build-cover" ]]; then
    build_httpserve_cover_bin
    echo "build go cover bin done"
    exit 0
fi

if [[ $1 == "cover-report" ]]; then
    create_func_cover_report
    create_html_cover_report
    echo "create cover report done"
    exit 0
fi

echo "invalid args: $1"
