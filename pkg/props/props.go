package props

import (
	"iter"
	"sync"
)

type Bag interface {
	Set(string, any)
	Get(string) (any, bool)
	Keys() iter.Seq[string]
	Merge(items map[string]any)
	Clear()
	Items() map[string]any
}

type mapBag struct {
	data *sync.Map
}

func (b *mapBag) Items() map[string]any {
	items := make(map[string]any)
	b.data.Range(func(key, value any) bool {
		keyStr := key.(string)
		items[keyStr] = value
		return true
	})
	return items
}

func NewBag() Bag {
	return &mapBag{
		data: &sync.Map{},
	}
}

func (b *mapBag) Set(key string, value any) {
	b.data.Store(key, value)
}

func (b *mapBag) Get(key string) (any, bool) {
	value, ok := b.data.Load(key)
	return value, ok
}

func (b *mapBag) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		b.data.Range(func(key, _ any) bool {
			keyStr := key.(string)
			yield(keyStr)
			return true
		})
		return
	}
}

func (b *mapBag) Clear() {
	b.data.Clear()
}

func (b *mapBag) Merge(items map[string]any) {
	for key, value := range items {
		b.Set(key, value)
	}
}
