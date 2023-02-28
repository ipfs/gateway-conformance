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
    /merge-fixtures
    ;;
  *)
    echo "Usage: $0 <frames|no-frames> [input XML file] [output HTML file or directory]"
    exit 1
    ;;
esac
