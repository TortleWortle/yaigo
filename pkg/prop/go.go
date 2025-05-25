package prop

import (
	"context"
)

// Go wraps GoAny with a prop helper
//
// GoAny evaluates the prop fn in a separate goroutine
func Go[T any](fn func(ctx context.Context) Result[T]) *LazyProp {
	return GoAny(func(ctx context.Context) (any, error) {
		return fn(ctx), nil
	})
}

// GoAny evaluates the prop fn in a separate goroutine
func GoAny(fn LazyPropFn) *LazyProp {
	return NewLazyProp(fn, false, false)
}

// DeferPropSync defers a prop to be loaded by inertia in a separate request
//
// The callback will be run sequentially
func DeferPropSync(fn LazyPropFn) *LazyProp {
	return NewLazyProp(fn, true, true)
}

// DeferAny defers a prop to be loaded by inertia in a separate request.
//
// The callback will be run concurrently with other normal DeferAny props
func DeferAny(fn LazyPropFn) *LazyProp {
	return NewLazyProp(fn, true, false)
}

// Defer wraps DeferAny with a prop helper
func Defer[T any](fn func(ctx context.Context) Result[T]) *LazyProp {
	return DeferAny(func(ctx context.Context) (any, error) {
		return fn(ctx), nil
	})
}

// Ok returns an OK value
func Ok[T any](value T) Result[T] {
	return Result[T]{
		Data:  &value,
		Error: nil,
	}
}

// Err returns an Err value
func Err[T any](message string, cause error) Result[T] {
	var causeMsg string
	if cause != nil {
		causeMsg = cause.Error()
	}
	return Result[T]{
		Data: nil,
		Error: &PropError{
			Message: message,
			Cause:   causeMsg,
		},
	}
}

type Result[T any] struct {
	Data  *T         `json:"data" or:"error"`
	Error *PropError `json:"error"`
}

type PropError struct {
	Message string `json:"message"`
	Cause   string `json:"cause"`
}
