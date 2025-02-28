package yaigo

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

type requestBag struct {
	inertiaHeader          string
	errorBagHeader         string
	versionHeader          string
	partialComponentHeader string
	partialOnlyHeader      string
	partialExceptHeader    string
	method                 string
}

func (r *requestBag) IsPartial(page string) bool {
	return r.partialComponentHeader == page
}

func newHeaderBag(r *http.Request) *requestBag {
	h := r.Header
	return &requestBag{
		inertiaHeader:          h.Get(HeaderInertia),
		errorBagHeader:         h.Get(HeaderErrorBag),
		versionHeader:          h.Get(HeaderVersion),
		partialComponentHeader: h.Get(HeaderPartialComponent),
		partialOnlyHeader:      h.Get(HeaderPartialOnly),
		partialExceptHeader:    h.Get(HeaderPartialExcept),
		method:                 r.Method,
	}
}

// RedirectIfVersionConflict redirects the request if the manifest version is outdated on the client, returns true if it has been redirected
func (r *requestBag) RedirectIfVersionConflict(w http.ResponseWriter, version string, target string) bool {
	if !r.IsInertia() {
		return false
	}

	if r.versionHeader == version {
		return false
	}

	w.Header().Set(headerLocation, target)
	w.WriteHeader(http.StatusConflict)
	return true
}

func (r *requestBag) IsInertia() bool {
	return r.inertiaHeader == "true"
}

func (r *requestBag) OnlyProps() []string {
	if r.partialOnlyHeader == "" {
		return []string{}
	}
	return strings.Split(r.partialOnlyHeader, ",")
}

func (r *requestBag) ExceptProps() []string {
	if r.partialExceptHeader == "" {
		return []string{}
	}
	return strings.Split(r.partialExceptHeader, ",")
}
