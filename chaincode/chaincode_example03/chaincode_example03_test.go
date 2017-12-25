package main

import (
  "fmt"
  "testing"

  "github.com/hyperledger/fabric/core/chaincode/shim"

  su "github.com/user/stringutil"
)

func checkInit(t *testing.T, scc *SimpleChaincode, stub *shim.MockStub, args []string) {
  res := stub.MockInit("1", su.StringArgsToBytesArgs(args))
  if res.Status != shim.OK {
    fmt.Println("Init failed", res.Message)
    t.FailNow()
  }
}

func checkState(t *testing.T, stub *shim.MockStub, name string, value string) {
  bytes := stub.State[name]
  if bytes == nil {
    fmt.Println("State", name, "failed to get value")
    t.FailNow()
  }
  if string(bytes) != value {
    fmt.Println("State value", name, "was not", value, "as expected")
    t.FailNow()
  }
}

func checkInvoke(t *testing.T, scc *SimpleChaincode, stub *shim.MockStub, args []string) {
  res := stub.MockInvoke("1", su.StringArgsToBytesArgs(args))
  if res.Status != shim.OK {
    fmt.Println("Query failed", string(res.Message))
    t.FailNow()
  }
}

func TestExample03_Init(t *testing.T) {
  scc := new(SimpleChaincode)
  stub := shim.NewMockStub("ex03", scc)

  // Init A=123 B=234
  checkInit(t, scc, stub, []string{"init", "A", "123"})

  checkState(t, stub, "A", "123")
}

func TestExample03_Invoke(t *testing.T) {
  scc := new(SimpleChaincode)
  stub := shim.NewMockStub("ex03", scc)

  // Init A=345 B=456
  checkInit(t, scc, stub, []string{"init", "A", "345"})

  // Invoke "query"
  checkInvoke(t, scc, stub, []string{"query", "A", "345"})
}
