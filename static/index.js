var inner = document.getElementById("inner");
var conn = new WebSocket(`ws://${location.host}/ws`);
conn.onclose = function (evt) {
  inner.innerHTML = "Connection closed";
  console.log("connection closed");
};
conn.onmessage = function (evt) {
  console.log("file updated");
  inner.innerHTML = evt.data;
};
