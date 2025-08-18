package golang

import "C"
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
	dst.Tab(1).Code("Init(ctx context.Context)\n\n")

	for _, method := range typ.Methods {
		if !isFast {
			dst.Code("\n")
		}
		isFast = false
		if nil != method.Doc && 0 < len(method.Doc.Text()) {
			dst.Tab(1).Code("//" + build.StringToHumpName(method.Name.Name) + " " + method.Doc.Text())
		}
		isMethod := method.Result.Type().(*ast.Ident).Name == "void"

		dst.Tab(1).Code("" + build.StringToHumpName(method.Name.Name))
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

	name := build.StringToUnderlineName(typ.Name.Name)
	dst.Code("func (s *Default" + serverName + ") ServerName() string {\n")
	dst.Tab(1).Code("return \"").Code(name).Code("\"\n")
	dst.Code("}\n\n")

	for _, method := range typ.Methods {
		if nil != method.Doc && 0 < len(method.Doc.Text()) {
			dst.Code("// " + build.StringToHumpName(method.Name.Name) + " " + method.Doc.Text())
		}
		isSub := method.Result.Type().(*ast.Ident).Name == "void"

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
				dst.Tab(1).Code("return ")
			} else {
				dst.Tab(1).Code("return nil,")
			}

			dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/erro", "")
			dst.Code(" erro.NewError(\"not find server " + build.StringToUnderlineName(typ.Name.Name) + "\")\n")
		} else {
			b.printBinding(dst, method, bind, isSub)
		}
		dst.Code("}\n\n")

		dst.Code("func (s *Default" + serverName + ") ")
		dst.Code(build.StringToHumpName(method.Name.Name)).Code("MethodName() string {\n")
		dst.Tab(1).Code("return \"").Code(name).Code("/").Code(build.StringToUnderlineName(typ.Name.Name)).Code("/").Code(build.StringToUnderlineName(method.Name.Name)).Code("\"\n")
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
		dst.Tab(1).Code("if err := req.Verify(ctx); err != nil {\n")
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/erro", "")
		dst.Tab(2).Code("return nil, erro.Wrap(err)\n")
		dst.Tab(1).Code("}\n")
	}
	pack := b.getPackage(dst, bind.Server.Name)
	if !isSub {
		dst.Tab(1).Code("resp, err := ")
	} else {
		dst.Tab(1).Code("err := ")
	}
	dst.Code(pack)
	dst.Code("Get").Code(bind.Server.Name.Name).Code("(ctx).").Code(bind.Method.Name.Name).Code("(ctx, &")
	dst.Code("req.")
	b.printType(dst, bind.Method.Param.Type(), true)
	dst.Code(")\n")

	dst.Tab(1).Code("if err != nil {\n")
	dst.Tab(2).Code("return nil, err\n")
	dst.Tab(1).Code("}\n")

	dst.Tab(1).Code("return ")
	if !isSub {
		dst.Code("&")
		b.printType(dst, method.Result.Type(), true)
		dst.Code("{")
		b.printType(dst, bind.Method.Result.Type(), true)
		dst.Code(": *resp}, ")
	}
	dst.Code("nil\n")
	return nil
}

func (b *Builder) printClient(dst *build.Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	dst.Code("type " + serverName + "Client struct {\n")
	dst.Tab(1).Code("client rpc.Client\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Client) Init(ctx context.Context) {\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Client) GetName() string {\n")
	dst.Tab(1).Code("return \"" + build.StringToUnderlineName(typ.Name.Name) + "\"\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Client) GetId() uint32 {\n")
	dst.Tab(1).Code("return 1\n")
	dst.Code("}\n\n")

	dst.Code("func New" + serverName + "Client(client rpc.Client) *" + serverName + "Client {\n")
	dst.Tab(1).Code("return &" + serverName + "Client{\n")
	dst.Tab(2).Code("client: client,\n")
	dst.Tab(1).Code("}\n")
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
			dst.Tab(1).Code("_")
		} else {
			dst.Tab(1).Code("ret")
		}
		dst.Code(", err := r.client.Invoke(ctx, req, \"" + name + "/" + build.StringToUnderlineName(typ.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\", &rpc.ClientInvoke{\n")
		if !isMethod {
			dst.Tab(2).Code("ToData: func(buf []byte) (hbuf.Data, error) {\n")
			dst.Tab(3).Code("var req ")
			b.printType(dst, method.Result.Type(), true)
			dst.Code("\n")
			dst.Tab(3).Code("return &req, json.Unmarshal(buf, &req)\n")
			dst.Tab(2).Code("},\n")
		}
		dst.Tab(2).Code("FormData: func(data hbuf.Data) ([]byte, error) {\n")
		dst.Tab(3).Code("return json.Marshal(&data)\n")
		dst.Tab(2).Code("},\n")
		dst.Tab(1).Code("}, 1, &rpc.ClientInvoke{})\n")
		if isMethod {
			dst.Tab(1).Code("if err != nil {\n")
			dst.Tab(2).Code("return err\n")
			dst.Tab(1).Code("}\n")
			dst.Tab(1).Code("return nil\n")
		} else {
			dst.Tab(1).Code("if err != nil {\n")
			dst.Tab(2).Code("return nil, err\n")
			dst.Tab(1).Code("}\n")
			dst.Tab(1).Code("return ret.(*")
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
	dst.Tab(1).Code("server " + serverName + "\n")
	dst.Tab(1).Code("names  map[string]*rpc.ServerInvoke\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetName() string {\n")
	dst.Tab(1).Code("return \"" + build.StringToUnderlineName(typ.Name.Name) + "\"\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetId() uint32 {\n")
	dst.Tab(1).Code("return 1\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetServer() rpc.Init {\n")
	dst.Tab(1).Code("return p.server\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Router) GetInvoke() map[string]*rpc.ServerInvoke {\n")
	dst.Tab(1).Code("return p.names\n")
	dst.Code("}\n\n")

	dst.Code("func New" + serverName + "Router(server " + serverName + ") *" + serverName + "Router {\n")
	dst.Tab(1).Code("return &" + serverName + "Router{\n")
	dst.Tab(2).Code("server: server,\n")
	dst.Tab(2).Code("names: map[string]*rpc.ServerInvoke{\n")
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")

		isMethod := method.Result.Type().(*ast.Ident).Name == "void"

		dst.Tab(3).Code("\"" + build.StringToUnderlineName(typ.Name.Name) + "/" + build.StringToUnderlineName(method.Name.Name) + "\": {\n")
		dst.Tab(4).Code("ToData: func(buf []byte) (hbuf.Data, error) {\n")
		dst.Tab(5).Code("var req ")
		b.printType(dst, method.Param, true)
		dst.Code("\n")
		dst.Import("encoding/json", "")
		dst.Tab(5).Code("return &req, json.Unmarshal(buf, &req)\n")
		dst.Tab(4).Code("},\n")
		if !isMethod {
			dst.Tab(4).Code("FormData: func(data hbuf.Data) ([]byte, error) {\n")
			dst.Tab(5).Code("return json.Marshal(&data)\n")
			dst.Tab(4).Code("},\n")
		}
		dst.Tab(4).Code("SetInfo: func(ctx context.Context) {\n")

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
		dst.Tab(4).Code("},\n")
		dst.Tab(4).Code("Invoke: func(ctx context.Context, data hbuf.Data) (hbuf.Data, error) {\n")
		dst.Tab(5).Code("return ")
		if isMethod {
			dst.Code("nil, ")
		}
		dst.Code("server." + build.StringToHumpName(method.Name.Name) + "(ctx, data.(*")
		b.printType(dst, method.Param, true)
		dst.Code("))\n")
		dst.Tab(4).Code("},\n")
		dst.Tab(3).Code("},\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Tab(2).Code("},\n")
	dst.Tab(1).Code("}\n")
	dst.Code("}\n\n")
}

func (b *Builder) printGetServerRouter(dst *build.Writer, typ *ast.ServerType) {
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/manage", "")
	serverName := build.StringToHumpName(typ.Name.Name)

	dst.Code("var NotFound" + serverName + " = &Default" + serverName + "{}\n\n")

	dst.Code("func Get" + serverName + "(ctx context.Context) " + serverName + " {\n")
	dst.Tab(1).Code("router := manage.GET(ctx).Get(&" + serverName + "Router{})\n")
	dst.Tab(1).Code("if nil == router {\n")
	dst.Tab(2).Code("return NotFound" + serverName + "\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("if val, ok := router.(" + serverName + "); ok {\n")
	dst.Tab(1).Code("	return val\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("return NotFound" + serverName + "\n")
	dst.Code("}\n\n")
}

func (b *Builder) printServerExtend(dst *build.Writer, extends []*ast.Extends, isFast *bool) {
	for _, v := range extends {
		if !*isFast {
			dst.Code("\n")
		}
		*isFast = false
		dst.Tab(1).Code("")
		pack := b.getPackage(dst, v.Name)
		dst.Code(pack)
		dst.Code("Default" + build.StringToHumpName(v.Name.Name))
		dst.Code("\n")
	}
}
