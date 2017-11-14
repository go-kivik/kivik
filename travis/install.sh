#!/bin/bash
set -euC
set -o xtrace

if [ "$TRAVIS_OS_NAME" == "osx" ]; then
    brew install glide
fi

glide install

function generate {
    go get -u github.com/jteeuwen/go-bindata/...
    go generate $(go list ./... | grep -v /vendor/)
}

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
    docker pull ${DOCKER_IMAGE}
    docker run -d -p 6000:5984 -e COUCHDB_USER=admin -e COUCHDB_PASSWORD=abc123 --name couchdb ${DOCKER_IMAGE}
    wait_for_server http://localhost:6000/
    url=http://admin:abc123@localhost:6000
    case "${DOCKER_IMAGE}" in
        *:1.6.1)
        curl --silent --fail -o /dev/null -X PUT ${url}/_config/replicator/connection_timeout -d '"5000"'
        ;;
        *:2.0.0)
        curl --silent --fail -o /dev/null -X PUT ${url}/_users
        curl --silent --fail -o /dev/null -X PUT ${url}/_replicator
        curl --silent --fail -o /dev/null -X PUT ${url}/_global_changes
        ;;
        *:2.1.0)
        curl --silent --fail -o /dev/null -X PUT ${url}/_users
        curl --silent --fail -o /dev/null -X PUT ${url}/_replicator
        curl --silent --fail -o /dev/null -X PUT ${url}/_global_changes
        curl --silent --fail -o /dev/null -X PUT ${url}/_node/nonode@nohost/_config/replicator/update_docs -H 'Content-Type: application/json' -d '"true"' # FIXME: https://github.com/flimzy/kivik/issues/215
        ;;
    esac
}

case "$1" in
    "standard")
        generate
    ;;
    "docker")
        setup_docker
        generate
    ;;
    "gopherjs")
        if [ "$TRAVIS_OS_NAME" == "linux" ]; then
            # Install nodejs and dependencies, but only for Linux
            curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash -
            sudo apt-get update -qq
            sudo apt-get install -y nodejs
        fi
        npm install
        # Install Go deps only needed by PouchDB driver/GopherJS
        glide -y glide.gopherjs.yaml install
        # Then install GopherJS and related dependencies
        go get -u github.com/gopherjs/gopherjs

        # Source maps (mainly to make GopherJS quieter; I don't really care
        # about source maps in Travis)
        npm install source-map-support

        # Set up GopherJS for syscalls
        (
            cd $GOPATH/src/github.com/gopherjs/gopherjs/node-syscall/
            npm install --global node-gyp
            node-gyp rebuild
            mkdir -p ~/.node_libraries/
            cp build/Release/syscall.node ~/.node_libraries/syscall.node
        )

        go get -u -d -tags=js github.com/gopherjs/jsbuiltin
        setup_docker
        generate
    ;;
    "linter")
        go get -u gopkg.in/alecthomas/gometalinter.v1
        gometalinter.v1 --install
    ;;
    "coverage")
        generate
    ;;
esac
