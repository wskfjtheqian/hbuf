package java

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printServerCode(dst *build.Writer, typ *ast.ServerType) {
	dst.Import("java.util.concurrent.CompletableFuture", "")
	dst.Import("com.hbuf.java.Data", "")
	dst.Import("com.hbuf.java.Server", "")

	b.printServer(dst, typ)
	b.printServerClient(dst, typ)
	b.printServerRouter(dst, typ)

}

func (b *Builder) printServer(dst *build.Writer, typ *ast.ServerType) {
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("///" + typ.Doc.Text())
	}
	dst.Code("\tinterface " + build.StringToHumpName(typ.Name.Name))
	if nil != typ.Extends {
		dst.Code(" extends ")
		b.printExtend(dst, typ.Extends, false)
	}
	dst.Code("{\n")
	for _, method := range typ.Methods {
		if nil != method.Doc && 0 < len(method.Doc.Text()) {
			dst.Code("///" + method.Doc.Text())
		}
		if build.CheckSuperMethod(method.Name.Name, typ) {
			dst.Code("\t@override\n")
		}
		dst.Code("\t\tCompletableFuture<")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("> " + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		b.printType(dst, method.Param, false)
		dst.Code(" " + build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", Server.Context ctx) throws Exception ;\n\n")
	}
	dst.Code("\t}\n\n")
}

func (b *Builder) printServerClient(dst *build.Writer, typ *ast.ServerType) {
	dst.Code("\tclass " + build.StringToHumpName(typ.Name.Name) + "Client extends Server.ClientRouter implements " + build.StringToHumpName(typ.Name.Name))

	dst.Code("{\n")

	dst.Code("\t\tpublic " + build.StringToHumpName(typ.Name.Name) + "Client(Server.Client client) {\n")
	dst.Code("\t\t\tsuper(client);\n")
	dst.Code("\t\t}\n\n")

	dst.Code("\t\t@Override\n")
	dst.Code("\t\tpublic String getName() {\n")
	dst.Code("\t\t\treturn \"" + build.StringToUnderlineName(typ.Name.Name) + "\";\n")
	dst.Code("\t\t}\n\n")
	dst.Code("\t\t@Override\n")
	dst.Code("\t\tpublic long getId() {\n")
	dst.Code("\t\t\treturn " + typ.Id.Value + ";\n")
	dst.Code("\t\t}\n\n")

	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Code("\t\t@Override\n")
		dst.Code("\t\tpublic CompletableFuture<")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("> " + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		b.printType(dst, method.Param, false)
		dst.Code(" " + build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", Server.Context ctx) throws Exception {\n")

		dst.Code("\t\t\treturn invoke(\"")
		dst.Code(build.StringToUnderlineName(server.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name))
		dst.Code("\", ")
		dst.Code(server.Id.Value + " << 32 | " + method.Id.Value)
		dst.Code(", ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", (data) -> Data.formJson.invoke(new String(data), ")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("Impl.class), (data)-> new ")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("Impl().formData(data));\n")

		dst.Code("\t\t}\n\n")
		return nil
	})
	dst.Code("\t}\n\n")
}

func (b *Builder) printServerRouter(dst *build.Writer, typ *ast.ServerType) {
	//dst.Code("class " + build.StringToHumpName(typ.Name.Name) + "Router extends ServerRouter")
	//
	//dst.Code("{\n")
	//dst.Code("\tfinal " + build.StringToHumpName(typ.Name.Name) + " server;\n\n")
	//
	//dst.Code("\t@override\n")
	//dst.Code("\tString get name => \"" + build.StringToUnderlineName(typ.Name.Name) + "\";\n\n")
	//
	//dst.Code("\t@override\n")
	//dst.Code("\tint get id => " + typ.Id.Values + ";\n\n")
	//
	//dst.Code("\tMap<String, ServerInvoke> _invokeNames = {};\n\n")
	//
	//dst.Code("\tMap<int, ServerInvoke> _invokeIds = {};\n\n")
	//
	//dst.Code("\t@override\n")
	//dst.Code("\tMap<String, ServerInvoke> get invokeNames => _invokeNames;\n\n")
	//
	//dst.Code("\t@override\n")
	//dst.Code("\tMap<int, ServerInvoke> get invokeIds => _invokeIds;\n\n")
	//
	//dst.Code("\t" + build.StringToHumpName(typ.Name.Name) + "Router(this.server){\n")
	//dst.Code("\t\t_invokeNames = {\n")
	//_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
	//	dst.Code("\t\t\t\"" + build.StringToUnderlineName(server.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\": ServerInvoke(\n")
	//	dst.Code("\t\t\t\ttoData: (List<int> buf) async {\n")
	//	dst.Code("\t\t\t\t\treturn ")
	//	b.printType(dst, method.Param.Type(), false)
	//	dst.Code(".fromMap(json.decode(utf8.decode(buf)));\n")
	//	dst.Code("\t\t\t\t},\n")
	//	dst.Code("\t\t\t\tformData: (Data data) async {\n")
	//	dst.Code("\t\t\t\t\t return utf8.encode(json.encode(data.toMap()));\n")
	//	dst.Code("\t\t\t\t},\n")
	//	dst.Code("\t\t\t\tinvoke: (Context ctx, Data data) async {\n")
	//	dst.Code("\t\t\t\t\t return await server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
	//	b.printType(dst, method.Param.Type(), false)
	//	dst.Code(", ctx);\n")
	//	dst.Code("\t\t\t\t},\n")
	//	dst.Code("\t\t\t),\n")
	//	return nil
	//})
	//dst.Code("\t\t};\n\n")
	//
	//dst.Code("\t\t_invokeIds = {\n")
	//_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
	//	dst.Code("\t\t\t\t" + server.Id.Values + " << 32 | " + method.Id.Values + ": ServerInvoke(\n")
	//	dst.Code("\t\t\t\ttoData: (List<int> buf) async {\n")
	//	dst.Code("\t\t\t\t\treturn ")
	//	b.printType(dst, method.Param.Type(), false)
	//	dst.Code(".fromData(ByteData.view(Uint8List.fromList(buf).buffer));\n")
	//	dst.Code("\t\t\t\t},\n")
	//	dst.Code("\t\t\t\tformData: (Data data) async {\n")
	//	dst.Code("\t\t\t\t\t return data.toData().buffer.asUint8List();\n")
	//	dst.Code("\t\t\t\t},\n")
	//	dst.Code("\t\t\t\tinvoke: (Context ctx, Data data) async {\n")
	//	dst.Code("\t\t 	\t return await server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
	//	b.printType(dst, method.Param.Type(), false)
	//	dst.Code(", ctx);\n")
	//	dst.Code("\t\t\t\t},\n")
	//	dst.Code("\t\t\t),\n")
	//	return nil
	//})
	//dst.Code("\t\t};\n\n")
	//
	//dst.Code("\t}\n\n")
	//
	////
	////dst.Code("  @override\n")
	////dst.Code("  ByteData invokeData(int id, ByteData data) {\n")
	////dst.Code("    switch (id) {\n")
	////_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
	////	dst.Code("      case " + server.Id.Values + " << 32 | " + method.Id.Values + " :\n")
	////	dst.Code("        return server." + build.StringToFirstLower(method.Name.Name) + "(")
	////	printType(dst, method.Param.Type(), false)
	////	dst.Code(".fromData(data)!).toData();\n")
	////	return nil
	////})
	////dst.Code("    }\n")
	////dst.Code("    return ByteData(0);\n")
	////dst.Code("  }\n\n")
	////
	////dst.Code("  @override\n")
	////dst.Code("  Map<String, dynamic> invokeMap(String name, Map<String, dynamic> map) {\n")
	////dst.Code("    switch (name) {\n")
	////
	////dst.Code("    }\n")
	////dst.Code("    return {};\n")
	////dst.Code("  }\n\n")
	//
	//dst.Code("}\n\n")
}
