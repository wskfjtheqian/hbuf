package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"sort"
)

func (b *Builder) printServerCode(dst *build.Writer, typ *ast.ServerType) {
	dst.Import("context", "")

	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/rpc", "")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/manage", "")
	b.printServer(dst, typ)
	b.printClient(dst, typ)
	b.printServerRouter(dst, typ)
	b.printServerDefault(dst, typ)
	b.printGetServerRouter(dst, typ)
}

func (b *Builder) printServer(dst *build.Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("//" + build.StringToHumpName(serverName) + " " + typ.Doc.Text())
	}
	dst.Code("type " + serverName)
	dst.Code(" interface {\n")
	isFast := true
	b.printDataExtend(dst, typ.Extends, &isFast)
	dst.Code("\tInit(ctx context.Context)\n\n")

	for _, method := range typ.Methods {
		if !isFast {
			dst.Code("\n")
		}
		isFast = false
		if nil != method.Doc && 0 < len(method.Doc.Text()) {
			dst.Code("\t//" + build.StringToHumpName(method.Name.Name) + " " + method.Doc.Text())
		}

		dst.Code("\t" + build.StringToHumpName(method.Name.Name))
		dst.Code("(ctx context.Context, ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(" *")
		b.printType(dst, method.Param, true)
		dst.Code(") (*")
		b.printType(dst, method.Result.Type(), true)
		dst.Code(", error)\n")
	}
	dst.Code("}\n\n")
}

func (b *Builder) printServerDefault(dst *build.Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("//" + build.StringToHumpName(serverName) + " " + typ.Doc.Text())
	}
	dst.Code("type Default" + serverName)
	dst.Code(" struct {\n")
	isFast := true
	b.printServerExtend(dst, typ.Extends, &isFast)
	dst.Code("}\n\n")

	dst.Code("func (s *Default" + serverName + ") Init(ctx context.Context) {\n")
	dst.Code("}\n\n")

	for _, method := range typ.Methods {
		if nil != method.Doc && 0 < len(method.Doc.Text()) {
			dst.Code("//" + build.StringToHumpName(method.Name.Name) + " " + method.Doc.Text())
		}

		dst.Import("errors", "")
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/erro", "")

		dst.Code("func (s *Default" + serverName + ") ")
		dst.Code(build.StringToHumpName(method.Name.Name))
		dst.Code("(ctx context.Context, ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(" *")
		b.printType(dst, method.Param, true)
		dst.Code(") (*")
		b.printType(dst, method.Result.Type(), true)
		dst.Code(", error) {\n")
		dst.Code("\treturn nil, erro.Wrap(errors.New(\"not find server\"))\n")

		dst.Code("}\n\n")
	}
}

func (b *Builder) printClient(dst *build.Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	dst.Code("type " + serverName + "Client struct {\n")
	dst.Code("\tclient rpc.Client\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Client) Init(ctx context.Context) {\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Client) GetName() string {\n")
	dst.Code("\treturn \"" + build.StringToUnderlineName(typ.Name.Name) + "\"\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Client) GetId() uint32 {\n")
	dst.Code("\treturn 1\n")
	dst.Code("}\n\n")

	dst.Code("func New" + serverName + "Client(client rpc.Client) *" + serverName + "Client {\n")
	dst.Code("\treturn &" + serverName + "Client{\n")
	dst.Code("\t\tclient: client,\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")
	name := build.StringToUnderlineName(typ.Name.Name)
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		if nil != method.Doc && 0 < len(method.Doc.Text()) {
			dst.Code("//" + build.StringToHumpName(method.Name.Name) + " " + method.Doc.Text())
		}
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")

		dst.Code("func (r *" + serverName + "Client) ")
		dst.Code(build.StringToHumpName(method.Name.Name))
		dst.Code("(ctx context.Context, ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(" *")
		b.printType(dst, method.Param, true)
		dst.Code(") (*")
		b.printType(dst, method.Result.Type(), true)
		dst.Code(", error) {\n")
		dst.Import("encoding/json", "")
		dst.Code("\tret, err := r.client.Invoke(ctx, req, \"" + name + "/" + build.StringToUnderlineName(typ.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\", &rpc.ClientInvoke{\n")
		dst.Code("\t\tToData: func(buf []byte) (hbuf.Data, error) {\n")
		dst.Code("\t\t\tvar req ")
		b.printType(dst, method.Result.Type(), true)
		dst.Code("\n")
		dst.Code("\t\t\treturn &req, json.Unmarshal(buf, &req)\n")
		dst.Code("\t\t},\n")
		dst.Code("\t\tFormData: func(data hbuf.Data) ([]byte, error) {\n")
		dst.Code("\t\t\treturn json.Marshal(&data)\n")
		dst.Code("\t\t},\n")
		dst.Code("\t}, 1, &rpc.ClientInvoke{})\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn nil, err\n")
		dst.Code("\t}\n")
		dst.Code("\treturn ret.(*")
		b.printType(dst, method.Result.Type(), true)
		dst.Code("), nil\n")
		dst.Code("}\n")
		dst.Code("\n")
		return nil
	})
	if err != nil {
		return
	}
}

type Tag map[string]string

func (b *Builder) getTag(tags []*ast.Tag) *Tag {
	val, ok := build.GetTag(tags, "tag")
	if !ok {
		return nil
	}
	au := make(Tag, 0)
	if nil != val.KV {
		for _, item := range val.KV {
			au[item.Name.Name] = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
		}
	}
	return &au
}

func (b *Builder) printServerRouter(dst *build.Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	dst.Code("type " + serverName + "Router struct {\n")
	dst.Code("\tserver " + serverName + "\n")
	dst.Code("\tnames  map[string]*rpc.ServerInvoke\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetName() string {\n")
	dst.Code("\treturn \"" + build.StringToUnderlineName(typ.Name.Name) + "\"\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetId() uint32 {\n")
	dst.Code("\treturn 1\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetServer() rpc.Init {\n")
	dst.Code("\treturn p.server\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetInvoke() map[string]*rpc.ServerInvoke {\n")
	dst.Code("\treturn p.names\n")
	dst.Code("}\n\n")

	dst.Code("func New" + serverName + "Router(server " + serverName + ") *" + serverName + "Router {\n")
	dst.Code("\treturn &" + serverName + "Router{\n")
	dst.Code("\t\tserver: server,\n")
	dst.Code("\t\tnames: map[string]*rpc.ServerInvoke{\n")
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")

		dst.Code("\t\t\t\"" + build.StringToUnderlineName(typ.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\": {\n")
		dst.Code("\t\t\t\tToData: func(buf []byte) (hbuf.Data, error) {\n")
		dst.Code("\t\t\t\t\tvar req ")
		b.printType(dst, method.Param, true)
		dst.Code("\n")
		dst.Import("encoding/json", "")
		dst.Code("\t\t\t\t\treturn &req, json.Unmarshal(buf, &req)\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tFormData: func(data hbuf.Data) ([]byte, error) {\n")
		dst.Code("\t\t\t\t\treturn json.Marshal(&data)\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tSetInfo: func(ctx context.Context) {\n")

		au := b.getTag(method.Tags)
		if nil != au {
			keys := build.GetKeysByMap(*au)
			sort.Strings(keys)
			for _, key := range keys {
				dst.Code("\t\t\t\t\trpc.SetTag(ctx, \"" + key + "\", \"" + (*au)[key] + "\")\n")
			}
		}
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tInvoke: func(ctx context.Context, data hbuf.Data) (hbuf.Data, error) {\n")
		dst.Code("\t\t\t\t\treturn server." + build.StringToHumpName(method.Name.Name) + "(ctx, data.(*")
		b.printType(dst, method.Param, true)
		dst.Code("))\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t},\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("\t\t},\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")
}

func (b *Builder) printGetServerRouter(dst *build.Writer, typ *ast.ServerType) {
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/manage", "")
	serverName := build.StringToHumpName(typ.Name.Name)

	dst.Code("var NotFound" + serverName + " = &Default" + serverName + "{}\n\n")

	dst.Code("func Get" + serverName + "(ctx context.Context) " + serverName + " {\n")
	dst.Code("\trouter := manage.GET(ctx).Get(&" + serverName + "Router{})\n")
	dst.Code("\tif nil == router {\n")
	dst.Code("\t\treturn NotFound" + serverName + "\n")
	dst.Code("\t}\n")
	dst.Code("\tif val, ok := router.(" + serverName + "); ok {\n")
	dst.Code("\t	return val\n")
	dst.Code("\t}\n")
	dst.Code("\treturn NotFound" + serverName + "\n")
	dst.Code("}\n\n")
}

func (b *Builder) printServerExtend(dst *build.Writer, extends []*ast.Extends, isFast *bool) {
	for _, v := range extends {
		if !*isFast {
			dst.Code("\n")
		}
		*isFast = false
		dst.Code("\t")
		pack := b.getPackage(dst, v.Name)
		dst.Code(pack)
		dst.Code("Default" + build.StringToHumpName(v.Name.Name))
		dst.Code("\n")
	}
}
