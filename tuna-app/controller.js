//SPDX-License-Identifier: Apache-2.0

/*
  This code is based on code written by the Hyperledger Fabric community.
  Original code can be found here: https://github.com/hyperledger/fabric-samples/blob/release/fabcar/query.js
  and https://github.com/hyperledger/fabric-samples/blob/release/fabcar/invoke.js
 */

import Fabric_Client from "fabric-client";
import path from "path";
import util from "util";

export default {
	get_member_user(fabric_client) {
		const store_path = path.join(__dirname, "hfc-key-store");
		console.log("Store path:" + store_path);
		var tx_id = null;

		// create the key value store as defined in the fabric-client/config/default.json 'key-value-store' setting
		return Fabric_Client.newDefaultKeyValueStore({
			path: store_path
		})
			.then(state_store => {
				// assign the store to the fabric client
				fabric_client.setStateStore(state_store);
				var crypto_suite = Fabric_Client.newCryptoSuite();
				// use the same location for the state store (where the users' certificate are kept)
				// and the crypto store (where the users' keys are kept)
				var crypto_store = Fabric_Client.newCryptoKeyStore({
					path: store_path
				});
				crypto_suite.setCryptoKeyStore(crypto_store);
				fabric_client.setCryptoSuite(crypto_suite);

				// get the enrolled user from persistence, this user will sign all requests
				return fabric_client.getUserContext("user1", true);
			})
			.then(user_from_store => {
				if (user_from_store && user_from_store.isEnrolled()) {
					console.log("Successfully loaded user1 from persistence");
					return user_from_store;
				} else {
					throw new Error("Failed to get user1.... run registerUser.js");
				}
			});
	},

	get_all_tuna(req, res) {
		console.log("getting all tuna from database: ");

		var fabric_client = new Fabric_Client();

		// setup the fabric network
		var channel = fabric_client.newChannel("mychannel");
		var peer = fabric_client.newPeer("grpcs://localhost:7051");
		channel.addPeer(peer);

		//
		let tx_id = null;

		// create the key value store as defined in the fabric-client/config/default.json 'key-value-store' setting
		this.get_member_user(fabric_client)
			.then(member_user => {
				// queryAllTuna - requires no arguments , ex: args: [''],
				const request = {
					chaincodeId: "tuna-app",
					txId: tx_id,
					fcn: "queryAllTuna",
					args: [""]
				};

				// send the query proposal to the peer
				return channel.queryByChaincode(request);
			})
			.then(query_responses => {
				console.log("Query has completed, checking results");
				// query_responses could have more than one  results if there multiple peers were used as targets
				if (query_responses && query_responses.length == 1) {
					if (query_responses[0] instanceof Error) {
						console.error("error from query = ", query_responses[0]);
					} else {
						console.log("Response is ", query_responses[0].toString());
						res.json(JSON.parse(query_responses[0].toString()));
					}
				} else {
					console.log("No payloads were returned from query");
				}
			})
			.catch(err => {
				console.error("Failed to query successfully :: " + err);
			});
	},

	add_tuna(req, res) {
		console.log("submit recording of a tuna catch: ");

		const array = req.params.tuna.split("-");
		console.log(util.format("Get sent data: %j ", array));
		const [key, location, timestamp, holder, vessel, weight] = array;

		var fabric_client = new Fabric_Client();

		// setup the fabric network
		var channel = fabric_client.newChannel("mychannel");
		var peer = fabric_client.newPeer("grpcs://localhost:7051");
		channel.addPeer(peer);
		var order = fabric_client.newOrderer("grpcs://localhost:7050");
		channel.addOrderer(order);

		let tx_id = null;

		this.get_member_user(fabric_client)
			.then(member_user => {
				// get a transaction id object based on the current user assigned to fabric client
				tx_id = fabric_client.newTransactionID();
				console.log("Assigning transaction_id: ", tx_id._transaction_id);

				// recordTuna - requires 5 args, ID, vessel, location, timestamp,holder - ex: args: ['10', 'Hound', '-12.021, 28.012', '1504054225', 'Hansel'],
				// send proposal to endorser
				const request = {
					//targets : --- letting this default to the peers assigned to the channel
					chaincodeId: "tuna-app",
					fcn: "recordTuna",
					args: [key, vessel, location, timestamp, holder, weight],
					chainId: "mychannel",
					txId: tx_id
				};

				// send the transaction proposal to the peers
				return channel.sendTransactionProposal(request);
			})
			.then(results => {
				var proposalResponses = results[0];
				var proposal = results[1];
				let isProposalGood = false;
				if (
					proposalResponses &&
					proposalResponses[0].response &&
					proposalResponses[0].response.status === 200
				) {
					isProposalGood = true;
					console.log("Transaction proposal was good");
				} else {
					console.error("Transaction proposal was bad");
				}
				if (isProposalGood) {
					console.log(
						util.format(
							'Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s"',
							proposalResponses[0].response.status,
							proposalResponses[0].response.message
						)
					);

					// build up the request for the orderer to have the transaction committed
					var request = {
						proposalResponses: proposalResponses,
						proposal: proposal
					};

					// set the transaction listener and set a timeout of 30 sec
					// if the transaction did not get committed within the timeout period,
					// report a TIMEOUT status
					var transaction_id_string = tx_id.getTransactionID(); //Get the transaction ID string to be used by the event processing
					var promises = [];

					var sendPromise = channel.sendTransaction(request);
					promises.push(sendPromise); //we want the send transaction first, so that we know where to check status

					// get an eventhub once the fabric client has a user assigned. The user
					// is required bacause the event registration must be signed
					let event_hub = fabric_client.newEventHub();
					event_hub.setPeerAddr("grpcs://localhost:7053");

					// using resolve the promise so that result status may be processed
					// under the then clause rather than having the catch clause process
					// the status
					let txPromise = new Promise((resolve, reject) => {
						let handle = setTimeout(() => {
							event_hub.disconnect();
							resolve({ event_status: "TIMEOUT" }); //we could use reject(new Error('Trnasaction did not complete within 30 seconds'));
						}, 3000);
						event_hub.connect();
						event_hub.registerTxEvent(
							transaction_id_string,
							(tx, code) => {
								// this is the callback for transaction event status
								// first some clean up of event listener
								clearTimeout(handle);
								event_hub.unregisterTxEvent(transaction_id_string);
								event_hub.disconnect();

								// now let the application know what happened
								var return_status = {
									event_status: code,
									tx_id: transaction_id_string
								};
								if (code !== "VALID") {
									console.error("The transaction was invalid, code = " + code);
									resolve(return_status); // we could use reject(new Error('Problem with the tranaction, event status ::'+code));
								} else {
									console.log(
										"The transaction has been committed on peer " +
											event_hub._ep._endpoint.addr
									);
									resolve(return_status);
								}
							},
							err => {
								//this is the callback if something goes wrong with the event registration or processing
								reject(
									new Error("There was a problem with the eventhub ::" + err)
								);
							}
						);
					});
					promises.push(txPromise);

					return Promise.all(promises);
				} else {
					console.error(
						"Failed to send Proposal or receive valid response. Response null or status is not 200. exiting..."
					);
					throw new Error(
						"Failed to send Proposal or receive valid response. Response null or status is not 200. exiting..."
					);
				}
			})
			.then(results => {
				console.log(
					"Send transaction promise and event listener promise have completed"
				);
				// check the results in the order the promises were added to the promise all list
				if (results && results[0] && results[0].status === "SUCCESS") {
					console.log("Successfully sent transaction to the orderer.");
					res.send(tx_id.getTransactionID());
				} else {
					console.error(
						"Failed to order the transaction. Error code: " + response.status
					);
				}

				if (results && results[1] && results[1].event_status === "VALID") {
					console.log(
						"Successfully committed the change to the ledger by the peer: " +
							tx_id.getTransactionID()
					);
				} else {
					console.log(
						"Transaction failed to be committed to the ledger due to ::" +
							results[1].event_status
					);
				}
			})
			.catch(err => {
				console.error("Failed to invoke successfully :: " + err);
			});
	},

	get_tuna(req, res) {
		var fabric_client = new Fabric_Client();
		var key = req.params.id;

		// setup the fabric network
		var channel = fabric_client.newChannel("mychannel");
		var peer = fabric_client.newPeer("grpcs://localhost:7051");
		channel.addPeer(peer);

		//
		let tx_id = null;

		this.get_member_user(fabric_client)
			.then(member_user => {
				// queryTuna - requires 1 argument, ex: args: ['4'],
				const request = {
					chaincodeId: "tuna-app",
					txId: tx_id,
					fcn: "queryTuna",
					args: [key]
				};

				// send the query proposal to the peer
				return channel.queryByChaincode(request);
			})
			.then(query_responses => {
				console.log("Query has completed, checking results");
				// query_responses could have more than one  results if there multiple peers were used as targets
				if (query_responses && query_responses.length == 1) {
					if (query_responses[0] instanceof Error) {
						console.error("error from query = ", query_responses[0]);
						res.send("Could not locate tuna");
					} else {
						console.log("Response is ", query_responses[0].toString());
						res.send(query_responses[0].toString());
					}
				} else {
					console.log("No payloads were returned from query");
					res.send("Could not locate tuna");
				}
			})
			.catch(err => {
				console.error("Failed to query successfully :: " + err);
				res.send("Could not locate tuna");
			});
	},
	change_holder(req, res) {
		console.log("changing holder of tuna catch: ");

		var array = req.params.holder.split("-");
		var key = array[0];
		var holder = array[1];

		var fabric_client = new Fabric_Client();

		// setup the fabric network
		var channel = fabric_client.newChannel("mychannel");
		var peer = fabric_client.newPeer("grpcs://localhost:7051");
		channel.addPeer(peer);
		var order = fabric_client.newOrderer("grpcs://localhost:7050");
		channel.addOrderer(order);

		let tx_id = null;

		// create the key value store as defined in the fabric-client/config/default.json 'key-value-store' setting
		this.get_member_user(fabric_client)
			.then(member_user => {
				// get a transaction id object based on the current user assigned to fabric client
				tx_id = fabric_client.newTransactionID();
				console.log("Assigning transaction_id: ", tx_id._transaction_id);

				// changeTunaHolder - requires 2 args , ex: args: ['1', 'Barry'],
				// send proposal to endorser
				var request = {
					//targets : --- letting this default to the peers assigned to the channel
					chaincodeId: "tuna-app",
					fcn: "changeTunaHolder",
					args: [key, holder],
					chainId: "mychannel",
					txId: tx_id
				};

				// send the transaction proposal to the peers
				return channel.sendTransactionProposal(request);
			})
			.then(results => {
				var proposalResponses = results[0];
				var proposal = results[1];
				let isProposalGood = false;
				if (
					proposalResponses &&
					proposalResponses[0].response &&
					proposalResponses[0].response.status === 200
				) {
					isProposalGood = true;
					console.log("Transaction proposal was good");
				} else {
					console.error("Transaction proposal was bad");
				}
				if (isProposalGood) {
					console.log(
						util.format(
							'Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s"',
							proposalResponses[0].response.status,
							proposalResponses[0].response.message
						)
					);

					// build up the request for the orderer to have the transaction committed
					var request = {
						proposalResponses: proposalResponses,
						proposal: proposal
					};

					// set the transaction listener and set a timeout of 30 sec
					// if the transaction did not get committed within the timeout period,
					// report a TIMEOUT status
					var transaction_id_string = tx_id.getTransactionID(); //Get the transaction ID string to be used by the event processing
					var promises = [];
					console.log(request, channel);
					var sendPromise = channel.sendTransaction(request);
					promises.push(sendPromise); //we want the send transaction first, so that we know where to check status

					// get an eventhub once the fabric client has a user assigned. The user
					// is required bacause the event registration must be signed
					let event_hub = fabric_client.newEventHub();
					event_hub.setPeerAddr("grpcs://localhost:7053");

					// using resolve the promise so that result status may be processed
					// under the then clause rather than having the catch clause process
					// the status
					let txPromise = new Promise((resolve, reject) => {
						let handle = setTimeout(() => {
							event_hub.disconnect();
							resolve({ event_status: "TIMEOUT" }); //we could use reject(new Error('Trnasaction did not complete within 30 seconds'));
						}, 3000);
						event_hub.connect();
						event_hub.registerTxEvent(
							transaction_id_string,
							(tx, code) => {
								// this is the callback for transaction event status
								// first some clean up of event listener
								clearTimeout(handle);
								event_hub.unregisterTxEvent(transaction_id_string);
								event_hub.disconnect();

								// now let the application know what happened
								var return_status = {
									event_status: code,
									tx_id: transaction_id_string
								};
								if (code !== "VALID") {
									console.error("The transaction was invalid, code = " + code);
									resolve(return_status); // we could use reject(new Error('Problem with the tranaction, event status ::'+code));
								} else {
									console.log(
										"The transaction has been committed on peer " +
											event_hub._ep._endpoint.addr
									);
									resolve(return_status);
								}
							},
							err => {
								//this is the callback if something goes wrong with the event registration or processing
								reject(
									new Error("There was a problem with the eventhub ::" + err)
								);
							}
						);
					});
					promises.push(txPromise);

					return Promise.all(promises);
				} else {
					console.error(
						"Failed to send Proposal or receive valid response. Response null or status is not 200. exiting..."
					);
					res.send("Error: no tuna catch found");
					// throw new Error('Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...');
				}
			})
			.then(results => {
				console.log(
					"Send transaction promise and event listener promise have completed"
				);
				// check the results in the order the promises were added to the promise all list
				if (results && results[0] && results[0].status === "SUCCESS") {
					console.log("Successfully sent transaction to the orderer.");
					res.send(tx_id.getTransactionID());
				} else {
					console.error(
						"Failed to order the transaction. Error code: " + response.status
					);
					res.send("Error: no tuna catch found");
				}

				if (results && results[1] && results[1].event_status === "VALID") {
					console.log(
						"Successfully committed the change to the ledger by the peer: " +
							tx_id.getTransactionID()
					);
				} else {
					console.log(
						"Transaction failed to be committed to the ledger due to ::" +
							results[1].event_status
					);
				}
			})
			.catch(err => {
				console.error("Failed to invoke successfully :: " + err);
				res.send("Error: no tuna catch found");
			});
	}
};
