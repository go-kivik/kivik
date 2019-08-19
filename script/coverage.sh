#!/bin/bash -e

echo "" > coverage.txt

TEST_PKGS=$(go list ./... | grep -v /test/)

for d in $TEST_PKGS; do
    go test -i $d
    go test -coverprofile=profile.out -covermode=set "$d"
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

curl -fs https://codecov.io/bash | bash -s -- -Z
