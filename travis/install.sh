#!/bin/bash
set -euC
set -o xtrace

# Install generic dependencies, needed for all builds
go get github.com/pborman/uuid \
    github.com/pkg/errors \
    golang.org/x/net/publicsuffix \
    github.com/flimzy/diff \
    golang.org/x/crypto/pbkdf2

function wait_for_server {
    printf "Waiting for $1"
    n=0
    until [ $n -gt 5 ]; do
        curl --output /dev/null --silent --head --fail $1 && break
        printf '.'
        n=$[$n+1]
        sleep 1
    done
    printf "ready!\n"
}

function setup_docker {
    if [ "$TRAVIS_OS_NAME" == "osx" ]; then
        return
    fi
    docker pull couchdb:1.6.1
    docker run -d -p 6000:5984 --name couchdb16 couchdb:1.6.1
    wait_for_server http://localhost:6000/
    curl -X PUT http://localhost:6000/_config/log/file -d '"/tmp/log"'
    curl -X PUT http://localhost:6000/_config/admins/admin -d '"abc123"'
    docker pull couchdb:latest
    docker run -d -p 6001:5984 --name couchdb20 couchdb:latest
    wait_for_server http://localhost:6001/
    curl -X PUT http://localhost:6001/_config/log/file -d '"/tmp/log"'
    curl -X PUT http://localhost:6001/_users
    curl -X PUT http://localhost:6001/_replicator
    curl -X PUT http://localhost:6001/_global_changes
    curl -X PUT http://localhost:6001/_config/admins/admin -d '"abc123"'
}

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
        go generate ./...
        setup_docker
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
        go get -u github.com/gopherjs/gopherjs
        go get -u -d -tags=js github.com/gopherjs/jsbuiltin
        setup_docker
    ;;
    "linter")
        go get -u gopkg.in/alecthomas/gometalinter.v1
        gometalinter.v1 --install
    ;;
esac
