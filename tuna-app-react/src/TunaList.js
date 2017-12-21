import React from "react";

export default ({ tunas }) => (
  <table className="table" align="center">
    <thead>
      <tr>
        <th>ID</th>
        <th>Timestamp</th>
        <th>Holder</th>
        <th>
          Catch Location <br />(Longitude, Latitude)
        </th>
        <th>Vessel</th>
        <th>Weight</th>
      </tr>
    </thead>
    <tbody>
      {tunas.map(({ Key, Record: tuna }) => (
        <tr key={Key}>
          <td>{Key}</td>
          <td>{tuna.timestamp}</td>
          <td>{tuna.holder}</td>
          <td>{tuna.location}</td>
          <td>{tuna.vessel}</td>
          <td>{tuna.weight}</td>
        </tr>
      ))}
    </tbody>
  </table>
);
