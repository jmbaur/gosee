package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmbaur/gosee/md"
	"github.com/radovskyb/watcher"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Poll file for changes with this period.
	filePeriod = 100 * time.Millisecond
)

var (
	bindIp   = flag.String("bind", "8080", "the port you would like to bind to")
	filename string
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	homeTempl = template.Must(template.New("").Parse(homeHTML))
)

func readFile() ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	p, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writer(ws *websocket.Conn) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()

	for {
		select {
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func watch(ws *websocket.Conn) {
	w := watcher.New()
	go func() {
		for {
			select {
			case <-w.Event:
				p, err := readFile()
				if err != nil {
					p = []byte(err.Error())
				}
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.TextMessage, md.Markdownify(p)); err != nil {
					return
				}
			case err := <-w.Error:
				log.Fatal(err)
			case <-w.Closed:
				return
			}
		}

	}()

	if err := w.Add(filename); err != nil {
		log.Fatal(err)
	}
	if err := w.Start(filePeriod); err != nil {
		log.Fatal(err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	go writer(ws)
	go watch(ws)
	reader(ws)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/"+filename {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	p, err := readFile()
	if err != nil {
		p = []byte(err.Error())
	}

	var v = struct {
		Host     string
		Filename string
		Data     template.HTML
	}{
		r.Host,
		filename,
		template.HTML(md.Markdownify(p)),
	}

	homeTempl.Execute(w, &v)
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("must specify a file")
	}
	filename = flag.Args()[0]
	if !strings.HasSuffix(filename, "md") {
		log.Fatal("must specify a markdown file")
	}
	path, err := filepath.Abs(filename)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("view at http://localhost:%s/%s\n", *bindIp, filename)

	http.Handle("/", http.FileServer(http.Dir(filepath.Dir(path))))
	http.HandleFunc("/"+filename, handler)
	http.HandleFunc("/ws", wsHandler)

	if err := http.ListenAndServe(":"+*bindIp, nil); err != nil {
		log.Fatal(err)
	}
}

const homeHTML = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/4.0.0/github-markdown.css"
    />
    <title>{{.Filename}}</title>
  </head>
  <body>
    <div id="root" class="markdown-body">{{.Data}}</div>
    <script type="text/javascript">
      (function() {
        var root = document.getElementById("root");
        var conn = new WebSocket("ws://{{.Host}}/ws");
        conn.onclose = function(evt) {
          root.innerHTML = "Connection closed";
          console.log("connection closed");
        };
        conn.onmessage = function(evt) {
          console.log("file updated");
          root.innerHTML = evt.data;
        };
      })();
    </script>
  </body>
</html>
`
