package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func printEnumCode(dst *Writer, typ *ast.EnumType) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("type " + name + " = int\n\n")
	for _, item := range typ.Items {
		dst.Code("const " + name + build.StringToHumpName(item.Name.Name) + " " + name + " = " + item.Id.Value + "\n\n")
	}

	dst.Code("func " + name + "ToName(value " + name + ") string {\n")
	dst.Code("	switch value {\n")
	for _, item := range typ.Items {
		dst.Code("		case " + name + build.StringToHumpName(item.Name.Name) + ":\n")
		dst.Code("			return \"" + build.StringToAllUpper(item.Name.Name) + "\"\n")
	}
	dst.Code("	}\n")
	dst.Code("	return \"\"\n")
	dst.Code("}\n\n")

	dst.Code("func AccountStatusOfName(name string) AccountStatus {\n")
	dst.Code("	switch name {\n")
	for _, item := range typ.Items {
		dst.Code("		case \"" + build.StringToAllUpper(item.Name.Name) + "\":\n")
		dst.Code("			return " + name + build.StringToHumpName(item.Name.Name) + "\n")
	}
	dst.Code("	}\n")
	dst.Code("	return 0\n")
	dst.Code("}\n")
}
