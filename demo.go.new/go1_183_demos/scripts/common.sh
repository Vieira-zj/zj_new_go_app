#!/bin/bash
set -eu

tmp_dir="/tmp/test"
go_project=$(git rev-parse --show-toplevel || get_project_root_path)

function get_project_root_path() {
    dirname $(pwd)
}

function is_bin_exist() {
    which $1 >/dev/null 2>&1
}

# colorful echo

function echo_info() {
    local color=$(tput setaf 2)
    local reset=$(tput sgr0)
    echo "${color}[INFO] $*${reset}"
}

function echo_warn() {
    local color=$(tput setaf 3)
    local reset=$(tput sgr0)
    echo "${color}[WARN] $*${reset}"
}

function echo_error() {
    # Visit this page for tput look up color code
    # https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
    local color=$(tput setaf 208)
    local reset=$(tput sgr0)
    echo "${color}[ERROR] $*${reset}"
}

# test

function is_bin_exist_test {
    local cmd="install"
    if is_bin_exist $cmd; then
        echo "$cmd exist"
    fi

    cmd="xxx"
    if ! is_bin_exist $cmd; then
        echo "$cmd not exist"
    fi
}

# is_bin_exist_test

echo "import shell utils."
