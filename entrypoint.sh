#!/usr/bin/env bash

set -euo pipefail

case "$1" in
  "test")
    echo "Running tests against $2 and writing JSON output to $3"
    jsonfile="$(realpath "$3")"
    export GATEWAY_URL="$2"
    pushd /app
    gotestsum --jsonfile "$jsonfile"
    popd
    ;;
  "extract-fixtures")
    echo "Extracting fixtures to $2"
    mkdir -p "$2"
    find /app/fixtures -name '*.car' -exec cp {} "${2}/" \;
    ;;
  "merge-fixtures")
    echo "Merging fixtures into $2"
    /merge-fixtures "$2"
    ;;
  *)
    echo "Usage: $0 test <gateway-url> <json-file>"
    echo "       $0 extract-fixtures <output-dir>"
    echo "       $0 merge-fixtures <output-car-file>"
    exit 1
    ;;
esac
