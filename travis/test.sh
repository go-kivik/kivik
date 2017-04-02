#!/bin/bash
set -euC

if [ "${TRAVIS_OS_NAME:-}" == "osx" ]; then
    # We don't have docker in OSX, so skip these tests
    unset KIVIK_TEST_DSN_COUCH16
    unset KIVIK_TEST_DSN_COUCH20
fi

case "$1" in
    "standard")
        go test $(go list ./... | grep -v /vendor/ | grep -v /pouchdb)
    ;;
    "gopherjs")
        gopherjs test $(go list ./... | grep -v /vendor/ | grep -v kivik/serve)
    ;;
    "linter")
        diff -u <(echo -n) <(gofmt -e -d $(find . -type f -name '*.go' -not -path "./vendor/*"))
        go install # to make gotype (run by gometalinter) happy
        gometalinter.v1 --deadline=30s --vendor \
            --exclude="Errors unhandled\..*\(gas\)"  # This is an annoying duplicate of errcheck
    ;;
esac
