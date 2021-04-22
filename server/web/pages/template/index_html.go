package template

import (
	_ "embed"
)

//go:embed pages/index.html
var IndexHtml []byte
