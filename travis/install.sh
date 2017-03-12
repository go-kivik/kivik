#!/bin/bash
set -euC

# Install generic dependencies, needed for all builds
go get github.com/pborman/uuid \
    github.com/pkg/errors \
    golang.org/x/net/publicsuffix \
    github.com/flimzy/diff \
    golang.org/x/crypto/pbkdf2

case "$1" in
    "standard")
        # These dependencies are only needed for the server, so no need to
        # install them for GopherJS.
        go get github.com/NYTimes/gziphandler \
            github.com/dimfeld/httptreemux \
            github.com/spf13/cobra \
            github.com/spf13/pflag \
            github.com/ajg/form \
            github.com/justinas/alice
        go get -u github.com/jteeuwen/go-bindata/...
        go-bindata -pkg serve -nocompress -prefix "serve/files" -o serve/files.go serve/files
    ;;
    "gopherjs")
        if [ "$TRAVIS_OS_NAME" == "linux" ]; then
            # Install nodejs and dependencies, but only for Linux
            curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash -
            sudo apt-get update -qq
            sudo apt-get install -y nodejs
        fi
        npm install
        # Install Go deps only needed by PouchDB driver
        go get github.com/imdario/mergo
        # Then install GopherJS and related dependencies
        go get -u github.com/gopherjs/gopherjs; \
        go get -u -d -tags=js github.com/gopherjs/jsbuiltin; \
    ;;
    "linter")
        go get -u gopkg.in/alecthomas/gometalinter.v1
        gometalinter.v1 --install
    ;;
esac
