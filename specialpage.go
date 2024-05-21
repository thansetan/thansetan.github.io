package main

import (
	"html/template"
	"slices"
)

func generateSpecialPage[T any](tmpl *template.Template, in, out string, data []T, cmpFunc func(T, T) int) error {
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
