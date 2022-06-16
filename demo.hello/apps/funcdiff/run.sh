#!/bin/bash
set -eu

root_dir="${GO_PROJECT_ROOT}"

function clear_format_files {
    set +e
    rm ${root_dir}/src1/main_format.go
    rm ${root_dir}/src2/main_format.go
}

function run_go_test_cover {
    local go_pkgs="demo.hello/apps/funcdiff/test/src1"
    local out_cov="/tmp/test/profile.cov"
    go test -v -timeout 10s ${go_pkgs} -cover -coverprofile ${out_cov}
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

# clear_format_files
run_go_test_cover
# get_diff_files

echo "funcs diff done"
