package main

import (
	"os"

	"github.com/raphael-goetz/lazysound/internal/entrypoints"
)

func main() {
	os.Exit(entrypoints.RunCtl(os.Args[1:], os.Stdout, os.Stderr))
}
