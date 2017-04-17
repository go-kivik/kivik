#!/bin/bash
set -euC

if [ "${TRAVIS_OS_NAME:-}" == "osx" ]; then
    # We don't have docker in OSX, so skip these tests
    unset KIVIK_TEST_DSN_COUCH16
    unset KIVIK_TEST_DSN_COUCH20
fi

function join_list {
    local IFS=","
    echo "$*"
}

case "$1" in
    "standard")
        go test -race $(go list ./... | grep -v /vendor/ | grep -v /pouchdb)
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
    "coverage")
        # Use only CouchDB 2.0 for the coverage tests, primarily because CouchDB
        # is sporadic with failures, and leads to fluctuating coverage stats.
        unset KIVIK_TEST_DSN_COUCH16
        echo "" > coverage.txt

        TEST_PKGS=$(find -name "*_test.go" | grep -v /vendor/ | grep -v /pouchdb | xargs dirname | sort -u | sed -e "s#^.#github.com/flimzy/kivik#" )

        for d in $TEST_PKGS; do
            go test -i $d
            DEPS=$((go list -f $'{{range $f := .TestImports}}{{$f}}\n{{end}}{{range $f := .Imports}}{{$f}}\n{{end}}' $d && echo $d) | sort -u | grep -v /vendor/ | grep -v /pouchdb | grep -v /kivik/test | grep ^github.com/flimzy/kivik | tr '\n' ' ')
            go test -coverprofile=profile.out -covermode=set -coverpkg=$(join_list $DEPS) $d
            if [ -f profile.out ]; then
                cat profile.out >> coverage.txt
                rm profile.out
            fi
        done

        bash <(curl -s https://codecov.io/bash)
esac
