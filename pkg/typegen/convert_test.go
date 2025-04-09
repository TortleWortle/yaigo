package typegen_test

import (
	_ "embed"
	"github.com/tortlewortle/yaigo/pkg/typegen"
	"github.com/tortlewortle/yaigo/pkg/typegen/internal/db"
	"testing"
)

var optString string
var optInt int
var optSlice []string

type BasicStruct struct {
	StringVal    string `json:"username,omitempty"`
	IntVal       int    `json:"age"`
	StringSlice  []string
	OptStringVal *string
	OptIntVal    *int
	BrokenIntVal uint
	Int8         int8
	Floaty       float32
	MapTime      map[string]int
	MapNillable  map[string]*int
	MapBad       map[int]string
	AliasMapTime AliasedMap
	NestedStruct NestedStruct
	privateField string
}
type AliasedMap = map[string]NestedStruct

type NestedStruct struct {
	BaseVal string
}

type Props = map[string]any

var testProps = Props{
	"basic": BasicStruct{
		StringVal:    "val",
		IntVal:       12,
		StringSlice:  []string{"hello", "world"},
		OptStringVal: nil,
		OptIntVal:    nil,
		privateField: "yah",
	},
	"basicButOptional": &BasicStruct{
		StringVal:    "val",
		IntVal:       12,
		StringSlice:  []string{"hello", "world"},
		OptStringVal: nil,
		OptIntVal:    nil,
	},
	"stringProp":          "value",
	"intProp":             12,
	"optString":           &optString,
	"optInt":              &optInt,
	"nilField":            nil,
	"stringSlice":         []string{"hello", "there"},
	"stringArray":         [2]string{"hi", "there"},
	"optStringSlice":      &optSlice,
	"basicStructSlice":    []BasicStruct{},
	"otherPkgStructSlice": []db.User{},
	"otherPkgStructMap":   map[string]db.Group{},
}

//go:embed data/WelcomeIndexProps.ts
var expectedOutput string

func TestGenerateTypeDefs(t *testing.T) {
	types, err := typegen.ConvertMap(testProps)
	if err != nil {
		t.Error(err)
	}

	parent := typegen.NewRootType("WelcomeIndexProps", types)

	typeDefs := typegen.GenerateTypeDef(parent, true)

	if typeDefs != expectedOutput {
		t.Errorf("output mismatch. got: \n%s\n expected:\n%s\n", typeDefs, expectedOutput)
	}
}
