#!/bin/sh
set -eu

function checkout_go_version {
	set +e
    local version=$1
    rm /usr/local/go
    ln -s /usr/local/go${version} /usr/local/go
    rm ${HOME}/Workspaces/.go
    ln -s ${HOME}/Workspaces/.go${version} ${HOME}/Workspaces/.go
}

checkout_go_version 1_165

echo "checkout done"
