package inertia

import (
	"errors"
	"sync"
)

type pageData struct {
	Component      string              `json:"component"`
	Url            string              `json:"url"`
	Props          Props               `json:"props"`
	Version        string              `json:"version"`
	EncryptHistory bool                `json:"encryptHistory"`
	ClearHistory   bool                `json:"clearHistory"`
	DeferredProps  map[string][]string `json:"deferredProps"`

	syncProps    map[string]*LazyProp
	asyncProps   map[string]*LazyProp
	asyncResults map[string]chan asyncPropResult // replace with syncmap?
	dirty        bool
}

func newPageData() *pageData {
	return &pageData{
		Component:      "",
		Url:            "",
		Props:          make(Props),
		Version:        "",
		EncryptHistory: false,
		ClearHistory:   false,
		DeferredProps:  make(map[string][]string),
		syncProps:      make(map[string]*LazyProp),
		asyncProps:     make(map[string]*LazyProp),
		asyncResults:   make(map[string]chan asyncPropResult),
		dirty:          false,
	}
}

func (data *pageData) moveDeferredProps() error {
	for name, value := range data.Props {
		switch value.(type) {
		case *LazyProp:
			prop, ok := value.(*LazyProp)
			if !ok {
				return errors.New("could not cast prop value to LazyProp")
			}
			if prop.deferred {
				data.DeferredProps[prop.group] = append(data.DeferredProps[prop.group], name)
				delete(data.Props, name)
			}
		}
	}
	return nil
}

type asyncPropResult struct {
	value any
	err   error
}

// todo: store these in pageData so it gets reset and pooled
func (data *pageData) evalLazyProps() error {

	var wg sync.WaitGroup

	// evaluate deferred props and set the values
	for k, v := range data.Props {
		switch v.(type) {
		case *LazyProp:
			prop, ok := v.(*LazyProp)
			if !ok {
				return errors.New("could not cast prop value to LazyProp")
			}
			if prop.sync {
				data.syncProps[k] = prop
			} else {
				data.asyncProps[k] = prop
			}
		}
	}

	wg.Add(len(data.asyncProps))
	// start evaluating async props first
	for name, prop := range data.asyncProps {
		ch := make(chan asyncPropResult, 1)
		data.asyncResults[name] = ch
		go func(chan asyncPropResult) {
			v, err := prop.fn()
			ch <- asyncPropResult{
				value: v,
				err:   err,
			}
			wg.Done()
		}(ch)
	}

	// evaluate sync props
	for name, prop := range data.syncProps {
		v, err := prop.fn()
		if err != nil {
			return err
		}
		data.Props[name] = v
	}

	wg.Wait()

	for name, result := range data.asyncResults {
		res := <-result
		if res.err != nil {
			return res.err
		}
		data.Props[name] = res.value
		close(result)
	}

	return nil
}

// resetIfDirty is called on render,
// this should only ever be called when you try to render a page as a result of a failed previous render.
// for example, if a prop fails to load
func (data *pageData) resetIfDirty() {
	if data.dirty {
		data.Reset()
	}
	data.dirty = true
}

func (data *pageData) Reset() {
	data.Component = ""
	data.Url = ""
	data.Version = ""
	data.EncryptHistory = false
	data.ClearHistory = false
	data.dirty = false
	data.resetProps()
}

func (data *pageData) resetProps() {
	for k := range data.Props {
		delete(data.Props, k)
	}

	for k := range data.DeferredProps {
		delete(data.DeferredProps, k)
	}

	for k := range data.asyncProps {
		delete(data.asyncProps, k)
	}

	for k := range data.syncProps {
		delete(data.syncProps, k)
	}

	for k := range data.asyncResults {
		delete(data.asyncResults, k)
	}
}
