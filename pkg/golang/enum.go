package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strings"
)

func printEnumCode(dst *build.Writer, typ *ast.EnumType) {
	name := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("//" + name + " " + typ.Doc.Text())
	}
	maxLen := 0
	dst.Code("type " + name + " int\n\n")
	for _, item := range typ.Items {
		itemName := build.StringToHumpName(item.Name.Name)
		l := len(itemName)
		if l > maxLen {
			maxLen = l
		}
		if nil != item.Doc && 0 < len(item.Doc.Text()) {
			dst.Code("//" + name + itemName + " " + item.Doc.Text())
		}
		dst.Code("const " + name + itemName + " " + name + " = " + item.Id.Value + "\n\n")
	}

	dst.Code("func (e " + name + ") Pointer() *" + name + " {\n")
	dst.Code("	pointer := e\n")
	dst.Code("	return &pointer\n")
	dst.Code("}\n\n")

	space := strings.Builder{}
	for i := 0; i < maxLen; i++ {
		space.WriteString(" ")
	}
	spaceText := space.String()
	dst.Code("var " + build.StringToFirstLower(typ.Name.Name) + "Map = map[" + name + "]string{\n")
	for _, item := range typ.Items {
		enumItem := build.StringToHumpName(item.Name.Name)
		dst.Code("\t" + name + enumItem + ": " + spaceText[:maxLen-len(enumItem)] + "\"" + build.StringToHumpName(item.Name.Name) + "\",\n")
	}
	dst.Code("}\n\n")

	dst.Code("func (e " + name + ") ToName() string {\n")
	dst.Code("\treturn " + build.StringToFirstLower(typ.Name.Name) + "Map[e]\n")
	dst.Code("}\n\n")

	dst.Code("var " + build.StringToFirstLower(typ.Name.Name) + "Values = map[string]" + name + "{\n")
	for _, item := range typ.Items {
		enumItem := build.StringToHumpName(item.Name.Name)
		dst.Code("\t\"" + enumItem + "\": " + spaceText[:maxLen-len(enumItem)] + name + build.StringToHumpName(item.Name.Name) + ",\n")
	}
	dst.Code("}\n\n")

	dst.Code("func " + name + "Values() map[string]" + name + "{\n")
	dst.Code("\treturn " + build.StringToFirstLower(typ.Name.Name) + "Values\n")
	dst.Code("}\n\n")

}
