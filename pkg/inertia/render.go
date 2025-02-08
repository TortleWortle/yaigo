package inertia

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

const (
	HeaderInertia          = "X-Inertia"
	HeaderErrorBag         = "X-Inertia-Error-Bag"
	HeaderLocation         = "X-Inertia-Location"
	HeaderVersion          = "X-Inertia-Version"
	HeaderPartialComponent = "X-Inertia-Partial-Component"
	HeaderPartialOnly      = "X-Inertia-Partial-Data"
	HeaderPartialExcept    = "X-Inertia-Partial-Except"
)

func Render(w http.ResponseWriter, r *http.Request, page string, pageProps Props) error {
	server, err := getServer(r)
	if err != nil {
		return err
	}
	return server.Render(w, r, page, pageProps)
}

func (s *Server) Render(w http.ResponseWriter, r *http.Request, page string, pageProps Props) error {
	isInertia := r.Header.Get(HeaderInertia) == "true"
	partialComponent := r.Header.Get(HeaderPartialComponent)
	isPartial := partialComponent == page

	// detect frontend version changes
	if isInertia && r.Method == http.MethodGet && r.Header.Get(HeaderVersion) != s.manifestVersion {
		w.Header().Set(HeaderLocation, r.URL.String())
		w.WriteHeader(http.StatusConflict)
		return nil
	}

	req, err := getRequest(r)
	if err != nil {
		req = s.requestPool.Get().(*request)
	}

	data := req.pageData
	// a little yank, but if we ever fail to render a page and wish to render an error-page, we need to clear the props of the request.
	data.resetIfDirty()

	bag := req.propBag
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

	// resolve deferred props
	if isPartial {
		return req.handlePartial(s, w, r, data)
	}

	// from this point there is no special logic in prop eval
	data.Props, data.DeferredProps, err = bag.GetProps()
	if err != nil {
		return fmt.Errorf("loading props: %w", err)
	}
	data.DeferredProps = bag.GetDeferredProps()

	if isInertia {
		return req.renderJson(w, data)
	}

	if s.ssrURL != "" {
		err = req.renderSSR(s, w, data)
		if err != nil {
			if errors.Is(err, errCommunicatingToSSRServer) {
				// render client side if ssr is unreachable
				return req.renderHtml(s, w, data)
			}
			return err
		}
		return nil
	}
	return req.renderHtml(s, w, data)
}

func (req *request) handlePartial(s *Server, w http.ResponseWriter, r *http.Request, data *pageData) error {
	bag := req.propBag
	onlyPropsStr := r.Header.Get(HeaderPartialOnly)
	if onlyPropsStr != "" {
		onlyProps := strings.Split(onlyPropsStr, ",")
		bag.Only(onlyProps)
	}

	// can't realistically be avoided
	exceptPropsStr := r.Header.Get(HeaderPartialExcept)
	if exceptPropsStr != "" {
		exceptProps := strings.Split(exceptPropsStr, ",")
		bag.Except(exceptProps)
	}
	var err error
	data.Props, _, err = bag.GetProps()
	if err != nil {
		return err
	}
	// eval props
	return req.renderJson(w, data)
}

type ssrResponse struct {
	Head []string `json:"head"`
	Body string   `json:"body"`
}

var errCommunicatingToSSRServer = errors.New("could not communicate with ssr server")

func (req *request) renderSSR(s *Server, w http.ResponseWriter, data *pageData) error {
	renderPath, err := url.JoinPath(s.ssrURL, "/render")
	if err != nil {
		return err
	}
	pData, err := json.Marshal(data)
	ssrReq, err := http.NewRequest("GET", renderPath, bytes.NewReader(pData))
	if err != nil {
		return errors.Join(errCommunicatingToSSRServer, err)
	}

	resp, err := s.ssrHTTPClient.Do(ssrReq)
	if err != nil {
		return errors.Join(errCommunicatingToSSRServer, err)
	}
	defer resp.Body.Close()

	var ssrRes ssrResponse
	err = json.NewDecoder(resp.Body).Decode(&ssrRes)
	if err != nil {
		return errors.Join(errCommunicatingToSSRServer, err)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(req.status)

	baseHead := s.inertiaBaseHead()
	return s.rootTemplate.Execute(w, rootTmplData{
		InertiaRoot: template.HTML(ssrRes.Body),
		InertiaHead: baseHead + "\n" + template.HTML(strings.Join(ssrRes.Head, "\n")), // this is for SSR later
	})
}

func (s *Server) inertiaBaseHead() template.HTML {
	if s.reactRefresh {
		return s.reactRefreshScript(nil)
	}
	return ""
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

func (req *request) renderHtml(s *Server, w http.ResponseWriter, data *pageData) error {
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

func (req *request) renderJson(w http.ResponseWriter, data *pageData) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Vary", HeaderInertia)
	w.Header().Set(HeaderInertia, "true")
	w.WriteHeader(req.status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return err
	}
	return nil
}
