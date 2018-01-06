#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0
# This code is based on code written by the Hyperledger Fabric community. 
# Original code can be found here: https://github.com/hyperledger/fabric-samples/blob/release/fabcar/startFabric.sh
#
# Exit on first error

set -e

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1

starttime=$(date +%s)

# launch network; create channel and join peer to channel
cd ../basic-network
./start.sh

# Now launch the CLI container in order to install, instantiate chaincode
# and prime the ledger with our 10 tuna catches
# docker-compose -f ./docker-compose.yml up -d cli

# Create the channel
docker exec cli peer channel create -o orderer.example.com:7050 -c mychannel -f /etc/hyperledger/configtx/channel.tx --tls true --cafile /etc/hyperledger/msp/orderer/msp/tlscacerts/tlsca.example.com-cert.pem
# Join peer0.org1.example.com to the channel.
docker exec cli peer channel join -b mychannel.block
# install chaincode
docker exec cli peer chaincode install -n tuna-app -v 1.0 -p github.com/tuna-app
docker exec cli peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n tuna-app -v 1.0 --tls true --cafile /etc/hyperledger/msp/orderer/msp/tlscacerts/tlsca.example.com-cert.pem -c '{"Args":[""]}'
# docker exec cli peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n tuna-app -v 1.0 --tls true --cafile /etc/hyperledger/msp/orderer/msp/tlscacerts/tlsca.example.com-cert.pem -c '{"Args":[""]}' -P "OR ('Org1MSP.member', 'Org2MSP.member')"
sleep 10
docker exec cli peer chaincode invoke -o orderer.example.com:7050 -C mychannel -n tuna-app --tls true --cafile /etc/hyperledger/msp/orderer/msp/tlscacerts/tlsca.example.com-cert.pem -c '{"function":"initLedger","Args":[""]}'

printf "\nTotal execution time : $(($(date +%s) - starttime)) secs ...\n\n"

docker run --name mysql -e MYSQL_ROOT_PASSWORD=123456 -p 3306:3306 -d mysql
docker cp ../blockchain-explorer/db mysql:/db
docker exec -i mysql mysql -uroot -p123456 < db/fabricexplorer.sql
