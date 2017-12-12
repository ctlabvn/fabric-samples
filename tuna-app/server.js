//SPDX-License-Identifier: Apache-2.0

// nodejs server setup

// call the packages we need
import express from "express"; // call express

import bodyParser from "body-parser";
import path from "path";

const app = express(); // define our app using express
// Load all of our middleware
// configure app to use bodyParser()
// this will let us get the data from a POST
// app.use(express.static(__dirname + '/client'));
app.use(bodyParser.urlencoded({ extended: true }));
app.use(bodyParser.json());

// this line requires and runs the code from our routes.js file and passes it app
app.use("/", require("./routes").default);

// set up a static file server that points to the "client" directory
app.use(express.static(path.join(__dirname, "./client")));

// Save our port
const port = process.env.PORT || 8000;

// Start the server and listen on port
app.listen(port, () => {
  console.log("Live on port: " + port);
});
