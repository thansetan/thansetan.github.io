package main

import (
	"time"
)

type postMeta struct {
	Date        time.Time
	Title, Path string
}

type pageMeta struct {
	Date          time.Time
	Title, layout string
}
type page[T any] struct {
	Content T
	Meta    pageMeta
}

type project struct {
	Name   string `yaml:"name"`
	Banner struct {
		URL     string `yaml:"url"`
		Alt     string `yaml:"alt"`
		Caption string `yaml:"caption"`
	} `yaml:"banner"`
	Description string `yaml:"description"`
	URL         struct {
		GitHub  string `yaml:"github"`
		Project string `yaml:"project"`
	} `yaml:"url"`
}

type projectsData struct {
	Projects []project `yaml:"projects"`
}
