#!/bin/bash
set -eu

function dk_build_goserver {
  docker build -t goserver-test:v1.0 -f deploy/goserver.dockerfile .
}

function dk_build_goccenter {
  docker build -t goccenter-test:v1.0 -f deploy/goccenter.dockerfile .
}

function create_cover_html_report {
    # coverage.xml
    local cover_dir="cobertura_reports"
    gocover-cobertura < coverprofile.cov > ${cover_dir}/coverage.xml

    # coverage.html
    local model_path=$(go list)
    local cover_path=${model_path}/${cover_dir}

    cd ${HOME}/Workspaces/zj_repos/zj_go2_project
    cp ${cover_path}/coverage.xml .
    pycobertura show --format html --output ${cover_path}/coverage.html coverage.xml
    rm coverage.xml
}

# dk_build_goserver
# dk_build_goccenter

create_cover_html_report

echo "Done"