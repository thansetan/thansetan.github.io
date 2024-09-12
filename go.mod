module ghpages

go 1.23.1

require (
	github.com/mangoumbrella/goldmark-figure v1.2.0
	github.com/mdigger/goldmark-attributes v0.0.0-20210529130523-52da21a6bf2b
	github.com/yuin/goldmark v1.7.4
	github.com/yuin/goldmark-meta v1.1.0
	gopkg.in/yaml.v3 v3.0.1
)

require gopkg.in/yaml.v2 v2.3.0 // indirect

replace github.com/mangoumbrella/goldmark-figure => github.com/thansetan/goldmark-figure v0.0.0-20240910091154-75de0c5e7033
