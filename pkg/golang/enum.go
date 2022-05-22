package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
)

func printEnum(dst io.Writer, typ *ast.EnumType) {
	name := build.StringToHumpName(typ.Name.Name)
	_, _ = dst.Write([]byte("type " + name + " = int\n\n"))
	for _, item := range typ.Items {
		_, _ = dst.Write([]byte("const " + name + build.StringToHumpName(item.Name.Name) + " " + name + " = " + item.Id.Value + "\n\n"))
	}

	_, _ = dst.Write([]byte("func " + name + "ToName(value " + name + ") string {\n"))
	_, _ = dst.Write([]byte("	switch value {\n"))
	for _, item := range typ.Items {
		_, _ = dst.Write([]byte("		case " + name + build.StringToHumpName(item.Name.Name) + ":\n"))
		_, _ = dst.Write([]byte("			return \"" + build.StringToAllUpper(item.Name.Name) + "\"\n"))
	}
	_, _ = dst.Write([]byte("	}\n"))
	_, _ = dst.Write([]byte("	return \"\"\n"))
	_, _ = dst.Write([]byte("}\n\n"))

	_, _ = dst.Write([]byte("func AccountStatusOfName(name string) AccountStatus {\n"))
	_, _ = dst.Write([]byte("	switch name {\n"))
	for _, item := range typ.Items {
		_, _ = dst.Write([]byte("		case \"" + build.StringToAllUpper(item.Name.Name) + "\":\n"))
		_, _ = dst.Write([]byte("			return " + name + build.StringToHumpName(item.Name.Name) + "\n"))
	}
	_, _ = dst.Write([]byte("	}\n"))
	_, _ = dst.Write([]byte("	return 0\n"))
	_, _ = dst.Write([]byte("}\n"))
}
