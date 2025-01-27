package main

import (
	"bytes"
	"fmt"
	"strings"
)

func resolveType(field *Field, forNew bool) string {
	switch field.Type {
	case "bytes":
		return "[]byte"
	case "bigint":
		return "big.Int"
	case "uvarint":
		return "uint64"
	case "varint":
		return "int64"
	case "chain":
		return "[32]byte"
	case "chainSet":
		return "[][32]byte"
	case "duration":
		return "time.Duration"
	case "time":
		return "time.Time"
	case "any":
		return "interface{}"
	case "slice":
		return "[]" + resolveType(field.Slice, false)
	}

	typ := field.Type
	if field.Pointer && !forNew {
		typ = "*" + typ
	}
	return typ
}

func jsonType(field *Field) string {
	switch field.Type {
	case "bytes":
		return "*string"
	case "chain":
		return "string"
	case "chainSet":
		return "[]string"
	case "duration":
		return "interface{}"
	case "slice":
		jt := jsonType(field.Slice)
		if jt != "" {
			return "[]" + jt
		}
	}
	return ""
}

func areEqual(w *bytes.Buffer, field *Field, varName, otherName string) error {
	var expr string
	switch field.Type {
	case "bool", "string", "chain", "uvarint", "varint", "duration", "time":
		expr = "%s == %s"
	case "bytes":
		expr = "bytes.Equal(%s, %s)"
	case "bigint":
		if field.Pointer {
			expr = "%s.Cmp(%s) == 0"
		} else {
			expr = "%s.Cmp(&%s) == 0"
		}
	case "slice", "chainSet":
		expr = "len(%s) == len(%s)"
	default:
		switch field.MarshalAs {
		case "reference":
			if field.Pointer {
				expr = "%s.Equal(%s)"
			} else {
				expr = "%s.Equal(&%s)"
			}
		case "value":
			if field.Pointer {
				expr = "*%s == *%s"
			} else {
				expr = "%s == %s"
			}
		default:
			return fmt.Errorf("field %q: cannot determine how to compare %s", field.Name, resolveType(field, false))
		}
	}

	expr = fmt.Sprintf(expr, varName, otherName)
	fmt.Fprintf(w, "\tif !(%s) { return false }\n\n", expr)

	switch field.Type {
	case "slice":
		fmt.Fprintf(w, "\tfor i := range %s {\n", varName)
		fmt.Fprintf(w, "\t\tv, u := %s[i], %s[i]\n", varName, otherName)
		err := areEqual(w, field.Slice, "v", "u")
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "\t}\n\n")

	case "chainSet":
		fmt.Fprintf(w, "\tfor i := range %s {\n", varName)
		fmt.Fprintf(w, "\t\tif %s[i] != %s[i] { return false }\n", varName, otherName)
		fmt.Fprintf(w, "\t}\n\n")

	default:
		fmt.Fprintf(w, "\n")
	}
	return nil
}

func binarySize(w *bytes.Buffer, field *Field, varName string) error {
	var expr string
	switch field.Type {
	case "bool", "bytes", "string", "chainSet", "uvarint", "varint", "duration", "time":
		expr = methodName(field.Type, "BinarySize") + "(%s)"
	case "bigint", "chain":
		expr = methodName(field.Type, "BinarySize") + "(&%s)"
	case "slice":
		expr = "encoding.UvarintBinarySize(uint64(len(%s)))"
	default:
		if field.MarshalAs != "reference" && field.MarshalAs != "value" {
			return fmt.Errorf("field %q: cannot determine how to marshal %s", field.Name, resolveType(field, false))
		}
		expr = "%s.BinarySize()"
	}

	expr = fmt.Sprintf(expr, varName)
	fmt.Fprintf(w, "\tn += %s\n\n", expr)

	if field.Type != "slice" {
		fmt.Fprintf(w, "\n")
		return nil
	}

	fmt.Fprintf(w, "\tfor _, v := range %s {\n", varName)
	err := binarySize(w, field.Slice, "v")
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "\t}\n\n")
	return nil
}

func binaryMarshalValue(w *bytes.Buffer, field *Field, varName, errName string, errArgs ...string) error {
	var expr string
	var canErr bool
	switch field.Type {
	case "bool", "bytes", "string", "chainSet", "uvarint", "varint", "duration", "time":
		expr, canErr = methodName(field.Type, "MarshalBinary")+"(%s)", false
	case "bigint", "chain":
		expr, canErr = methodName(field.Type, "MarshalBinary")+"(&%s)", false
	case "slice":
		expr, canErr = "encoding.UvarintMarshalBinary(uint64(len(%s)))", false
	default:
		if field.MarshalAs != "reference" && field.MarshalAs != "value" {
			return fmt.Errorf("field %q: cannot determine how to marshal %s", field.Name, resolveType(field, false))
		}
		expr, canErr = "%s.MarshalBinary()", true
	}

	expr = fmt.Sprintf(expr, varName)
	if canErr {
		err := fieldError("encoding", errName, errArgs...)
		fmt.Fprintf(w, "\tif b, err := %s; err != nil { return nil, %s } else { buffer.Write(b) }\n", expr, err)
	} else {
		fmt.Fprintf(w, "\tbuffer.Write(%s)\n", expr)
	}

	if field.Type != "slice" {
		fmt.Fprintf(w, "\n")
		return nil
	}

	fmt.Fprintf(w, "\tfor i, v := range %s {\n", varName)
	fmt.Fprintf(w, "\t\t_ = i\n")
	err := binaryMarshalValue(w, field.Slice, "v", errName+"[%d]", "i")
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "\t}\n\n")
	return nil
}

func binaryUnmarshalValue(w *bytes.Buffer, field *Field, varName, errName string, errArgs ...string) error {
	var expr, size, sliceName string
	var inPlace bool
	switch field.Type {
	case "bool", "bytes", "string", "chainSet", "uvarint", "varint", "duration", "time":
		expr, size, inPlace = methodName(field.Type, "UnmarshalBinary")+"(data)", methodName(field.Type, "BinarySize")+"(%s)", false
	case "bigint", "chain":
		expr, size, inPlace = methodName(field.Type, "UnmarshalBinary")+"(data)", methodName(field.Type, "BinarySize")+"(&%s)", false
	case "slice":
		sliceName, varName = varName, "len"+field.Name
		fmt.Fprintf(w, "var %s uint64\n", varName)
		expr, size, inPlace = "encoding.UvarintUnmarshalBinary(data)", "encoding.UvarintBinarySize(%s)", false
	default:
		if field.MarshalAs != "reference" && field.MarshalAs != "value" {
			return fmt.Errorf("field %q: cannot determine how to marshal %s", field.Name, resolveType(field, false))
		}
		expr, size, inPlace = "%s.UnmarshalBinary(data)", "%s.BinarySize()", true
	}

	size = fmt.Sprintf(size, varName)
	err := fieldError("decoding", errName, errArgs...)
	if inPlace {
		expr = fmt.Sprintf(expr, varName)
		fmt.Fprintf(w, "\tif err := %s; err != nil { return %s }\n", expr, err)
	} else if field.Type == "bigint" {
		fmt.Fprintf(w, "\tif x, err := %s; err != nil { return %s } else { %s.Set(x) }\n", expr, err, varName)
	} else {
		fmt.Fprintf(w, "\tif x, err := %s; err != nil { return %s } else { %s = x }\n", expr, err, varName)
	}
	fmt.Fprintf(w, "\tdata = data[%s:]\n\n", size)

	if field.Type != "slice" {
		return nil
	}

	fmt.Fprintf(w, "\t%s = make(%s, %s)\n", sliceName, resolveType(field, false), varName)
	fmt.Fprintf(w, "\tfor i := range %s {\n", sliceName)
	if field.Slice.Pointer {
		fmt.Fprintf(w, "\t\tx := new(%s)\n", resolveType(field.Slice, true))
		err := binaryUnmarshalValue(w, field.Slice, "x", errName+"[%d]", "i")
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "\t\t%s[i] = x", sliceName)
	} else {
		err := binaryUnmarshalValue(w, field.Slice, sliceName+"[i]", errName+"[%d]", "i")
		if err != nil {
			return err
		}
	}
	fmt.Fprintf(w, "\t}\n\n")
	return nil
}

func jsonVar(w *bytes.Buffer, typ *Record, varName string) {
	fmt.Fprintf(w, "\t%s := struct{\n", varName)
	if typ.Kind == "chain" {
		if flags.IsState {
			fmt.Fprintf(w, "\t\tChainHeader\n")
		} else {
			fmt.Fprintf(w, "\t\tstate.ChainHeader\n")
		}
	}
	for _, e := range typ.Embeddings {
		fmt.Fprintf(w, "\t\t%s\n", e)
	}
	for _, f := range typ.Fields {
		lcName := strings.ToLower(f.Name[:1]) + f.Name[1:]
		typ := jsonType(f)
		if typ == "" {
			typ = resolveType(f, false)
		}
		if f.KeepEmpty {
			fmt.Fprintf(w, "\t\t%s %s `json:\"%s\"`\n", f.Name, typ, lcName)
		} else {
			fmt.Fprintf(w, "\t\t%s %s `json:\"%s,omitempty\"`\n", f.Name, typ, lcName)
		}
	}
	fmt.Fprintf(w, "\t}{}\n")
}

func valueToJson(w *bytes.Buffer, field *Field, tgtName, srcName string) {
	switch field.Type {
	case "bytes", "chain", "chainSet", "duration":
		fmt.Fprintf(w, "\t%s = %s(%s)\n", tgtName, methodName(field.Type, "ToJSON"), srcName)
		return

	case "slice":
		if jsonType(field.Slice) == "" {
			break
		}

		fmt.Fprintf(w, "\t%s = make([]%s, len(%s))\n", tgtName, jsonType(field.Slice), srcName)
		fmt.Fprintf(w, "\tfor i, x := range %s {\n", srcName)
		valueToJson(w, field.Slice, tgtName+"[i]", "x")
		fmt.Fprintf(w, "\t}\n")
		return
	}

	// default:
	fmt.Fprintf(w, "\t%s = %s\n", tgtName, srcName)
}

func valueFromJson(w *bytes.Buffer, field *Field, tgtName, srcName, errName string, errArgs ...string) {
	err := fieldError("decoding", errName, errArgs...)
	switch field.Type {
	case "bytes", "chain", "chainSet", "duration":
		fmt.Fprintf(w, "\tif x, err := %s(%s); err != nil {\n\t\treturn %s\n\t} else {\n\t\t%s = x\n\t}\n", methodName(field.Type, "FromJSON"), srcName, err, tgtName)
		return

	case "slice":
		if jsonType(field.Slice) == "" {
			break
		}

		fmt.Fprintf(w, "\t%s = make([]%s, len(%s))\n", tgtName, resolveType(field.Slice, false), srcName)
		fmt.Fprintf(w, "\tfor i, x := range %s {\n", srcName)
		valueFromJson(w, field.Slice, tgtName+"[i]", "x", errName+"[%d]", "i")
		fmt.Fprintf(w, "\t}\n")
		return
	}

	// default:
	fmt.Fprintf(w, "\t%s = %s\n", tgtName, srcName)
}
