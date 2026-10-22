package tsgen

import (
	"fmt"
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"strconv"
	"strings"
)

// FileContext 一个 .hbuf 文件的模板根上下文。
type FileContext struct {
	*Ctx
	Package string         // ts 包名
	Datas   []*DataContext // data 类型列表(按 desc 依赖拓扑排序)
}

// NewFileContext 构建文件级模板上下文。
// 同文件内 data 按 desc 引用关系拓扑排序;环上节点用 NewSyncDescriptor 延迟构建。
func NewFileContext(ctx *Ctx, file *wire.File) *FileContext {
	// 每个 data 文件恒定依赖的运行时符号
	ctx.Import("DataDescriptor")
	ctx.TypeImport("HBufDescriptor")
	ctx.TypeImport("FieldDesc")
	ctx.TypeImport("HData")
	ctx.Import("NewDataDescriptor")
	ctx.Import("NewHBufDecoder")
	ctx.Import("NewHBufEncode")
	fc := &FileContext{Ctx: ctx, Package: ctx.Pack}
	for _, od := range orderDatas(ctx, file) {
		fc.Datas = append(fc.Datas, NewDataContext(ctx, od.Data, od.Sync))
	}
	return fc
}

// orderedData 拓扑排序结果。
type orderedData struct {
	Data *wire.DataType
	Sync bool
}

// orderDatas 同文件内 data 按 desc 依赖拓扑排序(被引用的 desc 必须先声明)。
// 存在环(自引用/互引用)时,取声明序最早的环上节点标记 Sync,
// 其 NewDataDescriptor 会被 NewSyncDescriptor 包裹以延迟求值,从而打破循环。
func orderDatas(ctx *Ctx, file *wire.File) []*orderedData {
	var datas []*wire.DataType
	for _, t := range file.Types {
		if t != nil && t.Kind == "data" && t.Data != nil {
			datas = append(datas, t.Data)
		}
	}
	// dependents[A] = {B...}:B 的声明依赖 A(即 A 必须在 B 之前声明)
	dependents := map[string]map[string]bool{}
	degree := map[string]int{}
	for _, d := range datas {
		deps := map[string]bool{}
		for _, e := range d.Extends {
			if e == nil {
				continue
			}
			if t, owner := ctx.index.Find(e.Name); t != nil && t.Kind == "data" && owner != nil && owner.Path == file.Path {
				deps[e.Name] = true
			}
		}
		for _, f := range d.Fields {
			if f == nil {
				continue
			}
			// 自引用也计入依赖:自环同样由环检测标记 Sync(NewSyncDescriptor 延迟构建)
			collectRefs(f.Type, func(name string) {
				if t, owner := ctx.index.Find(name); t != nil && t.Kind == "data" && owner != nil && owner.Path == file.Path {
					deps[name] = true
				}
			})
		}
		for dep := range deps {
			if dependents[dep] == nil {
				dependents[dep] = map[string]bool{}
			}
			dependents[dep][d.Name] = true
			degree[d.Name]++
		}
	}
	remaining := map[string]*wire.DataType{}
	for _, d := range datas {
		remaining[d.Name] = d
	}
	var out []*orderedData
	var queue []string
	for _, d := range datas {
		if degree[d.Name] == 0 {
			queue = append(queue, d.Name)
		}
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if remaining[n] == nil {
			continue
		}
		out = append(out, &orderedData{Data: remaining[n]})
		delete(remaining, n)
		for dname := range dependents[n] {
			if remaining[dname] == nil {
				continue
			}
			degree[dname]--
			if degree[dname] == 0 {
				queue = append(queue, dname)
			}
		}
		if len(queue) == 0 && len(remaining) > 0 {
			// 剩余节点构成环:取声明序最早的节点标记 Sync 并移除
			for _, d := range datas {
				if remaining[d.Name] == nil {
					continue
				}
				out = append(out, &orderedData{Data: d, Sync: true})
				delete(remaining, d.Name)
				for dname := range dependents[d.Name] {
					if remaining[dname] == nil {
						continue
					}
					degree[dname]--
					if degree[dname] == 0 {
						queue = append(queue, dname)
					}
				}
				break
			}
		}
	}
	// 防御:理论不可达
	for _, d := range datas {
		if remaining[d.Name] != nil {
			out = append(out, &orderedData{Data: d, Sync: true})
			delete(remaining, d.Name)
		}
	}
	return out
}

// collectRefs 收集类型树中的 data 引用名(嵌套数组/映射递归)。
func collectRefs(vt *wire.VarType, fn func(name string)) {
	if vt == nil {
		return
	}
	switch vt.Kind {
	case "ref":
		if vt.RefKind == "data" {
			fn(vt.Name)
		}
	case "array":
		collectRefs(vt.Element, fn)
	case "map":
		collectRefs(vt.Value, fn)
	}
}

// DataContext data 类型的模板上下文。
type DataContext struct {
	*Ctx
	Name         string           // 类名(无 Ts 后缀,与旧版 typescript 后端一致)
	DescName     string           // descriptor 单例名:xxxDesc
	DescOpen     string           // descriptor 声明开头(含类型注解与构造调用)
	Export       string           // 恒为 "export "(与旧版后端一致:所有 data 类与 desc 一律导出)
	Sync         bool             // 递归:NewSyncDescriptor 包裹
	Extends      []*ExtendContext // 直接父类 implements 列表 + extends_ Map 条目
	ExtendFields []*FieldContext  // 展平的祖先字段(递归,先父后自身)
	Fields       []*FieldContext  // 自身字段
	Doc          string
}

// NewDataContext 构建 data 类型的模板上下文。
// 继承采用与旧版 typescript 后端一致的方式:类 implements 直接父类,
// 祖先字段递归展平声明(desc 的 extends_ Map 仍按扩展 ID 挂父 desc 保证二进制兼容)。
// 所有类与 desc 一律 export:仅被同文件 extends 引用的父类也可能被业务代码 import,
// 旧版后端即全量导出,按需导出会破坏 API 兼容。
func NewDataContext(ctx *Ctx, data *wire.DataType, sync bool) *DataContext {
	dc := &DataContext{
		Ctx:      ctx,
		Name:     tpl.HumpName(data.Name),
		DescName: lowerCamel(data.Name) + "Desc",
		Export:   "export ",
		Sync:     sync,
		Doc:      data.Doc,
	}
	if sync {
		ctx.Import("NewSyncDescriptor")
		dc.DescOpen = fmt.Sprintf("%sconst %s: HBufDescriptor = NewSyncDescriptor(() => NewDataDescriptor(", dc.Export, dc.DescName)
	} else {
		dc.DescOpen = fmt.Sprintf("%sconst %s: DataDescriptor = NewDataDescriptor(", dc.Export, dc.DescName)
	}
	for _, e := range data.Extends {
		if e == nil {
			continue
		}
		ref := &wire.VarType{Kind: "ref", Name: e.Name, RefKind: "data"}
		ext := &ExtendContext{
			Ctx:  ctx,
			Type: ctx.refType(ref),
			Id:   strconv.FormatInt(e.Id, 10),
			Desc: ctx.refDesc(ref),
		}
		dc.Extends = append(dc.Extends, ext)
	}
	// 展平祖先字段(旧版后端 EnumField 语义:递归父类,父优先去重)
	visited := map[string]bool{}
	names := map[string]bool{}
	for _, e := range data.Extends {
		if e == nil {
			continue
		}
		collectExtendFields(ctx, e.Name, &dc.ExtendFields, visited, names)
	}
	for _, f := range data.Fields {
		if f == nil || f.Type == nil {
			continue
		}
		if names[f.Name] {
			// 父类已有同名字段,自身声明被覆盖(与旧版后端一致)
			continue
		}
		names[f.Name] = true
		dc.Fields = append(dc.Fields, newFieldContext(ctx, f))
	}
	return dc
}

// collectExtendFields 递归收集父类型的展平字段(先父的祖先、后父自身),按字段名去重。
func collectExtendFields(ctx *Ctx, name string, out *[]*FieldContext, visited, names map[string]bool) {
	if visited[name] {
		return
	}
	visited[name] = true
	t, _ := ctx.index.Find(name)
	if t == nil || t.Kind != "data" || t.Data == nil {
		return
	}
	for _, e := range t.Data.Extends {
		if e == nil {
			continue
		}
		collectExtendFields(ctx, e.Name, out, visited, names)
	}
	for _, f := range t.Data.Fields {
		if f == nil || f.Type == nil || names[f.Name] {
			continue
		}
		names[f.Name] = true
		*out = append(*out, newFieldContext(ctx, f))
	}
}

// newFieldContext 构建字段的模板上下文。
func newFieldContext(ctx *Ctx, f *wire.Field) *FieldContext {
	name := lowerCamel(f.Name)
	key := tpl.UnderlineName(f.Name)
	fc := &FieldContext{
		Ctx:      ctx,
		Name:     name,
		Id:       strconv.FormatInt(f.Id, 10),
		Type:     ctx.tsType(f.Type),
		Default:  ctx.tsDefault(f.Type),
		Desc:     ctx.tsDescriptor(f.Type),
		Comment:  strings.Trim(strings.ReplaceAll(f.Doc, "\n", " "), " "),
		JsonKey:  key,
		FromJSON: ctx.jsonFrom(f.Type, fmt.Sprintf("map[%q]", key)),
		ToJSON:   ctx.jsonTo(f.Type, "this."+name),
		CopyFrom: ctx.tsCopy(f.Type, "this."+name),
		Cmp:      ctx.tsCmp(f.Type, "this."+name, "val."+name),
		RawName:  f.Name,
	}
	// enum 字段为 enum 类实例,descriptor 仍按 int32 编码:
	// get 取 .value 数字,set 用 valueOf 还原实例
	if f.Type != nil && f.Type.Kind == "ref" && f.Type.RefKind == "enum" {
		if f.Type.Empty {
			fc.Get = "o." + name + "?.value"
		} else {
			fc.Get = "o." + name + ".value"
		}
		fc.Set = "o." + name + " = " + tpl.HumpName(f.Type.Name) + ".valueOf(v)"
		fc.SetType = "number"
	} else if f.Type != nil && f.Type.Kind == "base" && f.Type.Name == "decimal" {
		// decimal 字段为 Decimal 实例,descriptor 以字符串承载(wire 上十进制文本):
		// get 取 toString(),set 用 new Decimal(v) 还原实例
		ctx.decimalImport()
		if f.Type.Empty {
			fc.Get = "o." + name + "?.toString()"
		} else {
			fc.Get = "o." + name + ".toString()"
		}
		fc.Set = "o." + name + " = new Decimal(v)"
		fc.SetType = "string"
	} else if f.Type != nil && f.Type.Kind == "map" {
		// map 字段为 Record 对象(与旧版后端一致),descriptor 层仍操作 Map:
		// get 转 Map 供编码(key 恢复为 descriptor 键类型),set 转回 Record
		keyConv := "k"
		switch f.Type.Key {
		case "int64", "uint64":
			keyConv = "BigInt(k)"
		case "string", "decimal":
			// 字符串键原样
		default:
			// 数字类键(含 enum/data)转 number
			keyConv = "Number(k)"
		}
		mk := fmt.Sprintf("(() => { const m = new Map<any, any>(); const r: any = o.%s; for (const k of Object.keys(r)) { m.set(%s, r[k]) } return m })()", name, keyConv)
		if f.Type.Empty {
			fc.Get = "null == o." + name + " ? null : " + mk
		} else {
			fc.Get = mk
		}
		fc.Set = fmt.Sprintf("o.%s = (() => { const r: Record<string, any> = {}; for (const [k, e] of v) { r[String(k)] = e } return r })()", name)
		fc.SetType = "Map<any, any>"
	} else {
		fc.Get = "o." + name
		fc.Set = "o." + name + " = v"
		fc.SetType = fc.Type
	}
	return fc
}

// FieldContext 数据字段的模板上下文。
type FieldContext struct {
	*Ctx
	Name     string // 小驼峰属性名
	Id       string // id 字符串
	Type     string // TS 类型声明
	Default  string // 默认值表达式
	Desc     string // descriptor 构造表达式
	Get      string // descriptor get 闭包体(enum 字段取 .value)
	Set      string // descriptor set 闭包体(enum 字段 valueOf 还原)
	SetType  string // descriptor set 闭包参数类型(enum 字段恒为 number)
	Comment  string // 文档注释(已压缩单行)
	JsonKey  string // JSON 键(下划线形式)
	FromJSON string // JSON 读取表达式
	ToJSON   string // JSON 写出表达式
	CopyFrom string // clone 拷贝表达式
	Cmp      string // $getChangeField 比较表达式
	RawName  string // 原始字段名(IDL 声明名,用于 $fieldMaps)
}

// ExtendContext 直接父类的模板上下文(类 implements 列表与 extends_ Map 条目)。
type ExtendContext struct {
	*Ctx
	Type string // 父类类名
	Id   string // 扩展 ID
	Desc string // 父 desc 引用(descriptor 构造表达式)
}

// EnumFileContext 文件级 enum 模板上下文。
type EnumFileContext struct {
	*Ctx
	Enums []*EnumContext
}

// NewEnumFileContext 构建文件级 enum 上下文。
func NewEnumFileContext(ctx *Ctx, file *wire.File) *EnumFileContext {
	fc := &EnumFileContext{Ctx: ctx}
	for _, t := range file.Types {
		if t == nil || t.Kind != "enum" || t.Enum == nil {
			continue
		}
		fc.Enums = append(fc.Enums, NewEnumContext(ctx, t.Enum))
	}
	return fc
}

// EnumContext 枚举类型的模板上下文(类名无 Ts 后缀,与进程内 typescript 后端一致)。
type EnumContext struct {
	*Ctx
	Name  string // 类名
	Items []*EnumItemContext
	Doc   string
}

// NewEnumContext 构建枚举类型的模板上下文。
func NewEnumContext(ctx *Ctx, e *wire.EnumType) *EnumContext {
	ec := &EnumContext{Ctx: ctx, Name: tpl.HumpName(e.Name), Doc: e.Doc}
	for _, item := range e.Items {
		if item == nil {
			continue
		}
		typ, css, _ := uiEnum(item.Tags, item.Name)
		ec.Items = append(ec.Items, &EnumItemContext{
			Name:     tpl.AllUpperName(item.Name),
			Value:    strconv.FormatInt(item.Id, 10),
			ShowName: tpl.HumpName(item.Name),
			Type:     typ,
			CssClass: css,
			Doc:      strings.Trim(strings.ReplaceAll(item.Doc, "\n", " "), " "),
		})
	}
	return ec
}

// EnumItemContext 枚举项的模板上下文。
type EnumItemContext struct {
	Name     string // 常量名(全大写下划线)
	Value    string // id 字符串
	ShowName string // 驼峰显示名
	Type     string // ui type
	CssClass string // ui cssClass
	Doc      string // 文档注释(已压缩单行)
}

// ServerFileContext 文件级 server 模板上下文。
type ServerFileContext struct {
	*Ctx
	Servers []*ServerContext
}

// NewServerFileContext 构建文件级 server 上下文。
func NewServerFileContext(ctx *Ctx, file *wire.File) *ServerFileContext {
	// 每个 server 文件恒定依赖的运行时符号
	ctx.Import("Client")
	ctx.Import("Server")
	ctx.TypeImport("Option")
	ctx.TypeImport("RequestType")
	ctx.TypeImport("ResponseType")
	fc := &ServerFileContext{Ctx: ctx}
	for _, t := range file.Types {
		if t == nil || t.Kind != "server" || t.Server == nil {
			continue
		}
		fc.Servers = append(fc.Servers, NewServerContext(ctx, t.Server))
	}
	return fc
}

// ServerContext server 类型的模板上下文。
type ServerContext struct {
	*Ctx
	Name    string   // 接口名(无后缀)
	Route   string   // 下划线服务名(路由)
	Extends []string // 父接口名(接口 extends,跨文件时登记 import)
	Methods []*ServerMethodContext
	Doc     string
}

// NewServerContext 构建 server 类型的模板上下文。
func NewServerContext(ctx *Ctx, s *wire.ServerType) *ServerContext {
	sc := &ServerContext{Ctx: ctx, Name: tpl.HumpName(s.Name), Route: tpl.UnderlineName(s.Name), Doc: s.Doc}
	for _, e := range s.Extends {
		if e == nil {
			continue
		}
		ctx.serverImport(e.Name)
		sc.Extends = append(sc.Extends, tpl.HumpName(e.Name))
	}
	collectServerMethods(ctx, s, &sc.Methods)
	return sc
}

// collectServerMethods 递归展开方法(先自身后父类,与进程内后端 EnumMethod 一致)。
func collectServerMethods(ctx *Ctx, s *wire.ServerType, out *[]*ServerMethodContext) {
	for _, m := range s.Methods {
		if m == nil {
			continue
		}
		*out = append(*out, NewServerMethodContext(ctx, m))
	}
	for _, e := range s.Extends {
		if e == nil {
			continue
		}
		if t, _ := ctx.index.Find(e.Name); t != nil && t.Kind == "server" && t.Server != nil {
			collectServerMethods(ctx, t.Server, out)
		}
	}
}

// ServerMethodContext 服务方法的模板上下文。
type ServerMethodContext struct {
	*Ctx
	Name       string // 小驼峰方法名
	Route      string // 下划线方法名(路由)
	Id         string // 方法 id 字符串
	ParamName  string // 小驼峰参数名(无参/void 时为空)
	ParamType  string // 参数 TS 类型(void 参数为空)
	ResultType string // 返回 TS 类型
	FromParam  string // 参数 data 类名(服务端 fromMap),空表示无
	FromResult string // 结果 data 类名(客户端 invoke fromMap),空表示无
	Doc        string
}

// NewServerMethodContext 构建服务方法的模板上下文。
func NewServerMethodContext(ctx *Ctx, m *wire.Method) *ServerMethodContext {
	mc := &ServerMethodContext{
		Ctx:        ctx,
		Name:       lowerCamel(m.Name),
		Route:      tpl.UnderlineName(m.Name),
		Id:         strconv.FormatInt(m.Id, 10),
		Doc:        strings.Trim(strings.ReplaceAll(m.Doc, "\n", " "), " "),
		ResultType: "void",
	}
	// 入参:void / stream / 普通类型
	if p := m.Param; p != nil && p.Name != "void" {
		if p.Name == "stream" {
			mc.ParamName = m.ParamName
			if mc.ParamName == "" {
				mc.ParamName = "data"
			}
			mc.ParamType = "BufferType"
			ctx.TypeImport("BufferType")
		} else {
			mc.ParamName = lowerCamel(m.ParamName)
			if mc.ParamName == "" {
				mc.ParamName = "param"
			}
			mc.ParamType = ctx.tsType(p)
			if p.Kind == "ref" && p.RefKind == "data" {
				mc.FromParam = tpl.HumpName(p.Name)
				// data/server 是独立输出文件,同文件 data 也需登记 import
				ctx.serverDataImport(p.Name)
			}
		}
	}
	// 返回值:void(广播) / stream / 普通类型
	if r := m.Result; r != nil && r.Name != "void" {
		if r.Name == "stream" {
			mc.ResultType = "BufferType"
			ctx.TypeImport("BufferType")
		} else {
			mc.ResultType = ctx.tsType(r)
			if r.Kind == "ref" && r.RefKind == "data" {
				mc.FromResult = tpl.HumpName(r.Name)
				// data/server 是独立输出文件,同文件 data 也需登记 import
				ctx.serverDataImport(r.Name)
			}
		}
	}
	return mc
}
