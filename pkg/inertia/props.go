package inertia

import (
	"github.com/tortlewortle/go-inertia/internal/props"
	"github.com/tortlewortle/go-inertia/pkg/yaigo"
	"net/http"
)

type Props = yaigo.Props

func SetProp(r *http.Request, key string, value any) error {
	req, err := getRequest(r)
	if err != nil {
		return err
	}
	return req.SetProp(key, value)
}

// Resolve evaluates the prop fn in a separate goroutine
func Resolve(fn props.LazyPropFn) *props.LazyProp {
	return props.NewLazyProp(fn, false, false)
}

// DeferSync defers a prop to be loaded by inertia in a separate request
//
// The callback will be run sequentially
func DeferSync(fn props.LazyPropFn) *props.LazyProp {
	return props.NewLazyProp(fn, true, true)
}

// Defer defers a prop to be loaded by inertia in a separate request.
//
// The callback will be run concurrently with other normal Defer props
func Defer(fn props.LazyPropFn) *props.LazyProp {
	return props.NewLazyProp(fn, true, false)
}
