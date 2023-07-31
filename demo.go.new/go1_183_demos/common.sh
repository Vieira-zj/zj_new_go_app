#!/bin/bash
set -eu

project_dir="$(git rev-parse --show-toplevel)/demo.go.new/go1_183_demos"

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

echo "it's shell utils."
