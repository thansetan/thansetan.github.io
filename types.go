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
