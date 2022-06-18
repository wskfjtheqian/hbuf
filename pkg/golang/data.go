package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func printDataCode(dst *Writer, typ *ast.DataType) {
	dst.Import("encoding/json")

	dst.Code("type " + build.StringToHumpName(typ.Name.Name) + " struct")
	dst.Code(" {\n")
	printExtend(dst, typ.Extends)

	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if nil != field.Comment {
			dst.Code("\t//" + field.Comment.Text())
		}

		dst.Code("\t" + build.StringToHumpName(field.Name.Name) + " ")
		printType(dst, field.Type, false)

		dst.Code("\t`json:\"" + build.StringToUnderlineName(field.Name.Name) + "\"`")
		dst.Code("\n\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("}\n\n")

	dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") ToData() ([]byte, error) {\n")
	dst.Code("	return json.Marshal(g)\n")
	dst.Code("}\n\n")

	dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") FormData(data []byte) error {\n")
	dst.Code("	return json.Unmarshal(data, g)\n")
	dst.Code("}\n\n")
}

func printExtend(dst *Writer, extends []*ast.Ident) {
	for _, v := range extends {
		dst.Code("\t")
		dst.Code(build.StringToHumpName(v.Name))
		dst.Code("\n\n")
	}
}
