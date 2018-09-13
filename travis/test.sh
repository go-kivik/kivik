#!/bin/bash
set -euC

if [ "${TRAVIS_OS_NAME:-}" == "osx" ]; then
    # We don't have docker in OSX, so skip these tests
    unset KIVIK_TEST_DSN_COUCH16
    unset KIVIK_TEST_DSN_COUCH17
    unset KIVIK_TEST_DSN_COUCH20
    unset KIVIK_TEST_DSN_COUCH21
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
        diff -u <(echo -n) <(gofmt -e -d $(find . -type f -name '*.go' -not -path "./vendor/*"))
        go install # to make gotype (run by gometalinter) happy
        go test -i
        gometalinter.v1 --config=.linter.json
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
