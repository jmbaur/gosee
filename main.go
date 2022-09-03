package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
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
)

var (
	//go:embed md.webp
	icon string
	//go:embed index.html
	indexHTML string
	fullpath  string
	basename  string
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	homeTempl = template.Must(template.New("").Parse(indexHTML))
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

	for range pingTicker.C {
		ws.SetWriteDeadline(time.Now().Add(writeWait))
		if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			return
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
	v := struct {
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

func openDefault() bool {
	b := false
	if runtime.GOOS == "linux" {
		b = false
		_, display_ok := os.LookupEnv("DISPLAY")
		_, wl_display_ok := os.LookupEnv("WAYLAND_DISPLAY")
		if display_ok || wl_display_ok {
			b = true
		}
	}
	return b
}

func logic() error {
	host := flag.String("host", "[::1]:8080", "IP and port to bind to")
	tryFile := flag.String("file", "README.md", "File to use")
	open := flag.Bool("open", openDefault(), "Open preview in the default browser")
	flag.Parse()

	var err error
	fullpath, err = filepath.Abs(*tryFile)
	if err != nil {
		return fmt.Errorf("failed to get path to file: %v", err)
	}

	info, err := os.Stat(fullpath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("file is directory")
	}
	if fullpath == "" {
		return errors.New("could not find markdown file to use")
	}

	basename = filepath.Base(fullpath)
	if !strings.HasSuffix(basename, "md") {
		return errors.New("must specify a markdown file")
	}

	http.Handle("/", http.FileServer(http.Dir(filepath.Dir(fullpath))))
	http.HandleFunc("/"+basename, handler)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/webp")
		fmt.Fprintln(w, icon)
	})

	url := fmt.Sprintf("http://%s/%s", *host, basename)

	if *open {
		openCmd, ok := openCmds[runtime.GOOS]
		if ok {
			go func() {
				_ = exec.Command(openCmd, url).Run()
			}()
		}
	}

	fmt.Println("View preview at", url)
	return http.ListenAndServe(*host, nil)
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
