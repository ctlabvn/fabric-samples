import React from "react";
import logo from "./logo.svg";

export default () => (
  <header>
    <img src={logo} className="App-logo" alt="logo" />
    <div id="left_header">Hyperledger Fabric Tuna Application</div>
    <i id="right_header">
      Example Blockchain Application for Introduction to Hyperledger Fabric
      LFS171x
    </i>
  </header>
);
