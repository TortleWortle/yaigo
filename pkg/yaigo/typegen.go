package yaigo

import (
	"fmt"
	"github.com/tortlewortle/yaigo/internal/page"
	"github.com/tortlewortle/yaigo/pkg/typegen"
	"log/slog"
	"slices"
	"sync"
)

type typeGenerator struct {
	dirPath        string
	lock           *sync.Mutex
	propCache      map[string]Props
	optionalsCache map[string][]string
}

func (g *typeGenerator) Generate(inertiaPage *page.InertiaPage) (err error) {
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

	existingProps, ok := g.propCache[component]
	if ok {
		// if cache exists
		for k, v := range existingProps {
			propsForGen[k] = v
		}

		for k, v := range props {
			_, ok := propsForGen[k]
			if !ok {
				// prop is new, probably deferred, lets mark it forced optional
				forcedOptionals = append(forcedOptionals, k)
			}
			propsForGen[k] = v
		}
	} else {
		for k, v := range props {
			propsForGen[k] = v
		}
	}

	g.optionalsCache[component] = forcedOptionals
	g.propCache[component] = propsForGen

	cName, err := typegen.FormatComponentName(component)
	if err != nil {
		return fmt.Errorf("formatting comp name: %w", err)
	}

	root, err := typegen.ParseMap(typegen.Ident(cName), propsForGen)
	if err != nil {
		return fmt.Errorf("parsing propmap: %w", err)
	}

	var newProps []typegen.TsType
	for _, v := range root.Properties {
		if slices.Contains(forcedOptionals, v.Name) {
			v.Optional = true
		}
		newProps = append(newProps, v)
	}

	root.Properties = newProps

	err = typegen.GenerateTypeScriptForComponent(g.dirPath, root)
	if err != nil {
		return fmt.Errorf("generating typescript: %w", err)
	}

	return nil
}
