/*
Copyright IBM Corp. 2016 All Rights Reserved.
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
package main

import (
  "encoding/json"
  "fmt"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  "testing"
)

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
  res := stub.MockInit("1", args)
  if res.Status != shim.OK {
    fmt.Println("Init failed", string(res.Message))
    t.FailNow()
  }
}

func checkQuery(t *testing.T, stub *shim.MockStub, name string, value string) {
  res := stub.MockInvoke("1", [][]byte{[]byte("queryTuna"), []byte(name)})
  if res.Status != shim.OK {
    fmt.Println("Query", name, "failed", string(res.Message))
    t.FailNow()
  }
  if res.Payload == nil {
    fmt.Println("Query", name, "failed to get value")
    t.FailNow()
  }

  tuna := Tuna{}
  json.Unmarshal(res.Payload, &tuna)

  if tuna.Holder != value {
    fmt.Println("Query value", name, "was not", value, "as expected")
    t.FailNow()
  }
}

func checkInvoke(t *testing.T, stub *shim.MockStub, args [][]byte) {
  res := stub.MockInvoke("1", args)
  if res.Status != shim.OK {
    fmt.Println("Invoke", args, "failed", string(res.Message))
    t.FailNow()
  }
}

func Test_Chaincode(t *testing.T) {
  scc := new(SmartContract)
  stub := shim.NewMockStub("ex", scc)
  checkInit(t, stub, [][]byte{[]byte("init"), []byte("")})
  checkInvoke(t, stub, [][]byte{[]byte("recordTuna"), []byte("1"), []byte("923F"), []byte("67.0006, -70.5476"), []byte("1504054225"), []byte("Miriam"), []byte("110")})
  checkQuery(t, stub, "1", "Miriam")
  checkInvoke(t, stub, [][]byte{[]byte("changeTunaHolder"), []byte("1"), []byte("TuPT")})
  checkQuery(t, stub, "1", "TuPT")
}
