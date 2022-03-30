package dart

import (
	"go/printer"
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var _types = map[string]string{
	build.Int8: "int", build.Int16: "int", build.Int32: "int", build.Int64: "int", build.Uint8: "int",
	build.Uint16: "int", build.Uint32: "int", build.Uint64: "int", build.Bool: "bool", build.Float: "double",
	build.Double: "double", build.String: "String",
}

func Build(file *ast.File, out string) error {
	fc, err := os.Create(out + ".dart")
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
	case *ast.EnumType:
		printEnum(dst, expr.(*ast.EnumType))
	}
}

func printEnum(dst io.Writer, typ *ast.EnumType) {
	_, _ = dst.Write([]byte("class " + typ.Name.Name))
	_, _ = dst.Write([]byte("{\n"))
	_, _ = dst.Write([]byte("  final int value;\n"))
	_, _ = dst.Write([]byte("  final String name;\n\n"))

	_, _ = dst.Write([]byte("  const " + typ.Name.Name + "._(this.value, this.name);\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  bool operator ==(Object other) =>\n"))
	_, _ = dst.Write([]byte("      identical(this, other) ||\n"))
	_, _ = dst.Write([]byte("      other is Gender &&\n"))
	_, _ = dst.Write([]byte("          runtimeType == other.runtimeType &&\n"))
	_, _ = dst.Write([]byte("          value == other.value;\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  int get hashCode => value.hashCode;\n\n"))

	_, _ = dst.Write([]byte("  static Gender valueOf(int value) {\n"))
	_, _ = dst.Write([]byte("  	for (var item in values) {\n"))
	_, _ = dst.Write([]byte("  		if (item.value == value) {\n"))
	_, _ = dst.Write([]byte("  			return item;\n"))
	_, _ = dst.Write([]byte("  		}\n"))
	_, _ = dst.Write([]byte("  	}\n"))
	_, _ = dst.Write([]byte("  	throw 'Get Gender by value error, value=$value';\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("  static Gender nameOf(String name) {\n"))
	_, _ = dst.Write([]byte("  	for (var item in values) {\n"))
	_, _ = dst.Write([]byte("  		if (item.name == name) {\n"))
	_, _ = dst.Write([]byte("  			return item;\n"))
	_, _ = dst.Write([]byte("  		}\n"))
	_, _ = dst.Write([]byte("  	}\n"))
	_, _ = dst.Write([]byte("  	throw 'Get Gender by name error, name=$name';\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	for _, item := range typ.Items {
		_, _ = dst.Write([]byte("  static const " + item.Name.Name + " = Gender._(" + item.Id.Value + ", '" + item.Name.Name + "');\n"))
	}
	_, _ = dst.Write([]byte("\n"))
	_, _ = dst.Write([]byte("  static const List<Gender> values = [\n"))
	for _, item := range typ.Items {
		_, _ = dst.Write([]byte("    " + item.Name.Name + ",\n"))
	}
	_, _ = dst.Write([]byte("  ];\n\n"))

	_, _ = dst.Write([]byte("}\n\n"))
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
		for i, field := range fun.Params.List {
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
	_, _ = dst.Write([]byte("abstract class " + toClassName(typ.Name.Name) + " implements "))
	if nil != typ.Extends {
		printExtend(dst, typ.Extends)
	}
	_, _ = dst.Write([]byte("Data {\n"))
	for _, field := range typ.Fields.List {
		if nil != field.Comment {
			_, _ = dst.Write([]byte("  /// " + field.Comment.Text()))
		}

		_, _ = dst.Write([]byte("  "))
		printType(dst, field.Type)
		_, _ = dst.Write([]byte("? " + toFieldName(field.Name.Name)))
		_, _ = dst.Write([]byte(";\n\n"))
	}
	_, _ = dst.Write([]byte("  factory " + toClassName(typ.Name.Name) + "({\n"))
	err := build.EnumField(typ, func(field *ast.Field) error {
		_, _ = dst.Write([]byte("    "))
		printType(dst, field.Type)
		_, _ = dst.Write([]byte("? " + toFieldName(field.Name.Name)))
		_, _ = dst.Write([]byte(",\n"))
		return nil
	})
	if err != nil {
		return
	}
	_, _ = dst.Write([]byte("  }){\n"))
	_, _ = dst.Write([]byte("    return _" + toClassName(typ.Name.Name) + "(\n"))
	err = build.EnumField(typ, func(field *ast.Field) error {
		_, _ = dst.Write([]byte("      "))
		_, _ = dst.Write([]byte(toFieldName(field.Name.Name)))
		_, _ = dst.Write([]byte(": "))
		_, _ = dst.Write([]byte(toFieldName(field.Name.Name)))
		_, _ = dst.Write([]byte(",\n"))
		return nil
	})
	if err != nil {
		return
	}
	_, _ = dst.Write([]byte("    );\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("  static " + toClassName(typ.Name.Name) + "? fromMap(Map<String, dynamic> map){\n"))
	_, _ = dst.Write([]byte("    return _" + toClassName(typ.Name.Name) + ".fromMap(map);\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("}\n\n"))
}

func printDataEntity(dst io.Writer, typ *ast.DataType) {
	_, _ = dst.Write([]byte("class _" + toClassName(typ.Name.Name) + " implements " + toClassName(typ.Name.Name)))
	_, _ = dst.Write([]byte(" {\n"))

	err := build.EnumField(typ, func(field *ast.Field) error {
		_, _ = dst.Write([]byte("  @override\n"))
		_, _ = dst.Write([]byte("  "))
		printType(dst, field.Type)
		_, _ = dst.Write([]byte("? " + toFieldName(field.Name.Name)))
		_, _ = dst.Write([]byte(";\n\n"))
		return nil
	})
	if err != nil {
		return
	}

	_, _ = dst.Write([]byte("  _" + toClassName(typ.Name.Name) + "({\n"))
	err = build.EnumField(typ, func(field *ast.Field) error {
		_, _ = dst.Write([]byte("    this." + toFieldName(field.Name.Name)))
		_, _ = dst.Write([]byte(",\n"))
		return nil
	})
	if err != nil {
		return
	}
	_, _ = dst.Write([]byte("  });\n\n"))

	_, _ = dst.Write([]byte("  static _" + toClassName(typ.Name.Name) + "? fromMap(Map<String, dynamic> map){\n"))
	_, _ = dst.Write([]byte("    return _" + toClassName(typ.Name.Name) + "(\n"))

	err = build.EnumField(typ, func(field *ast.Field) error {
		_, _ = dst.Write([]byte("      " + toFieldName(field.Name.Name)))
		_, _ = dst.Write([]byte(": map[\"" + getJsonName(field) + "\"]"))
		_, _ = dst.Write([]byte(",\n"))
		return nil
	})
	if err != nil {
		return
	}

	_, _ = dst.Write([]byte("    );\n"))
	_, _ = dst.Write([]byte("  }\n"))

	_, _ = dst.Write([]byte("\n"))
	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  Map<String, dynamic> toMap() {\n"))
	_, _ = dst.Write([]byte("    return {\n"))
	err = build.EnumField(typ, func(field *ast.Field) error {
		_, _ = dst.Write([]byte("      \"" + getJsonName(field)))
		_, _ = dst.Write([]byte("\": " + toFieldName(field.Name.Name) + ",\n"))
		return nil
	})
	if err != nil {
		return
	}
	_, _ = dst.Write([]byte("    };\n"))
	_, _ = dst.Write([]byte("  }\n"))

	_, _ = dst.Write([]byte("}\n"))
}

func getJsonName(field *ast.Field) string {
	return field.Name.Name
}

func printExtend(dst io.Writer, extends []*ast.Ident) {
	for _, v := range extends {
		_, _ = dst.Write([]byte(toClassName(v.Name)))
		_, _ = dst.Write([]byte(", "))
	}
}

func printType(dst io.Writer, expr ast.Expr) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			_, _ = dst.Write([]byte((expr.(*ast.Ident)).Name))
		} else {
			_, _ = dst.Write([]byte(_types[(expr.(*ast.Ident)).Name]))
		}

	case *ast.ArrayType:
		_, _ = dst.Write([]byte("List<"))
		printType(dst, (expr.(*ast.ArrayType)).Elt.(*ast.Ident))
		_, _ = dst.Write([]byte("?>"))
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		_, _ = dst.Write([]byte("Map<"))
		printType(dst, ma.Key.(*ast.Ident))
		_, _ = dst.Write([]byte(", "))
		printType(dst, ma.Value.(*ast.Ident))
		_, _ = dst.Write([]byte("?>"))
	}
}

func toClassName(name string) string {
	var ret = ""
	for _, item := range strings.Split(name, "_") {
		if 0 < len(item) {
			ret += strings.ToUpper(item[0:1]) + item[1:]
		}
	}
	return ret
}

func toFieldName(name string) string {
	var ret = ""
	for i, item := range strings.Split(name, "_") {
		if 0 < len(item) {
			if 0 == i {
				ret += item
			} else {
				ret += strings.ToUpper(item[0:1]) + item[1:]
			}

		}
	}
	return ret
}
