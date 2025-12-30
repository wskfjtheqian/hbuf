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

		dst.Code(", opt?: h.Option): ")
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
	dst.Code("export class " + build.StringToHumpName(typ.Name.Name) + "Client implements ")
	dst.Code(b.getPackage(dst, typ.Name, ""))
	dst.Code(".")
	dst.Code(build.StringToHumpName(typ.Name.Name))

	dst.Code("{\n\n")

	dst.Tab(1).Code("protected client: h.Client\n\n")

	dst.Tab(1).Code("constructor(client: h.Client){\n")
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
			b.printType(dst, method.Param, false, false)
		}

		dst.Code(", opt?: h.Option): ")
		dst.Code("Promise<")
		if resultType == "void" {
			dst.Code("void")
		} else if resultType == "stream" {
			dst.Code("Blob | ArrayBuffer")
		} else {
			b.printType(dst, method.Result.Type(), false, false)
		}

		dst.Code("> {\n")

		dst.Tab(2).Code("return this.client.invoke<")
		if resultType == "void" {
			dst.Code("void")
		} else if resultType == "stream" {
			dst.Code("Blob | ArrayBuffer")
		} else {
			b.printType(dst, method.Result.Type(), false, false)
		}
		dst.Code(">(this.name, this.id << 32 | ")
		dst.Code(method.Id.Value)
		dst.Code(", \"")
		dst.Code(build.StringToUnderlineName(method.Name.Name))
		dst.Code("\", \"\", ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(", ")
		if resultType == "void" || resultType == "stream" {
			dst.Code("null, null)\n")
		} else {
			b.printType(dst, method.Result.Type(), false, false)
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
	dst.Code("export function Register").Code(serverName).Code("(r: h.Server, ")
	dst.Code("server: ").Code(serverName).Code(") {\n")
	dst.Tab(1).Code("r.register(0, \"").Code(build.StringToUnderlineName(typ.Name.Name)).Code("\", [\n")
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		//paramType := method.Param.Type().(*ast.Ident).Name
		dst.Tab(2).Code("{\n")
		dst.Tab(3).Code("id: 0,\n")
		dst.Tab(3).Code("name: \"").Code(build.StringToUnderlineName(method.Name.Name)).Code("\",\n")
		dst.Tab(3).Code("handler: (req: h.RequestType, opt?: h.Option): Promise<h.ResponseType> => {\n")
		dst.Tab(4).Code("return server.").Code(build.StringToFirstLower(method.Name.Name)).Code("(req as ")
		b.printType(dst, method.Param.Type(), false, false)
		dst.Code(", opt)\n")
		dst.Tab(3).Code("},\n")
		dst.Tab(3).Code("withContext: (opt?: h.Option): h.Option | undefined => {\n")
		dst.Tab(4).Code("return opt\n")
		dst.Tab(3).Code("},\n")
		dst.Tab(3).Code("from: ")
		b.printType(dst, method.Param.Type(), false, false)
		dst.Code(".fromMap,\n")
		dst.Tab(3).Code("tag: \"\",\n")
		dst.Tab(2).Code("},\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Tab(1).Code("])\n")
	dst.Code("}\n\n")

	dst.Code("export function UnRegister").Code(serverName).Code("(r: h.Server) {\n")
	dst.Tab(1).Code("r.unRegister(0, \"").Code(build.StringToUnderlineName(typ.Name.Name)).Code("\")\n")
	dst.Code("}\n\n")

	//dst.Code("export class " + build.StringToHumpName(typ.Name.Name) + "Router  {\n")
	//dst.Tab(1).Code("readonly server: " + build.StringToHumpName(typ.Name.Name) + "\n")
	//dst.Code("\n")
	//dst.Tab(1).Code("invoke: Record<string, h.ServerInvoke>\n")
	//dst.Code("\n")
	//dst.Tab(1).Code("getInvoke(): Record<string, h.ServerInvoke> {\n")
	//dst.Tab(2).Code("return this.invoke\n")
	//dst.Tab(1).Code("}\n")
	//dst.Code("\n")
	//dst.Tab(1).Code("getName(): string {\n")
	//dst.Tab(2).Code("return \"" + build.StringToUnderlineName(typ.Name.Name) + "\"\n")
	//dst.Tab(1).Code("}\n")
	//dst.Code("\n")
	//dst.Tab(1).Code("getId(): number {\n")
	//dst.Tab(2).Code("return 0\n")
	//dst.Tab(1).Code("}\n")
	//dst.Code("\n")
	//dst.Tab(1).Code("constructor(server: " + build.StringToHumpName(typ.Name.Name) + ") {\n")
	//dst.Tab(2).Code("this.server = server\n")
	//dst.Tab(2).Code("this.invoke = {\n")
	//err = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
	//
	//	paramType := method.Param.Type().(*ast.Ident).Name
	//
	//	dst.Tab(3).Code("\"" + build.StringToUnderlineName(method.Name.Name) + "\": {\n")
	//	dst.Tab(4).Code("formData(data: Blob | ArrayBuffer | Record<string, any>):  Blob | ArrayBuffer | h.Data {\n")
	//	dst.Tab(5).Code("return ")
	//	if paramType == "void" || paramType == "stream" {
	//		dst.Code("data as ArrayBuffer |h.Data\n")
	//	} else {
	//		b.printType(dst, method.Param.Type(), false, false)
	//		dst.Code(".fromMap(data as Record<string, any>)\n")
	//	}
	//
	//	dst.Tab(4).Code("},\n")
	//	dst.Tab(4).Code("toData(data: Blob | ArrayBuffer | h.Data): Blob | ArrayBuffer | Record<string, any> {\n")
	//	dst.Tab(5).Code("return ")
	//	if paramType == "void" || paramType == "stream" {
	//		dst.Code("data as Blob | ArrayBuffer\n")
	//	} else {
	//		dst.Code("(data as h.Data).toMap()\n")
	//	}
	//
	//	dst.Tab(4).Code("},\n")
	//	dst.Tab(4).Code("invoke(data: Blob | ArrayBuffer | h.Data, opt?: h.Option): Promise<h.Data | void> {\n")
	//	dst.Tab(5).Code("return server." + build.StringToFirstLower(method.Name.Name) + "(data as ")
	//	if paramType == "void" || paramType == "stream" {
	//		dst.Code("Blob | ArrayBuffer")
	//	} else {
	//		b.printType(dst, method.Param.Type(), false, false)
	//	}
	//	dst.Code(", ctx)\n")
	//	dst.Tab(4).Code("}\n")
	//	dst.Tab(3).Code("},\n")
	//	return nil
	//})
	//if err != nil {
	//	return
	//}

	//dst.Tab(2).Code("}\n")
	//dst.Tab(1).Code("}\n")
	//dst.Code("}\n")
}
