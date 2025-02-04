package inertia

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

func Render(w http.ResponseWriter, r *http.Request, page string, pageProps map[string]any) error {
	server, err := getServer(r)
	if err != nil {
		return err
	}
	return server.Render(w, r, page, pageProps)
}

type pageData struct {
	Component      string `json:"component"`
	Url            string `json:"url"`
	Props          Props  `json:"props"`
	Version        string `json:"version"`
	EncryptHistory bool   `json:"encryptHistory"`
	ClearHistory   bool   `json:"clearHistory"`
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
	}

	if r.Method == http.MethodGet && r.Header.Get("X-Inertia-Version") != "" && r.Header.Get("X-Inertia-Version") != s.manifestVersion {
		w.Header().Set("X-Inertia-Location", r.URL.String())
		w.WriteHeader(http.StatusConflict)
		return nil
	}

	if r.Header.Get("X-Inertia") == "true" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Vary", "X-Inertia")
		w.Header().Set("X-Inertia", "true")
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
