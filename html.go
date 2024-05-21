package main

import (
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
)

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

func toHTML[T any](tmpl *template.Template, layout, out string, data T) error {
	outFile, err := os.Create(out)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = tmpl.ExecuteTemplate(outFile, layout, data)
	if err != nil {
		return err
	}

	return nil
}
