package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"sort"
)

func (b *Builder) printServerCode(dst *build.Writer, typ *ast.ServerType) error {
	dst.Import("context", "")

	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/rpc", "")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/manage", "")
	b.printServer(dst, typ)
	b.printClient(dst, typ)
	b.printServerRouter(dst, typ)
	err := b.printServerDefault(dst, typ)
	if err != nil {
		return err
	}
	b.printGetServerRouter(dst, typ)
	return nil
}

func (b *Builder) printServer(dst *build.Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("// " + build.StringToHumpName(serverName) + " " + typ.Doc.Text())
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
		isMethod := method.Result.Type().(*ast.Ident).Name == "void"

		dst.Code("\t" + build.StringToHumpName(method.Name.Name))
		dst.Code("(ctx context.Context, ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(" *")
		b.printType(dst, method.Param, true)
		dst.Code(") ")
		if isMethod {
			dst.Code("error\n")
		} else {
			dst.Code("(*")
			b.printType(dst, method.Result.Type(), true)
			dst.Code(", error)\n")
		}
	}
	dst.Code("}\n\n")
}

func (b *Builder) printServerDefault(dst *build.Writer, typ *ast.ServerType) error {
	serverName := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("// " + build.StringToHumpName(serverName) + " " + typ.Doc.Text())
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
			dst.Code("// " + build.StringToHumpName(method.Name.Name) + " " + method.Doc.Text())
		}
		isSub := method.Result.Type().(*ast.Ident).Name == "void"

		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/erro", "")

		dst.Code("func (s *Default" + serverName + ") ")
		dst.Code(build.StringToHumpName(method.Name.Name))
		dst.Code("(ctx context.Context, ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(" *")
		b.printType(dst, method.Param, true)
		if isSub {
			dst.Code(") error {\n")
		} else {
			dst.Code(") (*")
			b.printType(dst, method.Result.Type(), true)
			dst.Code(", error) {\n")
		}

		bind, err := build.GetBinding(method.Tags, dst.File, b.GetDataType)
		if nil != err {
			return err
		}
		if bind == nil {
			if isSub {
				dst.Code("\treturn ")
			} else {
				dst.Code("\treturn nil,")
			}

			dst.Code(" erro.NewError(\"not find server " + build.StringToUnderlineName(typ.Name.Name) + "\")\n")
		} else {
			b.printBinding(dst, method, bind, isSub)
		}
		dst.Code("}\n\n")
	}
	return nil
}

func (b *Builder) printBinding(dst *build.Writer, method *ast.FuncType, bind *build.Binding, isSub bool) error {
	verify, err := build.GetVerify(method.Param.Type().(*ast.Ident).Obj.Decl.(*ast.TypeSpec).Type.(*ast.DataType).Tags, dst.File, b.GetDataType)
	if err != nil {
		return err
	}
	if nil != verify {
		dst.Code("\tif err := req.Verify(ctx); err != nil {\n")
		dst.Code("\t\treturn nil, erro.Wrap(err)\n")
		dst.Code("\t}\n")
	}
	pack := b.getPackage(dst, bind.Server.Name)
	if !isSub {
		dst.Code("\treps, err := ")
	} else {
		dst.Code("\terr := ")
	}
	dst.Code(pack)
	dst.Code("Get").Code(bind.Server.Name.Name).Code("(ctx).").Code(bind.Method.Name.Name).Code("(ctx, &")
	dst.Code("req.")
	b.printType(dst, bind.Method.Param.Type(), true)
	dst.Code(")\n")

	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn nil, erro.Wrap(err)\n")
	dst.Code("\t}\n")

	dst.Code("\treturn ")
	if !isSub {
		dst.Code("&")
		b.printType(dst, method.Result.Type(), true)
		dst.Code("{")
		b.printType(dst, bind.Method.Result.Type(), true)
		dst.Code(": *reps}, ")
	}
	dst.Code("nil\n")
	return nil
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
			dst.Code("// " + build.StringToHumpName(method.Name.Name) + " " + method.Doc.Text())
		}
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")

		isMethod := method.Result.Type().(*ast.Ident).Name == "void"

		dst.Code("func (r *" + serverName + "Client) ")
		dst.Code(build.StringToHumpName(method.Name.Name))
		dst.Code("(ctx context.Context, ")
		dst.Code(build.StringToFirstLower(method.ParamName.Name))
		dst.Code(" *")
		b.printType(dst, method.Param, true)
		dst.Code(") ")
		if isMethod {
			dst.Code("error {\n")
		} else {
			dst.Code("(*")
			b.printType(dst, method.Result.Type(), true)
			dst.Code(", error) {\n")
		}
		dst.Import("encoding/json", "")
		if isMethod {
			dst.Code("\t_")
		} else {
			dst.Code("\tret")
		}
		dst.Code(", err := r.client.Invoke(ctx, req, \"" + name + "/" + build.StringToUnderlineName(typ.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\", &rpc.ClientInvoke{\n")
		if !isMethod {
			dst.Code("\t\tToData: func(buf []byte) (hbuf.Data, error) {\n")
			dst.Code("\t\t\tvar req ")
			b.printType(dst, method.Result.Type(), true)
			dst.Code("\n")
			dst.Code("\t\t\treturn &req, json.Unmarshal(buf, &req)\n")
			dst.Code("\t\t},\n")
		}
		dst.Code("\t\tFormData: func(data hbuf.Data) ([]byte, error) {\n")
		dst.Code("\t\t\treturn json.Marshal(&data)\n")
		dst.Code("\t\t},\n")
		dst.Code("\t}, 1, &rpc.ClientInvoke{})\n")
		if isMethod {
			dst.Code("\tif err != nil {\n")
			dst.Code("\t\treturn err\n")
			dst.Code("\t}\n")
			dst.Code("\treturn nil\n")
		} else {
			dst.Code("\tif err != nil {\n")
			dst.Code("\t\treturn nil, err\n")
			dst.Code("\t}\n")
			dst.Code("\treturn ret.(*")
			b.printType(dst, method.Result.Type(), true)
			dst.Code("), nil\n")
		}
		dst.Code("}\n")
		dst.Code("\n")
		return nil
	})
	if err != nil {
		return
	}
}

type Tag map[string][]string

func (b *Builder) getTag(tags []*ast.Tag) *Tag {
	val, ok := build.GetTag(tags, "tag")
	if !ok {
		return nil
	}
	au := make(Tag, 0)
	if nil != val.KV {
		for _, item := range val.KV {
			list := make([]string, 0)
			for _, value := range item.Values {
				list = append(list, value.Value[1:len(value.Value)-1])
			}
			au[item.Name.Name] = list
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

		isMethod := method.Result.Type().(*ast.Ident).Name == "void"

		dst.Code("\t\t\t\"" + build.StringToUnderlineName(typ.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\": {\n")
		dst.Code("\t\t\t\tToData: func(buf []byte) (hbuf.Data, error) {\n")
		dst.Code("\t\t\t\t\tvar req ")
		b.printType(dst, method.Param, true)
		dst.Code("\n")
		dst.Import("encoding/json", "")
		dst.Code("\t\t\t\t\treturn &req, json.Unmarshal(buf, &req)\n")
		dst.Code("\t\t\t\t},\n")
		if !isMethod {
			dst.Code("\t\t\t\tFormData: func(data hbuf.Data) ([]byte, error) {\n")
			dst.Code("\t\t\t\t\treturn json.Marshal(&data)\n")
			dst.Code("\t\t\t\t},\n")
		}
		dst.Code("\t\t\t\tSetInfo: func(ctx context.Context) {\n")

		au := b.getTag(method.Tags)
		if nil != au {
			keys := build.GetKeysByMap(*au)
			sort.Strings(keys)
			for _, key := range keys {
				values := (*au)[key]
				if len(values) > 0 {
					dst.Tab(5).Code("rpc.SetTag(ctx, \"").Code(key)
					for _, val := range values {
						dst.Code("\", \"").Code(val)
					}
					dst.Code("\")\n")
				}
			}
		}
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tInvoke: func(ctx context.Context, data hbuf.Data) (hbuf.Data, error) {\n")
		dst.Code("\t\t\t\t\treturn ")
		if isMethod {
			dst.Code("nil, ")
		}
		dst.Code("server." + build.StringToHumpName(method.Name.Name) + "(ctx, data.(*")
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
