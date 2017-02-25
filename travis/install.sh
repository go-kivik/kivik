#!/bin/bash
set -euC

# Install generic dependencies, needed for all builds
go get github.com/pborman/uuid \
    github.com/dimfeld/httptreemux \
    github.com/pkg/errors \
    github.com/spf13/cobra \
    github.com/spf13/pflag

case "$1" in
    "gopherjs")
        # Install nodejs and dependencies
        curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash -
        sudo apt-get update -qq
        sudo apt-get install -y nodejs
        npm install
        # Then install GopherJS and related dependencies
        go get -u github.com/gopherjs/gopherjs; \
        go get -u -d -tags=js github.com/gopherjs/jsbuiltin; \
    ;;
    "linter")
        go get -u gopkg.in/alecthomas/gometalinter.v1
        gometalinter.v1 --install
    ;;
esac
