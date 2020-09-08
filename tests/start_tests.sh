#!/bin/bash
set -e 

exists=$(docker ps | egrep "solr|zookeeper" | awk '{ print $1 }')

if [ -z "${exists}" ]; then
  $GOPATH/src/github.com/uol/solr/scripts/start_solr.sh
fi

TESTHOME="github.com/uol/solr/tests"

go test -count=1 -v -timeout 5m $TESTHOME -run ^Test

docker rm -f $(docker ps -a -q)

printf "finalized tests\n"
