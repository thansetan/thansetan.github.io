package main

import (
	"cmp"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var (
	inputDir          = "src"
	outputDir         = "docs"
	postsDir          = "posts"
	projectsDir       = "projects"
	projectsFileName  = "projects.yaml"
	inputPostsDir     = filepath.Join(inputDir, postsDir)
	inputProjectsDir  = filepath.Join(inputDir, projectsDir)
	outputPostsDir    = filepath.Join(outputDir, postsDir)
	outputProjectsDir = filepath.Join(outputDir, projectsDir)
	templatesDir      = "templates"
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
	var (
		posts    []postMeta
		projects projectsData
	)

	tmpl, err := parseTemplates(templatesDir)
	if err != nil {
		return err
	}

	err = filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			err = os.MkdirAll(strings.Replace(path, inputDir, outputDir, 1), os.ModePerm)
		} else {
			if filepath.Ext(d.Name()) == ".md" {
				// ignore /posts and /projects path, will be treated later
				if path == filepath.Join(inputPostsDir, "index.md") || path == filepath.Join(inputProjectsDir, "index.md") {
					return nil
				}

				isPost := strings.HasPrefix(path, inputPostsDir)

				pageMeta, content, err := toPageData(path, isPost)
				if err != nil {
					return err
				}

				if isPost {
					filePath := filepath.Base(filepath.Dir(path))
					posts = append(posts, postMeta{
						Title: pageMeta.Title,
						Path:  filePath,
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
			} else if path == filepath.Join(inputProjectsDir, projectsFileName) {
				projects, err = parseYAML[projectsData](path)
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

	// create /posts page
	err = toListPage(tmpl, filepath.Join(inputPostsDir, "index.md"), filepath.Join(outputPostsDir, "index.html"), posts, func(a, b postMeta) int {
		return cmp.Compare(b.Date.Unix(), a.Date.Unix())
	})
	if err != nil {
		return err
	}

	// create /projects page
	err = toListPage(tmpl, filepath.Join(inputProjectsDir, "index.md"), filepath.Join(outputProjectsDir, "index.html"), projects.Projects, func(t1, t2 project) int { return 0 })
	if err != nil {
		return err
	}

	return nil
}

func toListPage[T any](tmpl *template.Template, in, out string, data []T, cmpFunc func(T, T) int) error {
	slices.SortFunc(data, cmpFunc)

	pageMeta, _, err := toPageData(in, false)
	if err != nil {
		return err
	}

	err = toHTML(tmpl, pageMeta.layout, out, page[[]T]{
		Meta:    pageMeta,
		Content: data,
	})
	if err != nil {
		return err
	}

	return nil
}
