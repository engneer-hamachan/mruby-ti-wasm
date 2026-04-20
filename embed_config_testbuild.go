//go:build !tinygo && !(js && wasm) && testbuild

package main

import (
	"embed"
	"io/fs"
)

//go:embed test/.ti-config
var rawConfigFS embed.FS

var tiConfigFS, _ = fs.Sub(rawConfigFS, "test")
