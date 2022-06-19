package golang

import (
	"errors"
	"go/printer"
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var _types = map[string]string{
	build.Int8: "int8", build.Int16: "int16", build.Int32: "int32", build.Int64: "int64", build.Uint8: "uint8",
	build.Uint16: "uint16", build.Uint32: "uint32", build.Uint64: "uint64", build.Bool: "bool", build.Float: "float32",
	build.Double: "float64", build.String: "string", build.Date: "hbuf.Time",
}

type Writer struct {
	imp      map[string]struct{}
	code     *strings.Builder
	packages string
	pack     string
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
	data     *Writer
	enum     *Writer
	server   *Writer
	database *Writer
	packages string
}

func (g *GoWriter) SetPackages(s string) {
	g.packages = s
	g.data.packages = s
	g.enum.packages = s
	g.server.packages = s
	g.database.packages = s

}

func NewGoWriter(pack string) *GoWriter {
	return &GoWriter{
		data:     NewWriter(pack),
		enum:     NewWriter(pack),
		server:   NewWriter(pack),
		database: NewWriter(pack),
	}
}

func Build(file *ast.File, fset *token.FileSet, param *build.Param) error {
	dst := NewGoWriter(param.GetPack())
	err := Node(dst, fset, file)
	if err != nil {
		return err
	}

	if 0 == len(dst.packages) {
		return errors.New("Not find package name")
	}

	dir, name := filepath.Split(param.GetOut())
	name = name[:len(name)-len(".hbuf")]
	packs := strings.Split(dst.packages, ".")
	for _, pack := range packs {
		dir = filepath.Join(dir, pack)
	}

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	if 0 < dst.data.code.Len() {
		err := writerFile(dst.data, dst.packages, filepath.Join(dir, name+".data.go"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.enum.code.Len() {
		err = writerFile(dst.enum, dst.packages, filepath.Join(dir, name+".enum.go"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.server.code.Len() {
		err = writerFile(dst.server, dst.packages, filepath.Join(dir, name+".server.go"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.database.code.Len() {
		err = writerFile(dst.database, dst.packages, filepath.Join(dir, name+".database.go"))
		if err != nil {
			return err
		}
	}
	return nil
}

func writerFile(data *Writer, packages string, out string) error {
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

	_, _ = fc.WriteString("package " + packages + "\n\n")

	if 0 < len(data.imp) {
		_, _ = fc.WriteString("import (\n")
		imps := make([]string, len(data.imp))

		i := 0
		for key, _ := range data.imp {
			imps[i] = key
			i++

		}
		sort.Strings(imps)
		for _, val := range imps {
			_, _ = fc.WriteString("\t\"" + val + "\"\n")
		}
		_, _ = fc.WriteString(")\n\n")
	}
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

	val, ok := file.Packages["go"]
	if !ok {
		return scanner.Error{
			Pos: fset.Position(file.Pos()),
			Msg: "Not find : package",
		}
	}

	dst.SetPackages(val.Value.Value[1 : len(val.Value.Value)-1])

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
		printDatabaseCode(dst.database, expr.(*ast.DataType))
	case *ast.ServerType:
		printServerCode(dst.server, expr.(*ast.ServerType))

	case *ast.EnumType:
		printEnumCode(dst.enum, expr.(*ast.EnumType))
	}
}

func printType(dst *Writer, expr ast.Expr, b bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			pack := getPackage(dst, expr)
			dst.Code(pack + (expr.(*ast.Ident)).Name)
		} else {
			if build.Date == (expr.(*ast.Ident)).Name {
				dst.Import("hbuf_golang/pkg/hbuf")
			}
			dst.Code(_types[(expr.(*ast.Ident)).Name])
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		dst.Code("[]")
		printType(dst, ar.VType, false)
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		dst.Code("map[")
		printType(dst, ma.Key, false)
		dst.Code("]")
		printType(dst, ma.VType, false)
	case *ast.VarType:
		t := expr.(*ast.VarType)
		if t.Empty {
			dst.Code("*")
		}
		printType(dst, t.Type(), false)
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

	val, ok := (file.(*ast.File)).Packages["go"]
	if !ok {
		return ""
	}

	pack := val.Value.Value[1 : len(val.Value.Value)-1]
	if 0 == len(pack) {
		return ""
	}

	if pack == dst.packages {
		return ""
	}

	packs := strings.Split(pack, ".")
	pack = packs[len(packs)-1]

	dst.Import(dst.pack + pack)
	return pack + "."
}
