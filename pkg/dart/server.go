package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
)

func printServer(dst io.Writer, typ *ast.ServerType) {
	_, _ = dst.Write([]byte("abstract class " + build.StringToHumpName(typ.Name.Name)))
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
		_, _ = dst.Write([]byte("  Future<"))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte("> " + build.StringToFirstLower(method.Name.Name)))
		_, _ = dst.Write([]byte("("))
		printType(dst, method.Param, false)
		_, _ = dst.Write([]byte(" " + build.StringToFirstLower(method.ParamName.Name)))
		_, _ = dst.Write([]byte(", [Context? ctx]);\n\n"))
	}
	_, _ = dst.Write([]byte("}\n\n"))
}

func printServerImp(dst io.Writer, typ *ast.ServerType) {
	_, _ = dst.Write([]byte("class " + build.StringToHumpName(typ.Name.Name) + "Client extends ServerClient implements " + build.StringToHumpName(typ.Name.Name)))

	_, _ = dst.Write([]byte("{\n"))

	_, _ = dst.Write([]byte("  " + build.StringToHumpName(typ.Name.Name) + "Client(Client client):super(client);\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  String get name => \"" + build.StringToHumpName(typ.Name.Name) + "\";\n\n"))
	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  int get id => " + typ.Id.Value + ";\n\n"))

	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		_, _ = dst.Write([]byte("  @override\n"))
		_, _ = dst.Write([]byte("  Future<"))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte("> " + build.StringToFirstLower(method.Name.Name)))
		_, _ = dst.Write([]byte("("))
		printType(dst, method.Param, false)
		_, _ = dst.Write([]byte(" " + build.StringToFirstLower(method.ParamName.Name)))
		_, _ = dst.Write([]byte(", [Context? ctx]){\n"))

		_, _ = dst.Write([]byte("    return invoke<"))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte(">(\""))
		_, _ = dst.Write([]byte(server.Name.Name + "/" + method.Name.Name))
		_, _ = dst.Write([]byte("\", "))
		_, _ = dst.Write([]byte(server.Id.Value + " << 32 | " + method.Id.Value))
		_, _ = dst.Write([]byte(", "))
		_, _ = dst.Write([]byte(build.StringToFirstLower(method.ParamName.Name)))
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

func printServerRouter(dst io.Writer, typ *ast.ServerType) {
	_, _ = dst.Write([]byte("class " + build.StringToHumpName(typ.Name.Name) + "Router extends ServerRouter"))

	_, _ = dst.Write([]byte("{\n"))
	_, _ = dst.Write([]byte("  final " + build.StringToHumpName(typ.Name.Name) + " server;\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  String get name => \"" + build.StringToHumpName(typ.Name.Name) + "\";\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  int get id => " + typ.Id.Value + ";\n\n"))

	_, _ = dst.Write([]byte("  Map<String, ServerInvoke> _invokeNames = {};\n\n"))

	_, _ = dst.Write([]byte("  Map<int, ServerInvoke> _invokeIds = {};\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  Map<String, ServerInvoke> get invokeNames => _invokeNames;\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  Map<int, ServerInvoke> get invokeIds => _invokeIds;\n\n"))

	_, _ = dst.Write([]byte("  " + build.StringToHumpName(typ.Name.Name) + "Router(this.server){\n"))
	_, _ = dst.Write([]byte("    _invokeNames = {\n"))
	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		_, _ = dst.Write([]byte("      \"" + server.Name.Name + "/" + method.Name.Name + "\": ServerInvoke(\n"))
		_, _ = dst.Write([]byte("        toData: (List<int> buf) async {\n"))
		_, _ = dst.Write([]byte("          return "))
		printType(dst, method.Param.Type(), false)
		_, _ = dst.Write([]byte(".fromMap(json.decode(utf8.decode(buf)));\n"))
		_, _ = dst.Write([]byte("        },\n"))
		_, _ = dst.Write([]byte("        formData: (Data data) async {\n"))
		_, _ = dst.Write([]byte("     	   return utf8.encode(json.encode(data.toMap()));\n"))
		_, _ = dst.Write([]byte("        },\n"))
		_, _ = dst.Write([]byte("        invoke: (Context ctx, Data data) async {\n"))
		_, _ = dst.Write([]byte("     	   return await server." + build.StringToFirstLower(method.Name.Name) + "(data as "))
		printType(dst, method.Param.Type(), false)
		_, _ = dst.Write([]byte(", ctx);\n"))
		_, _ = dst.Write([]byte("        },\n"))
		_, _ = dst.Write([]byte("      ),\n"))
		return nil
	})
	_, _ = dst.Write([]byte("    };\n\n"))

	_, _ = dst.Write([]byte("    _invokeIds = {\n"))
	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		_, _ = dst.Write([]byte("        " + server.Id.Value + " << 32 | " + method.Id.Value + ": ServerInvoke(\n"))
		_, _ = dst.Write([]byte("        toData: (List<int> buf) async {\n"))
		_, _ = dst.Write([]byte("          return "))
		printType(dst, method.Param.Type(), false)
		_, _ = dst.Write([]byte(".fromData(ByteData.view(Uint8List.fromList(buf).buffer));\n"))
		_, _ = dst.Write([]byte("        },\n"))
		_, _ = dst.Write([]byte("        formData: (Data data) async {\n"))
		_, _ = dst.Write([]byte("     	   return data.toData().buffer.asUint8List();\n"))
		_, _ = dst.Write([]byte("        },\n"))
		_, _ = dst.Write([]byte("        invoke: (Context ctx, Data data) async {\n"))
		_, _ = dst.Write([]byte("     	   return await server." + build.StringToFirstLower(method.Name.Name) + "(data as "))
		printType(dst, method.Param.Type(), false)
		_, _ = dst.Write([]byte(", ctx);\n"))
		_, _ = dst.Write([]byte("        },\n"))
		_, _ = dst.Write([]byte("      ),\n"))
		return nil
	})
	_, _ = dst.Write([]byte("    };\n\n"))

	_, _ = dst.Write([]byte("  }\n\n"))

	//
	//_, _ = dst.Write([]byte("  @override\n"))
	//_, _ = dst.Write([]byte("  ByteData invokeData(int id, ByteData data) {\n"))
	//_, _ = dst.Write([]byte("    switch (id) {\n"))
	//_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
	//	_, _ = dst.Write([]byte("      case " + server.Id.Value + " << 32 | " + method.Id.Value + " :\n"))
	//	_, _ = dst.Write([]byte("        return server." + build.StringToFirstLower(method.Name.Name) + "("))
	//	printType(dst, method.Param.Type(), false)
	//	_, _ = dst.Write([]byte(".fromData(data)!).toData();\n"))
	//	return nil
	//})
	//_, _ = dst.Write([]byte("    }\n"))
	//_, _ = dst.Write([]byte("    return ByteData(0);\n"))
	//_, _ = dst.Write([]byte("  }\n\n"))
	//
	//_, _ = dst.Write([]byte("  @override\n"))
	//_, _ = dst.Write([]byte("  Map<String, dynamic> invokeMap(String name, Map<String, dynamic> map) {\n"))
	//_, _ = dst.Write([]byte("    switch (name) {\n"))
	//
	//_, _ = dst.Write([]byte("    }\n"))
	//_, _ = dst.Write([]byte("    return {};\n"))
	//_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("}\n\n"))
}
