package yaigo

import (
	"context"
	"github.com/tortlewortle/yaigo/internal/page"
	"github.com/tortlewortle/yaigo/internal/props"
)

type contextKey int

const (
	configKey contextKey = iota
	requestInfoKey
	bagKey
	pageDataKey
)

// WithConfig sets the *yaigo.Config in the context
func WithConfig(ctx context.Context, config *Config) context.Context {
	return context.WithValue(ctx, configKey, config)
}

// WithRequestInfo sets the url in the context for inertia page data
func WithRequestInfo(ctx context.Context, info *RequestInfo) context.Context {
	return context.WithValue(ctx, requestInfoKey, info)
}

// WithPropBag provides a prop bag to keep props set between middlewares
func WithPropBag(ctx context.Context, bag *props.Bag) context.Context {
	return context.WithValue(ctx, bagKey, bag)
}

// WithInertiaPage provides page data struct
func WithInertiaPage(ctx context.Context, bag *page.InertiaPage) context.Context {
	return context.WithValue(ctx, pageDataKey, bag)
}
