package yaigo

import (
	"time"
)

type ServerOpts struct {
	ViteUrl      string
	SSRServerUrl string
	ReactRefresh bool
	SSRTimeout   time.Duration
}

type OptFunc = func(o *ServerOpts)

// WithViteDevServer loads the script from the url instead of the filesystem, this is for hot-reloading
func WithViteDevServer(url string, reactRefresh bool) OptFunc {
	return func(o *ServerOpts) {
		o.ViteUrl = url
		o.ReactRefresh = reactRefresh
	}
}

// WithSSR enables Server-side rendering using the provided ssr Server url and bundle bundlePath
func WithSSR(url string, timeout time.Duration) OptFunc {
	return func(o *ServerOpts) {
		o.SSRServerUrl = url
		o.SSRTimeout = timeout
	}
}
