package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

func parseYAML[T any](path string) (T, error) {
	var project T
	in, err := os.ReadFile(path)
	if err != nil {
		return project, err
	}
	err = yaml.Unmarshal(in, &project)
	if err != nil {
		return project, err
	}
	return project, nil
}
