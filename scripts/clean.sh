#!/usr/bin/env bash
set -e

function clean {
  echo "Cleaning build & bin directories"
  rm -rf ./bin ./build
  echo "Cleaning golang caches..."
  go clean -i -r -cache -testcache -modcache
}

clean;
