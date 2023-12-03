package main

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmHtml "github.com/yuin/goldmark/renderer/html"
)

type pageMeta struct {
	title      string
	layout     string
	modifiedAt time.Time
}
type Page struct {
	meta    pageMeta
	content string
}

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokailight"),
				highlighting.WithFormatOptions(
					html.WithLineNumbers(true),
				),
			),
			extension.Table,
			extension.Footnote,
		),
		goldmark.WithRendererOptions(
			gmHtml.WithUnsafe(),
			gmHtml.WithHardWraps(),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)
}

func toPageData(inputPath string) (Page, error) {
	var (
		data Page
		buf  bytes.Buffer
	)

	file, err := os.Open(inputPath)
	if err != nil {
		return data, err
	}
	defer file.Close()

	fi, _ := os.Stat(inputPath)
	mdBytes, _ := io.ReadAll(file)
	ctx := parser.NewContext()

	err = md.Convert(mdBytes, &buf, parser.WithContext(ctx))
	if err != nil {
		return data, err
	}

	metaData := meta.Get(ctx)
	data.content = buf.String()
	data.meta.modifiedAt = fi.ModTime()
	if metaData["Title"] != nil {
		data.meta.title = metaData["Title"].(string)
	} else {
		data.meta.title = "A Post"
	}
	if metaData["Layout"] != nil {
		data.meta.layout = metaData["Layout"].(string)
	} else {
		data.meta.layout = "post"
	}

	return data, nil
}
