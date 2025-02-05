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
func (data *pageData) evalProps() error {
	syncProps := make(map[string]*LazyProp)
	asyncProps := make(map[string]*LazyProp)
	asyncResults := make(map[string]chan asyncPropResult)

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
				syncProps[k] = prop
			} else {
				asyncProps[k] = prop
			}
		}
	}

	wg.Add(len(asyncProps))
	// start evaluating async props first
	for name, prop := range asyncProps {
		ch := make(chan asyncPropResult, 1)
		asyncResults[name] = ch
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
	for name, prop := range syncProps {
		v, err := prop.fn()
		if err != nil {
			return err
		}
		data.Props[name] = v
	}

	wg.Wait()

	for name, result := range asyncResults {
		res := <-result
		if res.err != nil {
			return res.err
		}
		data.Props[name] = res.value
		close(result)
	}

	return nil
}

func (data *pageData) Reset() {
	for k := range data.Props {
		delete(data.Props, k)
	}

	for k := range data.DeferredProps {
		delete(data.DeferredProps, k)
	}

	data.Url = ""
	data.Component = ""
	data.Version = ""
	data.ClearHistory = false
	data.EncryptHistory = false
}
