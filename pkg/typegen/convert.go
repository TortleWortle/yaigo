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
	Struct
	Map
	Array
)

type Ident string

const (
	TypeString  = "string"
	TypeNumber  = "number"
	TypeArray   = "array"
	TypeNull    = "null"
	TypeInvalid = "invalid"
	TypeStruct  = "struct"
	TypeDict    = "dictionary"
	TypeAny     = "any"
)

func (i Ident) String() string {
	return string(i)
}

type TsType struct {
	Kind     string
	Optional bool

	// Identifier,
	Ident Ident // same as Ident?

	// Name of the property it's assigned to
	Name       string
	Properties []TsType
}

func (t *TsType) MapKey() TsType {
	if len(t.Properties) != 2 || t.Kind != TypeDict {
		panic("Name can only be called on TypeDict")
	}

	return t.Properties[0]
}

func (t *TsType) Elem() TsType {
	if len(t.Properties) == 2 && t.Kind == TypeDict {
		return t.Properties[1]
	}

	if len(t.Properties) == 1 && t.Kind == TypeArray {
		return t.Properties[0]
	}

	panic("Elem can only be called on TypeArray or TypeDicts (children mismatch)")
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
		return TsType{
			Kind:  TypeAny,
			Ident: TypeAny,
		}, nil
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
		return TsType{
			Kind:       TypeDict,
			Ident:      TypeDict,
			Properties: []TsType{keyType, elemType},
		}, nil
	case reflect.Pointer:
		pt := v.Elem()
		t, err := getTsType(pt)
		if err != nil {
			return t, fmt.Errorf("converting pointer type: %w", err)
		}
		t.Optional = true
		return t, nil
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		pt := v.Elem()
		et, err := getTsType(pt)
		if err != nil {
			return t, fmt.Errorf("converting slice type: %w", err)
		}
		return TsType{
			Kind:       TypeArray,
			Ident:      TypeArray,
			Properties: []TsType{et},
		}, nil
	case reflect.Struct:
		children, err := ConvertStruct(v)
		if err != nil {
			return t, err
		}
		return TsType{
			Kind:       TypeStruct,
			Optional:   false,
			Ident:      Ident(fmt.Sprintf("%s%s", titleCaser.String(filepath.Base(v.PkgPath())), v.Name())),
			Properties: children,
		}, nil
	default:
		baseType := getBasicTsType(v)
		return TsType{
			Kind:     baseType.String(),
			Ident:    baseType,
			Optional: false,
		}, nil
	}
}

func NewRootType(name Ident, properties []TsType) TsType {
	return TsType{
		Kind:       TypeStruct,
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
			Kind:     TypeNull,
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

func ConvertStruct(v reflect.Type) ([]TsType, error) {
	if v.Kind() != reflect.Struct {
		return nil, errors.New("value has to be a struct")
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
			return nil, fmt.Errorf("getting field type %s: %w", key, err)
		}
		fieldType.Optional = fieldType.Optional || jsonOptional
		types = append(types, fieldType)
	}

	return types, nil
}

func ConvertMap(props map[string]any) ([]TsType, error) {
	var types []TsType
	for k, v := range props {
		t, err := getTypeFromValue(k, v)
		if err != nil {
			return nil, fmt.Errorf("converting %s: %w", k, err)
		}

		if t.Kind == TypeInvalid {
			// invalid type found, skip for now
			t.Kind = "never"
		}
		types = append(types, t)
	}

	sort.Slice(types, func(i, j int) bool {
		return strings.Map(unicode.ToUpper, types[i].Name) < strings.Map(unicode.ToUpper, types[j].Name)
	})
	return types, nil
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
		if v.Kind == TypeStruct {
			if _, ok := cache[v.Ident]; !ok {
				cache[v.Ident] = v.Properties
				getTypeDefs(cache, v.Properties)
			}
		}

		if v.Kind == TypeArray {
			cv := v.Elem()
			if cv.Kind == TypeStruct {
				if _, ok := cache[cv.Ident]; !ok {
					cache[cv.Ident] = cv.Properties
					getTypeDefs(cache, cv.Properties)
				}
			}
		}

		if v.Kind == TypeDict {
			cv := v.Elem()
			if cv.Kind == TypeStruct {
				if _, ok := cache[cv.Ident]; !ok {
					cache[cv.Ident] = cv.Properties
					getTypeDefs(cache, cv.Properties)
				}
			}
		}
	}
}
