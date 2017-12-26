/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke profile ====

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type profile struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Gender     string `json:"gender"`
	Age        int    `json:"age"`
	Language   string `json:"language"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	switch function {
	case "init":
		return t.initProfile(stub, args)
	case "delete":
		return t.deleteProfile(stub, args)
	case "read":
		return t.readProfile(stub, args)
	case "update":
		return t.updateProfile(stub, args)
	case "query":
		return t.queryProfiles(stub, args)
	case "queryByLanguage":
		return t.queryProfilesByLanguage(stub, args)
	case "history":
		return t.getHistoryForProfile(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initProfile - create a new profile, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initProfile(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       1       2     3
	// "tupt", "male", "35", "React"
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init profile")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	profileName := args[0]
	gender := strings.ToLower(args[1])
	language := strings.ToLower(args[3])
	age, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if profile already exists ====
	profileAsBytes, err := stub.GetState(profileName)
	if err != nil {
		return shim.Error("Failed to get profile: " + err.Error())
	} else if profileAsBytes != nil {
		fmt.Println("This profile already exists: " + profileName)
		return shim.Error("This profile already exists: " + profileName)
	}

	// ==== Create profile object and marshal to JSON ====
	objectType := "profile"
	profile := &profile{objectType, profileName, gender, age, language}
	profileJSONasBytes, err := json.Marshal(profile)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save profile to state ===
	err = stub.PutState(profileName, profileJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the profile to enable language-based range queries, e.g. return all C# profile ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~language~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~language~*
	languageNameIndexKey, err := stub.CreateCompositeKey("language~name", []string{profile.Language, profile.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the profile.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	stub.PutState(languageNameIndexKey, []byte{0x00})

	// ==== Profile saved and indexed. Return success ====
	fmt.Println("- end init profile")
	return shim.Success(nil)
}

// ===============================================
// readProfile - read a profile from chaincode state
// ===============================================
func (t *SimpleChaincode) readProfile(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the profile to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the profile from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Profile does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a profile key/value pair from state
// ==================================================
func (t *SimpleChaincode) deleteProfile(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var profileJSON profile
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	profileName := args[0]

	// to maintain the language~name index, we need to read the profile first and get its language
	valAsbytes, err := stub.GetState(profileName) //get the profile from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + profileName + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Profile does not exist: " + profileName + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &profileJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + profileName + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(profileName) //remove the profile from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	languageNameIndexKey, err := stub.CreateCompositeKey("language~name", []string{profileJSON.Language, profileJSON.Name})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(languageNameIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}

// ===========================================================
// update a profile via profile name on the profile
// ===========================================================
func (t *SimpleChaincode) updateProfile(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1        2
	// "name", "Language", "Java"
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	profileName := args[0]
	propertyName := strings.Title(args[1])

	var propertyValue string = ""

	switch propertyName {
	case "Name":
		return shim.Error("Can not update Profile Name, must be unique")
	case "Age", "Language", "Gender":
		propertyValue = strings.ToLower(args[2])
		break
	default:
		return shim.Error("Unknow Property Name: " + propertyName)
	}

	profileAsBytes, err := stub.GetState(profileName)
	if err != nil {
		return shim.Error("Failed to get profile:" + err.Error())
	} else if profileAsBytes == nil {
		return shim.Error("Profile does not exist")
	}

	profileToUpdate := profile{}
	err = json.Unmarshal(profileAsBytes, &profileToUpdate) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}

	// //change the field
	propertyField := reflect.ValueOf(&profileToUpdate).Elem().FieldByName(propertyName)
	if propertyField.Kind() == reflect.Int {
		propertyValueInt, err := strconv.Atoi(propertyValue)
		if err != nil {
			return shim.Error("Value for " + propertyName + " must be a numeric string")
		}
		propertyField.SetInt(int64(propertyValueInt))
	} else {
		propertyField.SetString(propertyValue)

		// update index for language~name
		if propertyName == "Language" {
			languageNameIndexKey, err := stub.CreateCompositeKey("language~name", []string{propertyValue, profileName})
			if err != nil {
				return shim.Error(err.Error())
			}
			stub.PutState(languageNameIndexKey, []byte{0x00})
		}

	}

	//rewrite the profile
	profileJSONAsBytes, _ := json.Marshal(profileToUpdate)
	err = stub.PutState(profileName, profileJSONAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end update Profile (success)")
	return shim.Success(nil)
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryProfilesByLanguage queries for profile based on a passed in language.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryProfilesByLanguage(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "c#"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	language := strings.ToLower(args[0])

	// Query the language~name index by language
	// This will execute a key range query on all keys starting with 'language'
	languageProfileResultsIterator, err := stub.GetStateByPartialCompositeKey("language~name", []string{language})
	if err != nil {
		return shim.Error(err.Error())
	}
	// run after exit function
	defer languageProfileResultsIterator.Close()

	// buffer is a JSON array containing historic values for the profile
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	// Iterate through result set and for each profile found, transfer to newOwner
	var i int
	for i = 0; languageProfileResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the profile name from the composite key
		responseRange, err := languageProfileResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the language and name from language~name composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedLanguage := compositeKeyParts[0]
		returnedProfileName := compositeKeyParts[1]
		fmt.Printf("- found a profile from index:%s language:%s name:%s\n", objectType, returnedLanguage, returnedProfileName)

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"Name\":")
		buffer.WriteString("\"")
		buffer.WriteString(returnedProfileName)
		buffer.WriteString("\"")
		buffer.WriteString("}")

		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")

	fmt.Printf("- get profiles returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ===== Example: Ad hoc rich query ========================================================
// queryProfiles uses a query string to perform a query for profile.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryProfilesByLanguage example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryProfiles(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// this method using blockchain technology to retrieve the pledger
func (t *SimpleChaincode) getHistoryForProfile(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	profileName := args[0]

	fmt.Printf("- start getHistoryForProfile: %s\n", profileName)

	resultsIterator, err := stub.GetHistoryForKey(profileName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the profile
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON profile)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForProfile returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}
