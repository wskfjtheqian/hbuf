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
		Req:     req,
		File:    file,
		Pack:    pack,
		Runtime: runtime,
		// 每个生成文件恒定依赖的运行时符号(见 hbuf-ts-codegen.md 第 3 节)
		Imports: map[string]bool{
			"DataDescriptor":    true,
			"HBufDescriptor":    true,
			"FieldDesc":         true,
			"HData":             true,
			"NewDataDescriptor": true,
		},
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
// 生成文件名是 <name>.data.ts,跨文件 import 需指向 <name>.data。
func (c *Ctx) relImport(owner *wire.File) string {
	if owner == nil {
		return ""
	}
	curPackDir := filepath.Join(strings.Split(c.Pack, ".")...)
	refPackDir := filepath.Join(strings.Split(packOf(owner, "ts"), ".")...)
	target := filepath.Join(refPackDir, baseName(owner.Path)+".data")
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

// refType 引用类型的 TS 类型:enum -> number,data -> XxxTs。
func (c *Ctx) refType(vt *wire.VarType) string {
	if vt.RefKind == "enum" {
		return "number"
	}
	name := tpl.HumpName(vt.Name) + "Ts"
	// 跨文件引用:类的类型名也需要 import
	if t, owner := c.index.Find(vt.Name); t != nil && owner != nil && owner.Path != c.File.Path {
		c.importFrom(c.relImport(owner), name)
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
		return "new " + tpl.HumpName(vt.Name) + "Ts()"
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
		c.importFrom(c.relImport(owner), desc)
	}
	return desc
}

// lowerCamel 驼峰转小驼峰:HBufSub1Test -> hbufSub1Test(连续大写段视为一个词)。
func lowerCamel(val string) string {
	if val == "" {
		return val
	}
	val = tpl.HumpName(val)
	words := splitWords(val)
	if len(words) == 0 {
		return strings.ToLower(val)
	}
	words[0] = strings.ToLower(words[0])
	return strings.Join(words, "")
}

// splitWords 按大写边界切词:HBufSub1Test -> ["HBuf", "Sub1", "Test"]。
func splitWords(val string) []string {
	var words []string
	start := 0
	for i := 1; i < len(val); i++ {
		if val[i] >= 'A' && val[i] <= 'Z' && (val[i-1] < 'A' || val[i-1] > 'Z') {
			words = append(words, val[start:i])
			start = i
		}
	}
	words = append(words, val[start:])
	return words
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
