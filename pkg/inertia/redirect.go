package inertia

import (
	"github.com/tortlewortle/yaigo/internal/errflash"
	"github.com/tortlewortle/yaigo/pkg/yaigo"
	"net/http"
)

// Redirect instructs inertia to redirect properly using http.StatusSeeOther
func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// Location redirects to external urls
func Location(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set(yaigo.HeaderLocation, url)
	w.WriteHeader(http.StatusConflict)
}

type FlashErrors = errflash.FlashErrors

func Back(w http.ResponseWriter, r *http.Request, errs errflash.FlashErrors) {
	if errs != nil {
		errflash.FlashError(w, r, errs)
	}
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}
