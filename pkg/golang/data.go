package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
)

func printDataEntity(dst io.Writer, typ *ast.DataType) {
	_, _ = dst.Write([]byte("type " + toClassName(typ.Name.Name) + " struct"))
	_, _ = dst.Write([]byte(" {\n"))
	printExtend(dst, typ.Extends)

	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if nil != field.Comment {
			_, _ = dst.Write([]byte("\t//" + field.Comment.Text()))
		}

		_, _ = dst.Write([]byte("\t" + toFieldName(field.Name.Name) + " "))
		printType(dst, field.Type, false)

		_, _ = dst.Write([]byte("\t`json:\"" + toJsonName(field.Name.Name) + "\"`"))
		_, _ = dst.Write([]byte("\n\n"))
		return nil
	})
	if err != nil {
		return
	}
	_, _ = dst.Write([]byte("}\n\n"))

	_, _ = dst.Write([]byte("func (g *" + toClassName(typ.Name.Name) + ") ToData() ([]byte, error) {\n"))
	_, _ = dst.Write([]byte("	return json.Marshal(g)\n"))
	_, _ = dst.Write([]byte("}\n\n"))

	_, _ = dst.Write([]byte("func (g *" + toClassName(typ.Name.Name) + ") FormData(data []byte) error {\n"))
	_, _ = dst.Write([]byte("	return json.Unmarshal(data, g)\n"))
	_, _ = dst.Write([]byte("}\n\n"))
}

func printExtend(dst io.Writer, extends []*ast.Ident) {
	for _, v := range extends {
		_, _ = dst.Write([]byte("\t"))
		_, _ = dst.Write([]byte(toClassName(v.Name)))
		_, _ = dst.Write([]byte("\n\n"))
	}
}
