#!/bin/bash
set -eu

function main_build {
    rm -r ./bin
    go build -o ./bin/job .

    # create link to bin for each job
    local jobs=("job_foo" "job_bar")
    for name in ${jobs[*]}; do
        echo "create link for ${name}"
        ln -sf ./job ./bin/${name}
    done
}

main_build

# run
# ./bin/job_bar
