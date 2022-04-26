package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
)

func printServer(dst io.Writer, typ *ast.ServerType) {
	_, _ = dst.Write([]byte("abstract class " + toClassName(typ.Name.Name)))
	if nil != typ.Extends {
		_, _ = dst.Write([]byte(" implements "))
		printExtend(dst, typ.Extends, false)
	}
	_, _ = dst.Write([]byte("{\n"))
	for _, method := range typ.Methods {
		if nil != method.Comment {
			_, _ = dst.Write([]byte("  /// " + method.Comment.Text()))
		}
		if build.CheckSuperMethod(method.Name.Name, typ) {
			_, _ = dst.Write([]byte("  @override\n"))
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

	_, _ = dst.Write([]byte("{\n"))
	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  String get name => \"" + toClassName(typ.Name.Name) + "\";\n\n"))
	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  int get id => " + typ.Id.Value + ";\n\n"))

	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
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
		_, _ = dst.Write([]byte(server.Name.Name + "/" + method.Name.Name))
		_, _ = dst.Write([]byte("\", "))
		_, _ = dst.Write([]byte(server.Id.Value + " << 32 | " + method.Id.Value))
		_, _ = dst.Write([]byte(", "))
		_, _ = dst.Write([]byte(toFieldName(method.ParamName.Name)))
		_, _ = dst.Write([]byte(", "))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte(".fromMap, "))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte(".fromData);\n"))

		_, _ = dst.Write([]byte("  }\n\n"))
		return nil
	})
	_, _ = dst.Write([]byte("}\n\n"))
}

func printServerRoute(dst io.Writer, typ *ast.ServerType) {
	_, _ = dst.Write([]byte("class " + toClassName(typ.Name.Name) + "Route extends ServerRoute"))

	_, _ = dst.Write([]byte("{\n"))
	_, _ = dst.Write([]byte("  final " + toClassName(typ.Name.Name) + " server;\n\n"))
	_, _ = dst.Write([]byte("  " + toClassName(typ.Name.Name) + "Route(this.server);\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  String get name => \"" + toClassName(typ.Name.Name) + "\";\n\n"))
	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  int get id => " + typ.Id.Value + ";\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  ByteData invokeData(int id, ByteData data) {\n"))
	_, _ = dst.Write([]byte("    switch (id) {\n"))
	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		_, _ = dst.Write([]byte("      case " + server.Id.Value + " << 32 | " + method.Id.Value + " :\n"))
		_, _ = dst.Write([]byte("        return server." + toFieldName(method.Name.Name) + "("))
		printType(dst, method.Param.Type(), false)
		_, _ = dst.Write([]byte(".fromData(data)!).toData();\n"))
		return nil
	})
	_, _ = dst.Write([]byte("    }\n"))
	_, _ = dst.Write([]byte("    return ByteData(0);\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  Map<String, dynamic> invokeMap(String name, Map<String, dynamic> map) {\n"))
	_, _ = dst.Write([]byte("    switch (name) {\n"))
	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		_, _ = dst.Write([]byte("      case \"" + server.Name.Name + "/" + method.Name.Name + "\":\n"))
		_, _ = dst.Write([]byte("        return server." + toFieldName(method.Name.Name) + "("))
		printType(dst, method.Param.Type(), false)
		_, _ = dst.Write([]byte(".fromMap(map)!).toMap();\n"))
		return nil
	})
	_, _ = dst.Write([]byte("    }\n"))
	_, _ = dst.Write([]byte("    return {};\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("}\n\n"))
}
