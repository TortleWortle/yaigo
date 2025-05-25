package inertia

import (
	"github.com/tortlewortle/yaigo/pkg/yaigo"
	"net/http"
)

func Page(component string, props Props) *yaigo.Page {
	return yaigo.NewPage(component, props)
}

func PageHandler(component string, props Props) http.HandlerFunc {
	page := Page(component, props)
	return func(w http.ResponseWriter, r *http.Request) {
		_ = page.Render(r.Context(), w)
	}
}
