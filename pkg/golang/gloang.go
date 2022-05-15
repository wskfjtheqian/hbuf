package golang

import (
	"go/printer"
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
	"io"
	"os"
	"strings"
)

var _types = map[string]string{
	build.Int8: "int8", build.Int16: "int16", build.Int32: "int32", build.Int64: "int64", build.Uint8: "uint8",
	build.Uint16: "uint16", build.Uint32: "uint32", build.Uint64: "uint64", build.Bool: "bool", build.Float: "float32",
	build.Double: "float64", build.String: "string",
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
	err = Node(fc, fset, file)
	if err != nil {
		return err
	}
	return nil
}

func Node(dst io.Writer, fset *token.FileSet, node interface{}) error {
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

	val, ok := file.Packages["go"]
	if !ok {
		return scanner.Error{
			Pos: fset.Position(file.Pos()),
			Msg: "Not find : package",
		}
	}

	_, _ = dst.Write([]byte("package " + val.Value.Value[1:len(val.Value.Value)-1] + "\n"))
	_, _ = dst.Write([]byte("import (\n"))
	_, _ = dst.Write([]byte("\t\"context\"\n"))
	_, _ = dst.Write([]byte("\t\"encoding/json\"\n"))
	_, _ = dst.Write([]byte("\t\"hbuf_golang/pkg/hbuf\"\n"))
	for _, s := range file.Imports {
		printImport(dst, s)
	}
	_, _ = dst.Write([]byte(")\n"))

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
	//_, file := filepath.Split(spec.Path.Value)
	//dst.Write([]byte("\t\"" + file + ".dart\";\n"))
}

func printTypeSpec(dst io.Writer, expr ast.Expr) {
	switch expr.(type) {
	case *ast.DataType:
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
		_, _ = dst.Write([]byte("[]"))
		printType(dst, ar.VType, false)
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		_, _ = dst.Write([]byte("map["))
		printType(dst, ma.Key, false)
		_, _ = dst.Write([]byte("]"))
		printType(dst, ma.VType, false)
	case *ast.VarType:
		t := expr.(*ast.VarType)
		if t.Empty {
			_, _ = dst.Write([]byte("*"))
		}
		printType(dst, t.Type(), false)
	}
}

func toClassName(name string) string {
	var ret = ""
	for _, item := range strings.Split(name, "_") {
		ret += strings.ToUpper(item[0:1]) + item[1:]
	}
	return ret
}

func toFieldName(name string) string {
	var ret = ""
	for _, item := range strings.Split(name, "_") {
		if 0 < len(item) {
			ret += strings.ToUpper(item[0:1]) + item[1:]
		}
	}
	return ret
}

func toJsonName(name string) string {
	var ret = ""
	for i, item := range strings.Split(name, "_") {
		if 0 < len(item) {
			if 0 == i {
				ret += strings.ToLower(item[0:1]) + item[1:]
			} else {
				ret += strings.ToUpper(item[0:1]) + item[1:]
			}

		}
	}
	return ret
}

func toParamName(name string) string {
	var ret = ""
	for i, item := range strings.Split(name, "_") {
		if 0 < len(item) {
			if 0 == i {
				ret += strings.ToLower(item[0:1]) + item[1:]
			} else {
				ret += strings.ToUpper(item[0:1]) + item[1:]
			}

		}
	}
	return ret
}
