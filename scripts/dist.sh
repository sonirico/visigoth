#!/usr/bin/env bash
set -e

# Get the version from the command line
if [ -z $VERSION ]; then
    echo "Please specify a version."
    exit 1
fi

if [ ! -d "./build" ]; then
    echo "Build directory does not exists. Please run 'make build' first."
    exit 1
fi

# Check dependencies
echo -n "## Checking dependencies... "
for name in md5sum shasum
do
  [[ $(command -v $name 2>/dev/null) ]] || { echo -en "\n$name needs to be installed.";deps=1; }
done
[[ $deps -ne 1 ]] && echo "OK" || { echo -en "\nInstall the above and rerun this script\n";exit 1; }


# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"


# Change into that dir because we expect that
cd $DIR

echo -n "## Removing old directory... "
rm -rf dist && mkdir dist
echo "OK"

# Packaging builds
echo "## Packaging..."
for build in $(ls ./build); do
    tar -czf ./dist/$build.tar.gz -C ./build/$build $(ls ./build/$build)
    echo "- $build finished"
done

# Make the checksums
echo "## Signing..."
for dist in $(ls ./dist); do
    name=${dist%.tar.gz}
    cd $DIR/dist
    shasum -a 256 $name.tar.gz > $name.sha256
    md5sum $name.tar.gz > $name.md5
    echo "- $name finished"
done