package typegen

import (
	"encoding"
	"encoding/json"
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

var jsonMarshalers = make(map[Ident]TsType)

// RegisterJsonMarshaler Marshals the struct and converts it into a map[string] to generate type info
func RegisterJsonMarshaler(val json.Marshaler) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()

	typ := reflect.TypeOf(val)
	data, err := json.Marshal(val)
	if err != nil {
		return
	}
	values := make(map[string]any)

	err = json.Unmarshal(data, &values)
	if err != nil {
		return
	}

	ident := makeIdent(typ)

	t, err := ParseMap(ident, values)
	if err != nil {
		return
	}

	_, ok := jsonMarshalers[ident]
	if ok {
		panic(fmt.Sprintf("%s already registered as JsonMarshaler", ident))
	}

	jsonMarshalers[ident] = t
}

type Kind uint

const (
	Invalid = iota
	Primitive
	Object
	InlineObject
	Map
	Array
	Null
)

type Ident string

const (
	TypeString  = "string"
	TypeNumber  = "number"
	TypeBool    = "boolean"
	TypeNull    = "null"
	TypeInvalid = "never"
	TypeAny     = "any"
)

func (i Ident) String() string {
	return string(i)
}

// TsType describes a Go type as what it would be in a TypeScript type
type TsType struct {
	Kind         Kind
	Optional     bool
	Ident        Ident
	PropertyName string
	Properties   []TsType

	PkgPath string
	Name    string

	Comment string // Optional comment to list next to the file

	export bool
}

func (t *TsType) Export(export bool) {
	if t.Kind != Object {
		panic("can only export Object")
	}
	t.export = export
}

func (t *TsType) MapKey() TsType {
	if len(t.Properties) != 2 || t.Kind != Map {
		panic("PropertyName can only be called on Map")
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
	if v.ConvertibleTo(reflect.TypeFor[bool]()) {
		return TypeBool
	}

	return TypeInvalid
}

func getDependencies(cm map[Ident]struct{}, t TsType) (deps []TsType) {
	for _, ct := range t.Properties {
		if ct.Kind == Object {
			if _, ok := cm[ct.Ident]; !ok {
				deps = append(deps, ct)
				cm[ct.Ident] = struct{}{}
			}
		} else if ct.Kind == InlineObject {
			deps = append(deps, getDependencies(cm, ct)...)
		}
	}

	return
}

func GetDependencies(t TsType) (deps []TsType) {
	cm := make(map[Ident]struct{})
	return getDependencies(cm, t)
}

func getTsType(t reflect.Type) (out TsType, err error) {
	if t == nil {
		return TsType{
			Kind:         Primitive,
			Ident:        TypeInvalid,
			Optional:     false,
			PropertyName: "this_is_a_bug",
		}, nil
	}

	switch t.Kind() {
	case reflect.Interface:
		out.Kind = Primitive
		ok := t.Implements(reflect.TypeFor[fmt.Stringer]())
		if ok {
			out.Ident = TypeString
		} else {
			out.Ident = TypeAny
		}
	case reflect.Map:
		if !t.Key().ConvertibleTo(reflect.TypeFor[string]()) {
			return out, errors.New("map key must be a string-able")
		}
		keyType, err := getTsType(t.Key())
		if err != nil {
			return out, fmt.Errorf("getting key type: %w", err)
		}
		elemType, err := getTsType(t.Elem())
		if err != nil {
			return out, fmt.Errorf("getting elem type: %w", err)
		}
		out.Optional = true
		out.Kind = Map
		out.Properties = []TsType{keyType, elemType}
	case reflect.Pointer:
		pe := t.Elem()
		pt, err := getTsType(pe)
		if err != nil {
			return out, fmt.Errorf("converting pointer type: %w", err)
		}
		pt.Optional = true
		out = pt
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		ok := t.Implements(reflect.TypeFor[encoding.TextMarshaler]())
		if ok {
			out.Kind = Primitive
			out.Ident = TypeString
			out.Comment = "implements encoding.TextMarshaler."
			break
		}
		pt := t.Elem()
		et, err := getTsType(pt)
		if err != nil {
			return out, fmt.Errorf("converting slice type: %w", err)
		}

		out.Kind = Array
		out.Properties = []TsType{et}
	case reflect.Struct:
		ok := t.Implements(reflect.TypeFor[encoding.TextMarshaler]())
		if ok {
			out.Kind = Primitive
			out.Ident = TypeString
			out.Comment = "implements encoding.TextMarshaler."
			break
		}
		ok = t.Implements(reflect.TypeFor[json.Marshaler]())
		if ok {
			goodType, ok := jsonMarshalers[makeIdent(t)]
			if ok {
				out = goodType
			} else {
				out.Kind = Primitive
				out.Ident = TypeAny
				out.Comment = "implements json.Marshaler, please register using typegen.RegisterJsonMarshaler()."
			}
			break
		}
		out, err = ParseStruct(t)
		if err != nil {
			return out, err
		}
	default:
		baseType := getBasicTsType(t)
		out.Kind = Primitive
		out.Ident = baseType
		out.Optional = false
	}

	out.PkgPath = t.PkgPath()
	out.Name = t.Name()
	sortTypes(out.Properties)
	return out, nil
}

func NewType(kind Kind, name Ident, properties []TsType) TsType {
	return TsType{
		Kind:         kind,
		Optional:     false,
		Ident:        name,
		PropertyName: "",
		Properties:   properties,
	}
}

func getTypeFromValue(key string, v any) (TsType, error) {
	t := reflect.TypeOf(v)
	tst, err := getTsType(t)
	if err != nil {
		return TsType{}, fmt.Errorf("gettype: %w", err)
	}
	tst.PropertyName = key
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
		fieldType.PropertyName = key
		if err != nil {
			return TsType{
				Kind: Invalid,
			}, fmt.Errorf("getting field type %s: %w", key, err)
		}
		fieldType.Optional = fieldType.Optional || jsonOptional
		types = append(types, fieldType)
	}

	var t Kind = Object
	if v.Name() == "" {
		t = InlineObject
	}

	root := NewType(t, makeIdent(v), types)
	if strings.Contains(v.Name(), "[") {
		name, _, _ := strings.Cut(strings.Split(v.Name(), "[")[1], "]")
		parts := strings.Split(name, ".")
		if len(parts) > 1 {
			pkgPath := strings.Join(parts[:len(parts)-1], ".")
			name = parts[len(parts)-1]
			name = fmt.Sprintf("%s%s", titleCaser.String(filepath.Base(pkgPath)), name)
		}
	}

	return root, nil
}

func makeIdent(t reflect.Type) Ident {
	name := t.Name()
	namePreGeneric, genericString, hasGeneric := strings.Cut(t.Name(), "[")
	if hasGeneric {
		name = namePreGeneric
		rawGenericNames := strings.Split(genericString[:len(genericString)-1], ",")

		var genericNames []string
		for _, n := range rawGenericNames {
			parts := strings.Split(n, ".")
			var genericName string
			if len(parts) > 1 {
				genericName = parts[len(parts)-1]
			} else {
				genericName = parts[0]
			}
			genericNames = append(genericNames, typeStrToTs(genericName))
		}

		name += strings.ReplaceAll(strings.Join(genericNames, ""), "*", "Opt")

	}

	return Ident(fmt.Sprintf("%s%s", titleCaser.String(filepath.Base(t.PkgPath())), name))
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
		return strings.Map(unicode.ToUpper, types[i].PropertyName) < strings.Map(unicode.ToUpper, types[j].PropertyName)
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

func sortTypes(slice []TsType) {
	sort.Slice(slice, func(i, j int) bool {
		return strings.Map(unicode.ToUpper, slice[i].PropertyName) < strings.Map(unicode.ToUpper, slice[j].PropertyName)
	})
}

// ExtractTypeDefs extracts all type defs from the root and properties, sorted alphabetically.
//
// Includes the root definition
func ExtractTypeDefs(root TsType) []TsType {
	typeDefCache := make(identCache)
	typeDefCache[root.Ident] = root
	extractAllTypeDefs(typeDefCache, root)

	// not actually sorted yet
	var sortedTypeDefs []TsType
	for _, subType := range typeDefCache {
		sortedTypeDefs = append(sortedTypeDefs, subType)
	}

	sortTypes(sortedTypeDefs)

	return sortedTypeDefs
}

type identCache = map[Ident]TsType

func extractAllTypeDefs(cache identCache, types TsType) {
	for _, v := range types.Properties {
		if v.Kind == Object {
			if _, ok := cache[v.Ident]; !ok {
				cache[v.Ident] = v
				extractAllTypeDefs(cache, v)
			}
		}

		if v.Kind == Array {
			cv := v.Elem()
			if cv.Kind == Object {
				if _, ok := cache[cv.Ident]; !ok {
					cache[cv.Ident] = cv
					extractAllTypeDefs(cache, cv)
				}
			}
		}

		if v.Kind == Map {
			cv := v.Elem()
			if cv.Kind == Object {
				if _, ok := cache[cv.Ident]; !ok {
					cache[cv.Ident] = cv
					extractAllTypeDefs(cache, cv)
				}
			}
		}
	}
}
