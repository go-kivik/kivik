#!/bin/bash

# This script is meant to run locally while testing Kivik. It starts various
# versions of CouchDB in docker, for testing.

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export COUCHDB_USER=admin
export COUCHDB_PASSWORD=abc123
export KIVIK_TEST_DSN_COUCH22=http://admin:abc123@localhost:6002/
export KIVIK_TEST_DSN_COUCH23=http://admin:abc123@localhost:6003/
export KIVIK_TEST_DSN_COUCH30=http://admin:abc123@localhost:6004/
export KIVIK_TEST_DSN_COUCH31=http://admin:abc123@localhost:6005/
export KIVIK_TEST_DSN_COUCH32=http://admin:abc123@localhost:6006/
export KIVIK_TEST_DSN_COUCH33=http://admin:abc123@localhost:6007/

echo "CouchDB 2.2.0"
docker pull couchdb:2.2.0
docker run --name couch22 -p 6002:5984/tcp -d --rm -e COUCHDB_USER -e COUCHDB_PASSWORD couchdb:2.2.0
${DIR}/complete_couch2.sh $KIVIK_TEST_DSN_COUCH22

echo "CouchDB 2.3.1"
docker pull apache/couchdb:2.3.1
docker run --name couch23 -p 6003:5984/tcp -d --rm -e COUCHDB_USER -e COUCHDB_PASSWORD apache/couchdb:2.3.1
${DIR}/complete_couch2.sh $KIVIK_TEST_DSN_COUCH23

echo "CouchDB 3.0.0"
docker pull couchdb:3.0.0
docker run --name couch30 -p 6004:5984/tcp -d --rm -e COUCHDB_USER -e COUCHDB_PASSWORD apache/couchdb:3.0.0
${DIR}/complete_couch2.sh $KIVIK_TEST_DSN_COUCH30

echo "CouchDB 3.1.2"
docker pull apache/couchdb:3.1.2
docker run --name couch31 -p 6005:5984/tcp -d --rm -e COUCHDB_USER -e COUCHDB_PASSWORD apache/couchdb:3.1.2
${DIR}/complete_couch2.sh $KIVIK_TEST_DSN_COUCH31

echo "CouchDB 3.2.3"
docker pull apache/couchdb:3.2.3
docker run --name couch32 -p 6006:5984/tcp -d --rm -e COUCHDB_USER -e COUCHDB_PASSWORD apache/couchdb:3.2.3
${DIR}/complete_couch2.sh $KIVIK_TEST_DSN_COUCH32

echo "CouchDB 3.3.2"
docker pull apache/couchdb:3.3.2
docker run --name couch33 -p 6007:5984/tcp -d --rm -e COUCHDB_USER -e COUCHDB_PASSWORD apache/couchdb:3.3.2
${DIR}/complete_couch2.sh $KIVIK_TEST_DSN_COUCH33
