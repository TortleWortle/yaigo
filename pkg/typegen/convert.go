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
)

type TsType struct {
	Type     string
	Optional bool
	Ident    string
	Key      string
	Children []TsType
}

func (t *TsType) MapKey() TsType {
	if len(t.Children) != 2 || t.Type != TypeDict {
		panic("Key can only be called on TypeDict")
	}

	return t.Children[0]
}

func (t *TsType) Elem() TsType {
	if len(t.Children) == 2 && t.Type == TypeDict {
		return t.Children[1]
	}

	if len(t.Children) == 1 && t.Type == TypeArray {
		return t.Children[0]
	}

	panic("Elem can only be called on TypeArray or TypeDicts (children mismatch)")
}

var titleCaser = cases.Title(language.English)

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

func getBasicTsType(v reflect.Type) string {
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
			Type:  TypeAny,
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
			Type:     TypeDict,
			Ident:    TypeDict,
			Children: []TsType{keyType, elemType},
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
			Type:     TypeArray,
			Ident:    TypeArray,
			Children: []TsType{et},
		}, nil
	case reflect.Struct:
		children, err := ConvertStruct(v)
		if err != nil {
			return t, err
		}
		return TsType{
			Type:     TypeStruct,
			Optional: false,
			Ident:    fmt.Sprintf("%s%s", titleCaser.String(filepath.Base(v.PkgPath())), v.Name()),
			Children: children,
		}, nil
	default:
		baseType := getBasicTsType(v)
		return TsType{
			Type:     baseType,
			Ident:    baseType,
			Optional: false,
		}, nil
	}
}

func NewRootType(name string, properties []TsType) (TsType, error) {
	typeName, err := FormatComponentName(name)
	if err != nil {
		return TsType{}, fmt.Errorf("formatting name: %w", err)
	}

	return TsType{
		Type:     TypeStruct,
		Optional: false,
		Ident:    typeName,
		Key:      "",
		Children: properties,
	}, nil
}

func getTypeFromValue(key string, v any) (TsType, error) {
	t := reflect.TypeOf(v)
	if v == nil {
		return TsType{
			Type:     TypeNull,
			Optional: false,
			Key:      key,
		}, nil
	}
	tst, err := getTsType(t)
	if err != nil {
		return TsType{}, fmt.Errorf("gettype: %w", err)
	}
	tst.Key = key
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
		fieldType.Key = key
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

		if t.Type == TypeInvalid {
			// invalid type found, skip for now
			t.Type = "never"
		}
		types = append(types, t)
	}

	sort.Slice(types, func(i, j int) bool {
		return strings.Map(unicode.ToUpper, types[i].Key) < strings.Map(unicode.ToUpper, types[j].Key)
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
	componentName.WriteString("Props")

	return componentName.String(), nil
}

type identCache = map[string][]TsType

func getTypeDefs(cache identCache, types []TsType) {
	for _, v := range types {
		if v.Type == TypeStruct {
			if _, ok := cache[v.Ident]; !ok {
				cache[v.Ident] = v.Children
				getTypeDefs(cache, v.Children)
			}
		}

		if v.Type == TypeArray {
			cv := v.Elem()
			if cv.Type == TypeStruct {
				if _, ok := cache[cv.Ident]; !ok {
					cache[cv.Ident] = cv.Children
					getTypeDefs(cache, cv.Children)
				}
			}
		}

		if v.Type == TypeDict {
			cv := v.Elem()
			if cv.Type == TypeStruct {
				if _, ok := cache[cv.Ident]; !ok {
					cache[cv.Ident] = cv.Children
					getTypeDefs(cache, cv.Children)
				}
			}
		}
	}
}
