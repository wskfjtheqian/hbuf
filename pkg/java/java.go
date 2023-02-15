package java

import (
	"go/printer"
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/token"
	"os"
	"path/filepath"
	"sort"
)

var _types = map[build.BaseType]string{
	build.Int8: "byte", build.Int16: "short", build.Int32: "int", build.Int64: "long", build.Uint8: "char",
	build.Uint16: "int", build.Uint32: "long", build.Uint64: "BigInteger", build.Bool: "boolean", build.Float: "float",
	build.Double: "double", build.String: "String", build.Date: "Date", build.Decimal: "BigDecimal",
}

var _nullTypes = map[build.BaseType]string{
	build.Int8: "Byte", build.Int16: "Short", build.Int32: "Integer", build.Int64: "Long", build.Uint8: "Character",
	build.Uint16: "Integer", build.Uint32: "Long", build.Uint64: "BigInteger", build.Bool: "Boolean", build.Float: "Float",
	build.Double: "Double", build.String: "String", build.Date: "Date", build.Decimal: "BigDecimal",
}

type JavaWriter struct {
	data     *build.Writer
	enum     *build.Writer
	server   *build.Writer
	path     string
	Packages string
}

func (g *JavaWriter) SetPath(s *ast.File) {
	g.path = s.Path
	g.data.File = s
	g.enum.File = s
	g.server.File = s
}

func (w *JavaWriter) SetPackages(s string) {
	w.Packages = s
	w.data.Packages = s
	w.enum.Packages = s
	w.server.Packages = s
}

func NewGoWriter() *JavaWriter {
	return &JavaWriter{
		data:   build.NewWriter(),
		enum:   build.NewWriter(),
		server: build.NewWriter(),
	}
}

type Builder struct {
	lang map[string]struct{}
	pkg  *ast.Package
}

func Build(file *ast.File, fset *token.FileSet, param *build.Param) error {
	b := Builder{
		lang: map[string]struct{}{},
		pkg:  param.GetPkg(),
	}
	dst := NewGoWriter()
	err := b.Node(dst, fset, file)
	if err != nil {
		return err
	}

	if 0 == len(dst.path) {
		return nil
	}

	dir, name := filepath.Split(param.GetOut())
	name = build.StringToHumpName(name[:len(name)-len(".hbuf")])

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	if 0 < dst.data.GetCode().Len() {
		err := writerFile(dst.data, dst.Packages, filepath.Join(dir, name+"Data.java"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.enum.GetCode().Len() {
		err = writerFile(dst.enum, dst.Packages, filepath.Join(dir, name+"Enum.java"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.server.GetCode().Len() {
		err = writerFile(dst.server, dst.Packages, filepath.Join(dir, name+"Server.java"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) GetDataType(file *ast.File, name string) *ast.Object {
	if obj := file.Scope.Lookup(name); nil != obj {
		switch obj.Decl.(type) {
		case *ast.TypeSpec:
			t := (obj.Decl.(*ast.TypeSpec)).Type
			switch t.(type) {
			case *ast.DataType:
				return obj
			case *ast.EnumType:
				return obj
			}
		}
	}
	for _, spec := range file.Imports {
		if f, ok := b.pkg.Files[spec.Path.Value]; ok {
			if obj := f.Scope.Lookup(name); nil != obj {
				switch obj.Decl.(type) {
				case *ast.TypeSpec:
					t := (obj.Decl.(*ast.TypeSpec)).Type
					switch t.(type) {
					case *ast.DataType:
						return obj
					case *ast.EnumType:
						return obj
					}
				}
			}
		}
	}
	return nil
}

func writerFile(data *build.Writer, packages string, out string) error {
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

	_, _ = fc.WriteString("package " + packages + ";\n\n")

	if 0 < len(data.GetImports()) {
		imps := make([]string, len(data.GetImports()))

		i := 0
		for key, _ := range data.GetImports() {
			imps[i] = key
			i++
		}
		sort.Strings(imps)
		for _, val := range imps {
			_, _ = fc.WriteString("import " + val + ";\n")
		}
	}
	_, _ = fc.WriteString("\n")
	_, _ = fc.WriteString(data.GetCode().String())
	return nil
}

func (b *Builder) Node(dst *JavaWriter, fset *token.FileSet, node interface{}) error {
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
	val, ok := file.Packages["java"]
	if !ok {
		return nil
	}

	dst.SetPath(file)
	dst.SetPackages(val.Value.Value[1 : len(val.Value.Value)-1])

	dst.data.Code("public interface UserData {\n")
	dst.server.Code("public interface UserServer {\n")
	dst.enum.Code("public interface UserEnum {\n")
	for _, s := range file.Specs {
		switch s.(type) {
		case *ast.ImportSpec:
		case *ast.TypeSpec:
			b.printTypeSpec(dst, (s.(*ast.TypeSpec)).Type)
		}
	}
	if 0 < dst.data.GetCode().Len() {
		dst.data.Code("}\n")
	}
	if 0 < dst.server.GetCode().Len() {
		dst.server.Code("}\n")
	}
	if 0 < dst.enum.GetCode().Len() {
		dst.enum.Code("}\n")
	}
	return nil
}

func (b *Builder) printTypeSpec(dst *JavaWriter, expr ast.Expr) {
	switch expr.(type) {
	case *ast.DataType:
		b.printDataCode(dst.data, expr.(*ast.DataType))
	case *ast.ServerType:
		b.printServerCode(dst.server, expr.(*ast.ServerType))
	case *ast.EnumType:
		b.printEnumCode(dst.enum, expr.(*ast.EnumType))
	}
}

func (b *Builder) printType(dst *build.Writer, expr ast.Expr, notEmpty bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			b.getPackage(dst, expr, "")
			dst.Code(expr.(*ast.Ident).Name)
		} else {
			if build.Decimal == build.BaseType((expr.(*ast.Ident).Name)) {
				dst.Import("java.math.BigDecimal", "")
			} else if build.Date == build.BaseType((expr.(*ast.Ident).Name)) {
				dst.Import("java.util.Date", "")
			} else if build.Int64 == build.BaseType((expr.(*ast.Ident).Name)) {
				dst.Import("java.math.BigInteger", "")
			}
			if notEmpty {
				dst.Code(_types[build.BaseType((expr.(*ast.Ident).Name))])
			} else {
				dst.Code(_nullTypes[build.BaseType((expr.(*ast.Ident).Name))])
			}
		}
	case *ast.ArrayType:
		dst.Import("java.util.List", "")
		ar := expr.(*ast.ArrayType)
		dst.Code("List<")
		b.printType(dst, ar.VType, false)
		dst.Code(">")
	case *ast.MapType:
		dst.Import("java.util.Map", "")
		ma := expr.(*ast.MapType)
		dst.Code("Map<")
		b.printType(dst, ma.Key, false)
		dst.Code(", ")
		b.printType(dst, ma.VType, false)
		dst.Code(">")
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printType(dst, t.Type(), notEmpty && !t.Empty)
	}
}

func (b *Builder) getPackage(dst *build.Writer, expr ast.Expr, s string) string {
	file := (expr.(*ast.Ident)).Obj.Data
	switch file.(type) {
	case *ast.File:
		break
	default:
		return ""
	}

	_, name := filepath.Split(file.(*ast.File).Path)
	name = dst.Packages + "." + build.StringToHumpName(name[:len(name)-len(".hbuf")])
	if 0 < len(s) {
		name = name + "." + s + ".dart"
	} else {
		switch (expr.(*ast.Ident)).Obj.Kind {
		case ast.Data:
			name = name + "Data.*"
		case ast.Enum:
			name = name + "Enum.*"
		case ast.Server:
			name = name + "Server.*"
		}
	}

	dst.Import(name, "")
	return ""
}
