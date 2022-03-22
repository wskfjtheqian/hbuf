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
	err = checkCode(fset, pkg)
	if err != nil {
		return err
	}
	return nil
}

func checkCode(fset *token.FileSet, pkg *ast.Package) error {
	for _, file := range pkg.Files {
		err := checkFile(fset, pkg, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkFile(fset *token.FileSet, pkg *ast.Package, file *ast.File) error {
	for _, s := range file.Specs {
		switch s.(type) {
		case *ast.TypeSpec:
			err := checkEnum(fset, pkg, (s.(*ast.TypeSpec)).Type)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func checkEnum(fset *token.FileSet, pkg *ast.Package, expr ast.Expr) error {
	switch expr.(type) {
	case *ast.EnumType:
		enum := expr.(*ast.EnumType)
		name := enum.Name.Name
		if obj := pkg.Scope.Lookup(name); nil != obj {
			return scanner.Error{
				Pos: fset.Position(enum.Name.Pos()),
				Msg: "Duplicate type: " + name,
			}
		}

		obj := ast.NewObj(ast.Enum, name)
		obj.Decl = enum
		pkg.Scope.Insert(obj)
	}
	return nil
}
