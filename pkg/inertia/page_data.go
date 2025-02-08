package inertia

type pageData struct {
	Component      string              `json:"component"`
	Url            string              `json:"url"`
	Props          map[string]any      `json:"props"`
	Version        string              `json:"version"`
	EncryptHistory bool                `json:"encryptHistory"`
	ClearHistory   bool                `json:"clearHistory"`
	DeferredProps  map[string][]string `json:"deferredProps"`

	dirty bool
}

func newPageData() *pageData {
	return &pageData{
		Component:      "",
		Url:            "",
		Version:        "",
		EncryptHistory: false,
		ClearHistory:   false,
	}
}

//func (data *pageData) evalLazyProps() error {
//// evaluate deferred props and set the values
//for k, v := range data.Props {
//switch v.(type) {
//case *LazyProp:
//prop, ok := v.(*LazyProp)
//if !ok {
//return errors.New("could not cast prop value to LazyProp")
//}
//if prop.sync {
//data.syncProps[k] = prop
//} else {
//data.asyncProps[k] = prop
//}
//}
//}
//
//data.wg.Add(len(data.asyncProps))
//// start evaluating async props first
//for name, prop := range data.asyncProps {
//ch := make(chan asyncPropResult, 1)
//data.asyncResults[name] = ch
//go func(chan asyncPropResult) {
//v, err := prop.fn()
//ch <- asyncPropResult{
//value: v,
//err:   err,
//}
//data.wg.Done()
//}(ch)
//}
//
//// evaluate sync props
//for name, prop := range data.syncProps {
//v, err := prop.fn()
//if err != nil {
//return err
//}
//data.Props[name] = v
//}
//data.wg.Wait()
//
//for name, result := range data.asyncResults {
//res := <-result
//if res.err != nil {
//return res.err
//}
//data.Props[name] = res.value
//close(result)
//}
//
//return nil
//}

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
	data.Props = nil
}
