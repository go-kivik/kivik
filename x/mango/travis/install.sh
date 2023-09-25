#!/bin/bash
set -euC
set -o xtrace

curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
dep ensure && dep status

# Only continue if we're on go 1.12; no need to run the linter for every case
if go version | grep -q go1.12; then
    go get -u gopkg.in/alecthomas/gometalinter.v1 && gometalinter.v1 --install
fi
