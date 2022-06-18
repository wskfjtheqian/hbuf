package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func printServerCode(dst *Writer, typ *ast.ServerType) {
	dst.Import("context")
	dst.Import("encoding/json")
	dst.Import("errors")
	dst.Import("hbuf_golang/pkg/hbuf")

	printServer(dst, typ)
	printServerImp(dst, typ)
	printServerRouter(dst, typ)
	printGetServerRouter(dst, typ)
}
func printServer(dst *Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	dst.Code("type " + serverName)
	dst.Code(" interface {\n")
	printExtend(dst, typ.Extends)

	for _, method := range typ.Methods {
		if nil != method.Comment {
			dst.Code("\t//" + build.StringToHumpName(method.Name.Name) + " " + method.Comment.Text())
		}

		dst.Code("\t" + build.StringToHumpName(method.Name.Name))
		dst.Code("(ctx context.Context, ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(" *")
		printType(dst, method.Param, false)
		dst.Code(") (*")
		printType(dst, method.Result.Type(), false)
		dst.Code(",error)\n\n")
	}
	dst.Code("}\n\n")
}

func printServerImp(dst *Writer, typ *ast.ServerType) {
	//
	//dst.Code("class " +serverName + "Client extends ServerClient implements " +serverName)
	//
	//dst.Code("{\n")
	//
	//dst.Code("  " +serverName + "Client(Client client):super(client);\n\n")
	//
	//dst.Code("  @override\n")
	//dst.Code("  String get name => \"" +serverName + "\";\n\n")
	//dst.Code("  @override\n")
	//dst.Code("  int get id => " + typ.Id.Value + ";\n\n")
	//
	//_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
	//	dst.Code("  @override\n")
	//	dst.Code("  Future<")
	//	printType(dst, method.Result.Type(), false)
	//	dst.Code("> " + build.StringToHumpName(method.Name.Name))
	//	dst.Code("(")
	//	printType(dst, method.Param, false)
	//	dst.Code(" " + build.StringToHumpName(method.ParamName.Name))
	//	dst.Code(", [Context? ctx]){\n")
	//
	//	dst.Code("    return invoke<")
	//	printType(dst, method.Result.Type(), false)
	//	dst.Code(">(\"")
	//	dst.Code(server.Name.Name + "/" + method.Name.Name)
	//	dst.Code("\", ")
	//	dst.Code(server.Id.Value + " << 32 | " + method.Id.Value)
	//	dst.Code(", ")
	//	dst.Code(build.StringToHumpName(method.ParamName.Name))
	//	dst.Code(", ")
	//	printType(dst, method.Result.Type(), false)
	//	dst.Code(".fromMap, ")
	//	printType(dst, method.Result.Type(), false)
	//	dst.Code(".fromData);\n")
	//
	//	dst.Code("  }\n\n")
	//	return nil
	//})
	//dst.Code("}\n\n")
}

func printServerRouter(dst *Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	dst.Code("type " + serverName + "Router struct {\n")
	dst.Code("	server " + serverName + "\n")
	dst.Code("	names  map[string]*hbuf.ServerInvoke\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetName() string {\n")
	dst.Code("	return \"" + build.StringToUnderlineName(typ.Name.Name) + "\"\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetId() uint32 {\n")
	dst.Code("	return 1\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetInvoke() map[string]*hbuf.ServerInvoke {\n")
	dst.Code("	return p.names\n")
	dst.Code("}\n\n")

	dst.Code("func New" + serverName + "Router(server " + serverName + ") *" + serverName + "Router {\n")
	dst.Code("	return &" + serverName + "Router{\n")
	dst.Code("		server: server,\n")
	dst.Code("		names: map[string]*hbuf.ServerInvoke{\n")
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Code("			\"" + build.StringToUnderlineName(typ.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\": {\n")
		dst.Code("				ToData: func(buf []byte) (hbuf.Data, error) {\n")
		dst.Code("					var req ")
		printType(dst, method.Param, false)
		dst.Code("\n")
		dst.Code("					return &req, json.Unmarshal(buf, &req)\n")
		dst.Code("				},\n")
		dst.Code("				FormData: func(data hbuf.Data) ([]byte, error) {\n")
		dst.Code("					return json.Marshal(&data)\n")
		dst.Code("				},\n")
		dst.Code("				Invoke: func(ctx context.Context, data hbuf.Data) (hbuf.Data, error) {\n")
		dst.Code("					return server." + build.StringToHumpName(method.Name.Name) + "(ctx, data.(*")
		printType(dst, method.Param, false)
		dst.Code("))\n")
		dst.Code("				},\n")
		dst.Code("			},\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("		},\n")
	dst.Code("	}\n")
	dst.Code("}\n\n")
}

func printGetServerRouter(dst *Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)

	dst.Code("func Get" + serverName + "(server hbuf.GetServer) (" + serverName + ", error) {\n")
	dst.Code("	router := server.Get(&" + serverName + "Router{})\n")
	dst.Code("	if nil == router {\n")
	dst.Code("		return nil, errors.New(\"Not find server\")\n")
	dst.Code("	}\n")
	dst.Code("	switch router.(type) {\n")
	dst.Code("	case *" + serverName + "Router:\n")
	dst.Code("		return router.(*" + serverName + "Router).server, nil\n")
	dst.Code("	}\n")
	dst.Code("	return nil, errors.New(\"Not find server\")\n")
	dst.Code("}\n")
}
