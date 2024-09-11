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
	inputDir     = "src"
	outputDir    = "docs"
	postsDir     = filepath.Join(inputDir, "posts")
	projectsDir  = filepath.Join(inputDir, "projects")
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
	var (
		posts    []postMeta
		projects []projectMeta
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
				if path == filepath.Join(postsDir, "index.md") || path == filepath.Join(projectsDir, "index.md") {
					return nil
				}

				isPost := strings.HasPrefix(path, postsDir)

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
			} else if filepath.Ext(d.Name()) == ".yaml" && strings.HasPrefix(path, projectsDir) {
				projectData, err := parseYAML[project](path)
				if err != nil {
					return err
				}
				projects = append(projects, projectMeta{
					Title:       projectData.Title,
					Path:        filepath.Base(filepath.Dir(path)),
					Timeframe:   projectData.Timeframe,
					Description: projectData.Description,
				})
				outputPath := filepath.Join(filepath.Dir(strings.Replace(path, inputDir, outputDir, 1)), "index.html")
				err = toHTML(tmpl, "project", outputPath, page[project]{
					Meta:    pageMeta{Title: projectData.Title},
					Content: projectData,
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

	// create /posts page
	err = toListPage(tmpl, filepath.Join(postsDir, "index.md"), filepath.Join(outputDir, "posts", "index.html"), posts, func(a, b postMeta) int {
		return cmp.Compare(b.Date.Unix(), a.Date.Unix())
	})
	if err != nil {
		return err
	}

	// create /projects page
	err = toListPage(tmpl, filepath.Join(projectsDir, "index.md"), filepath.Join(outputDir, "projects", "index.html"), projects, func(a, b projectMeta) int {
		if a.Timeframe.Start.Unix() == b.Timeframe.Start.Unix() {
			return cmp.Compare(b.Timeframe.End.Unix(), a.Timeframe.End.Unix())
		}
		return cmp.Compare(b.Timeframe.Start.Unix(), a.Timeframe.Start.Unix())
	})
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
