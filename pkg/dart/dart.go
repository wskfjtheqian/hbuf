package dart

import (
	"errors"
	"go/printer"
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var _types = map[string]string{
	build.Int8: "int", build.Int16: "int", build.Int32: "int", build.Int64: "int", build.Uint8: "int",
	build.Uint16: "int", build.Uint32: "int", build.Uint64: "int", build.Bool: "bool", build.Float: "double",
	build.Double: "double", build.String: "String", build.Date: "DateTime",
}

type Writer struct {
	imp  map[string]struct{}
	code *strings.Builder
	path string
	pack string
}

func (w *Writer) Import(text string) {
	w.imp[text] = struct{}{}
}

func (w *Writer) Code(text string) {
	_, _ = w.code.WriteString(text)
}

func NewWriter(pack string) *Writer {
	return &Writer{
		imp:  map[string]struct{}{},
		code: &strings.Builder{},
		pack: pack,
	}
}

type GoWriter struct {
	data   *Writer
	enum   *Writer
	server *Writer
	path   string
}

func (g *GoWriter) SetPath(s string) {
	g.path = s
	g.data.path = s
	g.enum.path = s
	g.server.path = s
}

func NewGoWriter(pack string) *GoWriter {
	return &GoWriter{
		data:   NewWriter(pack),
		enum:   NewWriter(pack),
		server: NewWriter(pack),
	}
}

func Build(file *ast.File, fset *token.FileSet, param *build.Param) error {
	dst := NewGoWriter(param.GetPack())
	err := Node(dst, fset, file)
	if err != nil {
		return err
	}

	if 0 == len(dst.path) {
		return errors.New("Not find package name")
	}

	dir, name := filepath.Split(param.GetOut())
	name = name[:len(name)-len(".hbuf")]

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	if 0 < dst.data.code.Len() {
		err := writerFile(dst.data, filepath.Join(dir, name+".data.dart"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.enum.code.Len() {
		err = writerFile(dst.enum, filepath.Join(dir, name+".enum.dart"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.server.code.Len() {
		err = writerFile(dst.server, filepath.Join(dir, name+".server.dart"))
		if err != nil {
			return err
		}
	}
	return nil
}

func writerFile(data *Writer, out string) error {
	fc, err := os.Create(out)
	if err != nil {
		return err
	}
	defer func(fc *os.File) {
		err := fc.Close()
		if err != nil {
			print(err)
		}
	}(fc)

	if 0 < len(data.imp) {
		imps := make([]string, len(data.imp))

		i := 0
		for key, _ := range data.imp {
			imps[i] = key
			i++

		}
		sort.Strings(imps)
		for _, val := range imps {
			_, _ = fc.WriteString("import '" + val + "';\n")
		}
	}
	_, _ = fc.WriteString("\n")
	_, _ = fc.WriteString(data.code.String())
	return nil
}

func Node(dst *GoWriter, fset *token.FileSet, node interface{}) error {
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

	dst.SetPath(file.Path)

	for _, s := range file.Specs {
		switch s.(type) {
		case *ast.ImportSpec:
		case *ast.TypeSpec:
			printTypeSpec(dst, (s.(*ast.TypeSpec)).Type)
		}
	}
	return nil
}

func printTypeSpec(dst *GoWriter, expr ast.Expr) {
	switch expr.(type) {
	case *ast.DataType:
		printDataCode(dst.data, expr.(*ast.DataType))
	case *ast.ServerType:
		printServerCode(dst.server, expr.(*ast.ServerType))

	case *ast.EnumType:
		printEnumCode(dst.enum, expr.(*ast.EnumType))
	}
}

func getJsonName(field *ast.Field) string {
	return field.Name.Name
}

func printType(dst *Writer, expr ast.Expr, b bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			getPackage(dst, expr)
			dst.Code(expr.(*ast.Ident).Name)
		} else {
			dst.Code(_types[(expr.(*ast.Ident).Name)])
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		dst.Code("List<")
		printType(dst, ar.VType, false)
		dst.Code(">")
		if ar.Empty {
			dst.Code("?")
		}
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		dst.Code("Map<")
		printType(dst, ma.Key, false)
		dst.Code(", ")
		printType(dst, ma.VType, false)
		dst.Code(">")
		if ma.Empty {
			dst.Code("?")
		}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		if !t.Empty && b {
			dst.Code("required ")
		}
		printType(dst, t.Type(), false)
		if t.Empty {
			dst.Code("?")
		}
	}
}

func getPackage(dst *Writer, expr ast.Expr) string {
	file := (expr.(*ast.Ident)).Obj.Data
	switch file.(type) {
	case *ast.File:
		break
	default:
		return ""
	}

	_, name := filepath.Split(file.(*ast.File).Path)
	name = name[:len(name)-len(".hbuf")]
	switch (expr.(*ast.Ident)).Obj.Kind {
	case ast.Data:
		name = name + ".data.dart"
	case ast.Enum:
		name = name + ".enum.dart"
	case ast.Server:
		name = name + ".server.dart"
	}

	dst.Import(name)
	return ""
}
