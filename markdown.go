package main

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	attributes "github.com/mdigger/goldmark-attributes"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmHtml "github.com/yuin/goldmark/renderer/html"
)

type pageMeta struct {
	title  string
	layout string
	date   time.Time
}
type Page struct {
	meta    pageMeta
	content string
}

var (
	md goldmark.Markdown
)

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
			extension.Strikethrough,
			attributes.Extension,
		),
		goldmark.WithRendererOptions(
			gmHtml.WithUnsafe(),
			gmHtml.WithHardWraps(),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
}

func replaceMarkdownURL(markdown []byte) []byte {
	re := regexp.MustCompile(`\[(.*?)\]\((.*?)\.md(.*?)\)`)
	return re.ReplaceAll(markdown, []byte(`[$1]($2.html$3)`))
}

func toPageData(inputPath string, isArticle bool) (Page, error) {
	var (
		data Page
		buf  bytes.Buffer
	)

	file, err := os.Open(inputPath)
	if err != nil {
		return data, err
	}
	defer file.Close()

	mdBytes, err := io.ReadAll(file)
	if err != nil {
		return data, err
	}
	mdBytes = replaceMarkdownURL(mdBytes)
	ctx := parser.NewContext()

	err = md.Convert(mdBytes, &buf, parser.WithContext(ctx))
	if err != nil {
		return data, err
	}

	metaData := meta.Get(ctx)
	data.content = buf.String()
	if v, ok := metaData["title"].(string); ok {
		data.meta.title = v
	}
	if v, ok := metaData["layout"].(string); ok {
		data.meta.layout = v
	} else {
		if isArticle {
			data.meta.layout = "article"
		} else {
			data.meta.layout = "page"
		}
	}
	if v, ok := metaData["date"].(string); ok {
		data.meta.date, err = time.Parse("2006-01-02", v)
		if err != nil {
			data.meta.date = time.Unix(0, 0)
		}
	} else {
		data.meta.date = time.Unix(0, 0)
	}

	return data, nil
}
