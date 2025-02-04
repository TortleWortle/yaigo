package inertia

import (
	"errors"
	"go.tortle.tech/go-inertia/pkg/vite"
	"html/template"
	"io/fs"
	"net/url"
	"sync"
)

type OptFunc = func(o *ServerOpts)

type Server struct {
	requestPool     *sync.Pool
	rootTemplate    *template.Template
	manifestVersion string
}

func NewServer(frontend fs.FS, optFns ...OptFunc) (*Server, error) {
	if frontend == nil {
		return nil, errors.New("frontend filesystem can not be nil")
	}

	// default opts
	opts := &ServerOpts{
		viteUrl:      "",
		viteScript:   "src/main.js",
		viteTemplate: "index.html",
	}

	for _, fn := range optFns {
		fn(opts)
	}

	manifest, err := loadManifest(frontend)
	if err != nil {
		return nil, err
	}

	version, err := manifest.Version()
	if err != nil {
		return nil, err
	}

	rootTmpl, err := generateRootTemplate(frontend, manifest, opts)
	if err != nil {
		return nil, err
	}

	server := &Server{
		manifestVersion: version,
		rootTemplate:    rootTmpl,
		requestPool: &sync.Pool{
			New: func() any {
				return newRequest()
			},
		},
	}

	return server, nil
}

func loadManifest(frontend fs.FS) (manifest *vite.Manifest, err error) {
	f, err := frontend.Open(".vite/manifest.json")
	if err != nil {
		return nil, err
	}
	defer func() {
		cErr := f.Close()
		if cErr != nil && err == nil {
			err = cErr
		}
	}()

	return vite.FromJSON(f)
}

func generateRootTemplate(frontend fs.FS, manifest *vite.Manifest, opts *ServerOpts) (*template.Template, error) {
	t := template.New(opts.viteTemplate)
	viteUrl, err := url.Parse(opts.viteUrl)

	if err != nil {
		return nil, err
	}

	t = t.Funcs(template.FuncMap{
		"vite": func(assetUrl string) (string, error) {
			item, err := manifest.GetItem(assetUrl)
			if err != nil {
				return "", err
			}
			if opts.viteUrl != "" {
				return viteUrl.JoinPath(assetUrl).String(), nil
			}

			return item.File, nil
		},
	})

	t, err = t.ParseFS(frontend, opts.viteTemplate)
	if err != nil {
		return nil, err
	}

	return t, nil
}
