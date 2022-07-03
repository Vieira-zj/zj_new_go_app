#!/bin/sh
set -eu

function install_deps_for_go1_183 {
    local deps=("golang.org/x/tools/gopls@latest" \
        "github.com/uudashr/gopkgs/v2/cmd/gopkgs" \
        "github.com/ramya-rao-a/go-outline" \
        "github.com/cweill/gotests/..." \
        "github.com/fatih/gomodifytags" \
        "github.com/josharian/impl" \
        "github.com/haya14busa/goplay/cmd/goplay" \
        "github.com/go-delve/delve/cmd/dlv" \
        "golang.org/x/lint/golint"
    )
    echo "GOPATH=${GOPATH}"
    for dep in ${deps[*]}; do
        echo "install dep bin: ${dep}"
        # go get ${dep}
        go install ${dep}
    done

    # GO111MODULE=on go install golang.org/x/tools/gopls@latest
}

function checkout_go_version {
	set +e
    local version=$1
    rm /usr/local/go
    ln -s /usr/local/go${version} /usr/local/go
    rm ${HOME}/Workspaces/.go
    ln -s ${HOME}/Workspaces/.go${version} ${HOME}/Workspaces/.go
}

# checkout_go_version 1_165
# checkout_go_version 1_1711

# install_deps_for_go1_183
# checkout_go_version 1_183

echo "done"
