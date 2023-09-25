#!/bin/bash -euC

pkgname=github.com/${CI_PROJECT_PATH}

# Find all packages that depend on this package. We can pass this to -coverpkg
# so that lines in these packages are counted as well.
find_deps() {
	(
		echo "$1"
		go list -f $'{{range $f := .Deps}}{{$f}}\n{{end}}' "$1"
		go list -f $'{{range $f := .TestImports}}{{$f}}\n{{end}}' "$1" |
			while read imp; do
				go list -f $'{{range $f := .Deps}}{{$f}}\n{{end}}' "$imp";
			done
	) | sort -u | grep ^$pkgname | grep -Ev "^$pkgname/(vendor|test)" |
		tr '\n' ' ' | sed 's/ $//' | tr ' ' ','
}

echo "" > coverage.txt

TEST_PKGS=$(go list ./... | grep -v /test)

# Cache
go test -i -cover -covermode=set $(go list ./...)

for pkg in $TEST_PKGS; do
    go test \
        -coverprofile=profile.out \
        -covermode=set \
        -coverpkg=$(find_deps "$pkg") \
        "$pkg" | grep -v "warning: no packages being tested depend on "

    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

curl -fs https://codecov.io/bash | bash -s -- -Z
