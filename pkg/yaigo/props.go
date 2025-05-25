package yaigo

import (
	"context"
	"github.com/tortlewortle/yaigo/internal/props"
)

type Props map[string]any

func NewDeferredProp(fn props.LazyPropFn, group string) *props.LazyProp {
	return props.NewLazyProp(fn, true, true)
}

func SetProp(ctx context.Context, key string, value any) {
	bag, ok := ctx.Value(bagKey).(*props.Bag)
	if !ok {
		panic("yaigo.SetProp: could not find bag in ctx")
	}
	bag.Set(key, value)
}
