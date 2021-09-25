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
	"os/exec"
	"path/filepath"
	"runtime"
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
	icon       = `iVBORw0KGgoAAAANSUhEUgAAAIAAAACACAMAAAD04JH5AAABdFBMVEUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAKUwSXAAAAe3RSTlMAAQIDBAUGBwkKDBETFRYXGhsfIiMnKSorLTAxMjM0Njc4P0VHTE1OUlNXWlxgY2RlZmprbW50d3yDhIWJiouNjo+QkpSWnZ+goqmrrK2ur7W2uLm9vsPIzM7Q0dLU1dba29/g4+Tm6Ors7e7v8PHy9PX29/j5+vv8/f6uVqFrAAAC9klEQVR4nOya91sTSxSGT3ZzrwEhUQSiBrCiWLD3hlIURLAhlgUsJFaUhKLJnn/eZxOTwJTdWXVmHvV8v2Xm7LyvySe7FKBQKBQKhUKhUCiUanJD3ic0mpX50d4Gvv2uWXg90+kav3vRDh+xkK3++9/Y4iMWMgBg6f2vZQog59sUwB4YssrHEfDsCsyB4f//bIpgl49IAiTwZwv4vu9XKpVKuVwu/zYB9lFhXHzdODsXdc5PC7QsiC5baDEmANkSf1Upy43pE4BB/qpBfkqjAF8DrgCaBdga8AXQLMDUQFAA3QKbayAogHaBjTUQFUC/QLMGwgLoF2jUQFwAAwL1GogLYEKgVgNJAYwIBDWQFcCIAGRL0gKYEYBBaQEMCYSGBEiABP5egRNi3kkR4Jia6/FYAqu9ojP61kQCnztU+DuWYwngYht/Rtui+C2edaP5yafxPgLERw675jyWfcaXogWuxuwAIl5m165IS/ZtTxR/fyW+gN+/eemQL2/569Zw/tY8y1MQwKXOjSudS7K5ILfDBe5xPBUBfJFqLqQ8+VyQI2H8UyxOUQAnmgsTYXOI+HG7nN/Ff6et+hXsTP312fA5RJxJyPj/P+f4ygLre2sv961HCuAFmcBNnq8sgPn24FW6EDUXyO4W8w+KfjCvfhOZcQHcJ9FziPgyJeKn3wn4MQTwGsB1lTlEHBMJPBDx4wjgwIDaHCL28/zzQn4sgS/LygLvM+z+rrVfF4gz95DZTs3LztEkgOc2b49Jz9ElsLpz4+5h2TH6BNDb0tzc9sGCAN5q7DkzUr5OAf9Afe+inK9TAN/+eJrs+6pFIJFIJBzHcVzXdZPJpGjmfvWE1lchfM2/MTkdCNwJHdErUOoCOBo+olcAn/3XsWRVAG/MRgzoFogMCcCKXX4RZPdpQ/Fg1K7AMPTaFcgBTNvkTwJApmCPn6/+MVfWmkG+u3bHzUzZ4U+mGzf9npG5oll40RvOsY8eFAqFQqFQKBQK5R/N9wAAAP//NnxaZ7cRZpgAAAAASUVORK5CYII=`
)

var (
	fullpath string
	basename string
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	homeTempl = template.Must(template.New("").Parse(homeHTML))
	tryFiles  = []string{"README.md", "readme.md", "CONTRIBUTING.md", "contributing.md", "NOTES.md", "notes.md"}
	openCmds  = map[string]string{"linux": "xdg-open", "darwin": "open"}
)

func markdownify(p []byte) ([]byte, error) {
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
		return nil, err
	}

	return buf.Bytes(), nil
}

func readFile() ([]byte, error) {
	file, err := os.Open(fullpath)
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
			p, err := readFile()
			if err != nil {
				p = []byte(err.Error())
			}
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			data, err := markdownify(p)
			if err != nil {
				log.Printf("failed to compile markdown: %+v\n", err)
			}
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("failed to write to websocket: %+v\n", err)
			}
		case err := <-w.Error:
			log.Fatalln(err)
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

	if err := watcher.Add(fullpath); err != nil {
		log.Fatal(err)
	}
	if err := watcher.Start(filePeriod); err != nil {
		log.Fatal(err)
	}

	go writer(ws)
	reader(ws)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/"+basename {
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

	var data []byte
	data, err = markdownify(p)
	if err != nil {
		data = []byte(err.Error())
	}
	var v = struct {
		Host     string
		Basename string
		Data     template.HTML
	}{
		r.Host,
		basename,
		template.HTML(data),
	}

	homeTempl.Execute(w, &v)
}

func main() {
	host := flag.String("host", ":8080", "IP and port to bind to")
	open := flag.Bool("open", true, "Open preview in the default browser")
	flag.Parse()

	for _, file := range append([]string{flag.Arg(0)}, tryFiles...) {
		var err error

		fullpath, err = filepath.Abs(file)
		if err != nil {
			log.Fatalf("Failed to get path to file: %v\n", err)
		}

		if info, err := os.Stat(fullpath); err != nil || info.IsDir() {
			continue
		} else {
			break
		}
	}

	if fullpath == "" {
		log.Fatal("could not find markdown file to use")
	}

	basename = filepath.Base(fullpath)
	if !strings.HasSuffix(basename, "md") {
		log.Fatalln("Must specify a markdown file")
	}

	http.Handle("/", http.FileServer(http.Dir(filepath.Dir(fullpath))))
	http.HandleFunc("/"+basename, handler)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(icon))
		w.Header().Set("Content-Type", "image/png")
		io.Copy(w, dec)
	})

	split := strings.Split(*host, ":")
	ip := split[0]
	if ip == "" {
		*host = "0.0.0.0" + *host
	}

	url := fmt.Sprintf("http://%s/%s", *host, basename)

	if *open {
		openCmd, ok := openCmds[runtime.GOOS]
		if !ok {
			log.Println("Could not find command to open preview with")
		}
		cmd := exec.Command(openCmd, url)
		if err := cmd.Run(); err != nil {
			log.Println(err)
		}
	}

	log.Println("View preview at", url)
	log.Fatal(http.ListenAndServe(*host, nil))
}

const homeHTML = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<link
			rel="stylesheet"
			href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/4.0.0/github-markdown.css"
		/>
		<title>{{.Basename}}</title>
	</head>
	<style>
		#outer {
			display: flex;
			justify-content: center;
		}
	</style>
	<body>
		<div id="outer" class="markdown-body"><div id="inner">{{.Data}}</div></div>
		<script type="text/javascript">
			(function () {
				var inner = document.getElementById("inner");
				var conn = new WebSocket("ws://{{.Host}}/ws");
				conn.onclose = function (evt) {
					inner.innerHTML = "Connection closed";
					console.log("connection closed");
				};
				conn.onmessage = function (evt) {
					console.log("file updated");
					inner.innerHTML = evt.data;
				};
			})();
		</script>
	</body>
</html>
`
