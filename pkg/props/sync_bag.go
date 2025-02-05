package props

import (
	"iter"
	"sync"
)

type syncMapBag struct {
	data *sync.Map
}

func (b *syncMapBag) Items() map[string]any {
	items := make(map[string]any)
	b.data.Range(func(key, value any) bool {
		keyStr := key.(string)
		items[keyStr] = value
		return true
	})
	return items
}

func (b *syncMapBag) Set(key string, value any) {
	b.data.Store(key, value)
}

func (b *syncMapBag) Get(key string) (any, bool) {
	value, ok := b.data.Load(key)
	return value, ok
}

func (b *syncMapBag) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		b.data.Range(func(key, _ any) bool {
			keyStr := key.(string)
			yield(keyStr)
			return true
		})
		return
	}
}

func (b *syncMapBag) Clear() {
	b.data.Clear()
}
