package entrypoints

import "github.com/raphael-goetz/lazysound/internal/app"

func mustConfigPath() string {
	path, err := app.ConfigPath()
	if err != nil {
		return ""
	}
	return path
}
