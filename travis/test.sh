#!/bin/bash
set -euC

if [ "${TRAVIS_OS_NAME:-}" == "osx" ]; then
    # We don't have docker in OSX, so skip these tests
    unset KIVIK_TEST_DSN_COUCH23
fi

function join_list {
    local IFS=","
    echo "$*"
}

case "$1" in
    "standard")
        ./travis/test_version.sh
        go test -race $(go list ./... | grep -v /vendor/)
    ;;
    "gopherjs")
        gopherjs test $(go list ./...)
    ;;
    "linter")
        golangci-lint run ./...
    ;;
    "coverage")
        echo "" > coverage.txt

        TEST_PKGS=$(go list ./... | grep -v /test)

        for d in $TEST_PKGS; do
            go test -i $d
            go test -coverprofile=profile.out -covermode=set "$d"
            if [ -f profile.out ]; then
                cat profile.out >> coverage.txt
                rm profile.out
            fi
        done

        bash <(curl -s https://codecov.io/bash)
esac
