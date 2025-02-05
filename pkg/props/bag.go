package props

import (
	"iter"
	"maps"
)

type Bag struct {
	data map[string]any
}

func NewBag() *Bag {
	return &Bag{
		data: make(map[string]any),
	}
}

// Items returns the underlying map
func (b *Bag) Items() map[string]any {
	return b.data
}

func (b *Bag) Set(key string, value any) {
	b.data[key] = value
}

func (b *Bag) Get(key string) (any, bool) {
	value, ok := b.data[key]
	return value, ok
}

func (b *Bag) Keys() iter.Seq[string] {
	return maps.Keys(b.data)
}

func (b *Bag) Clear() {
	for key := range b.data {
		delete(b.data, key)
	}
}
