package java

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printDataCode(dst *build.Writer, typ *ast.DataType) {
	dst.Import("com.hbuf.java.Data", "")

	b.printData(dst, typ)
	b.printDataEntity(dst, typ)
}

func (b *Builder) printData(dst *build.Writer, typ *ast.DataType) {
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("///" + typ.Doc.Text())
	}
	dst.Code("\tinterface " + build.StringToHumpName(typ.Name.Name) + " extends Data")
	if nil != typ.Extends {
		b.printExtend(dst, typ.Extends, true)
	}
	dst.Code(" {\n")
	for _, field := range typ.Fields.List {
		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("\t/// Get " + field.Doc.Text())
		}
		isSuper := build.CheckSuperField(field.Name.Name, typ)
		if isSuper {
			dst.Code("\t@override\n")
		}
		dst.Code("\t\t")
		b.printType(dst, field.Type, false)
		dst.Code(" get" + build.StringToHumpName(field.Name.Name))
		dst.Code("();\n\n")

		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("\t/// Set " + field.Doc.Text())
		}
		if isSuper {
			dst.Code("\t@override\n")
		}
		dst.Code("\t\tvoid set")
		dst.Code(build.StringToHumpName(field.Name.Name) + "(")
		b.printType(dst, field.Type, false)
		dst.Code(" value);\n\n")
	}

	dst.Code("\t\t" + build.StringToHumpName(typ.Name.Name) + " copy();\n")
	dst.Code("\t}\n\n")
}

func (b *Builder) printDataEntity(dst *build.Writer, typ *ast.DataType) {
	dst.Code("\tclass " + build.StringToHumpName(typ.Name.Name) + "Impl implements " + build.StringToHumpName(typ.Name.Name))
	if nil != typ.Extends {
		b.printExtend(dst, typ.Extends, true)
	}
	dst.Code(" {\n")
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("\t\t")
		b.printType(dst, field.Type, false)
		dst.Code(" " + build.StringToFirstLower(field.Name.Name) + ";\n\n")

		dst.Code("\t\t@Override\n")
		dst.Code("\t\tpublic ")
		b.printType(dst, field.Type, false)
		dst.Code(" get" + build.StringToHumpName(field.Name.Name))
		dst.Code("(){\n")
		dst.Code("\t\t\treturn this." + build.StringToFirstLower(field.Name.Name) + ";\n")
		dst.Code("\t\t}\n\n")

		dst.Code("\t\t@Override\n")
		dst.Code("\t\tpublic void set")
		dst.Code(build.StringToHumpName(field.Name.Name) + "(")
		b.printType(dst, field.Type, false)
		dst.Code(" value){\n")
		dst.Code("\t\t\tthis." + build.StringToFirstLower(field.Name.Name) + " = value;\n")
		dst.Code("\t\t}\n\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("\t\tpublic " + build.StringToHumpName(typ.Name.Name) + "Impl() {}\n\n")

	dst.Code("\t\t@Override\n")
	dst.Code("\t\tpublic " + build.StringToHumpName(typ.Name.Name) + " copy(){\n")
	dst.Code("\t\t\t" + build.StringToHumpName(typ.Name.Name) + "Impl ret = new " + build.StringToHumpName(typ.Name.Name) + "Impl();\n")
	dst.Code("\t\t\treturn ret;\n")
	dst.Code("\t\t}\n\n")

	dst.Code("\t\t@Override\n")
	dst.Code("\t\tpublic byte[] toData() throws Exception {\n")
	dst.Code("\t\t\treturn new byte[0];\n")
	dst.Code("\t\t}\n\n")

	dst.Code("\t\t@Override\n")
	dst.Code("\t\tpublic <T extends Data> T formData(byte[] data) throws Exception {\n")
	dst.Code("\t\t\treturn null;\n")
	dst.Code("\t\t}\n\n")

	dst.Code("\t}\n\n")
}

func (b *Builder) printCopy(dst *build.Writer, name string, expr ast.Expr, data *ast.DataType, empty bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				dst.Code(name)
			} else if ast.Data == t.Obj.Kind {
				if empty {
					dst.Code(name + "?.copy()")
				} else {
					dst.Code(name + ".copy()")
				}
			} else {
				dst.Code(name)
			}
		} else {
			switch build.BaseType(expr.(*ast.Ident).Name) {
			case build.Decimal:
				if empty {
					dst.Code("null == " + name + " ? null : Decimal.fromJson(" + name + "!.toJson())")
				} else {
					dst.Code("Decimal.fromJson(" + name + ".toJson())")
				}
			default:
				dst.Code(name)
			}
		}
	case *ast.ArrayType:
		t := expr.(*ast.ArrayType)
		empty = t.IsEmpty()
		if empty {
			dst.Code(name + "?.map((temp) => ")
			b.printCopy(dst, "temp", t.VType, data, empty)
			dst.Code(").toList()")
		} else {
			dst.Code(name + ".map((temp) => ")
			b.printCopy(dst, "temp", t.VType, data, empty)
			dst.Code(").toList()")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		if empty {
			dst.Code(name + "?.map((key,value) => MapEntry(")
			b.printCopy(dst, "key", t.Key, data, empty)
			dst.Code(",")
			b.printCopy(dst, "value", t.Key, data, empty)
			dst.Code(")")
		} else {
			dst.Code(name + ".map((key,value) => MapEntry(")
			b.printCopy(dst, "key", t.Key, data, empty)
			dst.Code(",")
			b.printCopy(dst, "value", t.Key, data, empty)
			dst.Code("))")
		}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printCopy(dst, name, t.Type(), data, t.Empty)
	}
}

func (b *Builder) printExtend(dst *build.Writer, extends []*ast.Extends, start bool) {
	for i, v := range extends {
		if 0 != i || start {
			dst.Code(", ")
		}
		dst.Code(build.StringToHumpName(v.Name.Name))
	}
}
