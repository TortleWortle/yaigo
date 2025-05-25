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

type RequestInfo struct {
	InertiaHeader          string
	ErrorBagHeader         string
	VersionHeader          string
	PartialComponentHeader string
	PartialOnlyHeader      string
	PartialExceptHeader    string
	Method                 string
	RequestURI             string
}

func (ri *RequestInfo) IsPartial(page string) bool {
	return ri.PartialComponentHeader == page
}

func (ri *RequestInfo) Fill(r *http.Request) {
	h := r.Header
	ri.InertiaHeader = h.Get(HeaderInertia)
	ri.ErrorBagHeader = h.Get(HeaderErrorBag)
	ri.VersionHeader = h.Get(HeaderVersion)
	ri.PartialComponentHeader = h.Get(HeaderPartialComponent)
	ri.PartialOnlyHeader = h.Get(HeaderPartialOnly)
	ri.PartialExceptHeader = h.Get(HeaderPartialExcept)
	ri.RequestURI = r.RequestURI
	ri.Method = r.Method
}

func (ri *RequestInfo) Empty() {
	ri.InertiaHeader = ""
	ri.ErrorBagHeader = ""
	ri.VersionHeader = ""
	ri.PartialComponentHeader = ""
	ri.PartialOnlyHeader = ""
	ri.PartialExceptHeader = ""
	ri.RequestURI = ""
	ri.Method = ""
}

// RedirectIfVersionConflict redirects the request if the manifest version is outdated on the client, returns true if it has been redirected
func (ri *RequestInfo) RedirectIfVersionConflict(w http.ResponseWriter, version string) bool {
	if !ri.IsInertia() {
		return false
	}

	if ri.VersionHeader == version {
		return false
	}

	return true
}

func (ri *RequestInfo) IsInertia() bool {
	return ri.InertiaHeader == "true"
}

func (ri *RequestInfo) OnlyProps() []string {
	if ri.PartialOnlyHeader == "" {
		return []string{}
	}
	return strings.Split(ri.PartialOnlyHeader, ",")
}

func (ri *RequestInfo) ExceptProps() []string {
	if ri.PartialExceptHeader == "" {
		return []string{}
	}
	return strings.Split(ri.PartialExceptHeader, ",")
}
