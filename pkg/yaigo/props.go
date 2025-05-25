package yaigo

import (
	"context"
	"github.com/tortlewortle/yaigo/pkg/prop"
)

type Props map[string]any

func SetProp(ctx context.Context, key string, value any) {
	bag, ok := ctx.Value(bagKey).(*prop.Bag)
	if !ok {
		panic("yaigo.SetProp: could not find bag in ctx")
	}
	bag.Set(key, value)
}
