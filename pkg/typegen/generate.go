package typegen

import (
	"fmt"
	"strings"
)

func writeIndentation(writer *strings.Builder, indentation int) {
	for i := 0; i < indentation; i++ {
		_, _ = writer.WriteString("\t")
	}
}

func typeStrToTs(in string) string {
	var out string
	switch in {
	case "uint8":
		fallthrough
	case "uint16":
		fallthrough
	case "uint32":
		fallthrough
	case "uint64":
		fallthrough
	case "int8":
		fallthrough
	case "int16":
		fallthrough
	case "int32":
		fallthrough
	case "int64":
		fallthrough
	case "float32":
		fallthrough
	case "float64":
		fallthrough
	case "byte":
		fallthrough
	case "uint":
		fallthrough
	case "int":
		out = "number"
	case "bool":
		out = "bool"
	case "string":
		out = "string"
	case "true":
		out = "true"
	case "false":
		out = "false"
	case "interface {}":
		fallthrough
	case "any":
		out = "any"
	default:
		out = in
	}
	if len(out) > 0 && strings.ToLower(string(out[0])) == string(out[0]) {
		out = titleCaser.String(out)
	}
	return out
}

func generateObjectDef(parent TsType, indentation int, forceNull string) string {
	var writer strings.Builder
	writer.WriteString("{\n")
	writeIndentation(&writer, indentation)
	for i, v := range parent.Properties {
		if i > 0 {
			writeIndentation(&writer, indentation)
		}
		writer.WriteString(v.PropertyName)
		if v.Optional {
			writer.WriteString("?")
		}
		writer.WriteString(": ")
		if v.PropertyName == forceNull {
			writer.WriteString(TypeNull)
		} else if v.Kind == Object {
			writer.WriteString(v.Ident.String())
		} else if v.Kind == InlineObject {
			writer.WriteString(generateObjectDef(v, indentation+1, ""))
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
	writeIndentation(&writer, indentation-1)
	writer.WriteString("}")
	return writer.String()
}

func GenerateTypeDef(parent TsType) string {
	if parent.Kind != Object {
		panic("can only generate typedefs for objects")
	}
	var writer strings.Builder

	if parent.export {
		writer.WriteString("export ")
	}
	writer.WriteString(fmt.Sprintf("type %s", parent.Ident))
	writer.WriteString(" = ")
	if len(parent.Union) == 2 {
		// todo: instead of force-null copy the parent and change the type in there.
		writer.WriteString(generateObjectDef(parent, 1, parent.Union[0]))
		writer.WriteString(" | ")
		writer.WriteString(generateObjectDef(parent, 1, parent.Union[1]))
	} else {
		writer.WriteString(generateObjectDef(parent, 1, ""))
	}
	writer.WriteString("\n")

	return writer.String()
}
