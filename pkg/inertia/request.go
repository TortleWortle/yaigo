package inertia

import (
	"github.com/tortlewortle/go-inertia/pkg/props"
	"html/template"
	"net/http"
)

type request struct {
	propBag  *props.Bag
	status   int
	tmpl     *template.Template
	pageData *pageData
}

func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func Location(w http.ResponseWriter, url string) {
	w.Header().Set(HeaderLocation, url)
	w.WriteHeader(http.StatusConflict)
}

func EncryptHistory(r *http.Request, encrypt bool) error {
	req, err := getRequest(r)
	if err != nil {
		return err
	}
	req.pageData.EncryptHistory = encrypt
	return nil
}

func ClearHistory(r *http.Request) error {
	req, err := getRequest(r)
	if err != nil {
		return err
	}
	req.pageData.ClearHistory = true
	return nil
}

func newRequest(tmpl *template.Template) *request {
	return &request{
		propBag:  props.NewBag(),
		status:   http.StatusOK,
		tmpl:     tmpl,
		pageData: newPageData(),
	}
}

func (req *request) Reset() {
	req.propBag.Clear()
	req.status = http.StatusOK
	req.pageData.Reset()
}

type rootTmplData struct {
	InertiaRoot template.HTML
	InertiaHead template.HTML
}
