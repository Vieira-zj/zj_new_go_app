#!/bin/bash
set -eu

root_dir="${HOME}/Downloads/tmps/go_funcdiff_space"

function clear_format_files() {
    set +e
    rm ${root_dir}/src1/main_format.go
    rm ${root_dir}/src2/main_format.go
}

function get_diff_files() {
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

clear_format_files
# get_diff_files

echo "funcs diff done"
