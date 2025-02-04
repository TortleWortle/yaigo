package inertia

import (
	"go.tortle.tech/go-inertia/pkg/props"
	"html/template"
)

type request struct {
	PropBag props.Bag
}

func newRequest() *request {
	return &request{
		PropBag: props.NewBag(),
	}
}

func (req *request) Reset() {
	req.PropBag.Clear()
}

type rootTmplData struct {
	InertiaRoot template.HTML
	InertiaHead template.HTML
}
