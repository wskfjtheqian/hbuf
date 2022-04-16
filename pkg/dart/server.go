package dart

import (
	"hbuf/pkg/ast"
	"io"
)

func printServer(dst io.Writer, typ *ast.ServerType) {
	_, _ = dst.Write([]byte("abstract class " + toClassName(typ.Name.Name)))
	if nil != typ.Extends {
		printExtend(dst, typ.Extends)
	}
	_, _ = dst.Write([]byte("{\n"))
	for _, method := range typ.Methods {
		if nil != method.Comment {
			_, _ = dst.Write([]byte("  /// " + method.Comment.Text()))
		}

		_, _ = dst.Write([]byte("  "))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte(" " + toFieldName(method.Name.Name)))
		_, _ = dst.Write([]byte("("))
		printType(dst, method.Param, false)
		_, _ = dst.Write([]byte(" " + toFieldName(method.ParamName.Name)))
		_, _ = dst.Write([]byte(");\n\n"))
	}
	_, _ = dst.Write([]byte("}\n\n"))
}

func printServerEntity(dst io.Writer, typ *ast.ServerType) {
	_, _ = dst.Write([]byte("abstract class " + toClassName(typ.Name.Name) + "Imp extends ServerImp implements " + toClassName(typ.Name.Name)))
	if nil != typ.Extends {
		printExtend(dst, typ.Extends)
	}
	_, _ = dst.Write([]byte("{\n"))
	for _, method := range typ.Methods {

		_, _ = dst.Write([]byte("  @override\n"))
		_, _ = dst.Write([]byte("  "))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte(" " + toFieldName(method.Name.Name)))
		_, _ = dst.Write([]byte("("))
		printType(dst, method.Param, false)
		_, _ = dst.Write([]byte(" " + toFieldName(method.ParamName.Name)))
		_, _ = dst.Write([]byte("){\n"))

		_, _ = dst.Write([]byte("    return invoke<"))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte(">(\""))
		_, _ = dst.Write([]byte(toFieldName(method.Name.Name)))
		_, _ = dst.Write([]byte("\", "))
		_, _ = dst.Write([]byte("0"))
		_, _ = dst.Write([]byte(", "))
		_, _ = dst.Write([]byte(toFieldName(method.ParamName.Name)))
		_, _ = dst.Write([]byte(", "))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte(".fromMap, "))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte(".fromData);\n"))

		_, _ = dst.Write([]byte("  }\n\n"))
	}
	_, _ = dst.Write([]byte("}\n\n"))
}
