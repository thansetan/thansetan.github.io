package main

import "time"

type articleMeta struct {
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

type timeFrame struct {
	Start time.Time `yaml:"start"`
	End   time.Time `yaml:"end"`
}

type projectMeta struct {
	Timeframe                timeFrame
	Title, Description, Path string
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
