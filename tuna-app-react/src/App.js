import React, { Component } from "react";
import TunaList from "./TunaList";
import Header from "./Header";

import { invoke, query } from "./api";
import "./App.css";

class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      tunas: [],
      errorMsg: null,
      txtId: null
    };
  }

  queryAllTuna = () => {
    [
      "getCreator",
      // "getBinding",
      "getSignedProposal"
      // "getTransient"
    ].forEach(cert =>
      query(cert, [], "mychannel", "shim-api").then(ret => console.log(ret))
    );

    query("queryAllTuna", []).then(tunas => this.setState({ tunas }));
  };

  changeHolder = () => {
    const key = document.querySelector("#holderId").value.trim();
    const holder = document.querySelector("#holderName").value.trim();
    invoke("changeTunaHolder", [key, holder]).then(data => {
      if (data && data.tx_id) {
        this.setState({ errorMsg: null, txtId: data.tx_id });
        this.queryAllTuna();
      } else {
        this.setState({
          errorMsg: "Error: Please enter a valid Tuna Id",
          txtId: null
        });
      }
    });
  };

  render() {
    const { tunas, txtId, errorMsg } = this.state;
    return (
      <div className="App">
        <Header />
        <div id="body">
          <button className="btn btn-primary mb-2" onClick={this.queryAllTuna}>
            Query All Tuna
          </button>

          <TunaList tunas={tunas} />

          <label>Change Tuna Holder</label>
          <br />
          {txtId && (
            <h5 className="text-success" id="success_holder">
              Success! Tx ID: change_holder {txtId}
            </h5>
          )}
          {errorMsg && (
            <h5 className="text-danger" id="error_holder">
              {errorMsg}
            </h5>
          )}
          <div className="row no-gutter">
            <div className="col">
              Enter a catch id between 1 and 10:{" "}
              <input
                className="form-control m-0"
                placeholder="Ex: 1"
                id="holderId"
              />
            </div>
            <div className="col">
              Enter name of new holder:{" "}
              <input
                className="form-control m-0"
                placeholder="Ex: Barry"
                id="holderName"
              />
            </div>
          </div>

          <button onClick={this.changeHolder} className="mt-2 btn btn-primary">
            Change
          </button>
        </div>
      </div>
    );
  }
}

export default App;
