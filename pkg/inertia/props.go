package inertia

import (
	"context"
	"errors"
	"net/http"

	"go.tortle.tech/go-inertia/pkg/props"
)

type Props map[string]any

func SetProp(r *http.Request, key string, value any) error {
	switch value.(type) {
	case *LazyProp:
		return errors.New("deferred props can only be used on the page render func")
	}
	return setPropCtx(r.Context(), key, value)
}

func GetProp[T any](r *http.Request, key string) (value T, err error) {
	return getPropCtx[T](r.Context(), key)
}

func GetPropBag(r *http.Request) (props.Bag, error) {
	return getPropBagCtx(r.Context())
}

func setPropCtx[T any](ctx context.Context, key string, value T) error {
	bag, err := getPropBagCtx(ctx)
	if err != nil {
		return err
	}

	bag.Set(key, value)
	return nil
}

func getPropCtx[T any](ctx context.Context, key string) (value T, err error) {
	bag, err := getPropBagCtx(ctx)
	if err != nil {
		return value, err
	}

	val, ok := bag.Get(key)
	if !ok {
		return value, errors.New("value nil")
	}

	castVal, ok := val.(T)
	if !ok {
		return value, errors.New("could not cast prop")
	}
	return castVal, nil
}

func getPropBagCtx(ctx context.Context) (props.Bag, error) {
	req, err := getRequestCtx(ctx)
	if err != nil {
		return nil, err
	}
	return req.propBag, nil
}

type LazyProp struct {
	group    string
	fn       LazyPropFn
	sync     bool
	deferred bool
}
type LazyPropFn = func() (any, error)

// Resolve evaluates the prop fn in a separate goroutine
func Resolve(fn LazyPropFn) *LazyProp {
	return &LazyProp{
		group:    "default",
		fn:       fn,
		sync:     false,
		deferred: false,
	}
}

// DeferSync defers a prop to be loaded by inertia in a separate request
//
// The callback will be run sequentially
func DeferSync(fn LazyPropFn) *LazyProp {
	return &LazyProp{
		group:    "default",
		fn:       fn,
		sync:     true,
		deferred: true,
	}
}

// Defer defers a prop to be loaded by inertia in a separate request.
//
// The callback will be run concurrently with other normal Defer props
func Defer(fn LazyPropFn) *LazyProp {
	return &LazyProp{
		group:    "default",
		fn:       fn,
		sync:     false,
		deferred: true,
	}
}

// Group sets the group name InertiaJS fetches each group in a separate request
//
// I believe this is more for the PHP world so you can concurrently load two things at the same time.
// We do not need this we can use the inertia.Resolve method instead to concurrently fetch props in the same request.
func (p *LazyProp) Group(name string) *LazyProp {
	p.group = name
	return p
}
