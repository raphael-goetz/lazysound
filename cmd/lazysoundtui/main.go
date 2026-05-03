package main

import (
	"os"

	"github.com/raphael-goetz/lazysound/internal/entrypoints"
)

func main() {
	os.Exit(entrypoints.RunTUI(os.Stderr))
}
