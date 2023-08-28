#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

lines=$(wc -l "${DIR}/header.txt" | cut -d' ' -f1)

find . -name '*.go' | while read file; do
    diff <(head -n $lines "${file}") "${DIR}/header.txt" >/dev/null || {
        echo "${file} missing license"
        exit 1
    }
done