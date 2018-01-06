#!/bin/bash

rm -rf /tmp/fabric-client-kvs_peerOrg*
docker run --name mysql -e MYSQL_ROOT_PASSWORD=123456 -p 3306:3306 -d mysql
docker cp ./db mysql:/db
sleep 2
docker exec -i mysql mysql -uroot -p123456 < db/fabricexplorer.sql
sleep 5
node main.js
# node main.js >log.log 2>&1 &
