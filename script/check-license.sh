#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

lines=$(wc -l "${DIR}/header.txt" | cut -d' ' -f1)

FAIL=0
find . -name '*.go' -not -path "./.git/*" | while read file; do
    diff <(head -n $lines "${file}") "${DIR}/header.txt" >/dev/null || {
        echo "${file} missing license"
        FAIL=1
    }
done

if [ FAIL == "1" ]; then
    exit 1
fi