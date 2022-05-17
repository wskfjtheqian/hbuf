package dart

import (
	"go/printer"
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/token"
	"io"
	"os"
	"path/filepath"
)

var _types = map[string]string{
	build.Int8: "int", build.Int16: "int", build.Int32: "int", build.Int64: "int", build.Uint8: "int",
	build.Uint16: "int", build.Uint32: "int", build.Uint64: "int", build.Bool: "bool", build.Float: "double",
	build.Double: "double", build.String: "String",
}

func Build(file *ast.File, fset *token.FileSet, out string) error {
	fc, err := os.Create(out + ".go")
	if err != nil {
		return err
	}
	defer func(fc *os.File) {
		err := fc.Close()
		if err != nil {
			print(err)
		}
	}(fc)
	err = Node(fc, file)
	if err != nil {
		return err
	}
	return nil
}

func Node(dst io.Writer, node interface{}) error {
	var file *ast.File
	switch n := node.(type) {
	case *ast.File:
		file = n
	case *printer.CommentedNode:
		if f, ok := n.Node.(*ast.File); ok {
			file = f
			//cnode = n
		}
	}

	_, _ = dst.Write([]byte("import 'dart:typed_data';\n"))
	_, _ = dst.Write([]byte("import 'dart:convert';\n"))
	_, _ = dst.Write([]byte("import 'package:hbuf_dart/hbuf_dart.dart';\n"))

	for _, s := range file.Imports {
		printImport(dst, s)
	}
	_, _ = dst.Write([]byte("\n"))
	for _, s := range file.Specs {
		switch s.(type) {
		case *ast.ImportSpec:
		case *ast.TypeSpec:
			printTypeSpec(dst, (s.(*ast.TypeSpec)).Type)
		}
	}
	return nil
}

func printImport(dst io.Writer, spec *ast.ImportSpec) {
	_, file := filepath.Split(spec.Path.Value)
	dst.Write([]byte("import \"" + file + ".dart\";\n"))
}

func printTypeSpec(dst io.Writer, expr ast.Expr) {
	switch expr.(type) {
	case *ast.DataType:
		printData(dst, expr.(*ast.DataType))
		printDataEntity(dst, expr.(*ast.DataType))
	case *ast.ServerType:
		printServer(dst, expr.(*ast.ServerType))
		printServerImp(dst, expr.(*ast.ServerType))
		printServerRouter(dst, expr.(*ast.ServerType))
	case *ast.EnumType:
		printEnum(dst, expr.(*ast.EnumType))
	}
}

func getJsonName(field *ast.Field) string {
	return field.Name.Name
}

func printType(dst io.Writer, expr ast.Expr, b bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			_, _ = dst.Write([]byte((expr.(*ast.Ident)).Name))
		} else {
			_, _ = dst.Write([]byte(_types[(expr.(*ast.Ident)).Name]))
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		_, _ = dst.Write([]byte("List<"))
		printType(dst, ar.VType, false)
		_, _ = dst.Write([]byte(">"))
		if ar.Empty {
			_, _ = dst.Write([]byte("?"))
		}
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		_, _ = dst.Write([]byte("Map<"))
		printType(dst, ma.Key, false)
		_, _ = dst.Write([]byte(", "))
		printType(dst, ma.VType, false)
		_, _ = dst.Write([]byte(">"))
		if ma.Empty {
			_, _ = dst.Write([]byte("?"))
		}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		if !t.Empty && b {
			_, _ = dst.Write([]byte("required "))
		}
		printType(dst, t.Type(), false)
		if t.Empty {
			_, _ = dst.Write([]byte("?"))
		}
	}
}
