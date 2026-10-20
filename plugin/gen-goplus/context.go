package goplus

import (
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"path/filepath"
	"strconv"
	"strings"
)

// TypeIndex 全包类型索引:类型名 -> 声明文件与声明。
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

// Ctx 单个 .hbuf 文件的生成上下文。
type Ctx struct {
	Req   *wire.Request
	File  *wire.File
	Pack  string // 本文件 go 包名
	index *TypeIndex
}

// NewCtx 创建生成上下文。
func NewCtx(req *wire.Request, file *wire.File, index *TypeIndex, pack string) *Ctx {
	return &Ctx{Req: req, File: file, Pack: pack, index: index}
}

// Find 按名称查找类型声明。
func (c *Ctx) Find(name string) (*wire.Type, *wire.File) {
	return c.index.Find(name)
}

// refPrefix 引用类型的前缀:同包为空,否则为"<末段包名>."并登记 import。
func (c *Ctx) refPrefix(refPack string, imp Imports) string {
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
	if imp != nil {
		imp.Import(pkg, "")
	}
	return seg + "."
}

// refHump 引用类型名(驼峰)含包前缀。
func (c *Ctx) refHump(refPack, name string, imp Imports) string {
	return c.refPrefix(refPack, imp) + tpl.HumpName(name)
}

// baseTypes 基础类型 -> Go 类型映射(与进程内 golang 后端一致)
var baseTypes = map[string]string{
	"int8": "int8", "int16": "int16", "int32": "int32", "int64": "int64",
	"uint8": "uint8", "uint16": "uint16", "uint32": "uint32", "uint64": "uint64",
	"bool": "bool", "float": "float32", "double": "float64",
	"string": "string", "date": "time.Time", "decimal": "decimal.Decimal", "bytes": "[]byte",
}

// baseType 基础类型映射,登记特殊 import(time/decimal)。
func (c *Ctx) baseType(name string, imp Imports) string {
	switch name {
	case "date":
		imp.Import("time", "")
	case "decimal":
		imp.Import("github.com/shopspring/decimal", "")
	}
	if v, ok := baseTypes[name]; ok {
		return v
	}
	return name
}

// goType 把 wire 类型渲染成 Go 类型。top 为 true 时顶层可空渲染为指针
// (数组/映射自身即可表示空,不加指针,与进程内后端一致)。
func (c *Ctx) goType(vt *wire.VarType, top bool, imp Imports) string {
	if vt == nil {
		return "any"
	}
	pre := ""
	if top && vt.Empty && vt.Kind != "array" && vt.Kind != "map" {
		pre = "*"
	}
	switch vt.Kind {
	case "base":
		return pre + c.baseType(vt.Name, imp)
	case "ref":
		return pre + c.refHump(vt.RefPack, vt.Name, imp)
	case "array":
		return pre + "[]" + c.goType(vt.Element, true, imp)
	case "map":
		return pre + "map[" + c.keyType(vt.Key, imp) + "]" + c.goType(vt.Value, true, imp)
	}
	return pre + "any"
}

// keyType map key 类型:基础类型或枚举引用。
func (c *Ctx) keyType(name string, imp Imports) string {
	if _, ok := baseTypes[name]; ok {
		return c.baseType(name, imp)
	}
	if t, f := c.Find(name); t != nil && t.Kind == "enum" && f != nil {
		return c.refPrefix(f.Packages["go"], imp) + tpl.HumpName(name)
	}
	return tpl.HumpName(name)
}

// enumField 遍历 data 的继承链字段(先父后子,同名子类覆盖),与进程内后端一致。
func (c *Ctx) enumField(typ *wire.DataType, call func(field *wire.Field, data *wire.DataType) error) error {
	return c.enumFieldRec(typ, map[string]bool{}, call)
}

func (c *Ctx) enumFieldRec(typ *wire.DataType, seen map[string]bool, call func(field *wire.Field, data *wire.DataType) error) error {
	for _, ext := range typ.Extends {
		if ext == nil {
			continue
		}
		t, _ := c.index.Find(ext.Name)
		if t == nil || t.Data == nil {
			continue
		}
		if err := c.enumFieldRec(t.Data, seen, call); err != nil {
			return err
		}
	}
	for _, field := range typ.Fields {
		if field == nil || seen[field.Name] {
			continue
		}
		seen[field.Name] = true
		if err := call(field, typ); err != nil {
			return err
		}
	}
	return nil
}

// baseName 从 .hbuf 文件路径取不带扩展名的基名。
func baseName(path string) string {
	name := filepath.Base(path)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// stringFillRight 右对齐补空格。
func stringFillRight(text string, fill byte, length int) string {
	if len(text) >= length {
		return text
	}
	return text + strings.Repeat(string(fill), length-len(text))
}

// getTag 查找指定名称的注解。
func getTag(tags []*wire.Tag, name string) (*wire.Tag, bool) {
	for _, tag := range tags {
		if tag != nil && tag.Name == name {
			return tag, true
		}
	}
	return nil, false
}

// getKV 查找注解参数。
func getKV(tag *wire.Tag, name string) (*wire.KeyValue, bool) {
	if tag == nil {
		return nil, false
	}
	for _, kv := range tag.KV {
		if kv != nil && kv.Name == name {
			return kv, true
		}
	}
	return nil, false
}

// isNil 类型是否可空。
func isNil(vt *wire.VarType) bool {
	return vt != nil && vt.Empty
}

// isArray 类型是否为数组。
func isArray(vt *wire.VarType) bool {
	return vt != nil && vt.Kind == "array"
}

// isMap 类型是否为映射。
func isMap(vt *wire.VarType) bool {
	return vt != nil && vt.Kind == "map"
}

// isEnum 类型是否为枚举引用。
func isEnum(vt *wire.VarType) bool {
	return vt != nil && vt.Kind == "ref" && vt.RefKind == "enum"
}

// isDate 类型是否为基础类型 date。
func isDate(vt *wire.VarType) bool {
	return vt != nil && vt.Kind == "base" && vt.Name == "date"
}

// baseType 基础类型名(引用类型返回空串)。
func baseType(vt *wire.VarType) string {
	if vt != nil && vt.Kind == "base" {
		return vt.Name
	}
	return ""
}

// 操作符(与进程内 build.Operator 一致)
type Operator int

const (
	OperatorGt       Operator = 1 + iota // 大于 >
	OperatorLt                           // 小于 <
	OperatorEq                           // 等于 ==
	OperatorGte                          // 大于等于 >=
	OperatorLte                          // 小于等于 <=
	OperatorNeq                          // 不等于 !=
	OperatorMatch                        // 匹配 =~
	OperatorNotMatch                     // 不匹配 !~
)

// Expr 校验表达式。
type Expr struct {
	Op  Operator
	Val string
}

// Format 字段的 format 注解。
type Format struct {
	Null bool
	Val  []Expr
	Len  []Expr
}

// getFormat 解析 format 注解(与进程内 build.GetFormat 一致)。
func getFormat(tags []*wire.Tag) *Format {
	val, ok := getTag(tags, "format")
	if !ok {
		return nil
	}
	f := &Format{}
	for _, kv := range val.KV {
		if kv == nil {
			continue
		}
		switch kv.Name {
		case "null":
			f.Null = len(kv.Values) > 0 && kv.Values[0] == "true"
		case "val":
			f.Val = parseExpr(kv.Values)
		case "len":
			f.Len = parseExpr(kv.Values)
		}
	}
	return f
}

// parseExpr 解析表达式列表(值已去除引号)。
func parseExpr(values []string) []Expr {
	var val []Expr
	for _, value := range values {
		v := parseExprItem(value)
		if v.Op > 0 {
			val = append(val, v)
		}
	}
	return val
}

func parseExprItem(value string) Expr {
	var e Expr
	switch {
	case strings.HasPrefix(value, "> "):
		e.Op, e.Val = OperatorGt, value[2:]
	case strings.HasPrefix(value, "< "):
		e.Op, e.Val = OperatorLt, value[2:]
	case strings.HasPrefix(value, "== "):
		e.Op, e.Val = OperatorEq, value[3:]
	case strings.HasPrefix(value, ">= "):
		e.Op, e.Val = OperatorGte, value[3:]
	case strings.HasPrefix(value, "<= "):
		e.Op, e.Val = OperatorLte, value[3:]
	case strings.HasPrefix(value, "!= "):
		e.Op, e.Val = OperatorNeq, value[3:]
	case strings.HasPrefix(value, "=~ "):
		e.Op, e.Val = OperatorMatch, value[3:]
	case strings.HasPrefix(value, "!~ "):
		e.Op, e.Val = OperatorNotMatch, value[3:]
	}
	return e
}

// DB 类型的 db 注解(与进程内 build.DB 一致)。
type DB struct {
	Name      string
	Key       bool
	Force     bool
	typ       string
	Insert    string
	Inserts   string
	Update    string
	Set       string
	Del       bool
	Get       string
	List      string
	ListAsync string
	Map       string
	Count     string
	Table     string
	Remove    bool
	Where     []string
	Offset    string
	Limit     string
	Order     string
	Converter string
	Group     string
	Change    string
	Changes   string
	Join      []string
}

// DBField db 字段:注解 + 字段声明。
type DBField struct {
	DB    *DB
	Field *wire.Field
}

// getDB 解析 db 注解,与进程内 build.GetDB 一致(name 为类型/字段名,默认库名)。
func getDB(name string, tags []*wire.Tag) *DB {
	val, ok := getTag(tags, "db")
	if !ok {
		return nil
	}
	db := &DB{}
	for _, kv := range val.KV {
		if kv == nil || len(kv.Values) == 0 {
			continue
		}
		first := kv.Values[0]
		switch kv.Name {
		case "name":
			db.Name = first
		case "converter":
			db.Converter = first
		case "order":
			db.Order = first
		case "key":
			db.Key = first == "true"
		case "typ":
			db.typ = first
		case "where":
			db.Where = kv.Values
		case "join":
			db.Join = kv.Values
		case "offset":
			db.Offset = first
		case "limit":
			db.Limit = first
		case "insert":
			db.Insert = first
		case "inserts":
			db.Inserts = first
		case "update":
			db.Update = first
		case "change":
			db.Change = first
		case "changes":
			db.Changes = first
		case "group":
			db.Group = first
		case "del":
			db.Del = strings.ToLower(first) == "true"
		case "get":
			db.Get = first
		case "list":
			db.List = first
		case "listAsync":
			db.ListAsync = first
		case "map":
			db.Map = strings.ToLower(first)
		case "table":
			db.Table = first
		case "set":
			db.Set = first
		case "count":
			db.Count = first
		case "force":
			db.Force = strings.ToLower(first) == "true"
		case "rm":
			db.Remove = strings.ToLower(first) == "true"
		}
	}
	if db.Name == "" {
		db.Name = tpl.UnderlineName(name)
	}
	return db
}

// cache 类型的 cache 注解(与进程内一致)。
type cache struct {
	min int
	max int
}

func getCache(tags []*wire.Tag) *cache {
	val, ok := getTag(tags, "cache")
	if !ok {
		return nil
	}
	c := &cache{min: 2 * 60 * 60, max: 3 * 60 * 60}
	for _, kv := range val.KV {
		if kv == nil || len(kv.Values) == 0 {
			continue
		}
		v, err := strconv.Atoi(kv.Values[0])
		if err != nil {
			// 与进程内一致:解析失败视为无缓存
			return nil
		}
		switch kv.Name {
		case "min":
			c.min = v
		case "max":
			c.max = v
		}
	}
	return c
}

// Export 类型的 export 注解。
type Export struct {
	Filter []string
}

// getExport 解析 export 注解。
func getExport(tags []*wire.Tag) *Export {
	val, ok := getTag(tags, "export")
	if !ok {
		return nil
	}
	e := &Export{}
	if kv, ok := getKV(val, "filter"); ok {
		e.Filter = kv.Values
	}
	return e
}

// VerifyEnum verify 注解中的枚举格式引用:"EnumName.ItemName"。
type VerifyEnum struct {
	Enum string
	Item string
}

// Verify 类型的 verify 注解。
type Verify struct {
	Format []VerifyEnum
	Fields string
}

// getVerify 解析 verify 注解(格式引用延后到生成期通过类型索引解析)。
func getVerify(tags []*wire.Tag) *Verify {
	val, ok := getTag(tags, "verify")
	if !ok {
		return nil
	}
	v := &Verify{}
	for _, kv := range val.KV {
		if kv == nil {
			continue
		}
		switch kv.Name {
		case "format":
			for _, item := range kv.Values {
				parts := strings.Split(item, ".")
				if len(parts) == 2 {
					v.Format = append(v.Format, VerifyEnum{Enum: parts[0], Item: parts[1]})
				}
			}
		case "fields":
			if len(kv.Values) > 0 {
				v.Fields = kv.Values[0]
			}
		}
	}
	return v
}
