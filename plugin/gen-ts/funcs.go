// Package tsgen 实现 hbuf-gen-ts 插件:把 wire 中性描述模型渲染成 TypeScript 数据类。
//
// 渲染采用模板引擎 plugin/core/tpl:
//   - 默认模板(templates/*.tmpl)通过 //go:embed 编译进插件二进制;
//   - 外部自定义模板目录(插件参数 template=<目录>)同名文件覆盖默认模板。
//
// 生成代码遵循 test/ts/hbuf-ts-codegen.md 规范:每个 data 生成一个实现 HData
// 接口的类与对应的 descriptor 单例,可直接对接 hbuf TS 运行时,与 Go 版字节级兼容。
package tsgen

import (
	"fmt"
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"path/filepath"
	"strings"
)

// baseTypes 基础类型 -> TS 类型映射(与 hbuf-ts-codegen.md 第 4 节一致)
var baseTypes = map[string]string{
	"int8": "number", "int16": "number", "int32": "number", "int64": "bigint",
	"uint8": "number", "uint16": "number", "uint32": "number", "uint64": "bigint",
	"bool": "boolean", "float": "number", "double": "number",
	"string": "string", "date": "Date | null", "decimal": "string", "bytes": "Uint8Array",
}

// baseDefaults 基础类型默认值表达式(与规范第 5.2 节一致)
var baseDefaults = map[string]string{
	"int8": "0", "int16": "0", "int32": "0", "int64": "0n",
	"uint8": "0", "uint16": "0", "uint32": "0", "uint64": "0n",
	"bool": "false", "float": "0", "double": "0",
	"string": `""`, "date": "null", "decimal": `""`, "bytes": "new Uint8Array(0)",
}

// descNames 基础类型 -> 运行时 descriptor 构造函数名
var descNames = map[string]string{
	"int8": "Int8", "int16": "Int16", "int32": "Int32", "int64": "Int64",
	"uint8": "Uint8", "uint16": "Uint16", "uint32": "Uint32", "uint64": "Uint64",
	"bool": "Bool", "float": "Float", "double": "Double",
	"string": "String", "date": "Time", "decimal": "Decimal", "bytes": "Bytes",
}

// TypeIndex 全包类型索引:类型名 -> 声明它的文件。
type TypeIndex struct {
	owner map[string]*wire.File
	typ   map[string]*wire.Type
}

// NewTypeIndex 基于请求携带的全部文件构建类型索引。
func NewTypeIndex(files []*wire.File) *TypeIndex {
	idx := &TypeIndex{owner: map[string]*wire.File{}, typ: map[string]*wire.Type{}}
	for _, f := range files {
		if f == nil {
			continue
		}
		for _, t := range f.Types {
			if t == nil {
				continue
			}
			name := typeName(t)
			if name == "" {
				continue
			}
			idx.owner[name] = f
			idx.typ[name] = t
		}
	}
	return idx
}

// Find 返回类型声明与其所属文件。
func (i *TypeIndex) Find(name string) (*wire.Type, *wire.File) {
	return i.typ[name], i.owner[name]
}

func typeName(t *wire.Type) string {
	switch t.Kind {
	case "enum":
		if t.Enum != nil {
			return t.Enum.Name
		}
	case "data":
		if t.Data != nil {
			return t.Data.Name
		}
	case "server":
		if t.Server != nil {
			return t.Server.Name
		}
	}
	return ""
}

// Ctx 单个 .hbuf 文件的渲染上下文。
type Ctx struct {
	Req      *wire.Request
	File     *wire.File
	Pack     string                     // 本文件 ts 包名
	Runtime  string                     // 运行时包 import 路径
	Imports  map[string]bool            // 运行时符号集合
	CrossRef map[string]map[string]bool // 跨文件 import:相对路径 -> 符号集合
	index    *TypeIndex
}

// NewCtx 创建渲染上下文。
func NewCtx(req *wire.Request, file *wire.File, index *TypeIndex, pack, runtime string) *Ctx {
	return &Ctx{
		Req:      req,
		File:     file,
		Pack:     pack,
		Runtime:  runtime,
		Imports:  map[string]bool{},
		CrossRef: map[string]map[string]bool{},
		index:    index,
	}
}

// clone 复制上下文:同一输出文件的符号集合共享(import 块按文件汇总)。
func (c *Ctx) clone() *Ctx {
	return &Ctx{
		Req:      c.Req,
		File:     c.File,
		Pack:     c.Pack,
		Runtime:  c.Runtime,
		Imports:  c.Imports,
		CrossRef: c.CrossRef,
		index:    c.index,
	}
}

// Import 登记运行时符号(幂等)。
func (c *Ctx) Import(symbol string) {
	if symbol != "" {
		c.Imports[symbol] = true
	}
}

// importFrom 登记跨文件 import:相对路径 -> 符号。
func (c *Ctx) importFrom(path, symbol string) {
	if path == "" || symbol == "" {
		return
	}
	if c.CrossRef[path] == nil {
		c.CrossRef[path] = map[string]bool{}
	}
	c.CrossRef[path][symbol] = true
}

// relImport 计算目标文件生成的相对 import 路径(不含 .ts 扩展名)。
// 生成文件名是 <name><suffix>.ts,如 <name>.data.ts / <name>.enum.ts / <name>.server.ts。
func (c *Ctx) relImport(owner *wire.File, suffix string) string {
	if owner == nil {
		return ""
	}
	curPackDir := filepath.Join(strings.Split(c.Pack, ".")...)
	refPackDir := filepath.Join(strings.Split(packOf(owner, "ts"), ".")...)
	target := filepath.Join(refPackDir, baseName(owner.Path)+suffix)
	rel, err := filepath.Rel(curPackDir, target)
	if err != nil {
		rel = target
	}
	rel = strings.ReplaceAll(rel, "\\", "/")
	if !strings.HasPrefix(rel, ".") {
		rel = "./" + rel
	}
	return rel
}

// tsType 把 wire 类型渲染成 TS 类型声明。
func (c *Ctx) tsType(vt *wire.VarType) string {
	if vt == nil {
		return "any"
	}
	// 规范:数组/映射不因可空改变声明(空容器本身即可表示空)
	null := ""
	if vt.Empty && vt.Kind != "array" && vt.Kind != "map" {
		null = " | null"
	}
	switch vt.Kind {
	case "base":
		if vt.Name == "date" {
			// 陷阱:date 恒为 Date | null(零时间用 null 表示)
			return baseTypes[vt.Name]
		}
		return c.baseType(vt.Name) + null
	case "ref":
		return c.refType(vt) + null
	case "array":
		elem := c.tsType(vt.Element)
		if strings.Contains(elem, "|") {
			elem = "(" + elem + ")"
		}
		return elem + "[]"
	case "map":
		return "Map<" + c.keyType(vt.Key) + ", " + c.tsType(vt.Value) + ">"
	}
	return "any"
}

// baseType 基础类型映射。
func (c *Ctx) baseType(name string) string {
	if v, ok := baseTypes[name]; ok {
		return v
	}
	return name
}

// refType 引用类型的 TS 类型:enum -> number,data -> Xxx(类名无 Ts 后缀,与旧版 typescript 后端一致)。
func (c *Ctx) refType(vt *wire.VarType) string {
	if vt.RefKind == "enum" {
		return "number"
	}
	name := tpl.HumpName(vt.Name)
	// 跨文件引用:类的类型名也需要 import
	if t, owner := c.index.Find(vt.Name); t != nil && owner != nil && owner.Path != c.File.Path {
		c.importFrom(c.relImport(owner, ".data"), name)
	}
	return name
}

// keyType map key 的 TS 类型(规范:K 仅 int64 或 string,其余按 number)。
func (c *Ctx) keyType(name string) string {
	switch name {
	case "int64", "uint64":
		return "bigint"
	case "string":
		return "string"
	}
	return "number"
}

// tsDefault 类型的默认值表达式(规范第 5.2 节)。
func (c *Ctx) tsDefault(vt *wire.VarType) string {
	if vt == nil {
		return "null"
	}
	// 可空字段默认 null;数组/映射外层可空不影响
	if vt.Empty && vt.Kind != "array" && vt.Kind != "map" {
		return "null"
	}
	switch vt.Kind {
	case "base":
		if v, ok := baseDefaults[vt.Name]; ok {
			return v
		}
		return "null"
	case "ref":
		if vt.RefKind == "enum" {
			return "0"
		}
		// 陷阱:非可空 data 引用默认零值实例(编码为 3 字节空结构)
		return "new " + tpl.HumpName(vt.Name) + "()"
	case "array":
		return "[]"
	case "map":
		return "new Map()"
	}
	return "null"
}

// tsDescriptor 渲染字段的 descriptor 构造表达式。
func (c *Ctx) tsDescriptor(vt *wire.VarType) string {
	isPtr := "false"
	if vt != nil && vt.Empty {
		isPtr = "true"
	}
	if vt == nil {
		c.Import("NewRefDescriptor")
		return "NewRefDescriptor(anyDesc, false)"
	}
	switch vt.Kind {
	case "base":
		c.Import("New" + descNames[vt.Name] + "Descriptor")
		return fmt.Sprintf("New%sDescriptor(%s)", descNames[vt.Name], isPtr)
	case "ref":
		if vt.RefKind == "enum" {
			// 与 Go 端一致:enum 按 int32 编码
			c.Import("NewInt32Descriptor")
			return fmt.Sprintf("NewInt32Descriptor(%s)", isPtr)
		}
		c.Import("NewRefDescriptor")
		return fmt.Sprintf("NewRefDescriptor(%s, %s)", c.refDesc(vt), isPtr)
	case "array":
		c.Import("NewListDescriptor")
		return fmt.Sprintf("NewListDescriptor(%s, false)", c.tsDescriptor(vt.Element))
	case "map":
		c.Import("NewMapDescriptor")
		return fmt.Sprintf("NewMapDescriptor(%s, %s, false)", c.keyDescriptor(vt.Key), c.tsDescriptor(vt.Value))
	}
	c.Import("NewRefDescriptor")
	return "NewRefDescriptor(anyDesc, false)"
}

// keyDescriptor map key 的 descriptor 表达式。
func (c *Ctx) keyDescriptor(name string) string {
	if name == "string" {
		c.Import("NewStringDescriptor")
		return "NewStringDescriptor(false)"
	}
	if d, ok := descNames[name]; ok {
		c.Import("New" + d + "Descriptor")
		return fmt.Sprintf("New%sDescriptor(false)", d)
	}
	// enum key 与 Go 端一致按 int32 编码
	c.Import("NewInt32Descriptor")
	return "NewInt32Descriptor(false)"
}

// refDesc 引用 data 的 descriptor 单例名(跨文件时登记 import)。
func (c *Ctx) refDesc(vt *wire.VarType) string {
	desc := lowerCamel(vt.Name) + "Desc"
	if t, owner := c.index.Find(vt.Name); t != nil && owner != nil && t.Kind == "data" && owner.Path != c.File.Path {
		c.importFrom(c.relImport(owner, ".data"), desc)
	}
	return desc
}

// lowerCamel 驼峰转小驼峰:HBufSub1Test -> hBufSub1Test(仅首字母小写,
// 与进程内 typescript 后端 StringToFirstLower 一致,保证方法/字段 API 兼容)
func lowerCamel(val string) string {
	if val == "" {
		return val
	}
	val = tpl.HumpName(val)
	if len(val) == 1 {
		return strings.ToLower(val)
	}
	return strings.ToLower(val[:1]) + val[1:]
}

// packOf 文件在目标语言下的包名。
func packOf(f *wire.File, lang string) string {
	if f == nil {
		return ""
	}
	return f.Packages[lang]
}

// baseName 从 .hbuf 文件路径取不带扩展名的基名。
func baseName(path string) string {
	name := filepath.Base(path)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// enumImport 登记 enum 类名的跨文件 import(生成文件后缀 .enum)。
func (c *Ctx) enumImport(name string) {
	t, owner := c.index.Find(name)
	if t == nil || owner == nil || t.Kind != "enum" || owner.Path == c.File.Path {
		return
	}
	c.importFrom(c.relImport(owner, ".enum"), tpl.HumpName(name))
}

// dataImport 登记 data 类名的跨文件 import(生成文件后缀 .data)。
func (c *Ctx) dataImport(name string) {
	t, owner := c.index.Find(name)
	if t == nil || owner == nil || t.Kind != "data" || owner.Path == c.File.Path {
		return
	}
	c.importFrom(c.relImport(owner, ".data"), tpl.HumpName(name))
}

// serverImport 登记 server 接口名的跨文件 import(生成文件后缀 .server)。
func (c *Ctx) serverImport(name string) {
	t, owner := c.index.Find(name)
	if t == nil || owner == nil || t.Kind != "server" || owner.Path == c.File.Path {
		return
	}
	c.importFrom(c.relImport(owner, ".server"), tpl.HumpName(name))
}

// serverDataImport 登记 server 方法引用的 data 类 import(生成文件后缀 .data)。
// data/server 是独立输出文件:即使 data 与 server 声明在同一 .hbuf 文件,也必须 import。
func (c *Ctx) serverDataImport(name string) {
	t, owner := c.index.Find(name)
	if t == nil || owner == nil || t.Kind != "data" {
		return
	}
	c.importFrom(c.relImport(owner, ".data"), tpl.HumpName(name))
}

// jsonFrom 渲染字段的 JSON 读取表达式(v 指代 JSON 值)。
// 与旧版 typescript 后端 fromMap 语义一致,适配新版类型系统:
// enum 字段按 int32 值(number)解析;data 引用递归 fromMap;数组/映射用工具函数。
func (c *Ctx) jsonFrom(vt *wire.VarType, v string) string {
	if vt == nil {
		return v
	}
	switch vt.Kind {
	case "base":
		return c.jsonFromBase(vt.Name, v, vt.Empty)
	case "ref":
		if vt.RefKind == "enum" {
			// 与类型声明一致:enum 字段按 int32 值(number)存储
			return c.jsonFromBase("int32", v, vt.Empty)
		}
		name := tpl.HumpName(vt.Name)
		c.dataImport(vt.Name)
		if vt.Empty {
			return fmt.Sprintf("null == %s ? null : %s.fromMap(%s as Record<string, any>)", v, name, v)
		}
		return fmt.Sprintf("null == %s ? %s.fromMap({}) : %s.fromMap(%s as Record<string, any>)", v, name, name, v)
	case "array":
		c.Import("convertArray")
		c.Import("isArray")
		inner := c.jsonFrom(vt.Element, "item")
		return fmt.Sprintf("null == %s ? [] : (!isArray(%s) ? [] : convertArray(%s, (item) => %s)!)", v, v, v, inner)
	case "map":
		c.Import("isRecord")
		key := c.mapKeyFrom(vt.Key, "_k")
		val := c.jsonFrom(vt.Value, "_v[_k]")
		return fmt.Sprintf("(() => { const ret = new Map<%s, %s>(); const _v = %s; if (isRecord(_v)) { for (const _k of Object.keys(_v)) { ret.set(%s, %s) } } return ret })()",
			c.keyType(vt.Key), c.tsType(vt.Value), v, key, val)
	}
	return v
}

// jsonFromBase 基础类型的 JSON 读取表达式。
func (c *Ctx) jsonFromBase(name, v string, empty bool) string {
	switch name {
	case "int8", "int16", "int32", "uint8", "uint16", "uint32", "float", "double":
		if empty {
			return fmt.Sprintf("null == %s ? null : Number(%s).valueOf()", v, v)
		}
		return fmt.Sprintf("null == %s ? 0 : (Number(%s).valueOf() ?? 0)", v, v)
	case "int64", "uint64":
		if empty {
			return fmt.Sprintf("null == %s ? null : BigInt(%s as string)", v, v)
		}
		return fmt.Sprintf("null == %s ? 0n : BigInt(%s as string)", v, v)
	case "string":
		if empty {
			return fmt.Sprintf("null == %s ? null : %s.toString()", v, v)
		}
		return fmt.Sprintf("null == %s ? \"\" : %s.toString()", v, v)
	case "bool":
		if empty {
			return fmt.Sprintf("null == %s ? null : (\"true\" === %s ? true : Boolean(%s))", v, v, v)
		}
		return fmt.Sprintf("null == %s ? false : (\"true\" === %s ? true : Boolean(%s))", v, v, v)
	case "date":
		if empty {
			return fmt.Sprintf("null == %s ? null : new Date(Number(%s).valueOf())", v, v)
		}
		return fmt.Sprintf("null == %s ? new Date(0) : new Date(Number(%s).valueOf())", v, v)
	case "decimal":
		if empty {
			return fmt.Sprintf("null == %s ? null : %s as string", v, v)
		}
		return fmt.Sprintf("null == %s ? \"\" : %s as string", v, v)
	case "bytes":
		if empty {
			return fmt.Sprintf("null == %s ? null : new Uint8Array([...atob(%s)].map(c => c.charCodeAt(0)))", v, v)
		}
		return fmt.Sprintf("null == %s ? new Uint8Array(0) : new Uint8Array([...atob(%s)].map(c => c.charCodeAt(0)))", v, v)
	}
	return v
}

// jsonTo 渲染字段的 JSON 写出表达式(v 指代字段值)。
func (c *Ctx) jsonTo(vt *wire.VarType, v string) string {
	if vt == nil {
		return v
	}
	switch vt.Kind {
	case "base":
		switch vt.Name {
		case "int64", "uint64":
			if vt.Empty {
				return fmt.Sprintf("%s?.toString()", v)
			}
			return fmt.Sprintf("%s.toString()", v)
		case "date":
			// 陷阱:date 的 TS 类型恒为 Date | null(零时间用 null 表示),统一可选链
			return fmt.Sprintf("%s?.getTime()", v)
		case "bytes":
			if vt.Empty {
				return fmt.Sprintf("%s == null ? null : btoa(String.fromCharCode(...%s))", v, v)
			}
			return fmt.Sprintf("btoa(String.fromCharCode(...%s))", v)
		}
		return v
	case "ref":
		if vt.RefKind == "enum" {
			// enum 字段按 int32 值(number)存储,JSON 原样输出
			return v
		}
		if vt.Empty {
			return fmt.Sprintf("%s?.toMap()", v)
		}
		return fmt.Sprintf("%s.toMap()", v)
	case "array":
		c.Import("convertArray")
		return fmt.Sprintf("convertArray(%s, (e) => %s)", v, c.jsonTo(vt.Element, "e"))
	case "map":
		key := "k"
		if vt.Key != "string" {
			key = "String(k)"
		}
		return fmt.Sprintf("(() => { const ret: Record<string, any> = {}; for (const [k, e] of %s) { ret[%s] = %s } return ret })()",
			v, key, c.jsonTo(vt.Value, "e"))
	}
	return v
}

// mapKeyFrom map key 的 JSON 键转换表达式(JSON 键是字符串)。
func (c *Ctx) mapKeyFrom(name, k string) string {
	switch name {
	case "string":
		return k
	case "int64", "uint64":
		return "BigInt(" + k + ")"
	}
	// 数字类基础类型与 enum:enum key 与 Go 端一致按 int32 值解析
	return "Number(" + k + ")"
}

// tsCopy 渲染字段的拷贝表达式(v 指代源值,与旧版 typescript 后端 printCopy 语义一致)。
// 适配新版类型系统:enum 是 number、decimal 是 string、int64/uint64 是 bigint,均不可变直接赋值;
// date 与 bytes 重新构造;data 引用递归 clone;数组/映射递归拷贝。
func (c *Ctx) tsCopy(vt *wire.VarType, v string) string {
	if vt == nil {
		return v
	}
	switch vt.Kind {
	case "base":
		switch vt.Name {
		case "date":
			// date 的 TS 类型恒为 Date | null
			return fmt.Sprintf("null == %s ? null : new Date(%s.getTime())", v, v)
		case "bytes":
			// 用 spread 构造避免 TS 泛型窄化(Uint8Array<ArrayBuffer> 不可赋给 Uint8Array<ArrayBufferLike>)
			if vt.Empty {
				return fmt.Sprintf("null == %s ? null : new Uint8Array([...%s])", v, v)
			}
			return fmt.Sprintf("new Uint8Array([...%s])", v)
		}
		// 其余基础类型不可变,直接赋值
		return v
	case "ref":
		if vt.RefKind == "enum" {
			return v
		}
		if vt.Empty {
			return fmt.Sprintf("null == %s ? null : %s.clone()", v, v)
		}
		return fmt.Sprintf("%s.clone()", v)
	case "array":
		c.Import("convertArray")
		return fmt.Sprintf("null == %s ? [] : convertArray(%s, (e) => %s)!", v, v, c.tsCopy(vt.Element, "e"))
	case "map":
		// key 为不可变类型直接复用;值递归拷贝
		return fmt.Sprintf("(() => { const ret = new Map<%s, %s>(); for (const [k, e] of %s) { ret.set(k, %s) } return ret })()",
			c.keyType(vt.Key), c.tsType(vt.Value), v, c.tsCopy(vt.Value, "e"))
	}
	return v
}

// tsCmp 渲染 $getChangeField 的变更比较表达式(self/val 为两个实例的同名字段)。
// 仅 date 需按时间戳比较,其余类型(含 enum number/decimal string/bigint)直接 !=。
func (c *Ctx) tsCmp(vt *wire.VarType, self, val string) string {
	if vt != nil && vt.Kind == "base" && vt.Name == "date" {
		return fmt.Sprintf("%s?.getTime() != %s?.getTime()", self, val)
	}
	return fmt.Sprintf("%s != %s", self, val)
}

// uiEnum 解析 enum 项的 ui 注解:type 与 class(与进程内 typescript 后端一致)。
func uiEnum(tags []*wire.Tag, itemName string) (typ, css string, ok bool) {
	typ = "info"
	for _, tag := range tags {
		if tag == nil || tag.Name != "ui" {
			continue
		}
		ok = true
		for _, kv := range tag.KV {
			if kv == nil {
				continue
			}
			switch kv.Name {
			case "type":
				if len(kv.Values) > 0 {
					typ = kv.Values[0]
				}
			case "class":
				css = strings.Join(kv.Values, " ") + " " + itemName
			}
		}
	}
	if ok && css == "" {
		css = itemName
	}
	return
}
