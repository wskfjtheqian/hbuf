package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
)

func printServer(dst io.Writer, typ *ast.ServerType) {
	_, _ = dst.Write([]byte("type " + toClassName(typ.Name.Name)))
	_, _ = dst.Write([]byte(" interface {\n"))
	printExtend(dst, typ.Extends)

	for _, method := range typ.Methods {
		if nil != method.Comment {
			_, _ = dst.Write([]byte("\t//" + toFieldName(method.Name.Name) + " " + method.Comment.Text()))
		}

		_, _ = dst.Write([]byte("\t" + toFieldName(method.Name.Name)))
		_, _ = dst.Write([]byte("(cxt *hbuf.Context, "))
		_, _ = dst.Write([]byte(toParamName(method.ParamName.Name)))
		_, _ = dst.Write([]byte(" *"))
		printType(dst, method.Param, false)
		_, _ = dst.Write([]byte(") (*"))
		printType(dst, method.Result.Type(), false)
		_, _ = dst.Write([]byte(",error)\n\n"))
	}
	_, _ = dst.Write([]byte("}\n\n"))
}

func printServerImp(dst io.Writer, typ *ast.ServerType) {
	//
	//_, _ = dst.Write([]byte("class " + toClassName(typ.Name.Name) + "Client extends ServerClient implements " + toClassName(typ.Name.Name)))
	//
	//_, _ = dst.Write([]byte("{\n"))
	//
	//_, _ = dst.Write([]byte("  " + toClassName(typ.Name.Name) + "Client(Client client):super(client);\n\n"))
	//
	//_, _ = dst.Write([]byte("  @override\n"))
	//_, _ = dst.Write([]byte("  String get name => \"" + toClassName(typ.Name.Name) + "\";\n\n"))
	//_, _ = dst.Write([]byte("  @override\n"))
	//_, _ = dst.Write([]byte("  int get id => " + typ.Id.Value + ";\n\n"))
	//
	//_ = build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
	//	_, _ = dst.Write([]byte("  @override\n"))
	//	_, _ = dst.Write([]byte("  Future<"))
	//	printType(dst, method.Result.Type(), false)
	//	_, _ = dst.Write([]byte("> " + toFieldName(method.Name.Name)))
	//	_, _ = dst.Write([]byte("("))
	//	printType(dst, method.Param, false)
	//	_, _ = dst.Write([]byte(" " + toFieldName(method.ParamName.Name)))
	//	_, _ = dst.Write([]byte(", [Context? ctx]){\n"))
	//
	//	_, _ = dst.Write([]byte("    return invoke<"))
	//	printType(dst, method.Result.Type(), false)
	//	_, _ = dst.Write([]byte(">(\""))
	//	_, _ = dst.Write([]byte(server.Name.Name + "/" + method.Name.Name))
	//	_, _ = dst.Write([]byte("\", "))
	//	_, _ = dst.Write([]byte(server.Id.Value + " << 32 | " + method.Id.Value))
	//	_, _ = dst.Write([]byte(", "))
	//	_, _ = dst.Write([]byte(toFieldName(method.ParamName.Name)))
	//	_, _ = dst.Write([]byte(", "))
	//	printType(dst, method.Result.Type(), false)
	//	_, _ = dst.Write([]byte(".fromMap, "))
	//	printType(dst, method.Result.Type(), false)
	//	_, _ = dst.Write([]byte(".fromData);\n"))
	//
	//	_, _ = dst.Write([]byte("  }\n\n"))
	//	return nil
	//})
	//_, _ = dst.Write([]byte("}\n\n"))
}

func printServerRouter(dst io.Writer, typ *ast.ServerType) {

	_, _ = dst.Write([]byte("type " + toClassName(typ.Name.Name) + "Router struct {\n"))
	_, _ = dst.Write([]byte("	server " + toClassName(typ.Name.Name) + "\n"))
	_, _ = dst.Write([]byte("	names  map[string]*hbuf.ServerInvoke\n"))
	_, _ = dst.Write([]byte("}\n\n"))

	_, _ = dst.Write([]byte("func (p *" + toClassName(typ.Name.Name) + "Router) GetName() string {\n"))
	_, _ = dst.Write([]byte("	return \"" + toClassName(typ.Name.Name) + "\"\n"))
	_, _ = dst.Write([]byte("}\n\n"))

	_, _ = dst.Write([]byte("func (p *" + toClassName(typ.Name.Name) + "Router) GetId() uint32 {\n"))
	_, _ = dst.Write([]byte("	return 1\n"))
	_, _ = dst.Write([]byte("}\n\n"))

	_, _ = dst.Write([]byte("func (p *" + toClassName(typ.Name.Name) + "Router) GetInvoke() map[string]*hbuf.ServerInvoke {\n"))
	_, _ = dst.Write([]byte("	return p.names\n"))
	_, _ = dst.Write([]byte("}\n\n"))

	_, _ = dst.Write([]byte("func New" + toClassName(typ.Name.Name) + "Router(server " + toClassName(typ.Name.Name) + ") *" + toClassName(typ.Name.Name) + "Router {\n"))
	_, _ = dst.Write([]byte("	return &" + toClassName(typ.Name.Name) + "Router{\n"))
	_, _ = dst.Write([]byte("		server: server,\n"))
	_, _ = dst.Write([]byte("		names: map[string]*hbuf.ServerInvoke{\n"))
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		_, _ = dst.Write([]byte("			\"" + toClassName(typ.Name.Name) + "/" + toFieldName(method.Name.Name) + "\": {\n"))
		_, _ = dst.Write([]byte("				ToData: func(buf []byte) (hbuf.Data, error) {\n"))
		_, _ = dst.Write([]byte("					var req "))
		printType(dst, method.Param, false)
		_, _ = dst.Write([]byte("\n"))
		_, _ = dst.Write([]byte("					return &req, json.Unmarshal(buf, &req)\n"))
		_, _ = dst.Write([]byte("				},\n"))
		_, _ = dst.Write([]byte("				FormData: func(data hbuf.Data) ([]byte, error) {\n"))
		_, _ = dst.Write([]byte("					return json.Marshal(&data)\n"))
		_, _ = dst.Write([]byte("				},\n"))
		_, _ = dst.Write([]byte("				Invoke: func(cxt *hbuf.Context, data hbuf.Data) (hbuf.Data, error) {\n"))
		_, _ = dst.Write([]byte("					return server." + toFieldName(method.Name.Name) + "(cxt, data.(*"))
		printType(dst, method.Param, false)
		_, _ = dst.Write([]byte("))\n"))
		_, _ = dst.Write([]byte("				},\n"))
		_, _ = dst.Write([]byte("			},\n"))
		return nil
	})
	if err != nil {
		return
	}
	_, _ = dst.Write([]byte("		},\n"))
	_, _ = dst.Write([]byte("	}\n"))
	_, _ = dst.Write([]byte("}\n\n"))
}
