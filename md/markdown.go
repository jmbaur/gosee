package md

import (
	"bytes"
	"log"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

func Markdownify(p []byte) []byte {
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
