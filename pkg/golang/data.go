package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printDataCode(dst *build.Writer, typ *ast.DataType) {
	dst.Import("encoding/json", "")
	name := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("//" + name + " " + typ.Doc.Text())
	}
	dst.Code("type " + name + " struct")
	dst.Code(" {\n")
	b.printExtend(dst, typ.Extends)

	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("\t//" + field.Doc.Text())
		}

		dst.Code("\t" + build.StringToHumpName(field.Name.Name) + " ")
		b.printType(dst, field.Type, false)

		dst.Code("\t`json:\"" + build.StringToUnderlineName(field.Name.Name) + "\"`")
		dst.Code("\n\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("}\n\n")

	dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") ToData() ([]byte, error) {\n")
	dst.Code("\treturn json.Marshal(g)\n")
	dst.Code("}\n\n")

	dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") FormData(data []byte) error {\n")
	dst.Code("\treturn json.Unmarshal(data, g)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printExtend(dst *build.Writer, extends []*ast.Ident) {
	for _, v := range extends {
		dst.Code("\t")
		dst.Code(build.StringToHumpName(v.Name))
		dst.Code("\n\n")
	}
}
