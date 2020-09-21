#!/usr/bin/env bash

CURDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cp $CURDIR/git/pre-commit $CURDIR/../.git/hooks/pre-commit

chmod 0755 $CURDIR/../.git/hooks/pre-commit
