#!/usr/bin/env bash
set -e

if [ -n "$1" ]; then
    ACTION=$1
fi

# If its dev mode, only build for ourself
if [ "$ACTION" == "dev" ]; then
    XC_OS=$(go env GOOS)
    XC_ARCH=$(go env GOARCH)
fi

# If its docker mode, only build linux-amd64
if [ "$ACTION" == "docker" ]; then
    XC_OS=linux
    XC_ARCH=amd64
fi

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that directory
cd "$DIR"

echo -n "## Recreate directory... "
rm -rf build && mkdir build
echo "OK"

# instruct to build statically linked binaries
export CGO_ENABLED=0

# build cmds
echo "## Build..."
if [ -d ./cmd ]; then
    for cmd in ./cmd/*; do
        if [ ! -d "$cmd" ]; then
            continue
        fi
        for OS in $XC_OS; do
            for ARCH in $XC_ARCH; do
                if ([ $OS == "darwin" ] && ([ $ARCH == "386" ] || [ $ARCH == "arm" ])) ||
                ([ $OS == "windows" ] && [ $ARCH == "arm" ])
                then
                    continue
                fi
                GOOS=$OS GOARCH=$ARCH go build -tags netgo -a -v -ldflags "$LD_FLAGS" -o build/$OS-$ARCH/${cmd##*/} ./cmd/${cmd##*/}
                echo "- $OS-$ARCH finished"
            done
        done
    done
fi

# If docker copy the binary to the root folder for dockerfile
if [ "$ACTION" == "docker" ]; then
    cp build/$XC_OS-$XC_ARCH/* ./
fi