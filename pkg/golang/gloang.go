package golang

import (
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
	build.Int8: "int8", build.Int16: "int16", build.Int32: "int32", build.Int64: "hbuf.Int64", build.Uint8: "uint8",
	build.Uint16: "uint16", build.Uint32: "uint32", build.Uint64: "hbuf.Uint64", build.Bool: "bool", build.Float: "float32",
	build.Double: "float64", build.String: "string", build.Date: "hbuf.Time", build.Decimal: "decimal.Decimal",
}

type GoWriter struct {
	data     *build.Writer
	enum     *build.Writer
	server   *build.Writer
	database *build.Writer
	packages string
}

func (w *GoWriter) SetPackages(s string) {
	w.data.Packages = s
	w.enum.Packages = s
	w.server.Packages = s
	w.database.Packages = s
	w.packages = s
}

func NewGoWriter() *GoWriter {
	return &GoWriter{
		data:     build.NewWriter(),
		enum:     build.NewWriter(),
		server:   build.NewWriter(),
		database: build.NewWriter(),
	}
}

type Builder struct {
	build    *build.Builder
	packages string
}

func Build(file *ast.File, fset *token.FileSet, param *build.Param) error {
	b := Builder{
		build: param.GetBuilder(),
	}
	b.packages = param.GetPack()
	dst := NewGoWriter()
	err := b.Node(dst, fset, file)
	if err != nil {
		return err
	}
	if 0 == len(dst.packages) {
		return nil
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

	if 0 < dst.data.GetCode().Len() {
		err := b.writerFile(dst.data, dst.data.Packages, filepath.Join(dir, name+".data.go"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.enum.GetCode().Len() {
		err = b.writerFile(dst.enum, dst.enum.Packages, filepath.Join(dir, name+".enum.go"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.server.GetCode().Len() {
		err = b.writerFile(dst.server, dst.server.Packages, filepath.Join(dir, name+".server.go"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.database.GetCode().Len() {
		err = b.writerFile(dst.database, dst.database.Packages, filepath.Join(dir, name+".database.go"))
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) writerFile(data *build.Writer, packages string, out string) error {
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

	if 0 < len(data.GetImports()) {
		_, _ = fc.WriteString("import (\n")
		imps := make([]string, len(data.GetImports()))

		i := 0
		for key, _ := range data.GetImports() {
			imps[i] = key
			i++
		}
		sort.Strings(imps)
		for _, val := range imps {
			key := data.GetImports()[val]
			_, _ = fc.WriteString("\t")
			if 0 < len(key) {
				_, _ = fc.WriteString(key + " ")
			}
			_, _ = fc.WriteString("\"" + val + "\"\n")
		}
		_, _ = fc.WriteString(")\n\n")
	}
	_, _ = fc.WriteString(data.GetCode().String())
	return nil
}

func (b *Builder) Node(dst *GoWriter, fset *token.FileSet, node interface{}) error {
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
		return nil
	}

	dst.SetPackages(val.Value.Value[1 : len(val.Value.Value)-1])

	for _, s := range file.Specs {
		switch s.(type) {
		case *ast.ImportSpec:
		case *ast.TypeSpec:
			b.printTypeSpec(dst, (s.(*ast.TypeSpec)).Type)
		}
	}
	return nil
}

func (b *Builder) printTypeSpec(dst *GoWriter, expr ast.Expr) {
	switch expr.(type) {
	case *ast.DataType:
		b.printDataCode(dst.data, expr.(*ast.DataType))
		b.printDatabaseCode(dst.database, expr.(*ast.DataType))
	case *ast.ServerType:
		b.printServerCode(dst.server, expr.(*ast.ServerType))

	case *ast.EnumType:
		printEnumCode(dst.enum, expr.(*ast.EnumType))
	}
}

func (b *Builder) printType(dst *build.Writer, expr ast.Expr, emp bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			pack := b.getPackage(dst, expr)
			dst.Code(pack + (expr.(*ast.Ident)).Name)
		} else {
			if build.Date == (expr.(*ast.Ident)).Name {
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
			} else if build.Decimal == (expr.(*ast.Ident)).Name {
				dst.Import("github.com/shopspring/decimal", "")
			} else if build.Int64 == (expr.(*ast.Ident).Name) || build.Uint64 == (expr.(*ast.Ident).Name) {
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
			}
			dst.Code(_types[(expr.(*ast.Ident)).Name])
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		dst.Code("[]")
		b.printType(dst, ar.VType, false)
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		dst.Code("map[")
		b.printType(dst, ma.Key, true)
		dst.Code("]")
		b.printType(dst, ma.VType, false)
	case *ast.VarType:
		t := expr.(*ast.VarType)
		if t.Empty || emp {
			dst.Code("*")
		}
		b.printType(dst, t.Type(), false)
	}
}

func (b *Builder) getPackage(dst *build.Writer, expr ast.Expr) string {
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

	if pack == dst.Packages {
		return ""
	}

	packs := strings.Split(pack, ".")
	pack = packs[len(packs)-1]

	dst.Import(b.packages+pack, "")
	return pack + "."
}

func (b *Builder) getFile(name *ast.Ident) *ast.File {
	return name.Obj.Data.(*ast.File)
}

func (b *Builder) getLimit(fields []*build.DBField) (*build.DBField, bool) {
	for _, field := range fields {
		if 0 < len(field.Dbs[0].Limit) {
			return field, true
		}
	}
	return nil, false
}

func (b *Builder) getOffset(fields []*build.DBField) (*build.DBField, bool) {
	for _, field := range fields {
		if 0 < len(field.Dbs[0].Offset) {
			return field, true
		}
	}
	return nil, false
}
