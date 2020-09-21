#!/usr/bin/env bash
set -e

# check version of golangci-lint
if [ -z "$LINT_VERSION" ]; then
    echo "Do not call this file directly - use the make command"
    exit 1
fi

function install {
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v"$LINT_VERSION"
}

if [ ! -f "./bin/golangci-lint" ]; then
    echo "Installing golangci-lint to current project"
    install
fi

if [[ $(./bin/golangci-lint --version) != *"$LINT_VERSION"* ]]; then
    echo "outdated version found - installing new"
    install
fi

./bin/golangci-lint run --tests=false --enable-all --disable=lll