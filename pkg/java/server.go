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
	dst.Tab(1).Code("interface " + build.StringToHumpName(typ.Name.Name))
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
			dst.Tab(1).Code("@override\n")
		}
		dst.Tab(2).Code("CompletableFuture<")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("> " + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		b.printType(dst, method.Param, false)
		dst.Code(" " + build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", Server.Context ctx) throws Exception ;\n\n")
	}
	dst.Tab(1).Code("}\n\n")
}

func (b *Builder) printServerClient(dst *build.Writer, typ *ast.ServerType) {
	dst.Tab(1).Code("class " + build.StringToHumpName(typ.Name.Name) + "Client extends Server.ClientRouter implements " + build.StringToHumpName(typ.Name.Name))

	dst.Code("{\n")

	dst.Tab(2).Code("public " + build.StringToHumpName(typ.Name.Name) + "Client(Server.Client client) {\n")
	dst.Tab(3).Code("super(client);\n")
	dst.Tab(2).Code("}\n\n")

	dst.Tab(2).Code("@Override\n")
	dst.Tab(2).Code("public String getName() {\n")
	dst.Tab(3).Code("return \"" + build.StringToUnderlineName(typ.Name.Name) + "\";\n")
	dst.Tab(2).Code("}\n\n")
	dst.Tab(2).Code("@Override\n")
	dst.Tab(2).Code("public long getId() {\n")
	//TODO dst.Tab(3).Code("return " + typ.Id.Value + ";\n")
	dst.Tab(3).Code("return 0;\n")
	dst.Tab(2).Code("}\n\n")

	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Tab(2).Code("@Override\n")
		dst.Tab(2).Code("public CompletableFuture<")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("> " + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		b.printType(dst, method.Param, false)
		dst.Code(" " + build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", Server.Context ctx) throws Exception {\n")

		dst.Tab(3).Code("return invoke(\"")
		dst.Code(build.StringToUnderlineName(server.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name))
		dst.Code("\", ")
		//TODO dst.Code(server.Id.Value + " << 32 | " + method.Id.Value)
		dst.Code("0 << 32 | " + method.Id.Value)
		dst.Code(", ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", (data) -> Data.formJson.invoke(new String(data), ")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("Impl.class), (data)-> new ")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("Impl().formData(data));\n")

		dst.Tab(2).Code("}\n\n")
		return nil
	})
	dst.Tab(1).Code("}\n\n")
}

func (b *Builder) printServerRouter(dst *build.Writer, typ *ast.ServerType) {
	//dst.Code("class " + build.StringToHumpName(typ.Name.Name) + "Router extends ServerRouter")
	//
	//dst.Code("{\n")
	//dst.Tab(1).Code("final " + build.StringToHumpName(typ.Name.Name) + " server;\n\n")
	//
	//dst.Tab(1).Code("@override\n")
	//dst.Tab(1).Code("String get name => \"" + build.StringToUnderlineName(typ.Name.Name) + "\";\n\n")
	//
	//dst.Tab(1).Code("@override\n")
	//dst.Tab(1).Code("int get id => " + typ.Id.Values + ";\n\n")
	//
	//dst.Tab(1).Code("Map<String, ServerInvoke> _invokeNames = {};\n\n")
	//
	//dst.Tab(1).Code("Map<int, ServerInvoke> _invokeIds = {};\n\n")
	//
	//dst.Tab(1).Code("@override\n")
	//dst.Tab(1).Code("Map<String, ServerInvoke> get invokeNames => _invokeNames;\n\n")
	//
	//dst.Tab(1).Code("@override\n")
	//dst.Tab(1).Code("Map<int, ServerInvoke> get invokeIds => _invokeIds;\n\n")
	//
	//dst.Tab(1).Code("" + build.StringToHumpName(typ.Name.Name) + "Router(this.server){\n")
	//dst.Tab(2).Code("_invokeNames = {\n")
	//_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
	//	dst.Tab(3).Code("\"" + build.StringToUnderlineName(server.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\": ServerInvoke(\n")
	//	dst.Tab(4).Code("toData: (List<int> buf) async {\n")
	//	dst.Tab(5).Code("return ")
	//	b.printType(dst, method.Param.Type(), false)
	//	dst.Code(".fromMap(json.decode(utf8.decode(buf)));\n")
	//	dst.Tab(4).Code("},\n")
	//	dst.Tab(4).Code("formData: (Data data) async {\n")
	//	dst.Tab(5).Code(" return utf8.encode(json.encode(data.toMap()));\n")
	//	dst.Tab(4).Code("},\n")
	//	dst.Tab(4).Code("invoke: (Context ctx, Data data) async {\n")
	//	dst.Tab(5).Code(" return await server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
	//	b.printType(dst, method.Param.Type(), false)
	//	dst.Code(", ctx);\n")
	//	dst.Tab(4).Code("},\n")
	//	dst.Tab(3).Code("),\n")
	//	return nil
	//})
	//dst.Tab(2).Code("};\n\n")
	//
	//dst.Tab(2).Code("_invokeIds = {\n")
	//_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
	//	dst.Tab(4).Code("" + server.Id.Values + " << 32 | " + method.Id.Values + ": ServerInvoke(\n")
	//	dst.Tab(4).Code("toData: (List<int> buf) async {\n")
	//	dst.Tab(5).Code("return ")
	//	b.printType(dst, method.Param.Type(), false)
	//	dst.Code(".fromData(ByteData.view(Uint8List.fromList(buf).buffer));\n")
	//	dst.Tab(4).Code("},\n")
	//	dst.Tab(4).Code("formData: (Data data) async {\n")
	//	dst.Tab(5).Code(" return data.toData().buffer.asUint8List();\n")
	//	dst.Tab(4).Code("},\n")
	//	dst.Tab(4).Code("invoke: (Context ctx, Data data) async {\n")
	//	dst.Tab(2).Code(" 	\t return await server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
	//	b.printType(dst, method.Param.Type(), false)
	//	dst.Code(", ctx);\n")
	//	dst.Tab(4).Code("},\n")
	//	dst.Tab(3).Code("),\n")
	//	return nil
	//})
	//dst.Tab(2).Code("};\n\n")
	//
	//dst.Tab(1).Code("}\n\n")
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
