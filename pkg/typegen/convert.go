package typegen

import (
	"errors"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"unicode"
)

type Kind uint

const (
	Primitive = iota
	Object
	Map
	Array
	Any
	Null
	Invalid
)

type Ident string

const (
	TypeString  = "string"
	TypeNumber  = "number"
	TypeNull    = "null"
	TypeInvalid = "never"
	TypeAny     = "any"
)

func (i Ident) String() string {
	return string(i)
}

// TsType describes a Go type as what it would be in a TypeScript type
type TsType struct {
	Kind       Kind
	Optional   bool
	Ident      Ident
	Name       string
	Properties []TsType
}

func (t *TsType) MapKey() TsType {
	if len(t.Properties) != 2 || t.Kind != Map {
		panic("Name can only be called on Map")
	}

	return t.Properties[0]
}

func (t *TsType) Elem() TsType {
	if len(t.Properties) == 2 && t.Kind == Map {
		return t.Properties[1]
	}

	if len(t.Properties) == 1 && t.Kind == Array {
		return t.Properties[0]
	}

	panic("Elem can only be called on Array or Map (children mismatch)")
}

var titleCaser = cases.Title(language.English)

func getBasicTsType(v reflect.Type) Ident {
	if v.ConvertibleTo(reflect.TypeFor[int]()) {
		return TypeNumber
	}
	if v.ConvertibleTo(reflect.TypeFor[string]()) {
		return TypeString
	}
	return TypeInvalid
}

func getTsType(v reflect.Type) (t TsType, err error) {
	switch v.Kind() {
	case reflect.Interface:
		t.Kind = Any
		t.Ident = TypeAny
	case reflect.Map:
		if !v.Key().ConvertibleTo(reflect.TypeFor[string]()) {
			return t, errors.New("map key must be a string-able")
		}
		keyType, err := getTsType(v.Key())
		if err != nil {
			return t, fmt.Errorf("getting key type: %w", err)
		}
		elemType, err := getTsType(v.Elem())
		if err != nil {
			return t, fmt.Errorf("getting elem type: %w", err)
		}
		t.Kind = Map
		t.Properties = []TsType{keyType, elemType}
	case reflect.Pointer:
		pe := v.Elem()
		pt, err := getTsType(pe)
		if err != nil {
			return t, fmt.Errorf("converting pointer type: %w", err)
		}
		pt.Optional = true
		t = pt
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		pt := v.Elem()
		et, err := getTsType(pt)
		if err != nil {
			return t, fmt.Errorf("converting slice type: %w", err)
		}

		t.Kind = Array
		t.Properties = []TsType{et}
	case reflect.Struct:
		t, err = ParseStruct(v)
		if err != nil {
			return t, err
		}
	default:
		baseType := getBasicTsType(v)
		t.Kind = Primitive
		t.Ident = baseType
		t.Optional = false
	}
	return t, nil
}

func NewType(kind Kind, name Ident, properties []TsType) TsType {
	return TsType{
		Kind:       kind,
		Optional:   false,
		Ident:      name,
		Name:       "",
		Properties: properties,
	}
}

func getTypeFromValue(key string, v any) (TsType, error) {
	t := reflect.TypeOf(v)
	if v == nil {
		return TsType{
			Kind:     Primitive,
			Ident:    TypeNull,
			Optional: false,
			Name:     key,
		}, nil
	}
	tst, err := getTsType(t)
	if err != nil {
		return TsType{}, fmt.Errorf("gettype: %w", err)
	}
	tst.Name = key
	return tst, nil
}

func ParseStruct(v reflect.Type) (TsType, error) {
	if v.Kind() != reflect.Struct {
		return TsType{
			Kind: Invalid,
		}, errors.New("value has to be a struct")
	}

	var types []TsType

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.IsExported() {
			continue
		}

		key := f.Name
		jsonTag := f.Tag.Get("json")
		jsonName, jsonOpts, _ := strings.Cut(jsonTag, ",")
		if jsonName != "" {
			key = jsonName
		}

		jsonOptional := slices.Contains(strings.Split(jsonOpts, ","), "omitempty")
		fieldType, err := getTsType(f.Type)
		fieldType.Name = key
		if err != nil {
			return TsType{
				Kind: Invalid,
			}, fmt.Errorf("getting field type %s: %w", key, err)
		}
		fieldType.Optional = fieldType.Optional || jsonOptional
		types = append(types, fieldType)
	}

	root := NewType(Object, Ident(fmt.Sprintf("%s%s", titleCaser.String(filepath.Base(v.PkgPath())), v.Name())), types)

	return root, nil
}

func ParseMap(ident Ident, props map[string]any) (TsType, error) {
	var types []TsType
	for k, v := range props {
		t, err := getTypeFromValue(k, v)
		if err != nil {
			return TsType{Kind: Invalid}, fmt.Errorf("converting %s: %w", k, err)
		}

		types = append(types, t)
	}

	sort.Slice(types, func(i, j int) bool {
		return strings.Map(unicode.ToUpper, types[i].Name) < strings.Map(unicode.ToUpper, types[j].Name)
	})

	root := NewType(Object, ident, types)

	return root, nil
}

func FormatComponentName(component string) (string, error) {
	if component == "" {
		return "", errors.New("component name cannot be empty")
	}

	var componentName strings.Builder
	for _, part := range strings.Split(component, "/") {
		componentName.WriteString(titleCaser.String(part))
	}

	return componentName.String(), nil
}

type identCache = map[Ident][]TsType

func getTypeDefs(cache identCache, types []TsType) {
	for _, v := range types {
		if v.Kind == Object {
			if _, ok := cache[v.Ident]; !ok {
				cache[v.Ident] = v.Properties
				getTypeDefs(cache, v.Properties)
			}
		}

		if v.Kind == Array {
			cv := v.Elem()
			if cv.Kind == Object {
				if _, ok := cache[cv.Ident]; !ok {
					cache[cv.Ident] = cv.Properties
					getTypeDefs(cache, cv.Properties)
				}
			}
		}

		if v.Kind == Map {
			cv := v.Elem()
			if cv.Kind == Object {
				if _, ok := cache[cv.Ident]; !ok {
					cache[cv.Ident] = cv.Properties
					getTypeDefs(cache, cv.Properties)
				}
			}
		}
	}
}
