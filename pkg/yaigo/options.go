package yaigo

import "time"

type ServerOpts struct {
	ViteUrl          string
	ViteScriptName   string
	ViteTemplateName string
	SSRServerUrl     string
	ReactRefresh     bool
	SSRTimeout       time.Duration
}

type OptFunc = func(o *ServerOpts)

// WithRootTemplateName sets the root template filename to look for inside the frontend filesystem
func WithRootTemplateName(template string) OptFunc {
	return func(o *ServerOpts) {
		o.ViteTemplateName = template
	}
}

// WithViteScriptName sets the script name to look for inside the vite manifest
func WithViteScriptName(script string) OptFunc {
	return func(o *ServerOpts) {
		o.ViteScriptName = script
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
