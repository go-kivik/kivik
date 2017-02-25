#!/bin/bash
set -euC

case "$1" in
    "standard")
        go test $(go list ./... | grep -v pouchdb)
    ;;
    "gopherjs")
        gopherjs test github.com/flimzy/kivik/test
    ;;
    "linter")
        diff -u <(echo -n) <(gofmt -d ./)
        go install # to make gotype (run by gometalinter) happy
        gometalinter.v1 --deadline=30s
    ;;
esac
