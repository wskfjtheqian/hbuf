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

const (
	Int8    string = "int8"
	Int16   string = "int16"
	Int32   string = "int32"
	Int64   string = "int64"
	Uint8   string = "uint8"
	Uint16  string = "uint16"
	Uint32  string = "uint32"
	Uint64  string = "uint64"
	Bool    string = "bool"
	Float   string = "float"
	Double  string = "double"
	Decimal string = "decimal"
	String  string = "string"
	Data    string = "data"
	Server  string = "server"
	Enum    string = "enum"
	Import  string = "import"
	Package string = "package"
	Date    string = "date"
)

type void struct {
}

var _types = map[string]void{
	Int8: {}, Int16: {}, Int32: {}, Int64: {}, Uint8: {}, Uint16: {}, Uint32: {}, Uint64: {}, Bool: {}, Float: {}, Double: {}, String: {}, Date: {}, Decimal: {},
}

var _keys = map[string]void{
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
	if _, ok := _keys[name]; ok {
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
		if _, ok := _types[varType.TypeExpr.(*ast.Ident).Name]; ok {
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

func EnumField(typ *ast.DataType, call func(field *ast.Field, data *ast.DataType) error) error {
	fields := map[string]int{}
	for _, field := range typ.Fields.List {
		err := call(field, typ)
		if err != nil {
			return err
		}
		fields[field.Name.Name] = 0
	}

	for _, extend := range typ.Extends {
		types := extend.Obj.Decl.(*ast.TypeSpec)
		data := types.Type.(*ast.DataType)
		for _, field := range data.Fields.List {
			if _, ok := fields[field.Name.Name]; ok {
				continue
			}
			err := call(field, data)
			if err != nil {
				return err
			}
			fields[field.Name.Name] = 0
		}
	}
	return nil
}

func CheckSuperField(name string, typ *ast.DataType) bool {
	for _, extend := range typ.Extends {
		types := extend.Obj.Decl.(*ast.TypeSpec)
		data := types.Type.(*ast.DataType)
		for _, field := range data.Fields.List {
			if name == field.Name.Name {
				return true
			}
		}
	}
	return false
}

func EnumMethod(typ *ast.ServerType, call func(method *ast.FuncType, server *ast.ServerType) error) error {
	fields := map[string]int{}
	for _, field := range typ.Methods {
		err := call(field, typ)
		if err != nil {
			return err
		}
		fields[field.Name.Name] = 0
	}

	for _, extend := range typ.Extends {
		types := extend.Obj.Decl.(*ast.TypeSpec)
		server := types.Type.(*ast.ServerType)
		for _, field := range server.Methods {
			if _, ok := fields[field.Name.Name]; ok {
				continue
			}
			err := call(field, server)
			if err != nil {
				return err
			}
			fields[field.Name.Name] = 0
		}
	}
	return nil
}

func CheckSuperMethod(name string, typ *ast.ServerType) bool {
	for _, extend := range typ.Extends {
		types := extend.Obj.Decl.(*ast.TypeSpec)
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

func IsNil(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.VarType:
		t := expr.(*ast.VarType)
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
