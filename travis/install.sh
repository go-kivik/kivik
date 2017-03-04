#!/bin/bash
set -euC

# Install generic dependencies, needed for all builds
go get github.com/pborman/uuid \
    github.com/dimfeld/httptreemux \
    github.com/pkg/errors \
    github.com/spf13/cobra \
    github.com/spf13/pflag \
    golang.org/x/net/publicsuffix

case "$1" in
    "standard")
        go get github.com/NYTimes/gziphandler
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
