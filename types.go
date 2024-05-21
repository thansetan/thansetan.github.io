package main

import "time"

type articleMeta struct {
	Title, Path string
	Date        time.Time
}

type pageMeta struct {
	Title, layout string
	Date          time.Time
}
type page[T any] struct {
	Meta    pageMeta
	Content T
}

type timeFrame struct {
	Start time.Time `yaml:"start"`
	End   time.Time `yaml:"end"`
}

type projectMeta struct {
	Title, Path string
	Timeframe   timeFrame
}
type tech struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type project struct {
	Title        string    `yaml:"title"`
	Timeframe    timeFrame `yaml:"timeframe"`
	Description  string    `yaml:"description"`
	GithubRepo   string    `yaml:"github_repo"`
	ProjectURL   string    `yaml:"project_url"`
	Technologies []tech    `yaml:"technologies"`
}
