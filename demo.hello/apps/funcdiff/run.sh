#!/bin/bash
set -eu

root_dir="${GO_PROJECT_ROOT}"

function run_go_test {
    local go_pkg="demo.hello/apps/funcdiff/pkg"
    local case="TestLinkProfileBlocksToFunc"
    go test -timeout 10s -run ^${case}$ ${go_pkg} -v -count=1
}

function run_go_test_cover {
    local go_pkg="demo.hello/apps/funcdiff/test/src1"
    local out_cov="/tmp/test/profile.cov"
    go test -v -timeout 10s ${go_pkg} -cover -coverprofile ${out_cov}
}

function get_diff_files {
    local go_repo="${HOME}/Downloads/tmps/go_repo_test"
    local commit1="b8acc5"
    local commit2="e09a77"

    cd ${go_repo}
    for fpath in $(git diff $commit1 $commit2 --name-only); do
        local commits=($commit1 $commit2)
        for commit in ${commits[*]}; do
            echo "cp $fpath to tmp folder for commit ${commit}"
            local fname=$(echo ${fpath} | awk -F '/' '{print $NF}')
            local diff_files="${diff_files} ${tmp_dir}/${commit}_${fname}"
            git checkout $commit
            cp $fpath ${tmp_dir}/${commit}_${fname}
        done
    done
}

run_go_test
# run_go_test_cover

# get_diff_files

echo "funcdiff done"
