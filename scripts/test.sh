#!/usr/bin/env bash
set -e

# check version of golangci-lint
if [ -z "$TEST_PATTERN" ]; then
    echo "Do not call this file directly - use the make command"
    exit 1
fi

if [ -z "$1" ]; then
    go test ${TEST_OPTIONS} ${SOURCE_FILES} -run ${TEST_PATTERN} -timeout=${TEST_TIMEOUT}
else
    go test "$TEST_OPTIONS" "$SOURCE_FILES" -run "$TEST_PATTERN" -timeout="$TEST_TIMEOUT" 2>&1 | go-junit-report > report.xml
fi
