package yaigo

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"sync"

	"github.com/tortlewortle/yaigo/internal/vite"
)

func NewServer(tfn func(*template.Template) (*template.Template, error), frontend fs.FS, optFns ...OptFunc) (*Server, error) {
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

	server := &Server{
		typeGen:         nil,
		manifestVersion: version,
		ssrHTTPClient: &http.Client{
			Timeout:   opts.SSRTimeout,
			Transport: ssrTransport,
		},
		ssrURL:       opts.SSRServerUrl,
		rootTemplate: rootTmpl,

		viteDevUrl:   opts.ViteUrl,
		reactRefresh: opts.ReactRefresh,
	}

	if opts.TypeGenOutput != "" {
		err := os.MkdirAll(opts.TypeGenOutput, 0700)
		if err != nil {
			return nil, fmt.Errorf("creating typegen output folder: %w", err)
		}

		server.typeGen = &typeGenerator{
			dirPath:        opts.TypeGenOutput,
			lock:           &sync.Mutex{},
			propCache:      make(map[string]Props),
			optionalsCache: make(map[string][]string),
		}
	}

	return server, nil
}

type Server struct {
	manifestVersion string

	rootTemplate *template.Template

	ssrHTTPClient *http.Client
	ssrURL        string

	reactRefresh bool
	viteDevUrl   string
	typeGen      *typeGenerator
}

// These methods are on the Server struct just to keep the api nice and tidy

func (*Server) Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (*Server) Location(w http.ResponseWriter, url string) {
	w.Header().Set(headerLocation, url)
	w.WriteHeader(http.StatusConflict)
}

func (s *Server) IsDevMode() bool {
	return s.viteDevUrl != ""
}
