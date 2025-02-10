package props

import (
	"errors"
	"slices"
	"sync"
)

type asyncPropResult struct {
	value any
	err   error
}

type Bag struct {
	deferredProps map[string][]string
	props         map[string]any

	valueProps []Prop[any]
	syncProps  []Prop[*LazyProp]
	asyncProps []Prop[*LazyProp]

	onlyProps   []string
	exceptProps []string

	wg           *sync.WaitGroup
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

		wg: &sync.WaitGroup{},
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

func filterPropSlice[T any](slice []Prop[T], check func(Prop[T]) bool) []Prop[T] {
	var i int
	for _, p := range slice {
		if check(p) {
			slice[i] = p
			i++
		}
	}
	return slice[:i]
}

// Rollback removes any props that are considered dirty
func (b *Bag) rollback() {
	b.asyncProps = filterPropSlice(b.asyncProps, func(p Prop[*LazyProp]) bool {
		return !p.dirty
	})

	b.syncProps = filterPropSlice(b.syncProps, func(p Prop[*LazyProp]) bool {
		return !p.dirty
	})
	b.valueProps = filterPropSlice(b.valueProps, func(p Prop[any]) bool {
		return !p.dirty
	})

	for k := range b.props {
		delete(b.props, k)
	}

	for k := range b.deferredProps {
		delete(b.deferredProps, k)
	}

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
func (b *Bag) GetProps() (map[string]any, error) {
	b.filterProps()

	// idea: syncmap for results
	// pros: less loops, simpler execution of props
	// cons: no error handling if a prop handler fails unless we keep track of errors separately.
	//
	// use sync errgroup
	// pros: track errors
	// cons: need to accept context inside of the callback functions now

	// copy value props over
	for _, prop := range b.valueProps {
		if b.includeProp(prop.name) {
			b.props[prop.name] = prop.value
		}
	}

	for _, p := range b.asyncProps {
		b.wg.Add(1)
		go func() {
			p.value.Execute()
			b.wg.Done()
		}()
	}

	for _, p := range b.syncProps {
		p.value.Execute()
		if p.value.err != nil {
			return b.props, p.value.err
		}
		b.props[p.name] = p.value.result
	}

	// wait for async props
	b.wg.Wait()

	for _, p := range b.asyncProps {
		if p.value.err != nil {
			return b.props, p.value.err
		}
		b.props[p.name] = p.value.result
	}

	return b.props, nil
}

// GetDeferredProps returns the props that were deferred after a GetProps call
func (b *Bag) GetDeferredProps() map[string][]string {
	return b.deferredProps
}

func (b *Bag) Set(key string, value any) error {
	switch value.(type) {
	case *LazyProp:
		p, ok := value.(*LazyProp)
		if !ok {
			return errors.New("could not cast prop as LazyProp")
		}

		if p.sync {
			b.syncProps = append(b.syncProps, Prop[*LazyProp]{
				name:     key,
				value:    p,
				dirty:    b.dirty,
				deferred: p.deferred,
			})
		} else {
			b.asyncProps = append(b.asyncProps, Prop[*LazyProp]{
				name:     key,
				value:    p,
				dirty:    b.dirty,
				deferred: p.deferred,
			})
		}
	default:
		b.valueProps = append(b.valueProps, Prop[any]{
			name:     key,
			value:    value,
			dirty:    b.dirty,
			deferred: false,
		})
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
	b.asyncProps = filterPropSlice(b.asyncProps, func(p Prop[*LazyProp]) bool {
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

	b.syncProps = filterPropSlice(b.syncProps, func(p Prop[*LazyProp]) bool {
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
