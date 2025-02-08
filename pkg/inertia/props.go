package inertia

import (
	"errors"
	"github.com/tortlewortle/go-inertia/internal/props"
	"net/http"
)

type Props map[string]any

func SetProp(r *http.Request, key string, value any) error {
	switch value.(type) {
	case *props.LazyProp:
		p, ok := value.(*props.LazyProp)
		if ok {
			if p.IsDeferred() {
				return errors.New("deferred props can only be used on the page render func")
			}
		}
		return errors.New("could not cast LazyProp")
	}
	bag, err := GetPropBag(r)
	if err != nil {
		return err
	}

	bag.Set(key, value)
	return nil
}

func GetPropBag(r *http.Request) (*props.Bag, error) {
	req, err := getRequestCtx(r.Context())
	if err != nil {
		return nil, err
	}
	return req.propBag, nil
}

// Resolve evaluates the prop fn in a separate goroutine
func Resolve(fn props.LazyPropFn) *props.LazyProp {
	return props.NewLazyProp(fn, "default", false, false)
}

// DeferSync defers a prop to be loaded by inertia in a separate request
//
// The callback will be run sequentially
func DeferSync(fn props.LazyPropFn) *props.LazyProp {
	return props.NewLazyProp(fn, "default", true, true)
}

// Defer defers a prop to be loaded by inertia in a separate request.
//
// The callback will be run concurrently with other normal Defer props
func Defer(fn props.LazyPropFn) *props.LazyProp {
	return props.NewLazyProp(fn, "default", true, false)
}
