package yaigo

import (
	"html/template"
	"time"
)

type ServerOpts struct {
	ViteUrl          string
	ViteTemplateName string
	SSRServerUrl     string
	ReactRefresh     bool
	SSRTimeout       time.Duration
	TemplateOptions  []func(t *template.Template)
}

type OptFunc = func(o *ServerOpts)

// WithRootTemplateName sets the root template filename to look for inside the frontend filesystem,
// defaults to index.html
func WithRootTemplateName(template string) OptFunc {
	return func(o *ServerOpts) {
		o.ViteTemplateName = template
	}
}

// WithTemplateOptions accepts functions to modify the template before parsing
func WithTemplateOptions(opts ...func(t *template.Template)) OptFunc {
	return func(o *ServerOpts) {
		o.TemplateOptions = opts
	}
}

// WithViteDevServer loads the script from the url instead of the filesystem, this is for hot-reloading
func WithViteDevServer(url string, react bool) OptFunc {
	return func(o *ServerOpts) {
		o.ViteUrl = url
		o.ReactRefresh = react
	}
}

// WithSSR enables Server-side rendering using the provided ssr Server url and bundle bundlePath
func WithSSR(url string, timeout time.Duration) OptFunc {
	return func(o *ServerOpts) {
		o.SSRServerUrl = url
		o.SSRTimeout = timeout
	}
}
