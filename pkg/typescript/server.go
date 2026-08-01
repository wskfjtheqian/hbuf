package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printServerCode(dst *build.Writer, typ *ast.ServerType) {
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
			b.printType(dst, method.Param, false, false, true)
		}

		dst.Import("hbuf_ts", "type Option")
		dst.Code(", _opt?: Option): ")
		dst.Code("Promise<")
		if resultType == "void" {
			dst.Code("void")
		} else if resultType == "stream" {
			dst.Code("Blob | ArrayBuffer")
		} else {
			b.printType(dst, method.Result.Type(), false, false, true)
		}
		dst.Code(">\n\n")
	}
	dst.Code("}\n\n")
}

func (b *Builder) printServerImp(dst *build.Writer, typ *ast.ServerType) {
	dst.Code("export class " + build.StringToHumpName(typ.Name.Name) + "Client implements ")
	b.getPackage(dst, typ.Name, "", false, false, "", "")
	dst.Code(build.StringToHumpName(typ.Name.Name))

	dst.Code("{\n\n")

	dst.Import("hbuf_ts", "type Client")
	dst.Tab(1).Code("protected client: Client\n\n")

	dst.Tab(1).Code("constructor(client: Client){\n")
	dst.Tab(2).Code("this.client = client\n")
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
			b.printType(dst, method.Param, false, false, true)
		}

		dst.Import("hbuf_ts", "type Option")
		dst.Code(", _opt?: Option): ")
		dst.Code("Promise<")
		if resultType == "void" {
			dst.Code("void")
		} else if resultType == "stream" {
			dst.Code("Blob | ArrayBuffer")
		} else {
			b.printType(dst, method.Result.Type(), false, false, true)
		}

		dst.Code("> {\n")

		dst.Tab(2).Code("return this.client.invoke<")
		if resultType == "void" {
			dst.Code("void")
		} else if resultType == "stream" {
			dst.Code("Blob | ArrayBuffer")
		} else {
			b.printType(dst, method.Result.Type(), false, false, false)
		}
		dst.Code(">(this.id << 32 | ")
		dst.Code(method.Id.Value)
		dst.Code(", this.name,  \"")
		dst.Code(build.StringToUnderlineName(method.Name.Name))
		dst.Code("\", \"\", ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", ")
		if resultType == "void" || resultType == "stream" {
			dst.Code("undefined)\n")
		} else {
			b.printType(dst, method.Result.Type(), false, false, true)
			dst.Code(".fromMap)\n")
			//b.printType(dst, method.Result.Type(), false, false)
			//dst.Code(".fromData)\n")
		}

		dst.Tab(1).Code("}\n\n")
		return nil
	})
	dst.Code("}\n\n")
}

func (b *Builder) printServerRouter(dst *build.Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	dst.Import("hbuf_ts", "type Server")
	dst.Code("export function Register").Code(serverName).Code("(r: Server, ")
	dst.Code("server: ").Code(serverName).Code(") {\n")
	dst.Tab(1).Code("r.register(0, \"").Code(build.StringToUnderlineName(typ.Name.Name)).Code("\", [\n")
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		paramType := method.Param.Type().(*ast.Ident).Name

		dst.Tab(2).Code("{\n")
		dst.Tab(3).Code("id: 0,\n")
		dst.Tab(3).Code("name: \"").Code(build.StringToUnderlineName(method.Name.Name)).Code("\",\n")

		dst.Import("hbuf_ts", "type Option", "type RequestType", "type ResponseType")
		dst.Tab(3).Code("handler: (req: RequestType, opt?: Option): Promise<ResponseType> => {\n")
		dst.Tab(4).Code("return server.").Code(build.StringToFirstLower(method.Name.Name)).Code("(")
		if paramType == "stream" {
			dst.Import("hbuf_ts", "type BufferType")
			dst.Code("req as BufferType,")
		} else if paramType == "void" {
		} else {
			dst.Code("req as ")
			b.printType(dst, method.Param.Type(), false, false, true)
			dst.Code(", ")
		}
		dst.Code("opt)\n")
		dst.Tab(3).Code("},\n")
		dst.Tab(3).Code("withContext: (opt?: Option): Option | undefined => {\n")
		dst.Tab(4).Code("return opt\n")
		dst.Tab(3).Code("},\n")
		if paramType == "void" || paramType == "stream" {

		} else {
			dst.Tab(3).Code("from: ")
			b.printType(dst, method.Param.Type(), false, false, false)
			dst.Code(".fromMap,\n")
		}
		dst.Tab(3).Code("tag: \"\",\n")
		dst.Tab(2).Code("},\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Tab(1).Code("])\n")
	dst.Code("}\n\n")

	dst.Import("hbuf_ts", "type Server")
	dst.Code("export function UnRegister").Code(serverName).Code("(r: Server) {\n")
	dst.Tab(1).Code("r.unRegister(0, \"").Code(build.StringToUnderlineName(typ.Name.Name)).Code("\")\n")
	dst.Code("}\n\n")
}
