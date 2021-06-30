package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
)

//go:embed asset
var rootRawFS embed.FS

var assetFS fs.FS

func init() {
	var err error
	assetFS, err = fs.Sub(rootRawFS, "asset")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot initialize assetFS: %s\n", err)
		panic("/asset/ not found")
	}
}
