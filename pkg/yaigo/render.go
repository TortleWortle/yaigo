package yaigo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tortlewortle/yaigo/pkg/typegen"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/tortlewortle/yaigo/internal/errflash"
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
	headerReset            = "X-Inertia-Reset"
)

type Props map[string]any

func (s *Server) Render(w http.ResponseWriter, r *http.Request, page string, pageProps Props) error {
	req := NewRequest()

	return s.RenderRequest(req, w, r, page, pageProps)
}

func (s *Server) RenderRequest(res *Request, w http.ResponseWriter, r *http.Request, page string, pageProps Props) error {
	hb := newHeaderBag(r)
	isPartial := hb.IsPartial(page)

	if hb.RedirectIfVersionConflict(w, s.manifestVersion) {
		errflash.Reflash(w, r)
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
		bag.Set(k, v)
	}

	data.Component = page
	data.Url = r.URL.Path
	data.Version = s.manifestVersion

	if isPartial {
		res.filterPartialProps(hb)
	}

	data.Props, err = bag.GetProps(r.Context())
	if err != nil {
		return fmt.Errorf("loading props: %w", err)
	}
	data.DeferredProps = bag.GetDeferredProps()

	if s.typeGenFilePath != "" {
		s.typeGenLock.Lock()
		defer s.typeGenLock.Unlock()

		propsForGen := Props{}
		forcedOptionals := s.typeGenOptionalsCache[page]

		existingProps, ok := s.typeGenPropCache[page]
		if ok {
			// if cache exists
			for k, v := range existingProps {
				propsForGen[k] = v
			}

			for k, v := range data.Props {
				_, ok := propsForGen[k]
				if !ok {
					// prop is new, probably deferred, lets mark it forced optional
					forcedOptionals = append(forcedOptionals, k)
				}
				propsForGen[k] = v
			}
		} else {
			for k, v := range data.Props {
				propsForGen[k] = v
			}
		}

		s.typeGenOptionalsCache[page] = forcedOptionals
		s.typeGenPropCache[page] = propsForGen

		err = typegen.GenerateTypeScriptForComponent(s.typeGenFilePath, page, propsForGen, forcedOptionals)
		if err != nil {
			return fmt.Errorf("typegen generation: %w", err)
		}
	}

	if hb.IsInertia() {
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

func (req *Request) filterPartialProps(rb *requestBag) {
	bag := req.propBag
	bag.LoadDeferred()
	onlyProps := rb.OnlyProps()
	if len(onlyProps) > 0 {
		bag.Only(onlyProps)
	}

	exceptProps := rb.ExceptProps()
	if len(exceptProps) > 0 {
		bag.Except(exceptProps)
	}
}

func (req *Request) renderJson(w http.ResponseWriter, data *page.InertiaPage) error {
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

func (req *Request) renderHtml(s *Server, w http.ResponseWriter, data *page.InertiaPage) error {
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

func (req *Request) renderSSR(s *Server, w http.ResponseWriter, data *page.InertiaPage) error {
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
