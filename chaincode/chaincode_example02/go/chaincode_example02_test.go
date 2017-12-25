package main

import (
  "fmt"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  su "github.com/user/stringutil"
  "testing"
)

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
  res := stub.MockInit("1", args)
  if res.Status != shim.OK {
    fmt.Println("Init failed", string(res.Message))
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

func checkQuery(t *testing.T, stub *shim.MockStub, name string, value string) {
  res := stub.MockInvoke("1", su.StringArgsToBytesArgs([]string{"query", name}))
  if res.Status != shim.OK {
    fmt.Println("Query", name, "failed", string(res.Message))
    t.FailNow()
  }

  if res.Payload == nil {
    fmt.Println("Query", name, "failed to get value")
    t.FailNow()
  }

  if string(res.Payload) != value {
    fmt.Println("Query value", name, "was not", value, "as expected")
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
  stub := shim.NewMockStub("ex02", scc)
  checkInit(t, stub, su.StringArgsToBytesArgs([]string{"init", "A", "123", "B", "234"}))

  checkState(t, stub, "A", "123")
  checkState(t, stub, "B", "234")
}

func Test_Query(t *testing.T) {
  scc := new(SimpleChaincode)
  stub := shim.NewMockStub("ex02", scc)
  checkInit(t, stub, su.StringArgsToBytesArgs([]string{"init", "A", "345", "B", "456"}))

  checkQuery(t, stub, "A", "345")
  checkQuery(t, stub, "B", "456")
}

func Test_Invoke(t *testing.T) {
  scc := new(SimpleChaincode)
  stub := shim.NewMockStub("ex02", scc)
  checkInit(t, stub, su.StringArgsToBytesArgs([]string{"init", "A", "567", "B", "678"}))

  checkInvoke(t, stub, []string{"invoke", "A", "B", "123"})
  checkQuery(t, stub, "A", "444")
  checkQuery(t, stub, "B", "801")

  checkInvoke(t, stub, []string{"invoke", "B", "A", "234"})
  checkQuery(t, stub, "A", "678")
  checkQuery(t, stub, "B", "567")
}
