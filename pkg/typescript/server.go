package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printServerCode(dst *build.Writer, typ *ast.ServerType) {
	dst.Import("hbuf_ts", "* as h")

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
			dst.Tab(1).Code("//" + method.Doc.Text())
		}

		resultType := method.Result.Type().(*ast.Ident).Name
		paramType := method.Param.Type().(*ast.Ident).Name

		dst.Tab(1).Code("" + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		dst.Code(build.StringToFirstLower(method.ParamName.Name) + ": ")

		if paramType == "stream" {
			dst.Code("Blob | ArrayBuffer")
		} else {
			b.printType(dst, method.Param, false, false)
		}

		dst.Code(", ctx?: h.Context): ")
		dst.Code("Promise<")
		if resultType == "void" {
			dst.Code("void")
		} else if resultType == "stream" {
			dst.Code("Blob | ArrayBuffer")
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

	dst.Tab(1).Code("constructor(client: h.Client){\n")
	dst.Tab(2).Code("super(client)\n")
	dst.Tab(1).Code("}\n")

	dst.Tab(1).Code("get name(): string {\n")
	dst.Tab(2).Code("return \"" + build.StringToUnderlineName(typ.Name.Name) + "\"\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("get id(): number {\n")
	dst.Tab(2).Code("return 0\t\n")
	dst.Tab(1).Code("}\n\n")

	_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		if nil != method.Doc && 0 < len(method.Doc.Text()) {
			dst.Tab(1).Code("//" + method.Doc.Text())
		}
		resultType := method.Result.Type().(*ast.Ident).Name
		paramType := method.Param.Type().(*ast.Ident).Name

		dst.Tab(1).Code("" + build.StringToFirstLower(method.Name.Name))
		dst.Code("(")
		dst.Code(build.StringToFirstLower(method.ParamName.Name) + ": ")
		if paramType == "stream" {
			dst.Code("Blob | ArrayBuffer")
		} else {
			b.printType(dst, method.Param, false, false)
		}

		dst.Code(", ctx?: h.Context): ")
		dst.Code("Promise<")
		if resultType == "void" {
			dst.Code("void")
		} else if resultType == "stream" {
			dst.Code("Blob | ArrayBuffer")
		} else {
			b.printType(dst, method.Result.Type(), false, false)
		}

		dst.Code("> {\n")

		dst.Tab(2).Code("return this.invoke<")
		if resultType == "void" {
			dst.Code("void")
		} else if resultType == "stream" {
			dst.Code("Blob | ArrayBuffer")
		} else {
			b.printType(dst, method.Result.Type(), false, false)
		}
		dst.Code(">(\"")
		dst.Code(build.StringToUnderlineName(method.Name.Name))
		dst.Code("\", ")
		dst.Code("0 << 32 | " + method.Id.Value)
		dst.Code(", ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", ")
		if resultType == "void" || resultType == "stream" {
			dst.Code("null, null);\n")
		} else {
			b.printType(dst, method.Result.Type(), false, false)
			dst.Code(".fromJson, ")
			b.printType(dst, method.Result.Type(), false, false)
			dst.Code(".fromData);\n")
		}

		dst.Tab(1).Code("}\n\n")
		return nil
	})
	dst.Code("}\n\n")
}

func (b *Builder) printServerRouter(dst *build.Writer, typ *ast.ServerType) {

	dst.Code("export class " + build.StringToHumpName(typ.Name.Name) + "Router implements h.ServerRouter {\n")
	dst.Tab(1).Code("readonly server: " + build.StringToHumpName(typ.Name.Name) + "\n")
	dst.Code("\n")
	dst.Tab(1).Code("invoke: Record<string, h.ServerInvoke>\n")
	dst.Code("\n")
	dst.Tab(1).Code("getInvoke(): Record<string, h.ServerInvoke> {\n")
	dst.Tab(2).Code("return this.invoke\n")
	dst.Tab(1).Code("}\n")
	dst.Code("\n")
	dst.Tab(1).Code("getName(): string {\n")
	dst.Tab(2).Code("return \"" + build.StringToUnderlineName(typ.Name.Name) + "\"\n")
	dst.Tab(1).Code("}\n")
	dst.Code("\n")
	dst.Tab(1).Code("getId(): number {\n")
	dst.Tab(2).Code("return 0\n")
	dst.Tab(1).Code("}\n")
	dst.Code("\n")
	dst.Tab(1).Code("constructor(server: " + build.StringToHumpName(typ.Name.Name) + ") {\n")
	dst.Tab(2).Code("this.server = server\n")
	dst.Tab(2).Code("this.invoke = {\n")
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {

		paramType := method.Param.Type().(*ast.Ident).Name

		dst.Tab(3).Code("\"" + build.StringToUnderlineName(method.Name.Name) + "\": {\n")
		dst.Tab(4).Code("formData(data: Blob | ArrayBuffer | Record<string, any>):  Blob | ArrayBuffer | h.Data {\n")
		dst.Tab(5).Code("return ")
		if paramType == "void" || paramType == "stream" {
			dst.Code("data as ArrayBuffer |h.Data\n")
		} else {
			b.printType(dst, method.Param.Type(), false, false)
			dst.Code(".fromJson(data)\n")
		}

		dst.Tab(4).Code("},\n")
		dst.Tab(4).Code("toData(data: Blob | ArrayBuffer | h.Data): Blob | ArrayBuffer | Record<string, any> {\n")
		dst.Tab(5).Code("return ")
		if paramType == "void" || paramType == "stream" {
			dst.Code("data as Blob | ArrayBuffer\n")
		} else {
			dst.Code("(data as h.Data).toJson()\n")
		}

		dst.Tab(4).Code("},\n")
		dst.Tab(4).Code("invoke(data: Blob | ArrayBuffer | h.Data, ctx?: h.Context): Promise<h.Data | void> {\n")
		dst.Tab(5).Code("return server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
		if paramType == "void" || paramType == "stream" {
			dst.Code("Blob | ArrayBuffer")
		} else {
			b.printType(dst, method.Param.Type(), false, false)
		}
		dst.Code(", ctx);\n")
		dst.Tab(4).Code("}\n")
		dst.Tab(3).Code("},\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Tab(2).Code("}\n")
	dst.Tab(1).Code("}\n")
	dst.Code("}\n")
}
