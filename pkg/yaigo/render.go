package yaigo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/tortlewortle/yaigo/internal/inertia"
	"github.com/tortlewortle/yaigo/internal/page"
)

const (
	headerInertia          = "X-Inertia"
	headerErrorBag         = "X-Inertia-Error-Bag"
	headerLocation         = "X-Inertia-Location"
	headerVersion          = "X-Inertia-Version"
	headerPartialComponent = "X-Inertia-Partial-Component"
	headerPartialOnly      = "X-Inertia-Partial-Data"
	headerPartialExcept    = "X-Inertia-Partial-Except"
)

type Props map[string]any

func (s *Server) Render(res *Response, w http.ResponseWriter, r *http.Request, page string, pageProps Props) error {
	ir := inertia.FromHTTPRequest(r)
	partialComponent := r.Header.Get(headerPartialComponent)
	isPartial := partialComponent == page

	// detect frontend version changes
	if ir.IsInertia() && r.Method == http.MethodGet && r.Header.Get(headerVersion) != s.manifestVersion {
		w.Header().Set(headerLocation, r.URL.String())
		w.WriteHeader(http.StatusConflict)
		return nil
	}
	var err error
	data := res.pageData
	bag := res.propBag

	// Remove any dirty props from a previous render attempt in the same request
	bag.Checkpoint()

	for k, v := range pageProps {
		err = bag.Set(k, v)
		if err != nil {
			return fmt.Errorf("transferring props: %w", err)
		}
	}

	data.Component = page
	data.Url = r.URL.Path
	data.Version = s.manifestVersion

	if isPartial {
		data.Props, err = res.filterPartialProps(r.Context(), ir)
		if err != nil {
			return fmt.Errorf("loading filtered props: %w", err)
		}
		data.DeferredProps = nil
	} else {
		data.Props, err = bag.GetProps(r.Context())
		if err != nil {
			return fmt.Errorf("loading props: %w", err)
		}
		data.DeferredProps = bag.GetDeferredProps()
	}

	if ir.IsInertia() {
		return res.renderJson(w, data)
	}

	if s.ssrURL != "" {
		err = res.renderSSR(s, w, data)
		if err != nil {
			if errors.Is(err, errCommunicatingWithSSRServer) {
				// render client side if ssr is unreachable
				return res.renderHtml(s, w, data)
			}
			return err
		}
		return nil
	}
	return res.renderHtml(s, w, data)
}

func (req *Response) filterPartialProps(ctx context.Context, r inertia.InertiaRequest) (map[string]any, error) {
	bag := req.propBag
	onlyProps := r.OnlyProps()
	if len(onlyProps) > 0 {
		bag.Only(onlyProps)
	}

	exceptProps := r.ExceptProps()
	if len(exceptProps) > 0 {
		bag.Except(exceptProps)
	}

	return bag.GetProps(ctx)
}

func (req *Response) renderJson(w http.ResponseWriter, data *page.InertiaPage) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Vary", headerInertia)
	w.Header().Set(headerInertia, "true")
	w.WriteHeader(req.status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) inertiaBaseHead() template.HTML {
	if s.reactRefresh {
		return s.reactRefreshScript(nil)
	}
	return ""
}

func (req *Response) renderHtml(s *Server, w http.ResponseWriter, data *page.InertiaPage) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(req.status)
	propStr, err := json.Marshal(data)
	if err != nil {
		return err
	}
	inertiaRoot := template.HTML(fmt.Sprintf("<div id=\"app\" data-page='%s'></div>", propStr))
	return s.rootTemplate.Execute(w, rootTmplData{
		InertiaRoot: inertiaRoot,
		InertiaHead: s.inertiaBaseHead(),
	})
}

var errCommunicatingWithSSRServer = errors.New("could not communicate with ssr Server")

type ssrResponse struct {
	Head []string `json:"head"`
	Body string   `json:"body"`
}

func (req *Response) renderSSR(s *Server, w http.ResponseWriter, data *page.InertiaPage) error {
	renderPath, err := url.JoinPath(s.ssrURL, "/render")
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

	resp, err := s.ssrHTTPClient.Do(ssrReq)
	if err != nil {
		return errors.Join(errCommunicatingWithSSRServer, err)
	}
	defer resp.Body.Close()

	var ssrRes ssrResponse
	err = json.NewDecoder(resp.Body).Decode(&ssrRes)
	if err != nil {
		return errors.Join(errCommunicatingWithSSRServer, err)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(req.status)

	baseHead := s.inertiaBaseHead()
	return s.rootTemplate.Execute(w, rootTmplData{
		InertiaRoot: template.HTML(ssrRes.Body),
		InertiaHead: baseHead + "\n" + template.HTML(strings.Join(ssrRes.Head, "\n")), // this is for SSR later
	})
}

func (s *Server) reactRefreshScript(attrs []template.HTMLAttr) template.HTML {
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
</script>`, attributes, s.viteDevUrl))
}
