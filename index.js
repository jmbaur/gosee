import "./index.css";
import "github-markdown-css";
const inner = document.getElementById("inner");
const conn = new WebSocket(`ws://${location.host}/ws`);
conn.onclose = (evt) => alert("connection closed");
conn.onmessage = (evt) => inner.innerHTML = evt.data;
