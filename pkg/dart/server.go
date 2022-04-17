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

func printServerImp(dst io.Writer, typ *ast.ServerType) {
	_, _ = dst.Write([]byte("class " + toClassName(typ.Name.Name) + "Imp extends ServerImp implements " + toClassName(typ.Name.Name)))
	if nil != typ.Extends {
		printExtend(dst, typ.Extends)
	}

	_, _ = dst.Write([]byte("{\n"))
	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  String get name => \"" + toClassName(typ.Name.Name) + "\";\n\n"))
	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  int get id => 0;\n\n"))

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
		_, _ = dst.Write([]byte(method.Name.Name))
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

func printServerRoute(dst io.Writer, typ *ast.ServerType) {
	_, _ = dst.Write([]byte("class " + toClassName(typ.Name.Name) + "Route extends ServerRoute"))
	if nil != typ.Extends {
		printExtend(dst, typ.Extends)
	}

	_, _ = dst.Write([]byte("{\n"))
	_, _ = dst.Write([]byte("  final " + toClassName(typ.Name.Name) + " server;\n\n"))
	_, _ = dst.Write([]byte("  " + toClassName(typ.Name.Name) + "Route(this.server);\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  String get name => \"" + toClassName(typ.Name.Name) + "\";\n\n"))
	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  int get id => 0;\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  ByteData invokeData(int id, ByteData data) {\n"))
	_, _ = dst.Write([]byte("    switch (id) {\n"))
	for _, method := range typ.Methods {
		_, _ = dst.Write([]byte("      case 0 :\n"))
		_, _ = dst.Write([]byte("        return server." + toFieldName(method.Name.Name) + "("))
		printType(dst, method.Param.Type(), false)
		_, _ = dst.Write([]byte(".fromData(data)!).toData();\n"))
	}
	_, _ = dst.Write([]byte("    }\n"))
	_, _ = dst.Write([]byte("    return ByteData(0);\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  Map<String, dynamic> invokeMap(String name, Map<String, dynamic> map) {\n"))
	_, _ = dst.Write([]byte("    switch (name) {\n"))
	for _, method := range typ.Methods {
		_, _ = dst.Write([]byte("      case \"" + method.Name.Name + "\":\n"))
		_, _ = dst.Write([]byte("        return server." + toFieldName(method.Name.Name) + "("))
		printType(dst, method.Param.Type(), false)
		_, _ = dst.Write([]byte(".fromMap(map)!).toMap();\n"))
	}
	_, _ = dst.Write([]byte("    }\n"))
	_, _ = dst.Write([]byte("    return {};\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("}\n\n"))
}
