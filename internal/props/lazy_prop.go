package props

type LazyProp struct {
	group    string
	fn       LazyPropFn
	sync     bool
	deferred bool

	// after executing
	result any
	err    error
}
type LazyPropFn = func() (any, error)

func NewLazyProp(fn LazyPropFn, group string, deferred, sync bool) *LazyProp {
	return &LazyProp{
		group:    group,
		fn:       fn,
		sync:     sync,
		deferred: deferred,
	}
}

// Execute the callback and populate the result and err fields.
//
// Technically should get a context.Context passed in, but it is expected that the functions provided already have those from the *http.Request
func (p *LazyProp) Execute() {
	val, err := p.fn()

	p.result = val
	p.err = err
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
