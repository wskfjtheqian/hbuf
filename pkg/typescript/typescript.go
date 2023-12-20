package ts

import (
	"go/printer"
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

var _types = map[build.BaseType]string{
	build.Int8: "number", build.Int16: "number", build.Int32: "number", build.Int64: "Long", build.Uint8: "number",
	build.Uint16: "number", build.Uint32: "number", build.Uint64: "Long", build.Bool: "boolean", build.Float: "number",
	build.Double: "number", build.String: "string", build.Date: "Date", build.Decimal: "d.Decimal",
}

type DartWriter struct {
	data   *build.Writer
	enum   *build.Writer
	server *build.Writer
	ui     *build.Writer
	verify *build.Writer
	path   string
}

func (g *DartWriter) SetPath(s *ast.File) {
	g.path = s.Path
	g.data.File = s
	g.enum.File = s
	g.server.File = s
	g.ui.File = s
	g.verify.File = s
}

func NewGoWriter() *DartWriter {
	return &DartWriter{
		data:   build.NewWriter(),
		enum:   build.NewWriter(),
		server: build.NewWriter(),
		ui:     build.NewWriter(),
		verify: build.NewWriter(),
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
	name = name[:len(name)-len(".hbuf")]

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	if 0 < dst.data.GetCode().Len() {
		err := writerFile(dst.data, filepath.Join(dir, name+".data.ts"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.enum.GetCode().Len() {
		err = writerFile(dst.enum, filepath.Join(dir, name+".enum.ts"))
		if err != nil {
			return err
		}
	}
	if 0 < dst.server.GetCode().Len() {
		err = writerFile(dst.server, filepath.Join(dir, name+".server.ts"))
		if err != nil {
			return err
		}
	}

	//printLanguge(dst.ui)
	//if 0 < dst.ui.GetCode().Len() {
	//	err = writerFile(dst.ui, filepath.Join(dir, name+".ui.dart"))
	//	if err != nil {
	//		return err
	//	}
	//}
	//if 0 < dst.verify.GetCode().Len() {
	//	err = writerFile(dst.verify, filepath.Join(dir, name+".verify.dart"))
	//	if err != nil {
	//		return err
	//	}
	//}
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

func writerFile(data *build.Writer, out string) error {
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
	if 0 < len(data.GetImports()) {
		imps := make([]string, len(data.GetImports()))

		temp := data.GetImports()
		i := 0
		for key, _ := range temp {
			imps[i] = key
			i++
		}
		sort.Strings(imps)
		for _, val := range imps {
			_, _ = fc.WriteString("import * as " + temp[val] + " from \"" + val + "\"\n")
		}
	}
	_, _ = fc.WriteString("\n")

	_, _ = fc.WriteString(data.GetCode().String())
	return nil
}

func (b *Builder) Node(dst *DartWriter, fset *token.FileSet, node interface{}) error {
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

	dst.SetPath(file)

	for _, s := range file.Specs {
		switch s.(type) {
		case *ast.ImportSpec:
		case *ast.TypeSpec:
			b.printTypeSpec(dst, (s.(*ast.TypeSpec)).Type)
		}
	}
	return nil
}

func (b *Builder) printTypeSpec(dst *DartWriter, expr ast.Expr) {
	switch expr.(type) {
	case *ast.DataType:
		b.printDataCode(dst.data, expr.(*ast.DataType))
		//b.printFormCode(dst.ui, expr)
		//b.printVerifyCode(dst.verify, expr.(*ast.DataType))
	case *ast.ServerType:
		b.printServerCode(dst.server, expr.(*ast.ServerType))

	case *ast.EnumType:
		b.printEnumCode(dst.enum, expr.(*ast.EnumType))
		//b.printFormCode(dst.ui, expr)
	}
}

func (b *Builder) printType(dst *build.Writer, expr ast.Expr, notEmpty bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			dst.Code(b.getPackage(dst, expr, ""))
			dst.Code(".")
			dst.Code(expr.(*ast.Ident).Name)
		} else {
			if build.Decimal == build.BaseType((expr.(*ast.Ident).Name)) {
				dst.Import("decimal.js", "d")
			} else if build.Int64 == build.BaseType((expr.(*ast.Ident).Name)) || build.Uint64 == build.BaseType((expr.(*ast.Ident).Name)) {
				dst.Import("long", "Long")
			}
			dst.Code(_types[build.BaseType((expr.(*ast.Ident).Name))])
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		b.printType(dst, ar.VType, false)
		dst.Code("[]")
		//if ar.Empty && !notEmpty {
		//	dst.Code("?")
		//}
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		dst.Code("Map<")
		b.printType(dst, ma.Key, false)
		dst.Code(", ")
		b.printType(dst, ma.VType, false)
		dst.Code(">")
		//if ma.Empty && !notEmpty {
		//	dst.Code("?")
		//}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printType(dst, t.Type(), false)
		//if t.Empty && !notEmpty {
		//	dst.Code("?")
		//}
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
	name = name[:len(name)-len(".hbuf")]
	if 0 < len(s) {
		name = name + "." + s
	} else {
		switch (expr.(*ast.Ident)).Obj.Kind {
		case ast.Data:
			name = name + ".data"
		case ast.Enum:
			name = name + ".enum"
		case ast.Server:
			name = name + ".server"
		}
	}

	return dst.Import("./"+name, "$"+strconv.Itoa(len(dst.GetImports())))
}
