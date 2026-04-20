//go:build !tinygo && !(js && wasm) && !testbuild

package main

import "embed"

//go:embed .ti-config
var tiConfigFS embed.FS
