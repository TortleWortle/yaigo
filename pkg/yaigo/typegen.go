package yaigo

import (
	"fmt"
	"github.com/tortlewortle/yaigo/internal/page"
	"github.com/tortlewortle/yaigo/pkg/typegen"
	"log/slog"
	"reflect"
	"slices"
	"sync"
)

type typeGenerator struct {
	dirPath        string
	lock           *sync.Mutex
	propCache      map[string]Props
	optionalsCache map[string][]string
}

func (g *typeGenerator) Generate(inertiaPage *page.InertiaPage) {
	// little dirty but we don't really care about these errors beyond logging
	var err error
	defer func() {
		if err != nil {
			slog.Warn("typegen error", slog.Any("error", err))
		}
	}()
	component := inertiaPage.Component
	props := inertiaPage.Props

	g.lock.Lock()
	defer g.lock.Unlock()

	propsForGen := Props{}
	forcedOptionals := g.optionalsCache[component]

	var updated bool
	existingProps, ok := g.propCache[component]
	if ok {
		// if cache exists
		for k, v := range existingProps {
			propsForGen[k] = v
		}

		for k, v := range props {
			existingProp, ok := propsForGen[k]
			if ok {
				// smoke test if the prop types are all still the same
				// (so two pages using the same component won't provide different types
				et := reflect.TypeOf(existingProp)
				vt := reflect.TypeOf(v)
				if et.Kind() != vt.Kind() {
					panic(fmt.Sprintf("prop %s for %s has conflicting types: %v != %v", k, inertiaPage.Component, et.Name(), vt.Name()))
				}
			} else {
				// prop is new, probably deferred, lets mark it forced optional
				forcedOptionals = append(forcedOptionals, k)
				updated = true
			}
			propsForGen[k] = v
		}
	} else {
		updated = true
		for k, v := range props {
			propsForGen[k] = v
		}
	}

	if !updated {
		// nothing new, let's skip!
		return
	}

	g.optionalsCache[component] = forcedOptionals
	g.propCache[component] = propsForGen

	cName, err := typegen.FormatComponentName(component)
	if err != nil {
		err = fmt.Errorf("formatting comp name: %w", err)
		return
	}

	root, err := typegen.ParseMap(typegen.Ident(fmt.Sprintf("%sProps", cName)), propsForGen)
	if err != nil {
		err = fmt.Errorf("parsing propmap: %w", err)
		return
	}

	root.Name = inertiaPage.Component
	root.PkgPath = "InertiaRender"

	var newProps []typegen.TsType
	for _, v := range root.Properties {
		if slices.Contains(forcedOptionals, v.PropertyName) {
			v.Optional = true
		}
		newProps = append(newProps, v)
	}

	root.Properties = newProps

	err = typegen.GenerateTypeScriptForComponent(g.dirPath, root)
	if err != nil {
		err = fmt.Errorf("generating typescript: %w", err)
		return
	}
}
