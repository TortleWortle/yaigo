package inertia

import (
	"go.tortle.tech/go-inertia/pkg/props"
	"html/template"
	"net/http"
)

type request struct {
	propBag  *props.Bag
	status   int
	tmpl     *template.Template
	pageData *pageData
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
