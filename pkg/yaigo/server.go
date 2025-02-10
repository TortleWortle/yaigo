package yaigo

import (
	"errors"
	"github.com/tortlewortle/go-inertia/internal/vite"
	"html/template"
	"io/fs"
	"net/http"
)

func NewServer(frontend fs.FS, optFns ...OptFunc) (*Server, error) {
	if frontend == nil {
		return nil, errors.New("frontend filesystem can not be nil")
	}

	// default opts
	opts := &ServerOpts{
		ViteUrl:          "",
		ViteScriptName:   "src/main.js",
		ViteTemplateName: "index.html",
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

	rootTmpl, err := generateRootTemplate(frontend, manifest, opts)
	if err != nil {
		return nil, err
	}

	server := &Server{
		manifestVersion: version,
		ssrHTTPClient: &http.Client{
			Timeout: opts.SSRTimeout,
		},
		ssrURL:       opts.SSRServerUrl,
		rootTemplate: rootTmpl,

		viteDevUrl:   opts.ViteUrl,
		reactRefresh: opts.ReactRefresh,
	}

	return server, nil
}

type Server struct {
	manifestVersion string

	rootTemplate  *template.Template
	ssrHTTPClient *http.Client
	ssrURL        string

	reactRefresh bool
	viteDevUrl   string
}

// These methods are on the Server struct just to keep the api nice and tidy

func (_ *Server) Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (_ *Server) Location(w http.ResponseWriter, url string) {
	w.Header().Set(headerLocation, url)
	w.WriteHeader(http.StatusConflict)
}
