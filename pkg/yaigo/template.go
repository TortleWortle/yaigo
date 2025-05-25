package yaigo

import (
	"fmt"
	"github.com/tortlewortle/yaigo/pkg/vite"
	"html/template"
	"net/url"
	"strings"
)

type rootTmplData struct {
	InertiaRoot template.HTML
	InertiaHead template.HTML
}

func generateRootTemplate(tfn func(*template.Template) (*template.Template, error), manifest *vite.Manifest, opts *ServerOpts) (*template.Template, error) {
	viteUrl, err := url.Parse(opts.ViteUrl)
	if err != nil {
		return nil, err
	}
	t := template.New("rootTemplate")

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
			// dev Config provides the css by itself
			if opts.ViteUrl != "" {
				return "", nil
			}
			var tb strings.Builder
			item, err := manifest.GetItem(scriptUrl)
			if err != nil {
				return "", err
			}
			for _, sheetUrl := range item.Css {
				tb.WriteString(fmt.Sprintf("<link rel=\"preload\" href=\"/%s\" as=\"style\"/>\n", sheetUrl))
			}
			tb.WriteString("\n")
			for _, sheetUrl := range item.Css {
				tb.WriteString(fmt.Sprintf("<link rel=\"stylesheet\" href=\"/%s\"/>\n", sheetUrl))
			}
			return template.HTML(tb.String()), nil
		},
	})

	return tfn(t)
}
