package main

import (
	"cmp"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	inputDir     = "src"
	outputDir    = "docs"
	articlesDir  = filepath.Join(inputDir, "articles")
	templatesDir = "templates"
)

func init() {
	// remove already generated files
	if err := os.RemoveAll(outputDir); err != nil {
		fmt.Printf("ERROR REMOVING EXISTING DIRECTORY: %s\n", err.Error())
	}
}

func main() {
	t0 := time.Now()
	err := buildWebsite()
	if err != nil {
		panic(err)
	}
	fmt.Printf("completed in %v\n", time.Since(t0))
}

func buildWebsite() error {
	var articles []articleMeta

	tmpl, err := parseTemplates(templatesDir)
	if err != nil {
		return err
	}

	err = filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			err = os.MkdirAll(strings.Replace(path, inputDir, outputDir, 1), os.ModePerm)
		} else {
			if filepath.Ext(d.Name()) == ".md" {
				// ignore /articles path, will be treated later
				if path == filepath.Join(articlesDir, "index.md") {
					return nil
				}

				isArticle := strings.HasPrefix(path, articlesDir)

				pageMeta, content, err := toPageData(path, isArticle)
				if err != nil {
					return err
				}

				if isArticle {
					articlePath := filepath.Base(filepath.Dir(path))
					articles = append(articles, articleMeta{
						Title: pageMeta.Title,
						Path:  articlePath,
						Date:  pageMeta.Date,
					})
				}

				outputPath := strings.TrimSuffix(strings.Replace(path, inputDir, outputDir, 1), ".md")
				outputPath = fmt.Sprintf("%s.html", outputPath)

				err = toHTML(tmpl, pageMeta.layout, outputPath, page[template.HTML]{
					Meta:    pageMeta,
					Content: template.HTML(content),
				})
				if err != nil {
					return err
				}
			} else {
				src, err := os.Open(path)
				if err != nil {
					return err
				}
				defer func() {
					if err := src.Close(); err != nil {
						fmt.Printf("ERROR CLOSING SOURCE FILE: %s\n", err.Error())
					}
				}()

				dst, err := os.Create(strings.Replace(path, inputDir, outputDir, 1))
				if err != nil {
					return err
				}
				defer func() {
					if err := dst.Close(); err != nil {
						fmt.Printf("ERROR CLOSING DESTINATION FILE: %s\n", err.Error())
					}
				}()

				_, err = io.Copy(dst, src)
				if err != nil {
					return err
				}
			}
		}
		return err
	})

	if err != nil {
		return err
	}

	// creating /articles page
	err = generateSpecialPage(tmpl, filepath.Join(articlesDir, "index.md"), filepath.Join(outputDir, "articles", "index.html"), articles, func(a, b articleMeta) int {
		return cmp.Compare(b.Date.Unix(), a.Date.Unix())
	})
	if err != nil {
		return err
	}

	return nil
}
