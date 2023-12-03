package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type post struct {
	Title, Path string
	Date        time.Time
}

type htmlPage struct {
	Content template.HTML
	Posts   []post
	Title   string
	Date    time.Time
}

func parseTemplates(dir string) (*template.Template, error) {
	tmpl := template.New("").Funcs(template.FuncMap{
		"assign": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid call")
			}

			dict := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				k, ok := values[i].(string)
				if !ok {
					return nil, errors.New("key must be string")
				}
				dict[k] = values[i+1]
			}
			return dict, nil
		},
	})

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			_, err = tmpl.ParseFiles(path)
			if err != nil {
				fmt.Println(err)
			}
		}

		return err
	})
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func toHTML(tmpl *template.Template, page Page, out string, posts []post) error {
	file, err := os.Create(out)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.ExecuteTemplate(file, page.meta.layout, htmlPage{
		Title:   page.meta.title,
		Content: template.HTML(page.content),
		Posts:   posts,
		Date:    page.meta.modifiedAt,
	})
	if err != nil {
		return err
	}

	return nil
}
