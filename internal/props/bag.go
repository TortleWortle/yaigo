package props

import (
	"context"
	"golang.org/x/sync/errgroup"
	"slices"
	"sync"
)

type Bag struct {
	deferredProps map[string][]string
	props         map[string]any

	valueProps []*Prop[any]
	syncProps  []*Prop[*LazyProp]
	asyncProps []*Prop[*LazyProp]

	onlyProps   []string
	exceptProps []string

	dirty        bool
	loadDeferred bool
}

type Prop[T any] struct {
	name  string
	value T

	deferred bool
	dirty    bool
}

func NewBag() *Bag {
	return &Bag{
		// re-usable ish
		deferredProps: make(map[string][]string),
		props:         make(map[string]any),
	}
}

// Checkpoint sets the bag as dirty and will mark any following incoming props as dirty as well
//
// Will remove any dirty
func (b *Bag) Checkpoint() {
	if b.dirty {
		b.rollback()
	}
	b.dirty = true
}

func filterPropSlice[T any](slice []*Prop[T], check func(*Prop[T]) bool) []*Prop[T] {
	var i int
	for _, p := range slice {
		if check(p) {
			slice[i] = p
			i++
		}
	}
	return slice[:i]
}

// Rollback removes any props that are considered dirty, will also clear onlyProps and exceptProps
func (b *Bag) rollback() {
	b.asyncProps = filterPropSlice(b.asyncProps, func(p *Prop[*LazyProp]) bool {
		return !p.dirty
	})

	b.syncProps = filterPropSlice(b.syncProps, func(p *Prop[*LazyProp]) bool {
		return !p.dirty
	})
	b.valueProps = filterPropSlice(b.valueProps, func(p *Prop[any]) bool {
		return !p.dirty
	})

	for k := range b.props {
		delete(b.props, k)
	}

	for k := range b.deferredProps {
		delete(b.deferredProps, k)
	}

	b.onlyProps = nil
	b.exceptProps = nil

	b.loadDeferred = false
	b.dirty = false
}

// Only limits it to only certain props
func (b *Bag) Only(propNames []string) *Bag {
	b.loadDeferred = true // we want to load deferred props when explicitly asking for them
	b.onlyProps = propNames
	return b
}

// Except filters out all other props
func (b *Bag) Except(propNames []string) *Bag {
	b.exceptProps = propNames
	return b
}

// GetProps calculates, evaluates, and returns the props for the current render cycle
//
// Deferred props will only be loaded when explicitly asked for.
func (b *Bag) GetProps(ctx context.Context) (map[string]any, error) {
	b.filterProps()
	var lock sync.Mutex

	g, ctx := errgroup.WithContext(ctx)

	// copy value props over
	for _, prop := range b.valueProps {
		if b.includeProp(prop.name) {
			b.props[prop.name] = prop.value
		}
	}

	for _, p := range b.asyncProps {
		g.Go(func() error {
			val, err := p.value.fn(ctx)
			if err != nil {
				return err
			}
			lock.Lock()
			b.props[p.name] = val
			lock.Unlock()
			return nil
		})
	}

	lock.Lock()
	for _, p := range b.syncProps {
		val, err := p.value.fn(ctx)
		if err != nil {
			return nil, err
		}
		b.props[p.name] = val
	}
	lock.Unlock()

	// wait for async props
	err := g.Wait()
	if err != nil {
		return nil, err
	}

	return b.props, nil
}

// GetDeferredProps returns the props deferred after a GetProps call
func (b *Bag) GetDeferredProps() map[string][]string {
	return b.deferredProps
}

func (b *Bag) Set(key string, value any) error {
	switch p := value.(type) {
	case *LazyProp:
		prop := &Prop[*LazyProp]{
			name:     key,
			value:    p,
			deferred: p.deferred,
			dirty:    b.dirty,
		}

		if p.sync {
			b.syncProps = append(b.syncProps, prop)
		} else {
			b.asyncProps = append(b.asyncProps, prop)
		}
	default:
		prop := &Prop[any]{
			name:     key,
			value:    value,
			deferred: false,
			dirty:    b.dirty,
		}

		b.valueProps = append(b.valueProps, prop)
	}
	return nil
}

func (b *Bag) Clear() {
	for key := range b.props {
		delete(b.props, key)
	}

	b.valueProps = nil

	for k := range b.deferredProps {
		delete(b.deferredProps, k)
	}

	b.asyncProps = nil
	b.syncProps = nil

	b.loadDeferred = false
	b.dirty = false
	b.onlyProps = nil
	b.exceptProps = nil
}

// filterProps throws out any props that are not meant to be loaded
// while keeping track of them in a map for inertia to use
func (b *Bag) filterProps() {
	b.asyncProps = filterPropSlice(b.asyncProps, func(p *Prop[*LazyProp]) bool {
		// skip deferred if we don't want deferred
		if p.deferred && !b.loadDeferred {
			b.deferredProps[p.value.group] = append(b.deferredProps[p.value.group], p.name)
			return false
		}

		if !b.includeProp(p.name) {
			return false
		}

		return true
	})

	b.syncProps = filterPropSlice(b.syncProps, func(p *Prop[*LazyProp]) bool {
		// skip deferred if we don't want deferred
		if p.deferred && !b.loadDeferred {
			b.deferredProps[p.value.group] = append(b.deferredProps[p.value.group], p.name)
			return false
		}

		if !b.includeProp(p.name) {
			return false
		}

		return true
	})
}

func (b *Bag) includeProp(name string) bool {
	if slices.Contains(b.exceptProps, name) {
		return false
	}

	if len(b.onlyProps) > 0 && !slices.Contains(b.onlyProps, name) {
		return false
	}

	return true
}
