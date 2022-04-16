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
	String  string = "string"
	Data    string = "data"
	Server  string = "server"
	Enum    string = "enum"
	Import  string = "import"
	Package string = "package"
)

type void struct {
}

var _types = map[string]void{
	Int8: {}, Int16: {}, Int32: {}, Int64: {}, Uint8: {}, Uint16: {}, Uint32: {}, Uint64: {}, Bool: {}, Float: {}, Double: {}, String: {},
}

var _keys = map[string]void{
	Int8: {}, Int16: {}, Int32: {}, Int64: {}, Uint8: {}, Uint16: {}, Uint32: {}, Uint64: {}, Bool: {}, Float: {}, Double: {}, String: {}, Data: {}, Server: {}, Enum: {}, Import: {}, Package: {},
}

var buildInits = map[string]func(file *ast.File, out string) error{}

func AddBuildType(name string, build func(file *ast.File, out string) error) {
	buildInits[name] = build
}
func CheckType(typ string) bool {
	_, ok := buildInits[typ]
	return ok
}

type Builder struct {
	fset  *token.FileSet
	pkg   *ast.Package
	build func(file *ast.File, out string) error
	out   string
}

func NewBuilder(build func(file *ast.File, out string) error, out string) *Builder {
	return &Builder{
		fset:  token.NewFileSet(),
		pkg:   ast.NewPackage(),
		build: build,
		out:   out,
	}
}

func Build(out string, in string, typ string) error {
	in = filepath.Clean(in)
	path := filepath.Dir(in)
	name := in[len(path)+1:]
	reg, err := regexp.Compile(strings.ReplaceAll(name, "*", "(.*)"))
	if err != nil {
		return err
	}

	build := NewBuilder(buildInits[typ], out)
	err = parser.ParseDir(build.fset, build.pkg, path, reg)
	if err != nil {
		return err
	}
	err = build.checkFiles()
	if err != nil {
		return err
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
		_, name := filepath.Split(path)
		err = b.build(file, filepath.Join(b.out, name))
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

func EnumField(typ *ast.DataType, call func(field *ast.Field) error) error {
	fields := map[string]int{}
	for _, field := range typ.Fields.List {
		err := call(field)
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
			err := call(field)
			if err != nil {
				return err
			}
			fields[field.Name.Name] = 0
		}
	}
	return nil
}
