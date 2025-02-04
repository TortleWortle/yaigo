package inertia

import (
	"go.tortle.tech/go-inertia/pkg/props"
	"html/template"
	"net/http"
)

type request struct {
	propBag props.Bag
	status  int
}

func newRequest() *request {
	return &request{
		propBag: props.NewBag(),
		status:  http.StatusOK,
	}
}

func (req *request) Reset() {
	req.propBag.Clear()
	req.status = http.StatusOK
}

type rootTmplData struct {
	InertiaRoot template.HTML
	InertiaHead template.HTML
}
