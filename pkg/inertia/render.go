package inertia

import (
	"encoding/json"
	"errors"
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

type pageData struct {
	Component      string              `json:"component"`
	Url            string              `json:"url"`
	Props          Props               `json:"props"`
	Version        string              `json:"version"`
	EncryptHistory bool                `json:"encryptHistory"`
	ClearHistory   bool                `json:"clearHistory"`
	DeferredProps  map[string][]string `json:"deferredProps"`
}

func (s *Server) Render(w http.ResponseWriter, r *http.Request, page string, pageProps Props) error {
	req, err := getRequest(r)
	if err != nil {
		return err
	}

	bag := req.PropBag
	bag.Merge(pageProps)

	data := pageData{
		Component:      page,
		Props:          bag.Items(),
		Url:            r.URL.Path, // todo: does this need query params?,
		Version:        s.manifestVersion,
		EncryptHistory: false, // seems to be hardcoded in the laravel implementation
		ClearHistory:   false, // seems to be hardcoded in the laravel implementation
		DeferredProps:  make(map[string][]string),
	}

	isInertia := r.Header.Get(HeaderInertia) == "true"
	partialComponent := r.Header.Get(HeaderPartialComponent)
	isPartial := partialComponent == data.Component

	// detect frontend changes
	if isInertia && r.Method == http.MethodGet && r.Header.Get(HeaderVersion) != s.manifestVersion {
		w.Header().Set(HeaderLocation, r.URL.String())
		w.WriteHeader(http.StatusConflict)
		return nil
	}

	// resolve deferred props
	if isInertia && isPartial {
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

		// evaluate deferred props and set the values
		for k, v := range data.Props {
			switch v.(type) {
			case *DeferredProp:
				prop, ok := v.(*DeferredProp)
				if !ok {
					return errors.New("could not cast prop value to DeferredProp")
				}
				v, err := prop.fn()
				if err != nil {
					return err
				}
				data.Props[k] = v
			}
		}
	}

	// remove any deferred props here and keep track of their names (todo: groups)
	maps.DeleteFunc(data.Props, func(k string, v any) bool {
		switch v.(type) {
		case *DeferredProp:
			if !isPartial {
				prop, ok := v.(*DeferredProp)
				if !ok {
					return true
				}
				data.DeferredProps[prop.group] = append(data.DeferredProps[prop.group], k)
			}
			return true
		default:
			return false
		}
	})

	// render logic
	if isInertia {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Vary", HeaderInertia)
		w.Header().Set(HeaderInertia, "true")
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			return err
		}
		return nil
	}

	propStr, err := json.Marshal(data)
	if err != nil {
		return err
	}
	inertiaRoot := template.HTML(fmt.Sprintf("<div id=\"app\" data-page='%s'></div>", propStr))
	return s.rootTemplate.Execute(w, rootTmplData{
		InertiaRoot: inertiaRoot,
		InertiaHead: template.HTML(""),
	})
}
