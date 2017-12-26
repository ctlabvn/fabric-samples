// base url from fabric explorer
const API_BASE = `${window.location.protocol}://${window.location
  .hostname}:8080`;
const DEFAULT_CHANNEL = "mychannel";
const DEFAULT_PEER = "peer1";
const DEFAULT_CHAINCODE = "tuna-app";

export const rejectErrors = res => {
  const { status } = res;
  if (status >= 200 && status < 300) {
    return res;
  }

  return Promise.reject({ message: res.statusText, status });
};

export const fetchJson = (url, options = {}, base = API_BASE) =>
  fetch(/^(?:https?)?:\/\//.test(url) ? url : base + url, options)
    .then(rejectErrors)
    // default return empty json when no content
    .then(res => res.json())
    .then(json => {
      let data = null;
      if (options.method === "POST") {
        // we have response and transaction id from event
        // return transaction
        data = json[1];
      } else {
        // if return is Buffer then parse it
        data =
          json[0].type === "Buffer"
            ? JSON.parse(String.fromCharCode.apply(null, json[0].data))
            : json[0];
      }
      return data;
    });

export const query = (
  fcn,
  args,
  channel = DEFAULT_CHANNEL,
  chaincodeName = DEFAULT_CHAINCODE,
  peer = DEFAULT_PEER
) =>
  fetchJson(`/apis/channels/${channel}/chaincodes/${chaincodeName}?peer=${peer}&fcn=${fcn}&args=
  ${JSON.stringify(args)}`);

export const invoke = (
  fcn,
  args,
  channel = DEFAULT_CHANNEL,
  chaincodeName = DEFAULT_CHAINCODE,
  peer = DEFAULT_PEER
) =>
  fetchJson(`/apis/channels/${channel}/chaincodes/${chaincodeName}`, {
    method: "POST",
    body: JSON.stringify({
      peers: [peer],
      fcn,
      args
    }),
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json"
    }
  });
