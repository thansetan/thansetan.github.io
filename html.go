package main

import (
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
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
	tmpl := template.New("")

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(d.Name()) == ".html" {
			_, err = tmpl.ParseFiles(path)
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
		Date:    page.meta.date,
	})
	if err != nil {
		return err
	}

	return nil
}
