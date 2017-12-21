/*
 Copyright ONECHAIN 2017 All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
 */
"use strict";
var path = require("path");
var fs = require("fs");
var util = require("util");
var hfc = require("fabric-client");
var Peer = require("fabric-client/lib/Peer.js");
var helper = require("./helper.js");
var log4js = require("log4js");
var logger = log4js.getLogger("channelRouter");
// var logger = helper.getLogger("invoke-chaincode");
var EventHub = require("fabric-client/lib/EventHub.js");

var hfc = require("fabric-client");
hfc.addConfigFile(path.join(__dirname, "../config.json"));
var ORGS = hfc.getConfigSetting("network-config");

var invokeChaincode = function(
    peers,
    channelName,
    chaincodeName,
    fcn,
    args,
    username,
    org
) {
    logger.debug(
        util.format(
            "\n============ invoke transaction on organization %s ============\n",
            org
        )
    );

    var client = helper.getClientForOrg(org);
    var channel = helper.getChannelForOrg(org, channelName);
    var targets = helper.newPeers(
        peers.map(peer => helper.getPeerAddressByName(org, peer))
    );
    var tx_id = null;
    var eh = null;

    return helper
        .getOrgAdmin(org)
        .then(admin => {
            tx_id = client.newTransactionID();
            logger.debug(util.format('Assigning transaction "%j"', tx_id));

            // send proposal to endorser
            var request = {
                targets: targets,
                chaincodeId: chaincodeName,
                fcn: fcn,
                args: args,
                chainId: channelName,
                txId: tx_id
            };

            return channel.sendTransactionProposal(request);
        })
        .then(results => {
            var proposalResponses = results[0];
            var proposal = results[1];
            var header = results[2];
            var isProposalGood = false;
            if (
                proposalResponses &&
                proposalResponses[0].response &&
                proposalResponses[0].response.status === 200
            ) {
                isProposalGood = true;
                logger.info("Transaction proposal was good");
            } else {
                logger.error("Transaction proposal was bad");
            }

            if (isProposalGood) {
                logger.info(
                    util.format(
                        'Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s", metadata - "%s"',
                        proposalResponses[0].response.status,
                        proposalResponses[0].response.message,
                        proposalResponses[0].response.payload
                    )
                );
                var request = {
                    proposalResponses: proposalResponses,
                    proposal: proposal,
                    header: header
                };
                // set the transaction listener and set a timeout of 30sec
                // if the transaction did not get committed within the timeout period,
                // fail the test
                var deployId = tx_id.getTransactionID();

                eh = client.newEventHub();
                // by default using first peer
                helper.setEventHub(org, peers[0], eh);
                eh.connect();

                let txPromise = new Promise((resolve, reject) => {
                    let handle = setTimeout(() => {
                        eh.disconnect();
                        reject({ event_status: "TIMEOUT" });
                    }, 30000);

                    eh.registerTxEvent(deployId, (tx, code) => {
                        logger.info(
                            "The chaincode instantiate transaction has been committed on peer " +
                                eh._ep._endpoint.addr
                        );
                        clearTimeout(handle);
                        eh.unregisterTxEvent(deployId);
                        eh.disconnect();

                        if (code !== "VALID") {
                            logger.error(
                                "The transaction was invalid, code = " + code
                            );
                            reject(code);
                        } else {
                            logger.info(
                                "The transaction has been committed on peer " +
                                    eh._ep._endpoint.addr
                            );
                            resolve({
                                event_status: code,
                                tx_id: deployId
                            });
                        }
                    });
                });
                var sendPromise = channel.sendTransaction(request);
                return Promise.all([sendPromise, txPromise]);
            } else {
                var errorMessage =
                    "Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...";
                logger.error(errorMessage);
                return errorMessage;
            }
        })
        .then(results => {
            logger.info(
                "Send transaction promise and event listener promise have completed"
            );
            // check the results in the order the promises were added to the promise all list
            if (results && results[0] && results[0].status === "SUCCESS") {
                logger.info("Successfully sent transaction to the orderer.");
            } else {
                logger.error(
                    "Failed to order the transaction. Error code: " +
                        results[0].status
                );
            }

            if (results && results[1] && results[1].event_status === "VALID") {
                logger.info(
                    "Successfully committed the change to the ledger by the peer: " +
                        tx_id.getTransactionID()
                );
            } else {
                logger.error(
                    "Transaction failed to be committed to the ledger due to ::" +
                        results[1].event_status
                );
            }

            return results;
        })
        .catch(err => {
            var errorMessage =
                "Failed to send transaction due to error: " + err.stack
                    ? err.stack
                    : err;
            logger.error(errorMessage);
            return errorMessage;
        });
};

exports.invokeChaincode = invokeChaincode;
