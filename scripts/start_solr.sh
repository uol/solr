#!/bin/bash -x

POD_ZOOKEEPER_NAME="zookeeper"
exists=$(docker ps | egrep "solr|zookeeper" | awk '{ print $1 }')

if [ -n "${exists}" ]; then
  docker rm -f solr1 solr2 solr3 zookeeper
fi

docker run -d --name "${POD_ZOOKEEPER_NAME}" --net=timeseriesNetwork zookeeper:3.4.11

sleep 2

echo "Zookeeper OK"

SOLRDOCKERPATH="$GOPATH/src/github.com/uol/solr/scripts"
cd ${SOLRDOCKERPATH}

zookeeperIP=$(docker inspect --format "{{ .NetworkSettings.Networks.timeseriesNetwork.IPAddress }}" zookeeper)

POD_NAME="solr"


for i in {1..3}; do
    docker run -d --name "${POD_NAME}${i}" "--net=timeseriesNetwork" -v "${SOLRDOCKERPATH}/solr-configs":/solr-configs --restart always solr:7.4.0-alpine -cloud -z ${zookeeperIP}
    sleep 2
done

docker exec solr1 /opt/solr/bin/solr zk cp file:/solr-configs/solr.xml zk:/solr.xml -z ${zookeeperIP}:2181

docker exec solr1 /opt/solr/bin/solr zk upconfig -n mycenae -d /solr-configs/mycenae -z ${zookeeperIP}:2181

echo "Solr OK"

