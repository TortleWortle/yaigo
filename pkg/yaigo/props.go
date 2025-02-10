package yaigo

import "github.com/tortlewortle/go-inertia/internal/props"

func NewDeferredProp(fn props.LazyPropFn, group string) *props.LazyProp {
	return props.NewLazyProp(fn, true, true)
}
