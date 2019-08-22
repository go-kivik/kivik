#!/bin/bash
set -euC

cur_ver=$(grep 'const Version =' constants.go | awk '{gsub(/"/, "", $4); print $4}')
echo "Current version: ${cur_ver}"

cur_tag=$(git describe --tags --exact-match 2>/dev/null || true)
if [ "$cur_tag"x == "x" ]; then
    # Only run this check if we're on an actual tag
    exit 0
fi

if [ "$cur_tag" != "$cur_ver" ]; then
    echo "Tag must match version"
    echo "${cur_tag} != ${cur_ver}"
    exit 1
fi
