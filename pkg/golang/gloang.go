package golang

import (
	"go/printer"
	ast "hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var _types = map[build.BaseType]string{
	build.Int8: "int8", build.Int16: "int16", build.Int32: "int32", build.Int64: "hbuf.Int64", build.Uint8: "uint8",
	build.Uint16: "uint16", build.Uint32: "uint32", build.Uint64: "hbuf.Uint64", build.Bool: "bool", build.Float: "float32",
	build.Double: "float64", build.String: "string", build.Date: "hbuf.Time", build.Decimal: "decimal.Decimal",
}
var _typesDefaultValue = map[build.BaseType]string{
	build.Int8:    "int8(0)",
	build.Int16:   "int16(0)",
	build.Int32:   "int32(0)",
	build.Int64:   "hbuf.Int64(0)",
	build.Uint8:   "uint8(0)",
	build.Uint16:  "uint16(0)",
	build.Uint32:  "uint32(0)",
	build.Uint64:  "hbuf.Uint64(0)",
	build.Bool:    "false",
	build.Float:   "float32(0)",
	build.Double:  "float64(0)",
	build.String:  "\"\"",
	build.Date:    "hbuf.Time{}",
	build.Decimal: "decimal.Zero",
}

type GoWriter struct {
	data     *build.Writer
	enum     *build.Writer
	server   *build.Writer
	database *build.Writer
	verify   *build.Writer

	packages string
}

func (w *GoWriter) SetPackages(s string) {
	w.data.Packages = s
	w.enum.Packages = s
	w.server.Packages = s
	w.database.Packages = s
	w.verify.Packages = s
	w.packages = s
}

func (w *GoWriter) SetPath(file *ast.File) {
	w.data.File = file
	w.enum.File = file
	w.server.File = file
	w.database.File = file
	w.verify.File = file
}

func NewGoWriter() *GoWriter {
	return &GoWriter{
		data:     build.NewWriter(),
		enum:     build.NewWriter(),
		server:   build.NewWriter(),
		database: build.NewWriter(),
		verify:   build.NewWriter(),
	}
}

type Builder struct {
	build    *build.Builder
	pkg      *ast.Package
	packages string
	fSet     *token.FileSet
}

func Build(file *ast.File, fSet *token.FileSet, param *build.Param) error {
	b := Builder{
		fSet:  fSet,
		build: param.GetBuilder(),
		pkg:   param.GetPkg(),
	}
	b.packages = param.GetPack()
	dst := NewGoWriter()
	err := b.Node(dst, fSet, file)
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
		err := b.writerFile(dst.data, dst.data.Packages, filepath.Join(dir, name+".data.go"), 1)
		if err != nil {
			return err
		}
	}
	if 0 < dst.enum.GetCode().Len() {
		err = b.writerFile(dst.enum, dst.enum.Packages, filepath.Join(dir, name+".enum.go"), 0)
		if err != nil {
			return err
		}
	}
	if 0 < dst.server.GetCode().Len() {
		err = b.writerFile(dst.server, dst.server.Packages, filepath.Join(dir, name+".server.go"), 0)
		if err != nil {
			return err
		}
	}
	if 0 < dst.database.GetCode().Len() {
		err = b.writerFile(dst.database, dst.database.Packages, filepath.Join(dir, name+".database.go"), 0)
		if err != nil {
			return err
		}
	}
	if 0 < dst.verify.GetCode().Len() {
		err = b.writerFile(dst.verify, dst.verify.Packages, filepath.Join(dir, name+".verify.go"), 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) writerFile(data *build.Writer, packages string, out string, i int) error {
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
	code := data.GetCode().String()
	_, _ = fc.WriteString(code[:len(code)-1])
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

	dst.SetPath(file)
	dst.SetPackages(val.Value.Value[1 : len(val.Value.Value)-1])

	for _, s := range file.Specs {
		switch s.(type) {
		case *ast.ImportSpec:
		case *ast.TypeSpec:
			err := b.printTypeSpec(dst, (s.(*ast.TypeSpec)).Type)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Builder) printTypeSpec(dst *GoWriter, expr ast.Expr) error {
	switch expr.(type) {
	case *ast.DataType:
		b.printDataCode(dst.data, expr.(*ast.DataType))
		err := b.printDatabaseCode(dst.database, expr.(*ast.DataType))
		if err != nil {
			return err
		}
		err = b.printVerifyCode(dst.verify, expr.(*ast.DataType))
		if err != nil {
			return err
		}
	case *ast.ServerType:
		err := b.printServerCode(dst.server, expr.(*ast.ServerType))
		if err != nil {
			return err
		}

	case *ast.EnumType:
		printEnumCode(dst.enum, expr.(*ast.EnumType))
	}
	return nil
}

func (b *Builder) printType(dst *build.Writer, expr ast.Expr, b2 bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			pack := b.getPackage(dst, expr)
			dst.Code(pack + (expr.(*ast.Ident)).Name)
		} else {
			if build.Date == build.BaseType((expr.(*ast.Ident)).Name) {
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
			} else if build.Decimal == build.BaseType((expr.(*ast.Ident)).Name) {
				dst.Import("github.com/shopspring/decimal", "")
			} else if build.Int64 == build.BaseType((expr.(*ast.Ident).Name)) || build.Uint64 == build.BaseType((expr.(*ast.Ident).Name)) {
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
			}
			dst.Code(_types[build.BaseType((expr.(*ast.Ident)).Name)])
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		dst.Code("[]")
		b.printType(dst, ar.VType, true)
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		dst.Code("map[")
		b.printType(dst, ma.Key, true)
		dst.Code("]")
		b.printType(dst, ma.VType, true)
	case *ast.VarType:
		t := expr.(*ast.VarType)
		if b2 && t.Empty {
			dst.Code("*")
		}
		b.printType(dst, t.Type(), true)
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

func (b *Builder) getKey(dst *build.Writer, fields []*build.DBField, name string) (*build.Writer, *build.Writer, bool) {
	param := build.NewWriter()
	param.Packages = dst.Packages
	where := build.NewWriter()
	where.Packages = dst.Packages

	for _, field := range fields {
		if field.Field.Name.Name == name {
			where.Code(build.StringToFirstLower(field.Field.Name.Name))
			b.printType(param, field.Field.Type, false)
			return param, where, true
		}
	}
	return param, where, false
}

func (b *Builder) converter(field *build.DBField, name string) string {
	fName := build.StringToHumpName(field.Field.Name.Name)
	if "json" == field.Dbs[0].Converter {
		return "db.NewJson(&" + name + "." + fName + ")"
	}
	return "&" + name + "." + fName
}
