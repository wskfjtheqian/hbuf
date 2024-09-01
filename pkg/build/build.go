package build

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/parser"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
	"path/filepath"
	"regexp"
	"strings"
)

type BaseType string

const (
	Int8    BaseType = "int8"
	Int16   BaseType = "int16"
	Int32   BaseType = "int32"
	Int64   BaseType = "int64"
	Uint8   BaseType = "uint8"
	Uint16  BaseType = "uint16"
	Uint32  BaseType = "uint32"
	Uint64  BaseType = "uint64"
	Bool    BaseType = "bool"
	Float   BaseType = "float"
	Double  BaseType = "double"
	Decimal BaseType = "decimal"
	String  BaseType = "string"
	Date    BaseType = "date"
	Enum    BaseType = "enum"
	Data    BaseType = "data"
	Server  BaseType = "server"
	Import  BaseType = "import"
	Package BaseType = "package"
)

type void struct {
}

var _types = map[BaseType]struct{}{
	Int8:    struct{}{},
	Int16:   struct{}{},
	Int32:   struct{}{},
	Int64:   struct{}{},
	Uint8:   struct{}{},
	Uint16:  struct{}{},
	Uint32:  struct{}{},
	Uint64:  struct{}{},
	Bool:    struct{}{},
	Float:   struct{}{},
	Double:  struct{}{},
	String:  struct{}{},
	Date:    struct{}{},
	Decimal: struct{}{},
}

var _keys = map[BaseType]void{
	Int8: {}, Int16: {}, Int32: {}, Int64: {}, Uint8: {}, Uint16: {}, Uint32: {}, Uint64: {}, Bool: {}, Float: {}, Double: {}, String: {}, Data: {}, Server: {}, Enum: {}, Import: {}, Package: {}, Date: {}, Decimal: {},
}

type Function = func(file *ast.File, fset *token.FileSet, param *Param) error

var buildInits = map[string]Function{}

func AddBuildType(name string, build Function) {
	buildInits[name] = build
}
func CheckType(typ string) bool {
	_, ok := buildInits[typ]
	return ok
}

type Param struct {
	out   string
	pack  string
	build *Builder
	pkg   *ast.Package
}

func (p *Param) GetOut() string {
	return p.out
}

func (p *Param) GetPack() string {
	return p.pack
}

func (p *Param) GetBuilder() *Builder {
	return p.build
}

func (p *Param) GetPkg() *ast.Package {
	return p.pkg
}

type Builder struct {
	fset  *token.FileSet
	pkg   *ast.Package
	build Function
	param *Param
}

func NewBuilder(build Function, param *Param) *Builder {
	return &Builder{
		fset:  token.NewFileSet(),
		pkg:   ast.NewPackage(),
		build: build,
		param: param,
	}
}

func Build(out string, in string, typ string, pack string) error {
	in = filepath.Clean(in)
	path := filepath.Dir(in)
	name := in[len(path)+1:]
	reg, err := regexp.Compile(strings.ReplaceAll(name, "*", "(.*)"))
	if err != nil {
		return err
	}

	build := NewBuilder(buildInits[typ], &Param{
		out:  out,
		pack: pack,
	})
	err = parser.ParseDir(build.fset, build.pkg, path, reg)
	if err != nil {
		return err
	}
	err = build.checkFiles()
	if err != nil {
		return err
	}

	for path, file := range build.pkg.Files {
		_, name := filepath.Split(path)
		err := build.build(file, build.fset, &Param{
			out:   filepath.Join(build.param.out, name),
			pkg:   build.pkg,
			pack:  build.param.pack,
			build: build,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) checkFiles() error {
	for path, file := range b.pkg.Files {
		imports := map[string]void{
			path: {},
		}
		err := b.checkFile(file, imports)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) checkFile(file *ast.File, imports map[string]void) error {
	for index, s := range file.Specs {
		switch s.(type) {
		case *ast.TypeSpec:
			err := b.checkType(file, (s.(*ast.TypeSpec)).Type, index)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Builder) checkType(file *ast.File, expr ast.Expr, index int) error {
	switch expr.(type) {
	case *ast.EnumType:
		err := b.checkEnum(file, expr.(*ast.EnumType), index)
		if err != nil {
			return err
		}
	case *ast.DataType:
		err := b.checkData(file, expr.(*ast.DataType), index)
		if err != nil {
			return err
		}
	case *ast.ServerType:
		err := b.checkServer(file, expr.(*ast.ServerType), index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) checkDuplicateType(file *ast.File, index int, name string) bool {
	for i := index + 1; i < len(file.Specs); i++ {
		s := file.Specs[i]
		switch s.(type) {
		case *ast.TypeSpec:
			t := (s.(*ast.TypeSpec)).Type
			switch t.(type) {
			case *ast.EnumType:
				n := (t.(*ast.EnumType)).Name.Name
				if n == name {
					return true
				}
			case *ast.DataType:
				n := (t.(*ast.DataType)).Name.Name
				if n == name {
					return true
				}
			case *ast.ServerType:
				n := (t.(*ast.ServerType)).Name.Name
				if n == name {
					return true
				}
			}
		}
	}

	for _, spec := range file.Imports {
		if f, ok := b.pkg.Files[spec.Path.Value]; ok {
			if obj := f.Scope.Lookup(name); nil != obj {
				return true
			}
		}
	}
	return false
}

func (b *Builder) registerServer(file *ast.File, enum *ast.ServerType) error {
	name := enum.Name.Name
	if _, ok := _keys[BaseType(name)]; ok {
		return scanner.Error{
			Pos: b.fset.Position(enum.Name.Pos()),
			Msg: "Invalid name: " + name,
		}
	}
	if obj := file.Scope.Lookup(name); nil != obj {
		return scanner.Error{
			Pos: b.fset.Position(enum.Name.Pos()),
			Msg: "Duplicate type: " + name,
		}
	}

	obj := ast.NewObj(ast.Server, name)
	obj.Decl = enum
	file.Scope.Insert(obj)
	return nil
}

func (b *Builder) checkDataMapKey(file *ast.File, varType *ast.VarType) error {
	if varType.Empty {
		return scanner.Error{
			Pos: b.fset.Position(varType.TypeExpr.End()),
			Msg: "Type cannot be empty",
		}
	}
	switch varType.TypeExpr.(type) {
	case *ast.Ident:
		if _, ok := _types[BaseType(varType.TypeExpr.(*ast.Ident).Name)]; ok {
			return nil
		}
		obj := b.GetDataType(file, varType.TypeExpr.(*ast.Ident).Name)
		if obj.Kind == ast.Enum {
			varType.TypeExpr.(*ast.Ident).Obj = obj
			return nil
		}
	}
	return scanner.Error{
		Pos: b.fset.Position(varType.TypeExpr.Pos()),
		Msg: "Map keys can only be of type",
	}
}

func (b *Builder) checkKeyValue(kvs []*ast.KeyValue) error {
	if nil == kvs {
		return nil
	}

	for i := 0; i < len(kvs)-1; i++ {
		for j := i + 1; j < len(kvs); j++ {
			if kvs[i].Name.Name == kvs[j].Name.Name {
				return scanner.Error{
					Pos: b.fset.Position(kvs[j].Pos()),
					Msg: "Repeated tag key",
				}
			}
		}
	}
	return nil
}
func (b *Builder) checkTags(tags []*ast.Tag) error {
	if nil == tags {
		return nil
	}

	for i := 0; i < len(tags)-1; i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[i].Name.Name == tags[j].Name.Name {
				return scanner.Error{
					Pos: b.fset.Position(tags[j].Pos()),
					Msg: "Repeated tag key",
				}
			}
		}
	}

	for _, tag := range tags {
		return b.checkKeyValue(tag.KV)
	}
	return nil
}

func GetTag(tags []*ast.Tag, key string) (*ast.Tag, bool) {
	if nil == tags {
		return nil, false
	}
	for _, tag := range tags {
		if tag.Name.Name == key {
			return tag, true
		}
	}
	return nil, false
}

func GetKeyValue(kvs []*ast.KeyValue, key string) (*ast.KeyValue, bool) {
	if nil == kvs {
		return nil, false
	}
	for _, kv := range kvs {
		if kv.Name.Name == key {
			return kv, true
		}
	}
	return nil, false
}

func enumField(typ *ast.DataType, fields map[string]struct{}, call func(field *ast.Field, data *ast.DataType) error) error {
	for _, extend := range typ.Extends {
		types := extend.Name.Obj.Decl.(*ast.TypeSpec)
		data := types.Type.(*ast.DataType)
		err := enumField(data, fields, call)
		if err != nil {
			return err
		}
	}

	for _, field := range typ.Fields.List {
		if _, ok := fields[field.Name.Name]; ok {
			continue
		}
		err := call(field, typ)
		if err != nil {
			return err
		}
		fields[field.Name.Name] = struct{}{}
	}
	return nil
}
func EnumField(typ *ast.DataType, call func(field *ast.Field, data *ast.DataType) error) error {
	fields := map[string]struct{}{}
	return enumField(typ, fields, call)
}

func CheckSuperField(name string, typ *ast.DataType) bool {
	for _, extend := range typ.Extends {
		types := extend.Name.Obj.Decl.(*ast.TypeSpec)
		data := types.Type.(*ast.DataType)
		for _, field := range data.Fields.List {
			if name == field.Name.Name {
				return true
			}
		}
	}
	return false
}

func enumMethod(typ *ast.ServerType, fields map[string]struct{}, call func(method *ast.FuncType, server *ast.ServerType) error) error {
	for _, field := range typ.Methods {
		if _, ok := fields[field.Name.Name]; ok {
			continue
		}
		err := call(field, typ)
		if err != nil {
			return err
		}
		fields[field.Name.Name] = struct{}{}
	}

	for _, extend := range typ.Extends {
		types := extend.Name.Obj.Decl.(*ast.TypeSpec)
		server := types.Type.(*ast.ServerType)
		err := enumMethod(server, fields, call)
		if err != nil {
			return err
		}
	}
	return nil
}

func EnumMethod(typ *ast.ServerType, call func(method *ast.FuncType, server *ast.ServerType) error) error {
	fields := map[string]struct{}{}
	return enumMethod(typ, fields, call)
}

func CheckSuperMethod(name string, typ *ast.ServerType) bool {
	for _, extend := range typ.Extends {
		types := extend.Name.Obj.Decl.(*ast.TypeSpec)
		data := types.Type.(*ast.ServerType)
		for _, field := range data.Methods {
			if name == field.Name.Name {
				return true
			}
		}
	}
	return false
}

// StringToHumpName 转驼峰名
func StringToHumpName(val string) string {
	if 0 == len(val) {
		return val
	}
	temp := strings.Split(val, "_")
	var ret string
	for _, item := range temp {
		ret += strings.ToUpper(item[:1])
		ret += item[1:]
	}
	return ret
}

// StringToFirstLower 转首字母小写
func StringToFirstLower(val string) string {
	if 0 == len(val) {
		return val
	}
	temp := strings.Split(val, "_")
	var ret string
	for _, item := range temp {
		ret += strings.ToUpper(item[:1])
		ret += item[1:]
	}
	return strings.ToLower(ret[:1]) + ret[1:]
}

// StringToUnderlineName 下划线
func StringToUnderlineName(val string) string {
	if 0 == len(val) {
		return val
	}

	rex := regexp.MustCompile(`[A-Z_]`)
	match := rex.FindAllStringSubmatchIndex(val, -1)
	if nil == match {
		return strings.ToLower(val)
	}
	var ret string
	var index = 0
	for _, item := range match {
		temp := strings.ToLower(val[index:item[0]])
		if 0 == strings.Index(temp, "_") {
			temp = temp[1:]
		}
		if 0 == len(temp) {
			continue
		}

		if 0 == len(ret) {
			ret += temp
		} else {
			ret += "_" + temp
		}
		index = item[0]
	}
	if index < len(val) {
		temp := strings.ToLower(val[index:])
		if 0 == strings.Index(temp, "_") {
			temp = temp[1:]
		}
		if 0 == len(ret) {
			ret += temp
		} else {
			ret += "_" + temp
		}
	}
	return ret
}

// StringToAllUpper 下划线加全部大写
func StringToAllUpper(val string) string {
	if 0 == len(val) {
		return val
	}

	rex := regexp.MustCompile(`[A-Z_]`)
	match := rex.FindAllStringSubmatchIndex(val, -1)
	if nil == match {
		return strings.ToUpper(val)
	}
	var ret string
	var index = 0
	for _, item := range match {
		temp := strings.ToUpper(val[index:item[0]])
		if 0 == strings.Index(temp, "_") {
			temp = temp[1:]
		}
		if 0 == len(temp) {
			continue
		}

		if 0 == len(ret) {
			ret += temp
		} else {
			ret += "_" + temp
		}
		index = item[0]
	}
	if index < len(val) {
		temp := strings.ToUpper(val[index:])
		if 0 == strings.Index(temp, "_") {
			temp = temp[1:]
		}
		if 0 == len(ret) {
			ret += temp
		} else {
			ret += "_" + temp
		}
	}
	return ret
}

func IsNumber(expr ast.Expr) bool {
	if t, ok := expr.(*ast.VarType); ok {
		if t, ok := t.TypeExpr.(*ast.Ident); ok {
			switch BaseType(t.Name) {
			case Int8, Uint8, Int16, Uint16, Int32, Uint32, Float, Double:
				return true
			}
		}
	}
	return false
}

func IsNil(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.VarType:
		t := expr.(*ast.VarType)
		if t.Empty {
			return true
		}
	case *ast.ArrayType:
		t := expr.(*ast.ArrayType)
		if t.Empty {
			return true
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		if t.Empty {
			return true
		}
	}
	return false
}

func IsArray(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.ArrayType:
		return true
	}
	return false
}

func IsMap(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.MapType:
		return true
	}
	return false
}

func IsEnum(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.EnumType:
		return true
	case *ast.Ident:
		i := expr.(*ast.Ident)
		if nil != i.Obj && i.Obj.Kind == ast.Enum {
			return true
		}
		return false
	case *ast.VarType:
		return IsEnum(expr.(*ast.VarType).Type())
	}
	return false
}

func GetBaseType(expr ast.Expr) BaseType {
	switch expr.(type) {
	case *ast.Ident:
		return BaseType(expr.(*ast.Ident).Name)
	case *ast.VarType:
		return GetBaseType(expr.(*ast.VarType).Type())
	}

	return ""
}

func GetFieldsByList[F any, E any](list []E, call func(item E) F) []F {
	field := make([]F, len(list))
	if nil == list {
		return field
	}
	for i, item := range list {
		field[i] = call(item)
	}
	return field
}

func GetKeysByMap[F string, E any](maps map[F]E) []F {
	keys := make([]F, len(maps))
	if nil == maps {
		return keys
	}
	i := 0
	for key, _ := range maps {
		keys[i] = key
		i++
	}
	return keys
}

func StringFillRight(text string, fill byte, length int) string {
	ret := strings.Builder{}
	ret.WriteString(text)
	for i := 0; i < length-len(text); i++ {
		ret.WriteByte(fill)
	}
	return ret.String()
}
