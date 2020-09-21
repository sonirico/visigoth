#!/usr/bin/env bash
set -e

echo "## Check and fix files with gofmt and goimports... "
files=$(find . -name '*.go' -not -wholename './vendor/*')

if [ -z "$files" ]; then
    echo "no files found - skipping"
    exit 0
fi

gofmt_files=$(gofmt -l $files)

for file in $gofmt_files 
do
    gofmt -w -s $file
    goimports -w $file
done