package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printServerCode(dst *Writer, typ *ast.ServerType) {
	dst.Import("dart:convert")
	dst.Import("dart:typed_data")
	dst.Import("package:hbuf_dart/hbuf_dart.dart")

	b.printServer(dst, typ)
	b.printServerImp(dst, typ)
	b.printServerRouter(dst, typ)
}

func (b *Builder) printServer(dst *Writer, typ *ast.ServerType) {
	dst.Code("abstract class " + build.StringToHumpName(typ.Name.Name))
	if nil != typ.Extends {
		dst.Code(" implements ")
		b.printExtend(dst, typ.Extends, false)
	}
	dst.Code("{\n")
	for _, method := range typ.Methods {
		if nil != method.Comment {
			dst.Code("  /// " + method.Comment.Text())
		}
		if build.CheckSuperMethod(method.Name.Name, typ) {
			dst.Code("  @override\n")
		}
		dst.Code("  Future<")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("> " + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		b.printType(dst, method.Param, false)
		dst.Code(" " + build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", [Context? ctx]);\n\n")
	}
	dst.Code("}\n\n")
}

func (b *Builder) printServerImp(dst *Writer, typ *ast.ServerType) {
	dst.Code("class " + build.StringToHumpName(typ.Name.Name) + "Client extends ServerClient implements " + build.StringToHumpName(typ.Name.Name))

	dst.Code("{\n")

	dst.Code("  " + build.StringToHumpName(typ.Name.Name) + "Client(Client client):super(client);\n\n")

	dst.Code("  @override\n")
	dst.Code("  String get name => \"" + build.StringToUnderlineName(typ.Name.Name) + "\";\n\n")
	dst.Code("  @override\n")
	dst.Code("  int get id => " + typ.Id.Value + ";\n\n")

	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Code("  @override\n")
		dst.Code("  Future<")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("> " + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		b.printType(dst, method.Param, false)
		dst.Code(" " + build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", [Context? ctx]){\n")

		dst.Code("    return invoke<")
		b.printType(dst, method.Result.Type(), false)
		dst.Code(">(\"")
		dst.Code(build.StringToUnderlineName(server.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name))
		dst.Code("\", ")
		dst.Code(server.Id.Value + " << 32 | " + method.Id.Value)
		dst.Code(", ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", ")
		b.printType(dst, method.Result.Type(), false)
		dst.Code(".fromMap, ")
		b.printType(dst, method.Result.Type(), false)
		dst.Code(".fromData);\n")

		dst.Code("  }\n\n")
		return nil
	})
	dst.Code("}\n\n")
}

func (b *Builder) printServerRouter(dst *Writer, typ *ast.ServerType) {
	dst.Code("class " + build.StringToHumpName(typ.Name.Name) + "Router extends ServerRouter")

	dst.Code("{\n")
	dst.Code("  final " + build.StringToHumpName(typ.Name.Name) + " server;\n\n")

	dst.Code("  @override\n")
	dst.Code("  String get name => \"" + build.StringToUnderlineName(typ.Name.Name) + "\";\n\n")

	dst.Code("  @override\n")
	dst.Code("  int get id => " + typ.Id.Value + ";\n\n")

	dst.Code("  Map<String, ServerInvoke> _invokeNames = {};\n\n")

	dst.Code("  Map<int, ServerInvoke> _invokeIds = {};\n\n")

	dst.Code("  @override\n")
	dst.Code("  Map<String, ServerInvoke> get invokeNames => _invokeNames;\n\n")

	dst.Code("  @override\n")
	dst.Code("  Map<int, ServerInvoke> get invokeIds => _invokeIds;\n\n")

	dst.Code("  " + build.StringToHumpName(typ.Name.Name) + "Router(this.server){\n")
	dst.Code("    _invokeNames = {\n")
	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Code("      \"" + build.StringToUnderlineName(server.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\": ServerInvoke(\n")
		dst.Code("        toData: (List<int> buf) async {\n")
		dst.Code("          return ")
		b.printType(dst, method.Param.Type(), false)
		dst.Code(".fromMap(json.decode(utf8.decode(buf)));\n")
		dst.Code("        },\n")
		dst.Code("        formData: (Data data) async {\n")
		dst.Code("     	   return utf8.encode(json.encode(data.toMap()));\n")
		dst.Code("        },\n")
		dst.Code("        invoke: (Context ctx, Data data) async {\n")
		dst.Code("     	   return await server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
		b.printType(dst, method.Param.Type(), false)
		dst.Code(", ctx);\n")
		dst.Code("        },\n")
		dst.Code("      ),\n")
		return nil
	})
	dst.Code("    };\n\n")

	dst.Code("    _invokeIds = {\n")
	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Code("        " + server.Id.Value + " << 32 | " + method.Id.Value + ": ServerInvoke(\n")
		dst.Code("        toData: (List<int> buf) async {\n")
		dst.Code("          return ")
		b.printType(dst, method.Param.Type(), false)
		dst.Code(".fromData(ByteData.view(Uint8List.fromList(buf).buffer));\n")
		dst.Code("        },\n")
		dst.Code("        formData: (Data data) async {\n")
		dst.Code("     	   return data.toData().buffer.asUint8List();\n")
		dst.Code("        },\n")
		dst.Code("        invoke: (Context ctx, Data data) async {\n")
		dst.Code("     	   return await server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
		b.printType(dst, method.Param.Type(), false)
		dst.Code(", ctx);\n")
		dst.Code("        },\n")
		dst.Code("      ),\n")
		return nil
	})
	dst.Code("    };\n\n")

	dst.Code("  }\n\n")

	//
	//dst.Code("  @override\n")
	//dst.Code("  ByteData invokeData(int id, ByteData data) {\n")
	//dst.Code("    switch (id) {\n")
	//_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
	//	dst.Code("      case " + server.Id.Value + " << 32 | " + method.Id.Value + " :\n")
	//	dst.Code("        return server." + build.StringToFirstLower(method.Name.Name) + "(")
	//	printType(dst, method.Param.Type(), false)
	//	dst.Code(".fromData(data)!).toData();\n")
	//	return nil
	//})
	//dst.Code("    }\n")
	//dst.Code("    return ByteData(0);\n")
	//dst.Code("  }\n\n")
	//
	//dst.Code("  @override\n")
	//dst.Code("  Map<String, dynamic> invokeMap(String name, Map<String, dynamic> map) {\n")
	//dst.Code("    switch (name) {\n")
	//
	//dst.Code("    }\n")
	//dst.Code("    return {};\n")
	//dst.Code("  }\n\n")

	dst.Code("}\n\n")
}
