package inertia

import "github.com/tortlewortle/yaigo/pkg/yaigo"

func Page(component string, props Props) *yaigo.Page {
	return yaigo.NewPage(component, props)
}
