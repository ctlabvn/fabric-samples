"use strict";
/*
* SPDX-License-Identifier: Apache-2.0
*/
/*
 * Chaincode Invoke

This code is based on code written by the Hyperledger Fabric community.
  Original code can be found here: https://gerrit.hyperledger.org/r/#/c/14395/4/fabcar/enrollAdmin.js
 */
var X509 = require("X509");
var Fabric_Client = require("fabric-client");
var Fabric_CA_Client = require("fabric-ca-client");
var forge = require("node-forge");
var path = require("path");
var util = require("util");
var os = require("os");

//
var fabric_client = new Fabric_Client();
var fabric_ca_client = null;
var admin_user = null;
var member_user = null;
// var store_path = path.join(os.homedir(), '.hfc-key-store');
var store_path = path.join(__dirname, "hfc-key-store");
console.log(" Store path:" + store_path);

function queryChaincode(channel, member_user) {
  // queryAllTuna - requires no arguments , ex: args: [''],
  const request = {
    chaincodeId: "shim-api",
    fcn: "getCreator",
    args: [""]
  };

  // send the query proposal to the peer
  channel
    .queryByChaincode(request)
    .then(query_responses => {
      console.log("Query has completed, checking results");
      // query_responses could have more than one  results if there multiple peers were used as targets
      if (query_responses && query_responses.length == 1) {
        if (query_responses[0] instanceof Error) {
          console.error("error from query = ", query_responses[0]);
        } else {
          console.log("Response is ", query_responses[0].toString());
          console.log(JSON.parse(query_responses[0].toString()));
        }
      } else {
        console.log("No payloads were returned from query");
      }
    })
    .catch(err => {
      console.error("Failed to query successfully :: " + err);
    });
}

// create the key value store as defined in the fabric-client/config/default.json 'key-value-store' setting
Fabric_Client.newDefaultKeyValueStore({
  path: store_path
})
  .then(state_store => {
    // assign the store to the fabric client
    fabric_client.setStateStore(state_store);
    var crypto_suite = Fabric_Client.newCryptoSuite();
    // use the same location for the state store (where the users' certificate are kept)
    // and the crypto store (where the users' keys are kept)
    var crypto_store = Fabric_Client.newCryptoKeyStore({ path: store_path });
    crypto_suite.setCryptoKeyStore(crypto_store);
    fabric_client.setCryptoSuite(crypto_suite);
    var tlsOptions = {
      trustedRoots: [],
      verify: false
    };
    // be sure to change the http to https when the CA is running TLS enabled
    fabric_ca_client = new Fabric_CA_Client(
      "https://localhost:7054",
      // null,
      // "",
      tlsOptions,
      "ca.example.com",
      crypto_suite
    );

    // first check to see if the admin is already enrolled
    return fabric_client.getUserContext("use4", true);
  })
  .then(user_from_store => {
    // user_from_store.attrs = [{ name: "hf.Registrar.Roles", value: "client" }];
    // user_from_store.setRoles(["client", "member"]);
    // fabric_client.saveUserToStateStore();
    // user_from_store.setAffiliation("org1.department1");
    // attrs: [{ name: "hf.Registrar.Roles", value: "client" }];
    // const cert = X509.parseCert(
    //   Fabric_CA_Client.normalizeX509(user_from_store.getIdentity()._certificate)
    // );

    // setup the fabric network
    // var channel = fabric_client.newChannel("mychannel");

    // let data = fs.readFileSync(ORGS[org][key]["tls_cacerts"]);
    //     peer = client.newPeer("grpcs://localhost:7051", {
    //       pem: Buffer.from(data).toString(),
    //       "ssl-target-name-override": "peer0"
    //     });

    // channel.addPeer(peer);
    // queryChaincode(channel, user_from_store)

    // // console.log(user_from_store._identity._certificate);
    // var cert1 = forge.pki.createCertificate();
    // cert1.setSubject(cert.subject);
    // cert1.setIssuer(cert.issuer);
    // cert1.subject.hash = cert.subjectHash;
    // cert1.serialNumber = cert.serial;
    // cert1.publicKey = cert.publicKey;
    // cert1.validity.notBefore = cert.notBefore;
    // cert1.validity.notAfter = cert.notAfter;
    // Object.keys(cert.extensions).forEach(key =>
    //   cert1.extensions.push({
    //     name: key,
    //     value: cert.extensions[key]
    //   })
    // );
    // var pem = forge.pki.certificateToAsn1(cert1);

    // queryChaincode();

    // console.log(cert);

    // setup the fabric network

    if (user_from_store && user_from_store.isEnrolled()) {
      return user_from_store;
    }

    return fabric_ca_client
      .enroll({
        enrollmentID: "user1",
        enrollmentSecret: "userpw",
        attr_reqs: [{ name: "permission", optional: true }]
      })
      .then(enrollment => {
        const cert = X509.parseCert(
          Fabric_CA_Client.normalizeX509(enrollment.certificate)
        );
        console.log(
          'Successfully enrolled admin user "user1"'
          // cert
        );

        console.log(cert.extensions);

        return fabric_client.createUser({
          username: "user1",
          mspid: "Org1MSP",
          affiliation: "org1.department1",
          cryptoContent: {
            privateKeyPEM: enrollment.key.toBytes(),
            signedCertPEM: enrollment.certificate
          }
        });
      });
  })
  .then(user => {
    fabric_client.saveUserToStateStore();
    console.log(user);
  })
  .catch(err => {
    console.error("Failed to enroll user1: " + err);
  });
