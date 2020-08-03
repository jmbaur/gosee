package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/radovskyb/watcher"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	filePeriod = 100 * time.Millisecond

	icon = `iVBORw0KGgoAAAANSUhEUgAAAIAAAACACAMAAAD04JH5AAABdFBMVEUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAKUwSXAAAAe3RSTlMAAQIDBAUGBwkKDBETFRYXGhsfIiMnKSorLTAxMjM0Njc4P0VHTE1OUlNXWlxgY2RlZmprbW50d3yDhIWJiouNjo+QkpSWnZ+goqmrrK2ur7W2uLm9vsPIzM7Q0dLU1dba29/g4+Tm6Ors7e7v8PHy9PX29/j5+vv8/f6uVqFrAAAC9klEQVR4nOya91sTSxSGT3ZzrwEhUQSiBrCiWLD3hlIURLAhlgUsJFaUhKLJnn/eZxOTwJTdWXVmHvV8v2Xm7LyvySe7FKBQKBQKhUKhUCiUanJD3ic0mpX50d4Gvv2uWXg90+kav3vRDh+xkK3++9/Y4iMWMgBg6f2vZQog59sUwB4YssrHEfDsCsyB4f//bIpgl49IAiTwZwv4vu9XKpVKuVwu/zYB9lFhXHzdODsXdc5PC7QsiC5baDEmANkSf1Upy43pE4BB/qpBfkqjAF8DrgCaBdga8AXQLMDUQFAA3QKbayAogHaBjTUQFUC/QLMGwgLoF2jUQFwAAwL1GogLYEKgVgNJAYwIBDWQFcCIAGRL0gKYEYBBaQEMCYSGBEiABP5egRNi3kkR4Jia6/FYAqu9ojP61kQCnztU+DuWYwngYht/Rtui+C2edaP5yafxPgLERw675jyWfcaXogWuxuwAIl5m165IS/ZtTxR/fyW+gN+/eemQL2/569Zw/tY8y1MQwKXOjSudS7K5ILfDBe5xPBUBfJFqLqQ8+VyQI2H8UyxOUQAnmgsTYXOI+HG7nN/Ff6et+hXsTP312fA5RJxJyPj/P+f4ygLre2sv961HCuAFmcBNnq8sgPn24FW6EDUXyO4W8w+KfjCvfhOZcQHcJ9FziPgyJeKn3wn4MQTwGsB1lTlEHBMJPBDx4wjgwIDaHCL28/zzQn4sgS/LygLvM+z+rrVfF4gz95DZTs3LztEkgOc2b49Jz9ElsLpz4+5h2TH6BNDb0tzc9sGCAN5q7DkzUr5OAf9Afe+inK9TAN/+eJrs+6pFIJFIJBzHcVzXdZPJpGjmfvWE1lchfM2/MTkdCNwJHdErUOoCOBo+olcAn/3XsWRVAG/MRgzoFogMCcCKXX4RZPdpQ/Fg1K7AMPTaFcgBTNvkTwJApmCPn6/+MVfWmkG+u3bHzUzZ4U+mGzf9npG5oll40RvOsY8eFAqFQqFQKBQK5R/N9wAAAP//NnxaZ7cRZpgAAAAASUVORK5CYII=`
)

var (
	bindIp   = flag.String("bind", "8080", "the port you would like to bind to")
	filename string
	filetype string
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	mdTempl = template.Must(template.New("").Parse(mdHTML))
)

func markdownify(p []byte) []byte {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
			),
			extension.GFM,
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe()),
	)
	var buf bytes.Buffer
	if err := markdown.Convert(p, &buf); err != nil {
		log.Fatalln(err)
	}

	return buf.Bytes()
}

func readFile(loc string) ([]byte, error) {
	file, err := os.Open(loc)
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

func watch(w *watcher.Watcher, ws *websocket.Conn) {
	for {
		select {
		case <-w.Event:
			p, err := readFile(filename)
			if err != nil {
				p = []byte(err.Error())
			}
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, markdownify(p)); err != nil {
				return
			}
		case err := <-w.Error:
			log.Fatal(err)
		case <-w.Closed:
			return
		}
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

	watcher := watcher.New()
	go watch(watcher, ws)

	if err := watcher.Add(filename); err != nil {
		log.Fatal(err)
	}
	if err := watcher.Start(filePeriod); err != nil {
		log.Fatal(err)
	}

	go writer(ws)
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

	switch filetype {
	case "markdown":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		p, err := readFile(filename)
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
			template.HTML(markdownify(p)),
		}

		mdTempl.Execute(w, &v)
	case "pdf":
		data, err := readFile(filename)
		if err != nil {
			io.Copy(w, strings.NewReader(err.Error()))
			return
		}
		src := base64.StdEncoding.EncodeToString(data)
		w.Header().Set("Content-Type", "application/pdf")
		io.Copy(w, strings.NewReader(src))
	}
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("must specify a file")
	}
	filename = flag.Args()[0]

	switch strings.Split(filename, ".")[1] {
	case "md":
		filetype = "markdown"
	case "pdf":
		filetype = "pdf"
	default:
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
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(icon))
		w.Header().Set("Content-Type", "image/png")
		io.Copy(w, dec)
	})

	if err := http.ListenAndServe(":"+*bindIp, nil); err != nil {
		log.Fatal(err)
	}
}

const mdHTML = `
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

const pdfHTML = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/pdf.js/2.5.207/pdf_viewer.min.css"
    />
    <title>{{.Filename}}</title>
  </head>
  <body>
    <canvas id="root" style="border:1px solid black;">{{.Data}}</canvas>
    <script
      type="text/javascript"
      src="https://cdnjs.cloudflare.com/ajax/libs/pdf.js/2.5.207/pdf.min.js"
    ></script>
    <script type="text/javascript">
      (function() {
	var scale = 1.5;
	var viewport = page.getViewport(scale);
	var context = canvas.getContext('2d');
	var canvas = document.getElementById("root");
	canvas.height = viewport.height;
	canvas.width = viewport.width;
	var renderContext = {
	  canvasContext: context,
	  viewport: viewport
	};
	page.render(renderContext);

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
