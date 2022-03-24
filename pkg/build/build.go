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
	Int8   string = "int8"
	Int16  string = "int16"
	Int32  string = "int32"
	Int64  string = "int64"
	Uint8  string = "uint8"
	Uint16 string = "uint16"
	Uint32 string = "uint32"
	Uint64 string = "uint64"
	Bool   string = "bool"
	Float  string = "float"
	Double string = "double"
	String string = "string"
	Data   string = "data"
	Server string = "server"
	Enum   string = "enum"
)

type void struct {
}

var _types = map[string]void{
	Int8: {}, Int16: {}, Int32: {}, Int64: {}, Uint8: {}, Uint16: {}, Uint32: {}, Uint64: {}, Bool: {}, Float: {}, Double: {}, String: {},
}

var _keys = map[string]void{
	Int8: {}, Int16: {}, Int32: {}, Int64: {}, Uint8: {}, Uint16: {}, Uint32: {}, Uint64: {}, Bool: {}, Float: {}, Double: {}, String: {}, Data: {}, Server: {}, Enum: {},
}

var buildInits = map[string]func(){
	"dart": dart.Init,
	"go":   golang.Init,
}

func CheckType(typ string) bool {
	_, ok := buildInits[typ]
	return ok
}

func Build(out string, in string, typ string) error {
	in = filepath.Clean(in)
	path := filepath.Dir(in)
	name := in[len(path)+1:]
	reg, err := regexp.Compile(strings.ReplaceAll(name, "*", "(.*)"))
	if err != nil {
		return err
	}

	fset := token.NewFileSet() // positions are relative to fset
	pkg := ast.NewPackage()    // positions are relative to fset
	err = parser.ParseDir(fset, pkg, path, reg)
	if err != nil {
		return err
	}
	pkg.Scope = ast.NewScope(nil)
	err = registerType(fset, pkg)
	if err != nil {
		return err
	}
	return nil
}

func registerType(fset *token.FileSet, pkg *ast.Package) error {
	for path, file := range pkg.Files {
		err := registerFile(fset, filepath.Dir(path), file, pkg)
		if err != nil {
			return err
		}
	}
	return nil
}

func registerFile(fset *token.FileSet, path string, file *ast.File, pkg *ast.Package) error {
	if nil != file.Imports {
		for _, i := range file.Imports {
			temp := filepath.Join(path, i.Path.Value[1:len(i.Path.Value)-1])
			err := registerFile(fset, filepath.Dir(temp), pkg.Files[temp], pkg)
			if err != nil {
				return err
			}
		}
	}
	for _, s := range file.Specs {
		switch s.(type) {
		case *ast.TypeSpec:
			err := registerEnum(fset, file, (s.(*ast.TypeSpec)).Type)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func registerEnum(fset *token.FileSet, file *ast.File, expr ast.Expr) error {
	switch expr.(type) {
	case *ast.EnumType:
		enum := expr.(*ast.EnumType)
		name := enum.Name.Name
		if _, ok := _keys[name]; ok {
			return scanner.Error{
				Pos: fset.Position(enum.Name.Pos()),
				Msg: "Invalid name: " + name,
			}
		}
		if obj := file.Scope.Lookup(name); nil != obj {
			return scanner.Error{
				Pos: fset.Position(enum.Name.Pos()),
				Msg: "Duplicate type: " + name,
			}
		}

		obj := ast.NewObj(ast.Enum, name)
		obj.Decl = enum
		file.Scope.Insert(obj)
	}
	return nil
}
