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
	if err := os.RemoveAll(outputDir); err != nil {
		fmt.Printf("ERROR REMOVING EXISTING DIRECTORY: %s\n", err.Error())
	}
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
			err = os.MkdirAll(strings.Replace(path, inputDir, outputDir, 1), os.ModePerm)
		} else {
			if filepath.Ext(d.Name()) == ".md" {
				// ignore /posts path, will be treated later
				if path == filepath.Join(postsDir, "index.md") {
					return nil
				}

				isPost := strings.HasPrefix(path, postsDir)

				page, err := toPageData(path, isPost)
				if err != nil {
					return err
				}

				if isPost {
					postPath := filepath.Base(filepath.Dir(path))
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

	// begin creating posts page

	// sort by file modification date, descending
	slices.SortFunc(posts, func(a, b post) int {
		if a.Date.Unix() > b.Date.Unix() {
			return -1
		} else if a.Date.Unix() == b.Date.Unix() {
			return 0
		}
		return 1
	})

	pageData, err := toPageData(filepath.Join(postsDir, "index.md"), false)
	if err != nil {
		return err
	}

	err = toHTML(tmpl, pageData, filepath.Join(outputDir, "posts", "index.html"), posts)
	if err != nil {
		return err
	}

	return nil
}
