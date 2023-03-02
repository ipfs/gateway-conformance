#!/usr/bin/env bash

case "$1" in
  "test")
    export GATEWAY_URL="$2"
    gotestsum --junitfile "$3"
    ;;
  "extract-fixtures")
    find /app/fixtures -name '*.car' -exec cp {} "${2}/" \;
    ;;
  "merge-fixtures")
    /merge-fixtures "$2"
    ;;
  *)
    echo "Usage: $0 test <gateway-url> <junit-file>"
    echo "       $0 extract-fixtures <output-dir>"
    echo "       $0 merge-fixtures <output-car-file>"
    exit 1
    ;;
esac
