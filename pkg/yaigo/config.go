package yaigo

import (
	"errors"
	"fmt"
	"github.com/tortlewortle/yaigo/pkg/vite"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
)

func New(tfn func(*template.Template) (*template.Template, error), frontend fs.FS, optFns ...OptFunc) (*Config, error) {
	if tfn == nil {
		return nil, errors.New("template can not be nil")
	}
	if frontend == nil {
		return nil, errors.New("frontend filesystem can not be nil")
	}

	// default opts
	opts := &ServerOpts{
		ViteUrl: "",
	}

	for _, fn := range optFns {
		fn(opts)
	}

	manifest, err := vite.FromDistFS(frontend)
	if err != nil {
		return nil, err
	}

	version, err := manifest.Version()
	if err != nil {
		return nil, err
	}

	rootTmpl, err := generateRootTemplate(tfn, manifest, opts)
	if err != nil {
		return nil, err
	}

	ssrTransport := http.DefaultTransport.(*http.Transport).Clone()
	ssrTransport.MaxIdleConns = 100
	ssrTransport.MaxConnsPerHost = 100
	ssrTransport.MaxIdleConnsPerHost = 100

	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	server := &Config{
		typeGenerator:   nil,
		manifestVersion: version,
		ssrHTTPClient: &http.Client{
			Timeout:   opts.SSRTimeout,
			Transport: ssrTransport,
		},
		ssrURL:       opts.SSRServerUrl,
		rootTemplate: rootTmpl,

		viteDevUrl:   opts.ViteUrl,
		reactRefresh: opts.ReactRefresh,
		logger:       opts.Logger,
	}

	if opts.TypeGen != nil {
		err := os.MkdirAll(opts.TypeGen.dirPath, 0700)
		if err != nil {
			return nil, fmt.Errorf("creating typegen output folder: %w", err)
		}

		server.typeGenerator = opts.TypeGen
	}

	return server, nil
}

type Config struct {
	manifestVersion string

	rootTemplate *template.Template

	ssrHTTPClient *http.Client
	ssrURL        string

	reactRefresh  bool
	viteDevUrl    string
	typeGenerator *TypeGenerator
	logger        *slog.Logger
}

func (s *Config) IsDevMode() bool {
	return s.viteDevUrl != ""
}
