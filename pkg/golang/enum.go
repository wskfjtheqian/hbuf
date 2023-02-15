package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func printEnumCode(dst *build.Writer, typ *ast.EnumType) {
	name := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("//" + name + " " + typ.Doc.Text())
	}
	dst.Code("type " + name + " int\n\n")
	for _, item := range typ.Items {
		itemName := name + build.StringToHumpName(item.Name.Name)
		if nil != item.Doc && 0 < len(item.Doc.Text()) {
			dst.Code("//" + itemName + " " + item.Doc.Text())
		}
		dst.Code("const " + itemName + " " + name + " = " + item.Id.Value + "\n\n")
	}

	dst.Code("func (e " + name + ") ToName() string {\n")
	dst.Code("\tswitch e {\n")
	for _, item := range typ.Items {
		dst.Code("\tcase " + name + build.StringToHumpName(item.Name.Name) + ":\n")
		dst.Code("\t\treturn \"" + build.StringToAllUpper(item.Name.Name) + "\"\n")
	}
	dst.Code("\t}\n")
	dst.Code("\treturn \"\"\n")
	dst.Code("}\n\n")

	dst.Code("func (e " + name + ") OfName(name string) (" + name + ", error) {\n")
	dst.Code("\tswitch name {\n")
	for _, item := range typ.Items {
		dst.Code("\tcase \"" + build.StringToAllUpper(item.Name.Name) + "\":\n")
		dst.Code("\t\treturn " + name + build.StringToHumpName(item.Name.Name) + ", nil\n")
	}
	dst.Code("\t}\n")
	dst.Import("errors", "")
	dst.Code("\treturn 0, errors.New(name + \" not to " + name + "\")\n")
	dst.Code("}\n\n")

	dst.Code("func (e " + name + ") Pointer() *" + name + " {\n")
	dst.Code("	pointer := e\n")
	dst.Code("	return &pointer\n")
	dst.Code("}\n\n")
}
