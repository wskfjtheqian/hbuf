package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printServerCode(dst *build.Writer, typ *ast.ServerType) {
	dst.Import("hbuf_ts", "h")

	b.printServer(dst, typ)
	b.printServerImp(dst, typ)
	b.printServerRouter(dst, typ)

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
		isMethod := method.Result.Type().(*ast.Ident).Name == "void"

		dst.Code("\t" + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		dst.Code(build.StringToFirstLower(method.ParamName.Name) + ": ")
		b.printType(dst, method.Param, false, false)

		dst.Code(", ctx?: h.Context): ")
		dst.Code("Promise<")
		if isMethod {
			dst.Code("void")
		} else {
			b.printType(dst, method.Result.Type(), false, false)
		}
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
		isMethod := method.Result.Type().(*ast.Ident).Name == "void"

		dst.Code("\t" + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		dst.Code(build.StringToFirstLower(method.ParamName.Name) + ": ")
		b.printType(dst, method.Param, false, false)

		dst.Code(", ctx?: h.Context): ")
		dst.Code("Promise<")
		if isMethod {
			dst.Code("void")
		} else {
			b.printType(dst, method.Result.Type(), false, false)
		}

		dst.Code("> {\n")

		dst.Code("\t\treturn this.invoke<")
		if isMethod {
			dst.Code("void")
		} else {
			b.printType(dst, method.Result.Type(), false, false)
		}
		dst.Code(">(\"")
		dst.Code(build.StringToUnderlineName(server.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name))
		dst.Code("\", ")
		dst.Code("0 << 32 | " + method.Id.Value)
		dst.Code(", ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", ")
		if isMethod {
			dst.Code("null, null);\n")
		} else {
			b.printType(dst, method.Result.Type(), false, false)
			dst.Code(".fromJson, ")
			b.printType(dst, method.Result.Type(), false, false)
			dst.Code(".fromData);\n")
		}

		dst.Code("\t}\n\n")
		return nil
	})
	dst.Code("}\n\n")
}

func (b *Builder) printServerRouter(dst *build.Writer, typ *ast.ServerType) {

	dst.Code("export class " + build.StringToHumpName(typ.Name.Name) + "Router implements h.ServerRouter {\n")
	dst.Code("\treadonly server: " + build.StringToHumpName(typ.Name.Name) + "\n")
	dst.Code("\n")
	dst.Code("\tinvoke: Record<string, h.ServerInvoke>\n")
	dst.Code("\n")
	dst.Code("\tgetInvoke(): Record<string, h.ServerInvoke> {\n")
	dst.Code("\t\treturn this.invoke\n")
	dst.Code("\t}\n")
	dst.Code("\n")
	dst.Code("\tgetName(): string {\n")
	dst.Code("\t\treturn \"" + build.StringToUnderlineName(typ.Name.Name) + "\"\n")
	dst.Code("\t}\n")
	dst.Code("\n")
	dst.Code("\tgetId(): number {\n")
	dst.Code("\t\treturn 0\n")
	dst.Code("\t}\n")
	dst.Code("\n")
	dst.Code("\tconstructor(server: " + build.StringToHumpName(typ.Name.Name) + ") {\n")
	dst.Code("\t\tthis.server = server\n")
	dst.Code("\t\tthis.invoke = {\n")
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Code("\t\t\t\"" + build.StringToUnderlineName(server.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\": {\n")
		dst.Code("\t\t\t\tformData(data: BinaryData | Record<string, any>): h.Data {\n")
		dst.Code("\t\t\t\t\treturn ")
		b.printType(dst, method.Param.Type(), false, false)
		dst.Code(".fromJson(data)\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\ttoData(data: h.Data): BinaryData | Record<string, any> {\n")
		dst.Code("\t\t\t\t\treturn data.toJson()\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tinvoke(data: h.Data, ctx?: h.Context): Promise<h.Data | void> {\n")
		dst.Code("\t\t\t\t\treturn server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
		b.printType(dst, method.Param.Type(), false, false)
		dst.Code(", ctx);\n")
		dst.Code("\t\t\t\t}\n")
		dst.Code("\t\t\t},\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("\t\t}\n")
	dst.Code("\t}\n")
	dst.Code("}\n")
}
