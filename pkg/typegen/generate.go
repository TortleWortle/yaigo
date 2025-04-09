package typegen

import (
	"fmt"
	"strings"
)

func GenerateTypeDef(parent TsType) string {
	if parent.Kind != Object {
		panic("can only generate typedefs for objects")
	}
	types := parent.Properties
	var writer strings.Builder

	if parent.export {
		writer.WriteString("export ")
	}
	writer.WriteString(fmt.Sprintf("type %s = {\n", parent.Ident))
	for _, v := range types {
		writer.WriteString("\t")
		writer.WriteString(v.PropertyName)
		if v.Optional {
			writer.WriteString("?")
		}
		writer.WriteString(": ")
		if v.Kind == Object {
			writer.WriteString(v.Ident.String())
		} else if v.Kind == Array {
			writer.WriteString(fmt.Sprintf("%s[]", v.Elem().Ident))
		} else if v.Kind == Map {
			writer.WriteString("{\n")
			writer.WriteString(fmt.Sprintf("\t\t[key: %s]: ", v.MapKey().Ident))
			writer.WriteString(v.Elem().Ident.String())
			if v.Elem().Optional {
				writer.WriteString(" | null")
			}
			writer.WriteString(";\n")
			writer.WriteString("\t}")
		} else if v.Kind == Primitive {
			writer.WriteString(v.Ident.String())
		} else {
			writer.WriteString("never")
		}
		writer.WriteString(";")
		if v.Comment != "" {
			writer.WriteString("// " + v.Comment)
		}
		writer.WriteString("\n")
	}

	writer.WriteString("}")
	writer.WriteString("\n")

	return writer.String()
}
