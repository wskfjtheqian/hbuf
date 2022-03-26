package build

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/dart"
	"hbuf/pkg/parser"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
	"path/filepath"
	"regexp"
	"strings"
)
import "hbuf/pkg/golang"

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

var buildInits = map[string]func(file *ast.File, out string) error{
	"dart": dart.Build,
	"go":   golang.Build,
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
		//case *ast.DataType:
		//	err := b.registerData(file, expr.(*ast.DataType))
		//	if err != nil {
		//		return err
		//	}
		//case *ast.ServerType:
		//	err := b.registerServer(file, expr.(*ast.ServerType))
		//	if err != nil {
		//		return err
		//	}
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

func (b *Builder) checkEnum(file *ast.File, enum *ast.EnumType, index int) error {
	name := enum.Name.Name
	if _, ok := _keys[name]; ok {
		return scanner.Error{
			Pos: b.fset.Position(enum.Name.Pos()),
			Msg: "Invalid name: " + name,
		}
	}

	if b.checkDuplicateType(file, index, name) {
		return scanner.Error{
			Pos: b.fset.Position(enum.Name.Pos()),
			Msg: "Duplicate type: " + name,
		}
	}

	err := b.checkEnumItem(file, enum)
	if err != nil {
		return err
	}

	obj := ast.NewObj(ast.Enum, name)
	obj.Decl = enum
	file.Scope.Insert(obj)
	return nil
}

func (b *Builder) checkEnumItem(file *ast.File, enum *ast.EnumType) error {
	for index, item := range enum.Items {
		if _, ok := _keys[item.Name.Name]; ok {
			return scanner.Error{
				Pos: b.fset.Position(enum.Name.Pos()),
				Msg: "Invalid name: " + item.Name.Name,
			}
		}
		if b.checkEnumDuplicateItem(enum, index, item.Name.Name) {
			return scanner.Error{
				Pos: b.fset.Position(item.Name.Pos()),
				Msg: "Duplicate item: " + item.Name.Name,
			}
		}
		if b.checkEnumDuplicateValue(enum, index, item.Id.Value) {
			return scanner.Error{
				Pos: b.fset.Position(item.Id.Pos()),
				Msg: "Duplicate item: " + item.Id.Value,
			}
		}
	}
	return nil
}

func (b *Builder) checkEnumDuplicateItem(enum *ast.EnumType, index int, name string) bool {
	for i := index + 1; i < len(enum.Items); i++ {
		s := enum.Items[i]
		if s.Name.Name == name {
			return true
		}
	}
	return false
}

func (b *Builder) checkEnumDuplicateValue(enum *ast.EnumType, index int, id string) bool {
	for i := index + 1; i < len(enum.Items); i++ {
		s := enum.Items[i]
		if s.Id.Value == id {
			return true
		}
	}
	return false
}

func (b *Builder) registerData(file *ast.File, enum *ast.DataType) error {
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

	obj := ast.NewObj(ast.Data, name)
	obj.Decl = enum
	file.Scope.Insert(obj)
	return nil
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
