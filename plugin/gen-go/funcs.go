// Package gogen 实现 hbuf-gen-go 插件:把 wire 中性描述模型渲染成 Go 代码。
//
// 渲染采用模板引擎 plugin/core/tpl:
//   - 默认模板(templates/*.tmpl)通过 //go:embed 编译进插件二进制;
//   - 外部自定义模板目录(插件参数 template=<目录>)同名文件覆盖默认模板。
//
// 生成文件名与进程内 golang 后端保持一致:<name>.data.go / <name>.enum.go / <name>.server.go。
package gogen

import (
	"fmt"
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"path/filepath"
	"sort"
	"strings"
)

// baseTypes 基础类型 -> Go 类型映射(与进程内 golang 后端一致)
var baseTypes = map[string]string{
	"int8": "int8", "int16": "int16", "int32": "int32", "int64": "int64",
	"uint8": "uint8", "uint16": "uint16", "uint32": "uint32", "uint64": "uint64",
	"bool": "bool", "float": "float32", "double": "float64",
	"string": "string", "date": "time.Time", "decimal": "decimal.Decimal", "bytes": "[]byte",
}

// baseDefaults 基础类型默认值表达式
var baseDefaults = map[string]string{
	"int8": "int8(0)", "int16": "int16(0)", "int32": "int32(0)", "int64": "int64(0)",
	"uint8": "uint8(0)", "uint16": "uint16(0)", "uint32": "uint32(0)", "uint64": "uint64(0)",
	"bool": "false", "float": "float32(0)", "double": "float64(0)",
	"string": `""`, "date": "time.Time{}", "decimal": "decimal.Zero", "bytes": "nil",
}

// descNames 基础类型 -> hbuf descriptor 构造函数名
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

// Ctx 单个 .hbuf 文件的渲染上下文(模板根对象)。
type Ctx struct {
	Req     *wire.Request
	File    *wire.File
	Pack    string          // 本文件 go 包名
	Imports map[string]bool // 渲染过程收集的 import 路径
	index   *TypeIndex
}

// NewCtx 创建渲染上下文。
func NewCtx(req *wire.Request, file *wire.File, index *TypeIndex, pack string) *Ctx {
	return &Ctx{
		Req:     req,
		File:    file,
		Pack:    pack,
		Imports: map[string]bool{},
		index:   index,
	}
}

// Import 登记一个 import 路径(幂等)。
func (c *Ctx) Import(path string) string {
	if path != "" {
		c.Imports[path] = true
	}
	return path
}

// clone 复制上下文:共享请求与类型索引,Imports 独立(按输出文件类别收集)。
func (c *Ctx) clone() *Ctx {
	return &Ctx{
		Req:     c.Req,
		File:    c.File,
		Pack:    c.Pack,
		Imports: map[string]bool{},
		index:   c.index,
	}
}

// refPrefix 引用类型的前缀:同包为空,否则为"<末段包名>."并登记 import。
func (c *Ctx) refPrefix(refPack string) string {
	if refPack == "" || refPack == c.Pack {
		return ""
	}
	seg := refPack
	if i := strings.LastIndex(refPack, "."); i >= 0 {
		seg = refPack[i+1:]
	}
	// 与进程内后端一致:目标包引用路径 = 构建包前缀 + "/" + 末段包名
	pkg := c.Req.Pack
	if pkg == "" {
		pkg = seg
	} else {
		pkg += "/" + seg
	}
	c.Import(pkg)
	return seg + "."
}

// refHump 引用类型名(驼峰)含包前缀。
func (c *Ctx) refHump(refPack, name string) string {
	return c.refPrefix(refPack) + tpl.HumpName(name)
}

// baseType 基础类型映射,登记特殊 import(time/decimal)。
func (c *Ctx) baseType(name string) string {
	switch name {
	case "date":
		c.Import("time")
	case "decimal":
		c.Import("github.com/shopspring/decimal")
	}
	if v, ok := baseTypes[name]; ok {
		return v
	}
	return name
}

// goType 把 wire 类型渲染成 Go 类型。top 为 true 时顶层可空渲染为指针。
func (c *Ctx) goType(vt *wire.VarType, top bool) string {
	if vt == nil {
		return "any"
	}
	pre := ""
	// 数组/映射自身即可表示空(nil slice/map),可空不加指针,与进程内后端一致
	if top && vt.Empty && vt.Kind != "array" && vt.Kind != "map" {
		pre = "*"
	}
	switch vt.Kind {
	case "base":
		return pre + c.baseType(vt.Name)
	case "ref":
		return pre + c.refHump(vt.RefPack, vt.Name)
	case "array":
		return pre + "[]" + c.goType(vt.Element, true)
	case "map":
		return pre + "map[" + c.keyType(vt.Key) + "]" + c.goType(vt.Value, true)
	}
	return pre + "any"
}

// keyType map key 类型:基础类型或枚举引用。
func (c *Ctx) keyType(name string) string {
	if _, ok := baseTypes[name]; ok {
		return c.baseType(name)
	}
	if t, f := c.index.Find(name); t != nil && t.Kind == "enum" && f != nil {
		return c.refPrefix(packOf(f, "go")) + tpl.HumpName(name)
	}
	return tpl.HumpName(name)
}

// goDefault 类型的默认值表达式(用于 Getter 空值返回)。
func (c *Ctx) goDefault(vt *wire.VarType) string {
	if vt == nil {
		return "nil"
	}
	switch vt.Kind {
	case "base":
		if v, ok := baseDefaults[vt.Name]; ok {
			if vt.Name == "date" {
				c.Import("time")
			}
			if vt.Name == "decimal" {
				c.Import("github.com/shopspring/decimal")
			}
			return v
		}
		return "nil"
	case "ref":
		if vt.RefKind == "enum" {
			return c.refHump(vt.RefPack, vt.Name) + "(0)"
		}
		return c.refHump(vt.RefPack, vt.Name) + "{}"
	case "array":
		return "[]" + c.goType(vt.Element, true) + "{}"
	case "map":
		return "map[" + c.keyType(vt.Key) + "]" + c.goType(vt.Value, true) + "{}"
	}
	return "nil"
}

// descType 泛型 descriptor 的元素类型(元素可空时带指针)。
func (c *Ctx) descType(vt *wire.VarType) string {
	return c.goType(vt, true)
}

// goDescriptor 渲染字段的 descriptor 构造表达式。
// structName/fieldName 非空时使用 unsafe.Offsetof 计算偏移。
func (c *Ctx) goDescriptor(vt *wire.VarType, structName, fieldName string) string {
	offset := "0"
	if structName != "" && fieldName != "" {
		offset = fmt.Sprintf("unsafe.Offsetof(%s.%s)", structName, fieldName)
	}
	isNull := "false"
	if vt != nil && vt.Empty {
		isNull = "true"
	}
	if vt == nil {
		return fmt.Sprintf("hbuf.NewRefDescriptor(&any{}, %s, %s)", offset, isNull)
	}
	switch vt.Kind {
	case "base":
		return fmt.Sprintf("hbuf.New%sDescriptor(%s, %s)", descNames[vt.Name], offset, isNull)
	case "ref":
		if vt.RefKind == "enum" {
			return fmt.Sprintf("hbuf.NewInt32Descriptor(%s, %s)", offset, isNull)
		}
		return fmt.Sprintf("hbuf.NewRefDescriptor(&%s{}, %s, %s)", c.refHump(vt.RefPack, vt.Name), offset, isNull)
	case "array":
		return fmt.Sprintf("hbuf.NewListDescriptor[%s](%s, %s, %s)",
			c.descType(vt.Element), offset, c.goDescriptor(vt.Element, "", ""), isNull)
	case "map":
		return fmt.Sprintf("hbuf.NewMapDescriptor[%s, %s](%s, %s, %s, %s)",
			c.keyType(vt.Key), c.descType(vt.Value), offset,
			c.goDescriptor(c.keyVar(vt.Key), "", ""),
			c.goDescriptor(vt.Value, "", ""), isNull)
	}
	return fmt.Sprintf("hbuf.NewRefDescriptor(&any{}, %s, %s)", offset, isNull)
}

// keyVar 把 map key 名字构造为 wire 类型(用于 descriptor 渲染)。
func (c *Ctx) keyVar(name string) *wire.VarType {
	if _, ok := baseTypes[name]; ok {
		return &wire.VarType{Kind: "base", Name: name}
	}
	if t, f := c.index.Find(name); t != nil && t.Kind == "enum" && f != nil {
		return &wire.VarType{Kind: "ref", Name: name, RefKind: "enum", RefPack: packOf(f, "go")}
	}
	return &wire.VarType{Kind: "base", Name: "string"}
}

// fieldTagMap 解析字段的 "tag" 注解:tag 键 -> 值列表(键小写,排除 hbuf)。
func fieldTagMap(tags []*wire.Tag) map[string][]string {
	for _, t := range tags {
		if t == nil || t.Name != "tag" {
			continue
		}
		m := map[string][]string{}
		for _, kv := range t.KV {
			if kv == nil {
				continue
			}
			name := strings.ToLower(kv.Name)
			if name == "hbuf" {
				continue
			}
			m[name] = append(m[name], kv.Values...)
		}
		return m
	}
	return nil
}

// goStructTag 渲染字段的 struct tag。
func (c *Ctx) goStructTag(field *wire.Field) string {
	tags := fieldTagMap(field.Tags)
	sb := strings.Builder{}
	sb.WriteString("`")
	if len(tags) > 0 {
		if _, ok := tags["json"]; !ok {
			sb.WriteString(`json:"` + tpl.UnderlineName(field.Name) + `,omitempty"`)
		}
		keys := make([]string, 0, len(tags))
		for key := range tags {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if sb.Len() > 1 {
				sb.WriteString(" ")
			}
			sb.WriteString(key + `:"` + strings.Join(tags[key], ",") + `"`)
		}
	} else {
		sb.WriteString(`json:"` + tpl.UnderlineName(field.Name) + `,omitempty"`)
	}
	sb.WriteString(fmt.Sprintf(` hbuf:"%d"`, field.Id))
	sb.WriteString("`")
	return sb.String()
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
