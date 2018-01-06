const x509 = require("x509");
var cert = process.argv[2];
// var issuer = x509.getIssuer(cert);

console.log({
  // issuer,
  // subject: x509.getSubject(cert),
  cert: x509.parseCert(cert)
});
