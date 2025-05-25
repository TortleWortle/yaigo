package yaigo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tortlewortle/yaigo/internal/page"
	"github.com/tortlewortle/yaigo/internal/props"
	"golang.org/x/net/html"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func NewPage(component string, props Props) *Page {
	return &Page{
		component: component,
		pageProps: props,
	}
}

type Page struct {
	component    string
	pageProps    Props
	clearHistory bool
}

func (p *Page) ClearHistory() {
	p.clearHistory = true
}

func (p *Page) Render(ctx context.Context, w io.Writer) error {
	config := ctx.Value(configKey).(*Config)
	requestInfo := ctx.Value(requestInfoKey).(*RequestInfo)
	var bag *props.Bag
	if bagVal := ctx.Value(bagKey); bagVal != nil {
		bag = bagVal.(*props.Bag)
	} else {
		bag = props.NewBag()
	}
	pageData := ctx.Value(pageDataKey).(*page.InertiaPage)
	pageData.Component = p.component
	pageData.ClearHistory = p.clearHistory

	bag.Checkpoint()

	for k, v := range p.pageProps {
		bag.Set(k, v)
	}

	if requestInfo.IsPartial(p.component) {
		bag.LoadDeferred()
		onlyProps := requestInfo.OnlyProps()
		if len(onlyProps) > 0 {
			bag.Only(onlyProps)
		}

		exceptProps := requestInfo.ExceptProps()
		if len(exceptProps) > 0 {
			bag.Except(exceptProps)
		}
	}

	var err error
	pageData.Props, err = bag.GetProps(ctx)
	if err != nil {
		return fmt.Errorf("loading props: %w", err)
	}
	pageData.DeferredProps = bag.GetDeferredProps()

	// todo: maybe move away?
	if config.typeGen != nil {
		start := time.Now()
		err := config.typeGen.Generate(pageData)
		if err != nil {
			config.logger.Warn("typegen failed", slog.String("component", p.component), slog.Any("error", err))
		}
		config.logger.Info("generated types", slog.String("component", p.component), slog.Duration("duration", time.Since(start)))
	}

	if requestInfo.IsInertia() {
		return p.renderJson(w, pageData)
	}

	if config.ssrURL != "" {
		err = p.renderSSR(config, w, pageData)
		if err != nil {
			if errors.Is(err, errCommunicatingWithSSRServer) {
				// render client side if ssr is unreachable
				return p.renderHtml(config, w, pageData)
			}
			return err
		}
		return nil
	}
	return p.renderHtml(config, w, pageData)
}

func (p *Page) renderJson(w io.Writer, data *page.InertiaPage) error {
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set(HeaderInertia, "true")
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Vary", HeaderInertia)
	}
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return err
	}
	return nil
}

func (p *Page) renderHtml(config *Config, w io.Writer, data *page.InertiaPage) error {
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "text/html")
	}
	propStr, err := json.Marshal(data)
	if err != nil {
		return err
	}

	inertiaRoot := template.HTML(fmt.Sprintf("<div id=\"app\" data-page='%s'></div>", html.EscapeString(string(propStr))))
	return config.rootTemplate.Execute(w, rootTmplData{
		InertiaRoot: inertiaRoot,
		InertiaHead: p.inertiaBaseHead(config),
	})
}

type ssrResponse struct {
	Head []string `json:"head"`
	Body string   `json:"body"`
}

var errCommunicatingWithSSRServer = errors.New("could not communicate with ssr Config")

func (p *Page) renderSSR(config *Config, w io.Writer, data *page.InertiaPage) error {
	renderPath, err := url.JoinPath(config.ssrURL, "/render")
	if err != nil {
		return err
	}
	pData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	ssrReq, err := http.NewRequest("GET", renderPath, bytes.NewReader(pData))
	if err != nil {
		return errors.Join(errCommunicatingWithSSRServer, err)
	}

	resp, err := config.ssrHTTPClient.Do(ssrReq)
	if err != nil {
		return errors.Join(errCommunicatingWithSSRServer, err)
	}
	defer resp.Body.Close()

	var ssrRes ssrResponse
	err = json.NewDecoder(resp.Body).Decode(&ssrRes)
	if err != nil {
		return errors.Join(errCommunicatingWithSSRServer, err)
	}

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "text/html")
	}

	baseHead := p.inertiaBaseHead(config)
	return config.rootTemplate.Execute(w, rootTmplData{
		InertiaRoot: template.HTML(ssrRes.Body),
		InertiaHead: baseHead + "\n" + template.HTML(strings.Join(ssrRes.Head, "\n")), // this is for SSR later
	})
}

func (p *Page) inertiaBaseHead(config *Config) template.HTML {
	if config.reactRefresh {
		return p.reactRefreshScript(config, nil)
	}
	return ""
}

func (p *Page) reactRefreshScript(config *Config, attrs []template.HTMLAttr) template.HTML {
	var attributes string
	if attrs != nil {
		var attrBuilder strings.Builder
		for _, a := range attrs {
			attrBuilder.WriteString(string(a))
			attrBuilder.WriteString(" ")
		}
		attributes = attrBuilder.String()
	}
	return template.HTML(fmt.Sprintf(`<script type="module" %s>
	import RefreshRuntime from '%s/@react-refresh'
	RefreshRuntime.injectIntoGlobalHook(window)
	window.$RefreshReg$ = () => {}
	window.$RefreshSig$ = () => (type) => type
	window.__vite_plugin_react_preamble_installed__ = true
</script>`, attributes, config.viteDevUrl))
}
