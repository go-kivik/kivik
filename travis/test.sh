#!/bin/bash
set -euC

if [ "$TRAVIS_OS_NAME" == "osx" ]; then
    # We don't have docker in OSX, so skip these tests
    unset KIVIK_TEST_DSN_COUCH16
    unset KIVIK_TEST_DSN_COUCH20
fi

case "$1" in
    "standard")
        go test $(go list ./... | grep -v /pouchdb)
    ;;
    "gopherjs")
        gopherjs test $(go list ./... | grep -v kivik/serve)
    ;;
    "linter")
        diff -u <(echo -n) <(gofmt -d ./)
        go install # to make gotype (run by gometalinter) happy
        gometalinter.v1 --deadline=30s
    ;;
esac
