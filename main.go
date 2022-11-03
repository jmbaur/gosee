package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
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
	"github.com/yuin/goldmark/renderer/html"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	filePeriod = 100 * time.Millisecond
)

var (
	//go:embed index.html.tmpl
	precompiledTmpl string
	tmpl            = template.Must(template.New("").Parse(precompiledTmpl))
	upgrader        = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	openCmds = map[string]string{"linux": "xdg-open", "darwin": "open"}
)

func markdownify(p []byte) ([]byte, error) {
	markdown := goldmark.New(
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

func readFile(fullpath string) ([]byte, error) {
	file, err := os.Open(fullpath)
	if err != nil {
		return nil, err
	}
	p, err := io.ReadAll(file)
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

func watch(fullpath string) func(w *watcher.Watcher, ws *websocket.Conn) {
	return func(w *watcher.Watcher, ws *websocket.Conn) {
		for {
			select {
			case <-w.Event:
				p, err := readFile(fullpath)
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
}

func wsHandler(fullpath string) func(w http.ResponseWriter, r *http.Request) {
	watchFunc := watch(fullpath)
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				log.Println(err)
			}
			return
		}

		mdWatcher := watcher.New()
		go watchFunc(mdWatcher, ws)

		if err := mdWatcher.Add(fullpath); err != nil {
			log.Fatal(err)
		}
		if err := mdWatcher.Start(filePeriod); err != nil {
			log.Fatal(err)
		}

		go writer(ws)
		reader(ws)
	}
}

func handler(fullpath string, basename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+basename {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		p, err := readFile(fullpath)
		if err != nil {
			p = []byte(err.Error())
		}

		var data []byte
		data, err = markdownify(p)
		if err != nil {
			data = []byte(err.Error())
		}
		v := struct {
			Basename string
			Data     template.HTML
		}{
			basename,
			template.HTML(data),
		}

		tmpl.Execute(w, &v)
	}
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
	addr := flag.String("addr", "[::1]:1234", "Address to bind on")
	open := flag.Bool("open", openDefault(), "Open preview in the default browser")
	flag.Parse()

	tryFile := flag.Arg(0)
	if tryFile == "" {
		tryFile = "README.md"
	}

	fullpath, err := filepath.Abs(tryFile)
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

	basename := filepath.Base(fullpath)
	if !strings.HasSuffix(basename, "md") {
		return errors.New("must specify a markdown file")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/"+basename, handler(fullpath, basename))
	mux.HandleFunc("/ws", wsHandler(fullpath))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	url := fmt.Sprintf("http://%s/%s", *addr, basename)

	if openCmd, ok := openCmds[runtime.GOOS]; *open && ok {
		go func() {
			_ = exec.Command(openCmd, url).Run()
		}()
	}

	fmt.Printf("View preview at %s\n", url)
	return http.ListenAndServe(*addr, mux)
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
