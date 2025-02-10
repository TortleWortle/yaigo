package yaigo

import (
	"fmt"
	"github.com/tortlewortle/go-inertia/internal/vite"
	"html/template"
	"io/fs"
	"net/url"
	"strings"
)

type rootTmplData struct {
	InertiaRoot template.HTML
	InertiaHead template.HTML
}

func generateRootTemplate(frontend fs.FS, manifest *vite.Manifest, opts *ServerOpts) (*template.Template, error) {
	t := template.New(opts.ViteTemplateName)
	viteUrl, err := url.Parse(opts.ViteUrl)

	if err != nil {
		return nil, err
	}

	t = t.Funcs(template.FuncMap{
		"vite": func(assetUrl string) (string, error) {
			item, err := manifest.GetItem(assetUrl)
			if err != nil {
				return "", err
			}
			if opts.ViteUrl != "" {
				return viteUrl.JoinPath(assetUrl).String(), nil
			}

			return "/" + item.File, nil
		},
		"viteCSS": func(scriptUrl string) (template.HTML, error) {
			// dev Server provides the css by itself
			if opts.ViteUrl != "" {
				return "", nil
			}
			var tb strings.Builder
			item, err := manifest.GetItem(scriptUrl)
			if err != nil {
				return "", err
			}
			for _, url := range item.Css {
				tb.WriteString(fmt.Sprintf("<link rel=\"stylesheet\" href=\"/%s\">\n", url))
			}
			return template.HTML(tb.String()), nil
		},
	})

	t, err = t.ParseFS(frontend, opts.ViteTemplateName)
	if err != nil {
		return nil, err
	}

	return t, nil
}
