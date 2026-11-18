package effectiveschoolboards

import (
	"embed"
	"io/fs"
)

//go:embed embedded/*
var assets embed.FS

var FS, _ = fs.Sub(assets, "embedded")
