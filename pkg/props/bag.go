package props

import (
	"iter"
)

type Bag interface {
	Set(string, any)
	Get(string) (any, bool)
	Keys() iter.Seq[string]
	Clear()
	Items() map[string]any
}

func NewBag() Bag {
	return &mapBag{
		data: make(map[string]any),
	}
}
