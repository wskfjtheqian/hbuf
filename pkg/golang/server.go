package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"sort"
)

func (b *Builder) printServerCode(dst *build.Writer, typ *ast.ServerType) error {
	dst.Import("context", "")

	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "")
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

			dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/herror", "")
			dst.Code(" herror.NewError(\"not find server " + build.StringToUnderlineName(typ.Name.Name) + "\")\n")
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
		dst.Tab(1).Code("if err := req.Verify(ctx); err != nil {\n")
		dst.Tab(2).Code("return nil, err\n")
		dst.Tab(1).Code("}\n")
	}
	pack := b.getPackage(dst, bind.Server.Name)
	if !isSub {
		dst.Tab(1).Code("reps, err := ")
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
		dst.Code(": *reps}, ")
	}
	dst.Code("nil\n")
	return nil
}

func (b *Builder) printClient(dst *build.Writer, typ *ast.ServerType) {
	serverName := build.StringToHumpName(typ.Name.Name)
	dst.Code("type " + serverName + "Client struct {\n")
	dst.Tab(1).Code("client *hrpc.Client\n")
	dst.Code("}\n\n")

	dst.Code("func (p *" + serverName + "Client) Init(ctx context.Context) {\n")
	dst.Code("}\n\n")

	dst.Code("func New" + serverName + "Client(client *hrpc.Client) " + serverName + " {\n")
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
		dst.Tab(1).Code("response, err := r.client.Invoke(ctx, 0, \"").Code(name).Code("\", \"")
		dst.Code(build.StringToUnderlineName(method.Name.Name)).Code("\", \"")
		dst.Code(b.getFilterTag(method))
		dst.Code("\", req, hrpc.NewResultResponse[*")
		b.printType(dst, method.Result.Type(), true)
		dst.Code("]())\n")
		dst.Tab(1).Code("if err != nil {\n")
		dst.Tab(2).Code("return nil, err\n")
		dst.Tab(1).Code("}\n")
		dst.Tab(1).Code("return response.(*")
		b.printType(dst, method.Result.Type(), true)
		dst.Code("), nil\n")
		dst.Code("}\n\n")
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
	dst.Tab(0).Code("func Register").Code(serverName).Code("(r hrpc.ServerRegister, server ").Code(serverName).Code(") {\n")
	dst.Tab(1).Code("r.Register(0, \"").Code(build.StringToUnderlineName(typ.Name.Name)).Code("\", ").Code("server").Code(",\n")
	err := build.EnumMethod(typ, func(method *ast.FuncType, server *ast.ServerType) error {

		dst.Tab(2).Code("&hrpc.Method{\n")
		dst.Tab(3).Code("Name: \"").Code(build.StringToUnderlineName(method.Name.Name)).Code("\",\n")
		dst.Tab(3).Code("Tag: \"").Code(b.getFilterTag(method)).Code("\",\n")
		dst.Tab(3).Code("WithContext: func(ctx context.Context) context.Context {\n")
		au := b.getTag(method.Tags)
		if nil != au {
			keys := build.GetKeysByMap(*au)
			sort.Strings(keys)
			for _, key := range keys {
				values := (*au)[key]
				if len(values) > 0 {
					dst.Tab(4).Code("hrpc.AddTag(ctx, \"").Code(key)
					for _, val := range values {
						dst.Code("\", \"").Code(val)
					}
					dst.Code("\")\n")
				}
			}
		}
		dst.Tab(4).Code("return ctx\n")
		dst.Tab(3).Code("},\n")
		dst.Tab(3).Code("Handler: func(ctx context.Context, req any) (any, error) {\n")
		dst.Tab(4).Code("return server.").Code(build.StringToHumpName(method.Name.Name)).Code("(ctx, req.(*")
		b.printType(dst, method.Param, true)
		dst.Code("))\n")
		dst.Tab(3).Code("},\n")
		dst.Tab(3).Code("Decode: func(decoder func(v hbuf.Data) (hbuf.Data, error)) (hbuf.Data, error) {\n")
		dst.Tab(4).Code("return decoder(&")
		b.printType(dst, method.Param, true)
		dst.Code("{})\n")
		dst.Tab(3).Code("},\n")
		dst.Tab(2).Code("},\n")

		return nil
	})
	if err != nil {
		return
	}
	dst.Tab(1).Code(")\n")
	dst.Code("}\n\n")
}

func (b *Builder) printGetServerRouter(dst *build.Writer, typ *ast.ServerType) {
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hservice", "")
	serverName := build.StringToHumpName(typ.Name.Name)

	dst.Code("var NotFound" + serverName + " = &Default" + serverName + "{}\n\n")

	dst.Code("func Get" + serverName + "(ctx context.Context) " + serverName + " {\n")
	dst.Tab(1).Code("router := hservice.GetClient(ctx, \"").Code(build.StringToUnderlineName(serverName)).Code("\")\n")
	dst.Tab(1).Code("if nil == router {\n")
	dst.Tab(2).Code("return NotFound" + serverName + "\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("if val, ok := router.(" + serverName + "); ok {\n")
	dst.Tab(1).Code("	return val\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("return NotFound" + serverName + "\n")
	dst.Code("}\n\n")

	dst.Code("func " + serverName + "Name() string {\n")
	dst.Tab(1).Code("return \"").Code(build.StringToUnderlineName(typ.Name.Name)).Code("\"\n")
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

func (b *Builder) getFilterTag(method *ast.FuncType) string {
	tag, ok := build.GetTag(method.Tags, "filter")
	if !ok || nil == tag.KV {
		return ""
	}
	for _, item := range tag.KV {
		if item.Name.Name == "tag" {
			return item.Values[0].Value[1 : len(item.Values[0].Value)-1]
		}
	}
	return ""
}
