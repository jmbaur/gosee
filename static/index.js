var inner = document.getElementById("inner");
var conn = new WebSocket(`ws://${location.host}/ws`);
conn.onclose = function (evt) {
  alert("connection closed");
};
conn.onmessage = function (evt) {
  inner.innerHTML = evt.data;
};
