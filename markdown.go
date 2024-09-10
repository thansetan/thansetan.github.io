package main

import (
	"bytes"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	figure "github.com/mangoumbrella/goldmark-figure"
	attributes "github.com/mdigger/goldmark-attributes"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			highlighting.NewHighlighting(
				highlighting.WithStyle("onedark"),
				highlighting.WithFormatOptions(
					html.WithLineNumbers(true),
				),
			),
			extension.Table,
			extension.Footnote,
			extension.Strikethrough,
			attributes.Extension,
			figure.Figure,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    &dotMdLinkTransformer{},
				Priority: 1000,
			}),
		),
	)
}

type dotMdLinkTransformer struct{}

func (*dotMdLinkTransformer) Transform(node *ast.Document, reader text.Reader, ctx parser.Context) {
	_ = ast.Walk(node, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if link, ok := node.(*ast.Link); ok &&
			strings.Contains(string(link.Destination), ".md") {
			url, err := url.Parse(string(link.Destination))
			if err != nil {
				return ast.WalkContinue, err
			}
			if url.Scheme == "" && filepath.Ext(url.Path) == ".md" { // only replace relative links and .md files
				url.Path = url.Path[:len(url.Path)-2] + "html"
				link.Destination = []byte(url.String())
			}
		}
		return ast.WalkContinue, nil
	})
}

func toPageData(inputPath string, isPost bool) (pageMeta, string, error) {
	var (
		pageMeta pageMeta
		buf      bytes.Buffer
	)

	file, err := os.Open(inputPath)
	if err != nil {
		return pageMeta, buf.String(), err
	}
	defer file.Close()

	mdBytes, err := io.ReadAll(file)
	if err != nil {
		return pageMeta, buf.String(), err
	}

	ctx := parser.NewContext()

	err = md.Convert(mdBytes, &buf, parser.WithContext(ctx))
	if err != nil {
		return pageMeta, buf.String(), err
	}

	metaData := meta.Get(ctx)
	if v, ok := metaData["title"].(string); ok {
		pageMeta.Title = v
	}
	if v, ok := metaData["layout"].(string); ok {
		pageMeta.layout = v
	} else {
		if isPost {
			pageMeta.layout = "post"
		} else {
			pageMeta.layout = "page"
		}
	}
	if v, ok := metaData["date"].(string); ok {
		pageMeta.Date, err = time.Parse("2006-01-02", v)
		if err != nil {
			pageMeta.Date = time.Unix(0, 0)
		}
	} else {
		pageMeta.Date = time.Unix(0, 0)
	}

	return pageMeta, buf.String(), nil
}
