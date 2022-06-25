package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func printEnumCode(dst *Writer, typ *ast.EnumType) {
	name := build.StringToHumpName(typ.Name.Name)
	if 0 < len(typ.Doc.Text()) {
		dst.Code("//" + name + " " + typ.Doc.Text())
	}
	dst.Code("type " + name + " int\n\n")
	for _, item := range typ.Items {
		itemName := name + build.StringToHumpName(item.Name.Name)
		if 0 < len(item.Doc.Text()) {
			dst.Code("//" + itemName + " " + item.Doc.Text())
		}
		dst.Code("const " + itemName + " " + name + " = " + item.Id.Value + "\n\n")
	}

	dst.Code("func (e " + name + ") ToName() string {\n")
	dst.Code("	switch e {\n")
	for _, item := range typ.Items {
		dst.Code("		case " + name + build.StringToHumpName(item.Name.Name) + ":\n")
		dst.Code("			return \"" + build.StringToAllUpper(item.Name.Name) + "\"\n")
	}
	dst.Code("	}\n")
	dst.Code("	return \"\"\n")
	dst.Code("}\n\n")

	dst.Code("func (e " + name + ") OfName(name string) " + name + " {\n")
	dst.Code("	switch name {\n")
	for _, item := range typ.Items {
		dst.Code("		case \"" + build.StringToAllUpper(item.Name.Name) + "\":\n")
		dst.Code("			return " + name + build.StringToHumpName(item.Name.Name) + "\n")
	}
	dst.Code("	}\n")
	dst.Code("	return 0\n")
	dst.Code("}\n")
}
