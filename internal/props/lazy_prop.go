package props

import (
	"context"
)

type LazyProp struct {
	group    string
	fn       LazyPropFn
	sync     bool
	deferred bool
}

type LazyPropFn = func(ctx context.Context) (any, error)

func NewLazyProp(fn LazyPropFn, deferred, sync bool) *LazyProp {
	return &LazyProp{
		fn:       fn,
		sync:     sync,
		deferred: deferred,
		group:    "default",
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

func (p *LazyProp) IsDeferred() bool {
	return p.deferred
}
