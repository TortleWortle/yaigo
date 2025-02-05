package web

import (
	"embed"
	"io/fs"
)

//go:generate npm run build

//go:embed dist/*
var dist embed.FS

func FrontendFS() (fs.FS, error) {
	frontend, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, err
	}
	return frontend, nil
}
