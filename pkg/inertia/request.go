package inertia

import (
	"net/http"
)

// EncryptHistory enables or disables page history encryption inside inertiajs
func EncryptHistory(r *http.Request, encrypt bool) error {
	req, err := getRequest(r)
	if err != nil {
		return err
	}
	req.EncryptHistory(encrypt)
	return nil
}

// ClearHistory tells inertiajs to roll the cache encryption key.
// This can be used to protect any sensitive information from being accessed after logout by using the back button.
func ClearHistory(r *http.Request) error {
	req, err := getRequest(r)
	if err != nil {
		return err
	}
	req.ClearHistory()
	return nil
}

// SetStatus of the http response
func SetStatus(r *http.Request, status int) error {
	req, err := getRequest(r)
	if err != nil {
		return err
	}
	req.SetStatus(status)
	return nil
}

const (
	headerLocation = "X-Inertia-Location"
)

func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func Location(w http.ResponseWriter, url string) {
	w.Header().Set(headerLocation, url)
	w.WriteHeader(http.StatusConflict)
}
