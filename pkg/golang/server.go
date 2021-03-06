package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printServerCode(dst *build.Writer, typ *ast.ServerType) {
	dst.Import("context", "")
	dst.Import("encoding/json", "")
	dst.Import("errors", "")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")

	b.printServer(dst, typ)
	b.printServerImp(dst, typ)
	b.printServerRouter(dst, typ)
	b.printGetServerRouter(dst, typ)
}
func (b *Builder) printServer(dst *build.Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	dst.Code("type " + serverName)
	dst.Code(" interface {\n")
	b.printExtend(dst, typ.Extends)

	for _, method := range typ.Methods {
		if nil != method.Comment {
			dst.Code("\t//" + build.StringToHumpName(method.Name.Name) + " " + method.Comment.Text())
		}

		dst.Code("\t" + build.StringToHumpName(method.Name.Name))
		dst.Code("(ctx context.Context, ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(" *")
		b.printType(dst, method.Param, false)
		dst.Code(") (*")
		b.printType(dst, method.Result.Type(), false)
		dst.Code(",error)\n\n")
	}
	dst.Code("}\n\n")
}

func (b *Builder) printServerImp(dst *build.Writer, typ *ast.ServerType) {
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

type auth map[string]string

func (b *Builder) getAuth(tags []*ast.Tag) *auth {
	val, ok := build.GetTag(tags, "auth")
	if !ok {
		return nil
	}
	au := make(auth, 0)
	if nil != val.KV {
		for _, item := range val.KV {
			au[item.Name.Name] = item.Value.Value[1 : len(item.Value.Value)-1]
		}
	}
	return &au
}

func (b *Builder) printServerRouter(dst *build.Writer, typ *ast.ServerType) {
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
	dst.Code("\treturn &" + serverName + "Router{\n")
	dst.Code("\t\tserver: server,\n")
	dst.Code("\t\tnames: map[string]*hbuf.ServerInvoke{\n")
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Code("\t\t\t\"" + build.StringToUnderlineName(typ.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\": {\n")
		dst.Code("\t\t\t\tToData: func(buf []byte) (hbuf.Data, error) {\n")
		dst.Code("\t\t\t\t\tvar req ")
		b.printType(dst, method.Param, false)
		dst.Code("\n")
		dst.Code("\t\t\t\t\treturn &req, json.Unmarshal(buf, &req)\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tFormData: func(data hbuf.Data) ([]byte, error) {\n")
		dst.Code("\t\t\t\t\treturn json.Marshal(&data)\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tSetInfo: func(ctx context.Context)  {\n")

		au := b.getAuth(method.Tags)
		if nil != au {
			for key, val := range *au {
				dst.Code("\t\t\t\t\thbuf.SetTag(ctx, \"" + key + "\", \"" + val + "\")\n")
			}
		}
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tInvoke: func(ctx context.Context, data hbuf.Data) (hbuf.Data, error) {\n")
		dst.Code("\t\t\t\t\treturn server." + build.StringToHumpName(method.Name.Name) + "(ctx, data.(*")
		b.printType(dst, method.Param, false)
		dst.Code("))\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t},\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("		},\n")
	dst.Code("	}\n")
	dst.Code("}\n\n")
}

func (b *Builder) printGetServerRouter(dst *build.Writer, typ *ast.ServerType) {
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
