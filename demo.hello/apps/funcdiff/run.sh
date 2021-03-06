#!/bin/bash
set -e

cur_dir=$(pwd)
go_repo="${HOME}/Downloads/tmps/go_repo_test"
tmp_dir="/tmp/test"
diff_files=""

function func_diff_test() {
    go run main.go -e -s ${tmp_dir}/b8acc5_test1.go -t ${tmp_dir}/e09a77_test1.go
}

function get_diff_files() {
    local commit1="b8acc5"
    local commit2="e09a77"

    cd ${go_repo}
    for fpath in $(git diff $commit1 $commit2 --name-only); do
        local commits=($commit1 $commit2)
        for commit in ${commits[*]}; do
            echo "cp $fpath to tmp folder for commit ${commit}"
            local fname=$(echo ${fpath} | awk -F '/' '{print $NF}')
            diff_files="${diff_files} ${tmp_dir}/${commit}_${fname}"
            git checkout $commit
            cp $fpath "${tmp_dir}/${commit}_${fname}"
        done
    done
}

function get_diff_funcs() {
    echo $diff_files
    file1=$(echo ${diff_files} | awk '{print $1}')
    file2=$(echo ${diff_files} | awk '{print $2}')
    cd $cur_dir
    go run main.go -e -s $file1 -t $file2
}

func_diff_test

# get_diff_files
# get_diff_funcs

echo "done."
