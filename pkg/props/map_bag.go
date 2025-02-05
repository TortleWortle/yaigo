package props

import (
	"iter"
	"maps"
)

type mapBag struct {
	data map[string]any
}

func (b *mapBag) Items() map[string]any {
	return maps.Clone(b.data)
}

func (b *mapBag) Set(key string, value any) {
	b.data[key] = value
}

func (b *mapBag) Get(key string) (any, bool) {
	value, ok := b.data[key]
	return value, ok
}

func (b *mapBag) Keys() iter.Seq[string] {
	return maps.Keys(b.data)
}

func (b *mapBag) Clear() {
	for key := range b.data {
		delete(b.data, key)
	}
}
