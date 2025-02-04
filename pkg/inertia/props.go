package inertia

import (
	"context"
	"errors"
	"net/http"

	"go.tortle.tech/go-inertia/pkg/props"
)

type Props map[string]any

func SetProp[T any](r *http.Request, key string, value T) error {
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
	return req.PropBag, nil
}

type DeferredProp struct {
	group string
	fn    DeferredPropFn
}
type DeferredPropFn = func() (any, error)

func Defer(fn DeferredPropFn) *DeferredProp {
	return &DeferredProp{
		group: "default",
		fn:    fn,
	}
}

func (p *DeferredProp) Group(name string) *DeferredProp {
	p.group = name
	return p
}
