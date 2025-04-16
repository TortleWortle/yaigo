package inertia

import (
	"context"
	"github.com/tortlewortle/yaigo/internal/props"
	"github.com/tortlewortle/yaigo/pkg/yaigo"
	"net/http"
)

type Props = yaigo.Props

func SetProp(r *http.Request, key string, value any) {
	req, err := getRequest(r)
	if err != nil {
		panic("SetProp: could not get *yaigo.Request from *http.Request context, is it wrapped in the middleware?")
	}
	req.SetProp(key, value)
}

// Resolve wraps ResolveProp with a prop helper
func Resolve[T any](fn func(ctx context.Context) PropValue[T]) *props.LazyProp {
	return ResolveProp(func(ctx context.Context) (any, error) {
		return fn(ctx), nil
	})
}

// ResolveProp evaluates the prop fn in a separate goroutine
func ResolveProp(fn props.LazyPropFn) *props.LazyProp {
	return props.NewLazyProp(fn, false, false)
}

// DeferPropSync defers a prop to be loaded by inertia in a separate request
//
// The callback will be run sequentially
func DeferPropSync(fn props.LazyPropFn) *props.LazyProp {
	return props.NewLazyProp(fn, true, true)
}

// DeferProp defers a prop to be loaded by inertia in a separate request.
//
// The callback will be run concurrently with other normal DeferProp props
func DeferProp(fn props.LazyPropFn) *props.LazyProp {
	return props.NewLazyProp(fn, true, false)
}

// Defer wraps DeferProp with a prop helper
func Defer[T any](fn func(ctx context.Context) PropValue[T]) *props.LazyProp {
	return DeferProp(func(ctx context.Context) (any, error) {
		return fn(ctx), nil
	})
}

// PropOk returns an OK value
func PropOk[T any](value T) PropValue[T] {
	return PropValue[T]{
		Data:  value,
		Error: PropError{},
	}
}

// PropErr returns an Error value
func PropErr[T any](reason, detail string) PropValue[T] {
	var value T
	return PropValue[T]{
		Data: value,
		Error: PropError{
			Reason: reason,
			Detail: detail,
		},
	}
}

type PropValue[T any] struct {
	Data  T         `json:"data" or:"error"`
	Error PropError `json:"error"`
}

type PropError struct {
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}
