package gogen

import (
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"strconv"
	"strings"
)

// FileContext 一个 .hbuf 文件的模板根上下文。
type FileContext struct {
	*Ctx
	Package string           // go 包名
	Datas   []*DataContext   // data 类型列表
	Enums   []*EnumContext   // enum 类型列表
	Servers []*ServerContext // server 类型列表

	// 各输出文件构建期登记的 import 集合(避免 data 的 import 污染 enum/server 文件)
	dataImports   map[string]bool
	enumImports   map[string]bool
	serverImports map[string]bool
}

// NewFileContext 构建文件级模板上下文。
// data/enum/server 三类上下文分别用独立的 import 集合构建,
// 各输出文件只携带自己类别实际用到的 import。
func NewFileContext(ctx *Ctx, file *wire.File) *FileContext {
	fc := &FileContext{
		Ctx:           ctx,
		Package:       ctx.Pack,
		dataImports:   map[string]bool{},
		enumImports:   map[string]bool{},
		serverImports: map[string]bool{},
	}
	for _, t := range file.Types {
		if t == nil {
			continue
		}
		switch t.Kind {
		case "data":
			if t.Data != nil {
				sub := ctx.clone()
				sub.Imports = fc.dataImports
				fc.Datas = append(fc.Datas, NewDataContext(sub, t.Data))
			}
		case "enum":
			if t.Enum != nil {
				sub := ctx.clone()
				sub.Imports = fc.enumImports
				fc.Enums = append(fc.Enums, NewEnumContext(sub, t.Enum))
			}
		case "server":
			if t.Server != nil {
				sub := ctx.clone()
				sub.Imports = fc.serverImports
				fc.Servers = append(fc.Servers, NewServerContext(sub, t.Server))
			}
		}
	}
	return fc
}

// FieldContext 数据字段的模板上下文。
type FieldContext struct {
	*Ctx
	Name    string // 驼峰字段名
	Id      string // id 字符串
	Type    string // 声明类型(含顶层指针)
	GetType string // Getter 返回类型
	Default string // 空值时的默认值表达式
	Desc    string // descriptor 构造表达式
	Tag     string // struct tag
	Comment string // 文档注释(已压缩单行)
	Ptr     bool   // 可空且非数组/映射:Getter 判空、Setter 取地址
	HasDB   bool   // 字段带 db 注解:Setter 需同步置 changeFields
}

// DataContext data 类型的模板上下文。
type DataContext struct {
	*Ctx
	Name      string // 驼峰类型名
	LowName   string // 小驼峰类型名
	Extends   []*ExtendContext
	Fields    []*FieldContext
	Doc       string
	HasChange bool // db 注解 change=self|parent:生成 changeFields 字段
}

// ExtendContext 继承项的模板上下文。
type ExtendContext struct {
	*Ctx
	Name   string // 引用类型(含包前缀,驼峰)
	Id     string
	DescId string // id 对齐后的 "id: " 部分
	Desc   string // descriptor 构造表达式
}

// NewDataContext 构建 data 类型的模板上下文。
func NewDataContext(ctx *Ctx, data *wire.DataType) *DataContext {
	dc := &DataContext{
		Ctx:     ctx,
		Name:    tpl.HumpName(data.Name),
		LowName: tpl.LowerFirstName(data.Name),
		Doc:     data.Doc,
	}
	// 与进程内 printDataStruct 一致:db 注解 change=self|parent 时生成 changeFields 字段
	if change, ok := tagKV(data.Tags, "db", "change"); ok {
		change = strings.ToLower(change)
		dc.HasChange = "self" == change || "parent" == change
	}
	for _, e := range data.Extends {
		if e == nil {
			continue
		}
		ec := &ExtendContext{
			Ctx:    ctx,
			Name:   ctx.refHump(extPack(ctx, e.Name), e.Name),
			Id:     strconv.FormatInt(e.Id, 10),
			DescId: padId(e.Id, data.Extends),
		}
		ec.Desc = ctx.goDescriptor(&wire.VarType{Kind: "ref", Name: e.Name, RefPack: extPack(ctx, e.Name)}, dc.LowName, tpl.HumpName(e.Name))
		dc.Extends = append(dc.Extends, ec)
	}
	for _, f := range data.Fields {
		if f == nil || f.Type == nil {
			continue
		}
		fc := &FieldContext{
			Ctx:     ctx,
			Name:    tpl.HumpName(f.Name),
			Id:      strconv.FormatInt(f.Id, 10),
			Type:    ctx.goType(f.Type, true),
			GetType: ctx.goType(f.Type, false),
			Default: ctx.goDefault(f.Type),
			Desc:    ctx.goDescriptor(f.Type, dc.LowName, tpl.HumpName(f.Name)),
			Tag:     ctx.goStructTag(f),
			Comment: strings.Trim(strings.ReplaceAll(f.Doc, "\n", " "), " "),
			Ptr:     f.Type.Empty && f.Type.Kind != "array" && f.Type.Kind != "map",
			HasDB:   hasTag(f.Tags, "db"),
		}
		dc.Fields = append(dc.Fields, fc)
	}
	return dc
}

// hasTag 判断注解列表中是否存在指定名称的注解。
func hasTag(tags []*wire.Tag, name string) bool {
	for _, t := range tags {
		if t != nil && t.Name == name {
			return true
		}
	}
	return false
}

// tagKV 取注解 name 中指定 key 的第一个值,如 [db:change="self"] 的 change。
func tagKV(tags []*wire.Tag, name, key string) (string, bool) {
	for _, t := range tags {
		if t == nil || t.Name != name {
			continue
		}
		for _, kv := range t.KV {
			if kv == nil || len(kv.Values) == 0 {
				continue
			}
			if kv.Name == key {
				return kv.Values[0], true
			}
		}
	}
	return "", false
}

// extPack 继承引用的目标包名(通过类型索引解析)。
func extPack(ctx *Ctx, name string) string {
	if _, f := ctx.index.Find(name); f != nil {
		return packOf(f, "go")
	}
	return ""
}

// padId 生成 id 对齐文本,如 "1:    "。
func padId(id int64, items interface{}) string {
	_ = items
	return strconv.FormatInt(id, 10)
}

// EnumItemContext 枚举项的模板上下文。
type EnumItemContext struct {
	*Ctx
	Name    string
	Id      string
	Comment string
}

// EnumContext enum 类型的模板上下文。
type EnumContext struct {
	*Ctx
	Name    string
	LowName string
	Items   []*EnumItemContext
	Doc     string
}

// NewEnumContext 构建 enum 类型的模板上下文。
func NewEnumContext(ctx *Ctx, enum *wire.EnumType) *EnumContext {
	ec := &EnumContext{
		Ctx:     ctx,
		Name:    tpl.HumpName(enum.Name),
		LowName: tpl.LowerFirstName(enum.Name),
		Doc:     enum.Doc,
	}
	for _, item := range enum.Items {
		if item == nil {
			continue
		}
		ec.Items = append(ec.Items, &EnumItemContext{
			Ctx:     ctx,
			Name:    tpl.HumpName(item.Name),
			Id:      strconv.FormatInt(item.Id, 10),
			Comment: strings.Trim(strings.ReplaceAll(item.Doc, "\n", " "), " "),
		})
	}
	return ec
}

// MethodContext 服务方法的模板上下文。
type MethodContext struct {
	*Ctx
	Name       string // 驼峰方法名
	LowName    string // 下划线方法名(注册/调用用)
	ParamName  string // 小驼峰参数名
	ParamDecl  string // 参数声明:io.Reader 或 *Type
	ParamType  string // 参数类型(不含星)
	ResultType string // 结果类型(不含星,void/stream 为特殊名)
	Sig        string // 返回签名:error / (io.ReadCloser, error) / (*Type, error)
	Void       bool   // 无返回(void 或广播)
	StreamIn   bool   // 流式入参
	StreamOut  bool   // 流式返回
	Tags       []*TagContext
	Bind       *BindContext
	Comment    string
}

// TagContext hrpc.AddTag 参数。
type TagContext struct {
	Key    string
	Values []string
}

// BindContext bind 注解解析结果(Default 方法转发目标)。
type BindContext struct {
	Call    string // 转发调用表达式:GetServer(ctx).Method(ctx, &req.Type{})
	RespVar bool   // 是否有响应变量
	Result  string // 返回表达式主体
}

// ServerContext server 类型的模板上下文。
type ServerContext struct {
	*Ctx
	Name    string // 驼峰服务名
	LowName string // 下划线服务名
	Extends []*ExtendContext
	Methods []*MethodContext
	Doc     string
}

// NewServerContext 构建 server 类型的模板上下文。
func NewServerContext(ctx *Ctx, server *wire.ServerType) *ServerContext {
	sc := &ServerContext{
		Ctx:     ctx,
		Name:    tpl.HumpName(server.Name),
		LowName: tpl.UnderlineName(server.Name),
		Doc:     server.Doc,
	}
	for _, e := range server.Extends {
		if e == nil {
			continue
		}
		sc.Extends = append(sc.Extends, &ExtendContext{
			Ctx:    ctx,
			Name:   ctx.refHump(extPack(ctx, e.Name), e.Name),
			Id:     strconv.FormatInt(e.Id, 10),
			DescId: padId(e.Id, server.Extends),
		})
	}
	for _, m := range server.Methods {
		if m == nil {
			continue
		}
		mc := &MethodContext{
			Ctx:       ctx,
			Name:      tpl.HumpName(m.Name),
			LowName:   tpl.UnderlineName(m.Name),
			ParamName: tpl.LowerFirstName(m.ParamName),
			Comment:   strings.Trim(strings.ReplaceAll(m.Doc, "\n", " "), " "),
		}
		if m.Param != nil {
			if isStream(m.Param) {
				mc.StreamIn = true
				mc.ParamDecl = " io.Reader"
			} else {
				mc.ParamType = ctx.goType(m.Param, false)
				mc.ParamDecl = " *" + mc.ParamType
			}
		}
		switch {
		case m.Result == nil || isVoid(m.Result):
			mc.Void = true
		case isStream(m.Result):
			mc.StreamOut = true
		default:
			mc.ResultType = ctx.goType(m.Result, false)
		}
		switch {
		case mc.Void:
			mc.Sig = "error"
		case mc.StreamOut:
			mc.Sig = "(io.ReadCloser, error)"
		default:
			mc.Sig = "(*" + mc.ResultType + ", error)"
		}
		for _, t := range m.Tags {
			if t == nil || t.Name != "tag" {
				continue
			}
			for _, kv := range t.KV {
				if kv == nil {
					continue
				}
				mc.Tags = append(mc.Tags, &TagContext{Key: kv.Name, Values: kv.Values})
			}
		}
		mc.Bind = resolveBind(ctx, m)
		sc.Methods = append(sc.Methods, mc)
	}
	return sc
}

func isStream(vt *wire.VarType) bool {
	return vt != nil && vt.Kind == "base" && vt.Name == "stream"
}

func isVoid(vt *wire.VarType) bool {
	return vt != nil && vt.Kind == "base" && vt.Name == "void"
}

// resolveBind 解析方法的 bind 注解,生成转发调用表达式。
func resolveBind(ctx *Ctx, m *wire.Method) *BindContext {
	var target string
	for _, t := range m.Tags {
		if t == nil || t.Name != "bind" {
			continue
		}
		for _, kv := range t.KV {
			if kv == nil || kv.Name != "value" || len(kv.Values) == 0 {
				continue
			}
			target = kv.Values[0]
		}
	}
	if target == "" {
		return nil
	}
	parts := strings.Split(target, ".")
	if len(parts) != 2 {
		return nil
	}
	typ, file := ctx.index.Find(parts[0])
	if typ == nil || typ.Kind != "server" || typ.Server == nil {
		return nil
	}
	serverName := parts[0]
	methodName := parts[1]
	pack := ""
	if file != nil {
		pack = ctx.refPrefix(packOf(file, "go"))
	}
	bind := &BindContext{}
	// 找到 bind 目标方法的参数与结果类型用于构造转发
	var bindParam, result *wire.VarType
	for _, bm := range typ.Server.Methods {
		if bm != nil && bm.Name == methodName {
			bindParam = bm.Param
			result = bm.Result
		}
	}
	if m.Param != nil && isStream(m.Param) {
		bind.Call = pack + "Get" + serverName + "(ctx)." + methodName + "(ctx, req)"
	} else if bindParam != nil {
		// 转发参数取 req 中嵌入的目标参数类型字段(嵌入字段名不带包前缀)
		bind.Call = pack + "Get" + serverName + "(ctx)." + methodName + "(ctx, &req." + tpl.HumpName(bindParam.Name) + ")"
	} else {
		bind.Call = pack + "Get" + serverName + "(ctx)." + methodName + "(ctx, &req." + tpl.HumpName(m.Param.Name) + ")"
	}
	if m.Result == nil || isVoid(m.Result) {
		bind.Result = "nil"
	} else if result != nil && !isVoid(result) && !isStream(result) {
		bind.RespVar = true
		// 响应包装:当前结果内嵌目标结果,复合字面量 key 用纯类型名
		bind.Result = "&" + ctx.goType(m.Result, false) + "{" + tpl.HumpName(result.Name) + ": *resp}, nil"
	} else {
		bind.RespVar = true
		bind.Result = "nil, nil"
	}
	return bind
}
