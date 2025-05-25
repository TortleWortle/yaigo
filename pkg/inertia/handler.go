package inertia

import (
	"github.com/tortlewortle/yaigo/internal/errflash"
	"net/http"
)

const (
	headerLocation = "X-Inertia-Location"
)

func PageHandler(component string, props Props) http.HandlerFunc {
	page := Page(component, props)
	return func(w http.ResponseWriter, r *http.Request) {
		_ = page.Render(r.Context(), w)
	}
}

// Redirect instructs inertia to redirect properly using http.StatusSeeOther
func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// Location redirects to external urls
func Location(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set(headerLocation, url)
	w.WriteHeader(http.StatusConflict)
}

type FlashErrors = errflash.FlashErrors

func Back(w http.ResponseWriter, r *http.Request, errs errflash.FlashErrors) {
	if errs != nil {
		errflash.FlashError(w, r, errs)
	}
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}
