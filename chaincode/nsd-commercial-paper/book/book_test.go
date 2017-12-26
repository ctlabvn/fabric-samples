package main

import (
  "fmt"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  "testing"
  // "github.com/olegabu/nsd-commercial-paper/chaincode/go/security"
  su "github.com/user/stringutil"
)

func checkInit(t *testing.T, stub *shim.MockStub, args []string) {
  res := stub.MockInit("1", su.StringArgsToBytesArgs(args))
  if res.Status != shim.OK {
    fmt.Println("Init failed", string(res.Message))
    t.FailNow()
  }
}

func checkState(t *testing.T, stub *shim.MockStub, expectedStatus int32, args []string) {
  bytes := stub.MockInvoke("1", su.StringArgsToBytesArgs(args))
  if bytes.Status != expectedStatus {
    fmt.Println("Wrong status. Current value: ", bytes.Status, ", Expected value: ", expectedStatus, ".")
    t.FailNow()
  }
}

func TestBook_Init(t *testing.T) {
  scc := new(BookChaincode)
  stub := shim.NewMockStub("bookChaincode", scc)

  checkInit(t, stub, []string{"init", "[{\"account\":\"AC0689654902\",\"division\":\"87680000045800005\",\"security\":\"RU000ABC0001\",\"quantity\":\"100\"},{\"account\":\"AC0689654902\",\"division\":\"87680000045800005\",\"security\":\"RU000ABC0002\",\"quantity\":\"42\"}]"})

  //Correct transaction
  checkState(t, stub, 200, []string{"check", "AC0689654902", "87680000045800005", "RU000ABC0001", "90"})
  //Wrong number of arguments
  checkState(t, stub, 400, []string{"check", "AC0689654902"})
  // Record not found
  checkState(t, stub, 404, []string{"check", "AAA", "BBB", "CCC", "200"})
  // Quantity less than current balance
  checkState(t, stub, 409, []string{"check", "AC0689654902", "87680000045800005", "RU000ABC0001", "200"})
}

//TODO: uncomment when package for security changed to  "security"
// func TestRedeem(t *testing.T) {
//   sccSecurity := new(security.SecurityChaincode)
//   stubSecurity := shim.NewMockStub("security", sccSecurity)
//   stubSecurity.MockInit("1", []string{"init", "RU000ABC0001", "active", "AAA689654902", "87680000045800005"})

//   fmt.Println(stubSecurity.State)

//   sccBook := new(BookChaincode)
//   stub := shim.NewMockStub("book", sccBook)

//   stub.MockPeerChaincode("security/common", stubSecurity)
//   checkInit(t, stub, []string{"init", "[{\"account\":\"BBB689654902\",\"division\":\"87680000045800005\",\"security\":\"RU000ABC0001\",\"quantity\":\"100\"}]"})

//   stub.MockInvoke("1", []string{"redeem", "RU000ABC0001", "some message"})

//   //AAA should have at least 100
//   checkState(t, stub, 200, []string{"check", "AAA689654902", "87680000045800005", "RU000ABC0001", "90"})
//   //BBB should have nothing on it's balance
//   checkState(t, stub, 409, []string{"check", "BBB689654902", "87680000045800005", "RU000ABC0001", "90"})
//   //Second redeem is impossible
//   checkState(t, stub, 400, []string{"redeem", "RU000ABC0001"})
// }
