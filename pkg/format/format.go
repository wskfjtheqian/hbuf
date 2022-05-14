package format

import (
	"go/printer"
	"hbuf/pkg/ast"
	"io"
)

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

	dst.Write([]byte("package " + file.Package.Value.Value + "\n\n"))
	for _, s := range file.Imports {
		printImport(dst, s)
	}
	dst.Write([]byte("\n"))
	for _, s := range file.Specs {
		switch s.(type) {
		case *ast.ImportSpec:
		case *ast.TypeSpec:
			printTypeSpec(dst, (s.(*ast.TypeSpec)).Type)
		}
	}
	return nil
}

func printTypeSpec(dst io.Writer, expr ast.Expr) {
	switch expr.(type) {
	case *ast.DataType:
		printData(dst, expr.(*ast.DataType))
	case *ast.ServerType:
		printServer(dst, expr.(*ast.ServerType))
	case *ast.EnumType:
		printEnum(dst, expr.(*ast.EnumType))

	}
}

func printEnum(dst io.Writer, typ *ast.EnumType) {
	dst.Write([]byte("enum " + typ.Name.Name))
	dst.Write([]byte("{\n"))
	for _, item := range typ.Items {
		dst.Write([]byte("    " + item.Name + "\n"))
	}
	dst.Write([]byte("}\n\n"))
}

func printServer(dst io.Writer, typ *ast.ServerType) {
	dst.Write([]byte("server " + typ.Name.Name))
	if nil != typ.Extends {
		printExtend(dst, typ.Extends)
	}
	dst.Write([]byte("{\n"))
	for _, field := range typ.Methods.List {
		dst.Write([]byte("    "))
		var fun = field.Type.(*ast.FuncType)
		printType(dst, *fun.Result)
		dst.Write([]byte(" " + field.Name.Name))
		dst.Write([]byte("("))
		for i, field := range fun.Param.List {
			if 0 != i {
				dst.Write([]byte(", "))
			}
			printType(dst, field.Type)
			dst.Write([]byte(" " + field.Name.Name))
			if nil != field.Id {
				dst.Write([]byte(" = " + field.Id.Value))
			}
			if nil != field.Tag {
				dst.Write([]byte(" " + field.Tag.Value))
			}
		}
		dst.Write([]byte(")"))
		dst.Write([]byte("\n"))
	}
	dst.Write([]byte("}\n\n"))
}

func printData(dst io.Writer, typ *ast.DataType) {
	dst.Write([]byte("data " + typ.Name.Name))
	if nil != typ.Extends {
		printExtend(dst, typ.Extends)
	}
	dst.Write([]byte("{\n"))
	for _, field := range typ.Fields.List {
		dst.Write([]byte("    "))
		printType(dst, field.Type)
		dst.Write([]byte(" " + field.Name.Name))
		if nil != field.Id {
			dst.Write([]byte(" = " + field.Id.Value))
		}
		if nil != field.Tag {
			dst.Write([]byte(" " + field.Tag.Value))
		}
		dst.Write([]byte("\n"))
	}
	dst.Write([]byte("}\n\n"))
}

func printExtend(dst io.Writer, extends []*ast.Ident) {
	dst.Write([]byte(": "))
	for i, v := range extends {
		if 0 != i {
			dst.Write([]byte(", "))
		}
		dst.Write([]byte(v.Name))
	}
}

func printType(dst io.Writer, expr ast.Expr) {
	switch expr.(type) {
	case *ast.Ident:
		dst.Write([]byte((expr.(*ast.Ident)).Name))
	case *ast.ArrayType:
		dst.Write([]byte(((expr.(*ast.ArrayType)).Elt.(*ast.Ident)).Name + "[]"))
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		dst.Write([]byte((ma.Value.(*ast.Ident)).Name))
		if nil != ma.Key {
			dst.Write([]byte("<" + (ma.Key.(*ast.Ident)).Name + ">"))
		}
	}
}

func printImport(dst io.Writer, spec *ast.ImportSpec) {
	dst.Write([]byte("import " + spec.Path.Value + "\n"))
}
