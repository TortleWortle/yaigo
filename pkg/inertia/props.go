package inertia

import (
	"context"
	"github.com/tortlewortle/yaigo/internal/props"
	"github.com/tortlewortle/yaigo/pkg/yaigo"
)

type Props = yaigo.Props

func SetProp(ctx context.Context, key string, value any) {
	yaigo.SetProp(ctx, key, value)
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

// PropResult returns an OK value
func PropResult[T any](value T) PropValue[T] {
	return PropValue[T]{
		Data:  &value,
		Error: nil,
	}
}

// PropErr returns an Error value
func PropErr[T any](message string, cause error) PropValue[T] {
	var causeMsg string
	if cause != nil {
		causeMsg = cause.Error()
	}
	return PropValue[T]{
		Data: nil,
		Error: &PropError{
			Message: message,
			Cause:   causeMsg,
		},
	}
}

type PropValue[T any] struct {
	Data  *T         `json:"data" or:"error"`
	Error *PropError `json:"error"`
}

type PropError struct {
	Message string `json:"message"`
	Cause   string `json:"cause"`
}
