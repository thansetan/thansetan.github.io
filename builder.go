package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var (
	inputDir     = "src"
	outputDir    = "docs"
	postsDir     = filepath.Join(inputDir, "posts")
	templatesDir = "templates"
)

func init() {
	// remove already generated files
	os.RemoveAll(outputDir)
}

func main() {
	err := buildWebsite()
	if err != nil {
		panic(err)
	}
}

func buildWebsite() error {
	var posts []post

	tmpl, err := parseTemplates(templatesDir)
	if err != nil {
		return err
	}

	err = filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			err := os.MkdirAll(strings.Replace(path, inputDir, outputDir, 1), os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			if filepath.Ext(path) == ".md" {
				if path == filepath.Join(postsDir, "index.md") {
					return nil
				}

				page, err := toPageData(path)
				if err != nil {
					return err
				}

				if page.meta.layout == "post" {
					postPath := filepath.Join(filepath.Base(filepath.Dir(path)))
					posts = append(posts, post{
						Title: page.meta.title,
						Path:  postPath,
						Date:  page.meta.modifiedAt,
					})
				}

				outputPath := strings.TrimSuffix(strings.Replace(path, inputDir, outputDir, 1), ".md")
				outputPath = fmt.Sprintf("%s.html", outputPath)

				err = toHTML(tmpl, page, outputPath, nil)
				if err != nil {
					return err
				}
			} else {
				src, err := os.Open(path)
				if err != nil {
					return err
				}
				defer src.Close()

				dst, err := os.Create(strings.Replace(path, inputDir, outputDir, 1))
				if err != nil {
					return err
				}
				defer dst.Close()

				_, err = io.Copy(dst, src)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// begin creating posts page

	// sort by file modification date, ascending
	slices.SortFunc(posts, func(a, b post) int {
		if a.Date.Unix() > b.Date.Unix() {
			return 1
		} else if a.Date.Unix() == b.Date.Unix() {
			return 0
		}
		return -1
	})

	pageData, err := toPageData(filepath.Join(postsDir, "index.md"))
	if err != nil {
		return err
	}

	err = toHTML(tmpl, pageData, filepath.Join(outputDir, "posts", "index.html"), posts)
	if err != nil {
		return err
	}

	return nil
}
