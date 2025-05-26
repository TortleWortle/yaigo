package inertia

import (
	"context"
	"github.com/tortlewortle/yaigo/pkg/yaigo"
)

type Props = yaigo.Props

func SetProp(ctx context.Context, key string, value any) {
	yaigo.SetProp(ctx, key, value)
}
