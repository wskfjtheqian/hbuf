package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strings"
)

func printEnumCode(dst *build.Writer, typ *ast.EnumType) {
	name := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("// " + name + " " + typ.Doc.Text())
	}
	maxLen := 0
	dst.Code("type " + name + " int32\n\n")

	dst.Import("strconv", "", 0)
	dst.Import("strings", "", 0)
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/herror", "", 0)

	dst.Code("func (t ").Code(name).Code(") MarshalJSON() ([]byte, error) {\n")
	dst.Tab(1).Code("return []byte(strconv.FormatInt(int64(t), 10)), nil\n")
	dst.Code("}\n")
	dst.Code("\n")
	dst.Code("func (e *").Code(name).Code(") UnmarshalJSON(data []byte) error {\n")
	dst.Tab(1).Code("parseInt, err := strconv.ParseInt(strings.Trim(string(data), \"\\\"\"), 10, 64)\n")
	dst.Tab(1).Code("if err != nil {\n")
	dst.Tab(1).Code("parseInt = 0\n")
	dst.Tab(1).Code("herror.PrintStack(err)\n")
	dst.Code("}\n")
	dst.Tab(1).Code("*e = ").Code(name).Code("(parseInt)\n")
	dst.Tab(1).Code("return nil\n")
	dst.Code("}\n")

	for _, item := range typ.Items {
		itemName := build.StringToHumpName(item.Name.Name)
		l := len(itemName)
		if l > maxLen {
			maxLen = l
		}
		if nil != item.Doc && 0 < len(item.Doc.Text()) {
			dst.Code("// " + name + itemName + " " + item.Doc.Text())
		}
		dst.Code("const " + name + itemName + " " + name + " = " + item.Id.Value + "\n\n")
	}

	dst.Code("func (e " + name + ") Pointer() *" + name + " {\n")
	dst.Tab(1).Code("pointer := e\n")
	dst.Tab(1).Code("return &pointer\n")
	dst.Code("}\n\n")

	space := strings.Builder{}
	for i := 0; i < maxLen; i++ {
		space.WriteString(" ")
	}
	spaceText := space.String()
	dst.Code("var " + build.StringToFirstLower(typ.Name.Name) + "Map = map[" + name + "]string{\n")
	for _, item := range typ.Items {
		enumItem := build.StringToHumpName(item.Name.Name)
		dst.Tab(1).Code("" + name + enumItem + ": " + spaceText[:maxLen-len(enumItem)] + "\"" + build.StringToHumpName(item.Name.Name) + "\",\n")
	}
	dst.Code("}\n\n")

	dst.Code("func (e " + name + ") ToName() string {\n")
	dst.Tab(1).Code("return " + build.StringToFirstLower(typ.Name.Name) + "Map[e]\n")
	dst.Code("}\n\n")

	dst.Code("var " + build.StringToFirstLower(typ.Name.Name) + "Values = map[string]" + name + "{\n")
	for _, item := range typ.Items {
		enumItem := build.StringToHumpName(item.Name.Name)
		dst.Tab(1).Code("\"" + enumItem + "\": " + spaceText[:maxLen-len(enumItem)] + name + build.StringToHumpName(item.Name.Name) + ",\n")
	}
	dst.Code("}\n\n")

	dst.Code("func " + name + "Values() map[string]" + name + " {\n")
	dst.Tab(1).Code("return " + build.StringToFirstLower(typ.Name.Name) + "Values\n")
	dst.Code("}\n\n")

}
