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
	inputDir            = "src"
	outputDir           = "docs"
	postsDir            = "posts"
	projectsDir         = "projects"
	otherSidesDir       = "other-sides"
	projectsFileName    = "projects.yaml"
	otherSidesFileName  = "other-sides.yaml"
	inputPostsDir       = filepath.Join(inputDir, postsDir)
	inputProjectsDir    = filepath.Join(inputDir, projectsDir)
	inputOtherSidesDir  = filepath.Join(inputDir, otherSidesDir)
	outputPostsDir      = filepath.Join(outputDir, postsDir)
	outputProjectsDir   = filepath.Join(outputDir, projectsDir)
	outputOtherSidesDir = filepath.Join(outputDir, otherSidesDir)
	templatesDir        = "templates"
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
		others   otherSidesData
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
				// ignore /posts, /projects, and /other-sides path, will be treated later
				if path == filepath.Join(inputPostsDir, "index.md") || path == filepath.Join(inputProjectsDir, "index.md") || path == filepath.Join(inputOtherSidesDir, "index.md") {
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
			} else if path == filepath.Join(inputOtherSidesDir, otherSidesFileName) {
				others, err = parseYAML[otherSidesData](path)
				if err != nil {
					return err
				}
			} else {
				dstPath := strings.Replace(path, inputDir, outputDir, 1)
				if err := copyFile(path, dstPath); err != nil {
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

	// create /other-sides page
	err = toListPage(tmpl, filepath.Join(inputOtherSidesDir, "index.md"), filepath.Join(outputOtherSidesDir, "index.html"), others.Sites, func(t1, t2 otherSite) int { return 0 })
	if err != nil {
		return err
	}

	return nil
}

func copyFile(srcPath, dstPath string) (err error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := src.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dst.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(dst, src)
	return err
}

func toListPage[T any](tmpl *template.Template, in, out string, data []T, cmpFunc func(T, T) int) error {
	slices.SortStableFunc(data, cmpFunc)

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
