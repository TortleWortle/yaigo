package inertia

import (
	"encoding/json"
	"fmt"
	"html/template"
	"maps"
	"net/http"
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

func Render(w http.ResponseWriter, r *http.Request, page string, pageProps map[string]any) error {
	server, err := getServer(r)
	if err != nil {
		return err
	}
	return server.Render(w, r, page, pageProps)
}

func (s *Server) Render(w http.ResponseWriter, r *http.Request, page string, pageProps Props) error {
	req, err := getRequest(r)
	if err != nil {
		return err
	}

	bag := req.propBag
	props := maps.Clone(bag.Items())
	for k, v := range pageProps {
		props[k] = v
	}

	data := &pageData{
		Component:      page,
		Props:          props,
		Url:            r.URL.Path, // todo: does this need query params?,
		Version:        s.manifestVersion,
		EncryptHistory: false, // seems to be hardcoded in the laravel implementation
		ClearHistory:   false, // seems to be hardcoded in the laravel implementation
		DeferredProps:  make(map[string][]string),
	}

	isInertia := r.Header.Get(HeaderInertia) == "true"
	partialComponent := r.Header.Get(HeaderPartialComponent)
	isPartial := partialComponent == data.Component

	// detect frontend version changes
	if isInertia && r.Method == http.MethodGet && r.Header.Get(HeaderVersion) != s.manifestVersion {
		w.Header().Set(HeaderLocation, r.URL.String())
		w.WriteHeader(http.StatusConflict)
		return nil
	}

	// resolve deferred props
	if isPartial {
		return handlePartial(w, r, req.status, data)
	}

	// filter out deferred props and add them to the deferred props object
	err = data.moveDeferredProps()
	if err != nil {
		return fmt.Errorf("moving deferred props: %w", err)
	}

	if isInertia {
		return renderJson(w, req.status, data)
	}

	return renderHtml(s.rootTemplate, w, req.status, data)
}

func handlePartial(w http.ResponseWriter, r *http.Request, status int, data *pageData) error {
	onlyPropsStr := r.Header.Get(HeaderPartialOnly)
	if onlyPropsStr != "" {
		onlyProps := strings.Split(onlyPropsStr, ",")
		maps.DeleteFunc(data.Props, func(k string, v any) bool {
			return !slices.Contains(onlyProps, k)
		})
	}

	exceptPropsStr := r.Header.Get(HeaderPartialExcept)
	if exceptPropsStr != "" {
		exceptProps := strings.Split(exceptPropsStr, ",")
		maps.DeleteFunc(data.Props, func(k string, v any) bool {
			return slices.Contains(exceptProps, k)
		})
	}

	err := data.evalProps()
	if err != nil {
		return fmt.Errorf("evaluating props: %w", err)
	}

	return renderJson(w, status, data)
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
		InertiaHead: template.HTML(""),
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
