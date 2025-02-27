package inertia

import (
	"net/http"
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

type InertiaRequest struct {
	inertiaHeader          string
	errorBagHeader         string
	versionHeader          string
	partialComponentHeader string
	partialOnlyHeader      string
	partialExceptHeader    string
}

func FromHTTPRequest(r *http.Request) InertiaRequest {
	h := r.Header
	return InertiaRequest{
		inertiaHeader:          h.Get(HeaderInertia),
		errorBagHeader:         h.Get(HeaderErrorBag),
		versionHeader:          h.Get(HeaderVersion),
		partialComponentHeader: h.Get(HeaderPartialComponent),
		partialOnlyHeader:      h.Get(HeaderPartialOnly),
		partialExceptHeader:    h.Get(HeaderPartialExcept),
	}
}

func (r *InertiaRequest) IsInertia() bool {
	return r.inertiaHeader == "true"
}

func (r *InertiaRequest) OnlyProps() []string {
	return strings.Split(r.partialOnlyHeader, ",")
}

func (r *InertiaRequest) ExceptProps() []string  {
	return strings.Split(r.partialExceptHeader, ",")
}