#!/bin/bash
set -eu

#
# Build, Lint
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

function go_generate {
  go generate main.go
}

#
# Build Docker Image
#

function build_echoserver_img {
  # run on minikube master
  docker build -t echoserver-test:v1.0 -f echo/deploy/echoserver.dockerfile .
}

function build_goccenter_img {
  # run on minikube master
  docker build -t goccenter-test:v1.0 -f echo/deploy/goccenter.dockerfile .
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
master_branch="remotes/origin/master"

function run_diff_cover_debug {
  # debug by local py module
  # pre-condition: pip uninstall diff-cover
  local diffcover="python ${HOME}/Workspaces/github_repos/diff_cover/diff_cover/main.py"
  ${diffcover} ${ut_cov_xml} --compare-branch=${master_branch} --html-report=${report_dir}/diff_cov.html
}

function go_test_with_diff_cover {
  rm -r ${report_dir}/*

  local ut_cov_file="${report_dir}/ut.cov"
  local go_pkgs="demo.hello/echoserver/handlers"
  go test -v -timeout 10s ${go_pkgs} -cover -coverprofile ${ut_cov_file} > ${report_dir}/ut_results.out
  go tool cover -func=${ut_cov_file} -o ${report_dir}/func_cover.out

  gocov convert ${ut_cov_file} | gocov-xml > ${ut_cov_xml}
  diff-cover ${ut_cov_xml} --compare-branch=${master_branch} --html-report=${report_dir}/diff_cover.html > ${report_dir}/diff_cover.out
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

function clear_goc_coverage {
  rm ${coverprofiles_dir}/*
  rm ${coverreport_dir}/*
  rm ${cobertura_dir}/*
  # clear server side coverage data
  goc clear --center=${goccenter_addr} --address=${gocagent_addr}
}

function get_goc_coverage {
  goc profile --center=${goccenter_addr} --address=${gocagent_addr} -o ${coverprofiles_dir}/profile_$(date +%s).cov
}

function merge_coverage_files {
  local target_file=$1
  local count=$(ls ${coverprofiles_dir} | wc -l)

  if [[ ${count} -gt 1 ]]; then
    goc merge ${coverprofiles_dir}/* -o ${target_file}
  else
    cp ${coverprofiles_dir}/* ${target_file}
  fi
}

function show_func_coverage {
  # manual set coverage file
  go tool cover -func=echoserver/profile.cov
}

function create_coverage_html {
  local src_file=$1
  go tool cover -html=${src_file} -o ${coverreport_dir}/coverage_$(date +%Y%m%d_%H%M).html
}

function create_cobertura_cover_html {
  local src_file=$1
  local project_path="${HOME}/Workspaces/zj_repos/zj_go2_project"
  local cobertura_xml_path="${project_path}/cobertura_cover.xml"

  gocover-cobertura < ${src_file} > ${cobertura_xml_path}
  pycobertura show --format html --output ${cobertura_dir}/cobertura_$(date +%Y%m%d_%H%M).html ${cobertura_xml_path}
  rm ${cobertura_xml_path}
}

#
# Main
#

merge_coverage_fname="merge_cover_$(date +%Y%m%d_%H%M).cov"
merge_coverage_fpath="${app_dir}/${merge_coverage_fname}"

if [[ $1 == "clear" ]]; then
  clear_goc_coverage
elif [[ $1 == "sync" ]]; then
  get_goc_coverage
elif [[ $1 == "show" ]]; then
  show_func_coverage
elif [[ $1 == "report" ]]; then
  merge_coverage_files ${merge_coverage_fpath}
  create_coverage_html ${merge_coverage_fpath}
  create_cobertura_cover_html ${merge_coverage_fpath}
  rm ${merge_coverage_fpath}
elif [[ $1 == "ut" ]]; then
  go_test_with_diff_cover
elif [[ $1 == "main" ]]; then
  # build_linux_bin
  build_cover_bin
  # golint_check
  # govet_check
else
  echo "Invalid args [$1] and run test."
  # build_echoserver_img
  # build_goccenter_img
fi

# fix go lib conflict
# go mod edit -replace google.golang.org/grpc@v1.37.0=google.golang.org/grpc@v1.26.0

echo "Done"
