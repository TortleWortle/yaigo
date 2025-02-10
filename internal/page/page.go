package page

type InertiaPage struct {
	Component      string              `json:"component"`
	Url            string              `json:"url"`
	Props          map[string]any      `json:"props"`
	DeferredProps  map[string][]string `json:"deferredProps"`
	Version        string              `json:"version"`
	EncryptHistory bool                `json:"encryptHistory"`
	ClearHistory   bool                `json:"clearHistory"`

	dirty bool
}

func New() *InertiaPage {
	return &InertiaPage{
		Component:      "",
		Url:            "",
		Version:        "",
		EncryptHistory: false,
		ClearHistory:   false,
	}
}

// ResetIfDirty is called on render,
// this should only ever be called when you try to render a page as a result of a failed previous render.
// for example, if a prop fails to load
func (data *InertiaPage) ResetIfDirty() {
	if data.dirty {
		data.Reset()
	}
	data.dirty = true
}

func (data *InertiaPage) Reset() {
	data.Component = ""
	data.Url = ""
	data.Version = ""
	data.EncryptHistory = false
	data.ClearHistory = false
	data.Props = nil
	data.DeferredProps = nil
	data.dirty = false
}
