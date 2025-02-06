package inertia

type ServerOpts struct {
	viteUrl      string
	viteScript   string
	viteTemplate string
	ssrServerUrl string
}

func WithRootTemplate(template string) OptFunc {
	return func(o *ServerOpts) {
		o.viteTemplate = template
	}
}

func WithScript(script string) OptFunc {
	return func(o *ServerOpts) {
		o.viteScript = script
	}
}

// WithViteDevServer loads the script from the url instead of the filesystem, this is for hot-reloading
func WithViteDevServer(url string) OptFunc {
	return func(o *ServerOpts) {
		o.viteUrl = url
	}
}

// WithSSR enables server-side rendering using the provided ssr server url and bundle bundlePath

func WithSSR(url string) OptFunc {
	return func(o *ServerOpts) {
		o.ssrServerUrl = url
	}
}
