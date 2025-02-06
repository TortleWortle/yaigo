package inertia

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"maps"
	"net/http"
	"net/url"
	"slices"
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

	for k, v := range req.propBag.Items() {
		data.Props[k] = v
	}

	for k, v := range pageProps {
		data.Props[k] = v
	}

	data.Component = page
	data.Url = r.URL.Path
	data.Version = s.manifestVersion

	// resolve deferred props
	if isPartial {
		return handlePartial(w, r, req.status, data)
	}

	// filter out deferred props and add them to the deferred props object
	err = data.moveDeferredProps()
	if err != nil {
		return fmt.Errorf("moving deferred props: %w", err)
	}

	// evaluate any remaining LazyProps
	err = data.evalLazyProps()
	if err != nil {
		return fmt.Errorf("evaluating deferred props: %w", err)
	}

	if isInertia {
		return renderJson(w, req.status, data)
	}
	if s.ssrURL != "" {
		err = renderSSR(req.tmpl, s.ssrURL, w, req.status, data)
		if err != nil {
			if errors.Is(err, errCommunicatingToSSRServer) {
				return renderHtml(req.tmpl, w, req.status, data)
			}
			return err
		}
		return nil
	}
	return renderHtml(req.tmpl, w, req.status, data)
}

func handlePartial(w http.ResponseWriter, r *http.Request, status int, data *pageData) error {
	// can't be avoided ig
	onlyPropsStr := r.Header.Get(HeaderPartialOnly)
	if onlyPropsStr != "" {
		onlyProps := strings.Split(onlyPropsStr, ",")
		maps.DeleteFunc(data.Props, func(k string, v any) bool {
			return !slices.Contains(onlyProps, k)
		})
	}

	// can't realistically be avoided
	exceptPropsStr := r.Header.Get(HeaderPartialExcept)
	if exceptPropsStr != "" {
		exceptProps := strings.Split(exceptPropsStr, ",")
		maps.DeleteFunc(data.Props, func(k string, v any) bool {
			return slices.Contains(exceptProps, k)
		})
	}

	err := data.evalLazyProps()
	if err != nil {
		return fmt.Errorf("evaluating props: %w", err)
	}

	return renderJson(w, status, data)
}

type ssrResponse struct {
	Head []string `json:"head"`
	Body string   `json:"body"`
}

var ssrHTTPClient http.Client

var errCommunicatingToSSRServer = errors.New("could not communicate with ssr server")

func renderSSR(rootTemplate *template.Template, ssrUrl string, w http.ResponseWriter, status int, data *pageData) error {
	renderPath, err := url.JoinPath(ssrUrl, "/render")
	if err != nil {
		return err
	}
	pData, err := json.Marshal(data)
	req, err := http.NewRequest("GET", renderPath, bytes.NewReader(pData))
	if err != nil {
		return errors.Join(errCommunicatingToSSRServer, err)
	}

	resp, err := ssrHTTPClient.Do(req)
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
	w.WriteHeader(status)

	//inertiaRoot := template.HTML(fmt.Sprintf("<div id=\"app\" data-page='%s'></div>", propStr))
	return rootTemplate.Execute(w, rootTmplData{
		InertiaRoot: template.HTML(ssrRes.Body),
		InertiaHead: template.HTML(strings.Join(ssrRes.Head, "\n")), // this is for SSR later
	})
}

func renderHtml(rootTemplate *template.Template, w http.ResponseWriter, status int, data *pageData) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	propStr, err := json.Marshal(data)
	if err != nil {
		return err
	}
	inertiaRoot := template.HTML(fmt.Sprintf("<div id=\"app\" data-page='%s'></div>", propStr))
	return rootTemplate.Execute(w, rootTmplData{
		InertiaRoot: inertiaRoot,
		InertiaHead: template.HTML(""), // this is for SSR later
	})
}

func renderJson(w http.ResponseWriter, status int, data *pageData) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Vary", HeaderInertia)
	w.Header().Set(HeaderInertia, "true")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return err
	}
	return nil
}
