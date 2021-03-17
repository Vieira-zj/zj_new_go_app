#!/bin/bash
set -eu

#
# Code Check
#

function build_linux_bin {
  GOOS=linux GOARCH=amd64 go build main.go
}

function build_cover_bin {
  cd echoserver/; goc build -o echoserver_cover .
}

function golint_check {
  # golint -h
  golint demos apps
}

function govet_check {
  # go vet help
  # go vet demo.hello/demos demo.hello/apps/etcd
  go vet demos/*.go
}

#
# Docker Image
#

function build_echoserver_img {
  # run on minikube master
  docker build -t echoserver-test:v1.0 -f echoserver/deploy/echoserver.dockerfile .
}

function build_goccenter_img {
  # run on minikube master
  docker build -t goccenter-test:v1.0 -f echoserver/deploy/goccenter.dockerfile .
}

#
# Unit Test, and Coverage
#

all_go_pkgs=''

function get_all_go_pkgs {
    local pkgs=''
    for pkg in $(go list ./...); do
        pkgs="${pkgs},${pkg}"
    done
    all_go_pkgs=$(echo ${pkgs} | awk -F ',' '{for(i=1; i<=NF; i++) print $i;}')
    echo ${all_go_pkgs}
}

report_dir="echoserver/ut_reports"
ut_cov_xml="${report_dir}/ut_cover.xml"

function run_diff_cover_deubg {
    # local debug (pre-condition: pip uninstall diff-cover)
    local diffcover="python ${HOME}/Workspaces/github_repos/diff_cover/diff_cover/main.py"
    ${diffcover} ${ut_cov_xml} --compare-branch=remotes/origin/master --html-report=${report_dir}/diff_cov.html
}

function go_test_with_diff_cover {
  rm -r ${report_dir}/*

  local ut_cov_file="${report_dir}/ut.cov"
  local go_pkgs="demo.hello/echoserver/handlers"
  go test -v -timeout 10s ${go_pkgs} -cover -coverprofile ${ut_cov_file} > ${report_dir}/ut.results
  go tool cover -func=${ut_cov_file} -o ${report_dir}/cover_all.out

  local git_branch="remotes/origin/master"
  gocov convert ${ut_cov_file} | gocov-xml > ${ut_cov_xml}
  diff-cover ${ut_cov_xml} --compare-branch=${git_branch} --html-report=${report_dir}/cover_diff.html > ${report_dir}/cover_diff.out
}

#
# Goc Coverage
#

app_dir="echoserver"
coverprofiles_dir="${app_dir}/cover_profiles"
coverreport_dir="${app_dir}/cover_reports"
cobertura_dir="${app_dir}/cobertura_reports"

goccenter_addr="http://goccenter.default.k8s"
#$(goc list --center=${goccenter_addr} | jq ".echoserver[0]")
gocagent_addr="http://127.0.0.1:8080"

function clear_coverage_data {
  rm ${coverprofiles_dir}/*
  rm ${coverreport_dir}/*
  rm ${cobertura_dir}/*
  # clear server side coverage data
  goc clear --center=${goccenter_addr} --address=${gocagent_addr}
}

function get_goc_coverage {
  goc profile --center=${goccenter_addr} --address=${gocagent_addr} -o ${coverprofiles_dir}/coverprofile_$(date +%s).cov
}

function merge_goc_coverages {
  local target_file=$1
  local count=$(ls ${coverprofiles_dir} | wc -l)

  if [[ ${count} -gt 1 ]]; then
    goc merge ${coverprofiles_dir}/* -o ${target_file}
  else
    cp ${coverprofiles_dir}/* ${target_file}
  fi
}

function show_goc_coverage {
  # manual set coverage file
  go tool cover -func=echoserver/coverprofile.cov
}

function create_goc_coverage_html {
  local src_file=$1
  go tool cover -html=${src_file} -o ${coverreport_dir}/coverage_$(date +%Y%m%d_%H%M).html
}

function create_cobertura_thml {
  local src_file=$1
  local project_path="${HOME}/Workspaces/zj_repos/zj_go2_project"
  local cobertura_xml_path="${project_path}/cobertura.xml"

  gocover-cobertura < ${src_file} > ${cobertura_xml_path}
  pycobertura show --format html --output ${cobertura_dir}/cobertura_$(date +%Y%m%d_%H%M).html ${cobertura_xml_path}
  rm ${cobertura_xml_path}
}

#
# main
#

merge_coverage_fname="merge_coverprofile_$(date +%Y%m%d_%H%M).cov"
merge_coverage_fpath="${app_dir}/${merge_coverage_fname}"

if [[ $1 == "clear" ]]; then
  clear_coverage_data
elif [[ $1 == "sync" ]]; then
  get_goc_coverage
elif [[ $1 == "show" ]]; then
  show_goc_coverage # for test
elif [[ $1 == "report" ]]; then
  merge_goc_coverages ${merge_coverage_fpath}
  create_goc_coverage_html ${merge_coverage_fpath}
  create_cobertura_thml ${merge_coverage_fpath}
  rm ${merge_coverage_fpath}
elif [[ $1 == "ut" ]]; then
  # ut ci workflow
  go_test_with_diff_cover
  # gen-diff-cover-html-report
elif [[ $1 == "main" ]]; then
  # main
  # build_linux_bin
  build_cover_bin
  # golint_check
  # govet_check
else
  echo "Invalid args [$1] and run test."
  # build_echoserver_img
  # build_goccenter_img
fi

echo "Done"
