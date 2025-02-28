package page

type InertiaPage struct {
	Component      string              `json:"component"`
	Url            string              `json:"url"`
	Props          map[string]any      `json:"props"`
	DeferredProps  map[string][]string `json:"deferredProps"`
	Version        string              `json:"version"`
	EncryptHistory bool                `json:"encryptHistory"`
	ClearHistory   bool                `json:"clearHistory"`
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

func (data *InertiaPage) Reset() {
	data.Component = ""
	data.Url = ""
	data.Version = ""
	data.EncryptHistory = false
	data.ClearHistory = false
	data.Props = nil
	data.DeferredProps = nil
}
