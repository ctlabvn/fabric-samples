package main

import (
  "fmt"
  "testing"

  "github.com/hyperledger/fabric/core/chaincode/shim"
  ex02 "github.com/hyperledger/fabric/examples/chaincode/go/chaincode_example02"
  su "github.com/user/stringutil"
)

var chaincodeName = "ex02"

// chaincode_example05 looks like it wanted to return a JSON response to Query()
// it doesn't actually do this though, it just returns the sum value
func jsonResponse(name string, value string) string {
  return fmt.Sprintf("jsonResponse = \"{\"Name\":\"%v\",\"Value\":\"%v\"}", name, value)
}

func checkInit(t *testing.T, stub *shim.MockStub, args []string) {
  res := stub.MockInit("1", su.StringArgsToBytesArgs(args))
  if res.Status != shim.OK {
    fmt.Println("Init failed", string(res.Message))
    t.FailNow()
  }
}

func checkState(t *testing.T, stub *shim.MockStub, name string, expect string) {
  bytes := stub.State[name]
  if bytes == nil {
    fmt.Println("State", name, "failed to get value")
    t.FailNow()
  }
  if string(bytes) != expect {
    fmt.Println("State value", name, "was not", expect, "as expected")
    t.FailNow()
  }
}

func checkQuery(t *testing.T, stub *shim.MockStub, args []string, expect string) {
  res := stub.MockInvoke("1", su.StringArgsToBytesArgs(args))
  if res.Status != shim.OK {
    fmt.Println("Query", args, "failed", string(res.Message))
    t.FailNow()
  }
  if res.Payload == nil {
    fmt.Println("Query", args, "failed to get result")
    t.FailNow()
  }
  if string(res.Payload) != expect {
    fmt.Println("Query result ", string(res.Payload), "was not", expect, "as expected")
    t.FailNow()
  }
}

func checkInvoke(t *testing.T, stub *shim.MockStub, args []string) {
  res := stub.MockInvoke("1", su.StringArgsToBytesArgs(args))
  if res.Status != shim.OK {
    fmt.Println("Invoke", args, "failed", string(res.Message))
    t.FailNow()
  }
}

func Test_Init(t *testing.T) {
  scc := new(SimpleChaincode)
  stub := shim.NewMockStub("ex05", scc)

  // Init A=123 B=234
  checkInit(t, stub, []string{"init", "sumStoreName", "432"})

  checkState(t, stub, "sumStoreName", "432")
}

func Test_Query(t *testing.T) {
  scc := new(SimpleChaincode)
  stub := shim.NewMockStub("ex05", scc)

  ccEx2 := new(ex02.SimpleChaincode)
  stubEx2 := shim.NewMockStub(chaincodeName, ccEx2)
  checkInit(t, stubEx2, []string{"init", "a", "111", "b", "222"})
  stub.MockPeerChaincode(chaincodeName, stubEx2)

  checkInit(t, stub, []string{"init", "sumStoreName", "0"})

  // a + b = 111 + 222 = 333
  checkQuery(t, stub, []string{"query", chaincodeName, "sumStoreName", ""}, "333") // example05 doesn't return JSON?
}

func TestExample05_Invoke(t *testing.T) {
  scc := new(SimpleChaincode)
  stub := shim.NewMockStub("ex05", scc)

  ccEx2 := new(ex02.SimpleChaincode)
  stubEx2 := shim.NewMockStub(chaincodeName, ccEx2)
  checkInit(t, stubEx2, []string{"init", "a", "222", "b", "333"})
  stub.MockPeerChaincode(chaincodeName, stubEx2)

  checkInit(t, stub, []string{"init", "sumStoreName", "0"})

  // a + b = 222 + 333 = 555
  checkInvoke(t, stub, []string{"invoke", chaincodeName, "sumStoreName", ""})
  checkQuery(t, stub, []string{"query", chaincodeName, "sumStoreName", ""}, "555") // example05 doesn't return JSON?
  checkQuery(t, stubEx2, []string{"query", "a"}, "222")
  checkQuery(t, stubEx2, []string{"query", "b"}, "333")

  // update A-=10 and B+=10
  checkInvoke(t, stubEx2, []string{"invoke", "a", "b", "10"})

  // a + b = 212 + 343 = 555
  checkInvoke(t, stub, []string{"invoke", chaincodeName, "sumStoreName", ""})
  checkQuery(t, stub, []string{"query", chaincodeName, "sumStoreName", ""}, "555") // example05 doesn't return JSON?
  checkQuery(t, stubEx2, []string{"query", "a"}, "212")
  checkQuery(t, stubEx2, []string{"query", "b"}, "343")
}
