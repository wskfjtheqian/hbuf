package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printServerCode(dst *build.Writer, typ *ast.ServerType) {
	dst.Import("hbuf_ts", "h")

	b.printServer(dst, typ)
	b.printServerImp(dst, typ)
	//b.printServerRouter(dst, typ)

}

func (b *Builder) printServer(dst *build.Writer, typ *ast.ServerType) {
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("///" + typ.Doc.Text())
	}
	dst.Code("export interface " + build.StringToHumpName(typ.Name.Name))
	if nil != typ.Extends {
		dst.Code(" extends ")
		b.printExtend(dst, typ.Extends, false)
	}
	dst.Code(" {\n")
	for _, method := range typ.Methods {
		if nil != method.Doc && 0 < len(method.Doc.Text()) {
			dst.Code("\t//" + method.Doc.Text())
		}

		dst.Code("\t" + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		dst.Code(build.StringToFirstLower(method.ParamName.Name) + ": ")
		b.printType(dst, method.Param, false)

		dst.Code(", ctx?: h.Context): ")
		dst.Code("Promise<")
		b.printType(dst, method.Result.Type(), false)
		dst.Code(">\n\n")
	}
	dst.Code("}\n\n")
}

func (b *Builder) printServerImp(dst *build.Writer, typ *ast.ServerType) {
	dst.Code("export class " + build.StringToHumpName(typ.Name.Name) + "Client extends h.ServerClient implements ")
	dst.Code(b.getPackage(dst, typ.Name, ""))
	dst.Code(".")
	dst.Code(build.StringToHumpName(typ.Name.Name))

	dst.Code("{\n")

	dst.Code("\tconstructor(client: h.Client){\n")
	dst.Code("\t\tsuper(client)\n")
	dst.Code("\t}\n")

	dst.Code("\tget name(): string {\n")
	dst.Code("\t\treturn \"" + build.StringToUnderlineName(typ.Name.Name) + "\"\n")
	dst.Code("\t}\n\n")

	dst.Code("\tget id(): number {\n")
	dst.Code("\t\treturn 0\t\n")
	dst.Code("\t}\n\n")

	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		if nil != method.Doc && 0 < len(method.Doc.Text()) {
			dst.Code("\t//" + method.Doc.Text())
		}

		dst.Code("\t" + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		dst.Code(build.StringToFirstLower(method.ParamName.Name) + ": ")
		b.printType(dst, method.Param, false)

		dst.Code(", ctx?: h.Context): ")
		dst.Code("Promise<")
		b.printType(dst, method.Result.Type(), false)
		dst.Code("> {\n")

		dst.Code("\t\treturn this.invoke<")
		b.printType(dst, method.Result.Type(), false)
		dst.Code(">(\"")
		dst.Code(build.StringToUnderlineName(server.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name))
		dst.Code("\", ")
		dst.Code("0 << 32 | " + method.Id.Value)
		dst.Code(", ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", ")
		b.printType(dst, method.Result.Type(), false)
		dst.Code(".fromJson, ")
		b.printType(dst, method.Result.Type(), false)
		dst.Code(".fromData);\n")

		dst.Code("\t}\n\n")
		return nil
	})
	dst.Code("}\n\n")
}

func (b *Builder) printServerRouter(dst *build.Writer, typ *ast.ServerType) {
	dst.Code("class " + build.StringToHumpName(typ.Name.Name) + "Router extends h.ServerRouter")

	dst.Code("{\n")
	dst.Code("\tfinal " + build.StringToHumpName(typ.Name.Name) + " server;\n\n")

	dst.Code("\t@override\n")
	dst.Code("\tString get name => \"" + build.StringToUnderlineName(typ.Name.Name) + "\";\n\n")

	dst.Code("\t@override\n")
	//TODO dst.Code("  int get id => " + typ.Id.Value + ";\n\n")
	dst.Code("\tint get id => 0;\n\n")

	dst.Code("\tMap<String, ServerInvoke> _invokeNames = {};\n\n")

	dst.Code("\tMap<int, ServerInvoke> _invokeIds = {};\n\n")

	dst.Code("\t@override\n")
	dst.Code("\tMap<String, ServerInvoke> get invokeNames => _invokeNames;\n\n")

	dst.Code("\t@override\n")
	dst.Code("\tMap<int, ServerInvoke> get invokeIds => _invokeIds;\n\n")

	dst.Code("\t" + build.StringToHumpName(typ.Name.Name) + "Router(this.server){\n")
	dst.Code("\t\t_invokeNames = {\n")
	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Code("\t\t\t\"" + build.StringToUnderlineName(server.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\": ServerInvoke(\n")
		dst.Code("\t\t\t\ttoData: (List<int> buf) async {\n")
		dst.Code("\t\t\t\t\treturn ")
		b.printType(dst, method.Param.Type(), false)
		dst.Code(".fromMap(json.decode(utf8.decode(buf)));\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tformData: (Data data) async {\n")
		dst.Code("\t\t\t\t\t return utf8.encode(json.encode(data.toMap()));\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tinvoke: (Context ctx, Data data) async {\n")
		dst.Code("\t\t\t\t\t return await server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
		b.printType(dst, method.Param.Type(), false)
		dst.Code(", ctx);\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t),\n")
		return nil
	})
	dst.Code("\t\t};\n\n")

	dst.Code("\t\t_invokeIds = {\n")
	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		//dst.Code("        " + server.Id.Value + " << 32 | " + method.Id.Value + ": ServerInvoke(\n")
		dst.Code("\t\t\t\t0 << 32 | " + method.Id.Value + ": ServerInvoke(\n")
		dst.Code("\t\t\t\ttoData: (List<int> buf) async {\n")
		dst.Code("\t\t\t\t\treturn ")
		b.printType(dst, method.Param.Type(), false)
		dst.Code(".fromData(ByteData.view(Uint8List.fromList(buf).buffer));\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tformData: (Data data) async {\n")
		dst.Code("\t\t\t\t\t return data.toData().buffer.asUint8List();\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tinvoke: (Context ctx, Data data) async {\n")
		dst.Code("\t\t\t\t\t return await server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
		b.printType(dst, method.Param.Type(), false)
		dst.Code(", ctx);\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t),\n")
		return nil
	})
	dst.Code("\t\t};\n\n")

	dst.Code("\t}\n\n")

	dst.Code("}\n\n")
}
