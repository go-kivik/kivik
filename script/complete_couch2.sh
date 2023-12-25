#!/bin/sh -e

for db in _users _replicator _global_changes; do
echo ${1}/${db}
    status=$(curl --silent --write-out "%{http_code}" -o /dev/null -u ${COUCHDB_USER}:${COUCHDB_PASSWORD} -X PUT "${1}/${db}")
    case ${status} in
        2*)
            # Success!
        ;;
        412)
            # Already exists, nothing to do.
        ;;
        *)
            echo "Failed to create ${db}: ${status}"
            exit 1
        ;;
    esac
done
curl --silent --fail -o /dev/null -u ${COUCHDB_USER}:${COUCHDB_PASSWORD} -X PUT "${1}/_node/nonode@nohost/_config/replicator/interval" -d '"1000"'
curl --silent --fail -o /dev/null -u ${COUCHDB_USER}:${COUCHDB_PASSWORD} -X PUT "${1}/_node/nonode@nohost/_config/cluster/n" -d '"1"'
