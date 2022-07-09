package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printDataCode(dst *Writer, typ *ast.DataType) {
	dst.Import("dart:typed_data")
	dst.Import("package:hbuf_dart/hbuf_dart.dart")

	b.printData(dst, typ)
	b.printDataEntity(dst, typ)
}
func (b *Builder) printData(dst *Writer, typ *ast.DataType) {
	dst.Code("abstract class " + build.StringToHumpName(typ.Name.Name) + " implements Data")
	if nil != typ.Extends {
		b.printExtend(dst, typ.Extends, true)
	}
	dst.Code("{\n")
	for _, field := range typ.Fields.List {
		if nil != field.Comment {
			dst.Code("  /// " + field.Comment.Text())
		}
		isSuper := build.CheckSuperField(field.Name.Name, typ)
		if isSuper {
			dst.Code("  @override\n")
		}
		dst.Code("  ")
		b.printType(dst, field.Type, false)
		dst.Code(" get " + build.StringToFirstLower(field.Name.Name))
		dst.Code(";\n\n")

		if isSuper {
			dst.Code("  @override\n")
		}
		dst.Code("  set ")
		dst.Code(build.StringToFirstLower(field.Name.Name) + "(")
		b.printType(dst, field.Type, false)
		dst.Code(" value);\n\n")
	}
	isParam := false
	dst.Code("  factory " + build.StringToHumpName(typ.Name.Name) + "(")
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if !isParam {
			dst.Code("{\n")
			isParam = true
		}
		dst.Code("    ")
		if !field.Type.IsEmpty() {
			dst.Code("required ")
		}
		b.printType(dst, field.Type, false)
		dst.Code(" " + build.StringToFirstLower(field.Name.Name))
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	if isParam {
		dst.Code("}")
	}
	dst.Code("  ){\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + "(\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("      ")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(": ")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("    );\n")
	dst.Code("  }\n\n")

	dst.Code("  static " + build.StringToHumpName(typ.Name.Name) + " fromMap(Map<String, dynamic> map){\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + ".fromMap(map);\n")
	dst.Code("  }\n\n")

	dst.Code("  static " + build.StringToHumpName(typ.Name.Name) + " fromData(ByteData data){\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + ".fromData(data);\n")
	dst.Code("  }\n\n")

	dst.Code("}\n\n")
}

func (b *Builder) printDataEntity(dst *Writer, typ *ast.DataType) {
	dst.Code("class _" + build.StringToHumpName(typ.Name.Name) + " implements " + build.StringToHumpName(typ.Name.Name))
	dst.Code(" {\n")

	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("  @override\n")
		dst.Code("  ")
		b.printType(dst, field.Type, false)
		dst.Code(" " + build.StringToFirstLower(field.Name.Name))
		dst.Code(";\n\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("  _" + build.StringToHumpName(typ.Name.Name) + "(")
	isParam := false
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if !isParam {
			dst.Code("{\n")
			isParam = true
		}
		dst.Code("    ")
		if !field.Type.IsEmpty() {
			dst.Code("required ")
		}
		dst.Code("this." + build.StringToFirstLower(field.Name.Name))
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("  ")
	if isParam {
		dst.Code("}")
	}
	dst.Code(");\n\n")

	dst.Code("  static _" + build.StringToHumpName(typ.Name.Name) + " fromMap(Map<String, dynamic> map){\n")
	dst.Code("     dynamic temp;\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + "(\n")

	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("      " + build.StringToFirstLower(field.Name.Name) + ": ")
		jsonName := build.StringToUnderlineName(field.Name.Name)
		b.printJsonValue(dst, "(temp = map[\""+jsonName+"\"])", field.Type, data, false)
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("    );\n")
	dst.Code("  }\n")

	dst.Code("\n")
	dst.Code("  @override\n")
	dst.Code("  Map<String, dynamic> toMap() {\n")
	dst.Code("    return {\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("      \"" + build.StringToUnderlineName(field.Name.Name))
		dst.Code("\": " + build.StringToFirstLower(field.Name.Name) + ",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("    };\n")
	dst.Code("  }\n")

	dst.Code("  static _" + build.StringToHumpName(typ.Name.Name) + " fromData(ByteData data){\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + ".fromMap({});\n")

	//err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
	//	dst.Code("      " + build.StringToFirstLower(field.Name.Name))
	//	dst.Code(": map[\"" + getJsonName(field) + "\"]")
	//	dst.Code(",\n")
	//	return nil
	//})
	//if err != nil {
	//	return
	//}

	//dst.Code("    );\n")
	dst.Code("  }\n")

	dst.Code("\n")
	dst.Code("  @override\n")
	dst.Code("  ByteData toData() {\n")
	dst.Code("    return ByteData.view(Uint8List(12).buffer);\n")
	//err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
	//	dst.Code("      \"" + getJsonName(field))
	//	dst.Code("\": " + build.StringToFirstLower(field.Name.Name) + ",\n")
	//	return nil
	//})
	//if err != nil {
	//	return
	//}
	//dst.Code("    };\n")
	dst.Code("  }\n")

	dst.Code("}\n\n")
}

func (b *Builder) printJsonValue(dst *Writer, name string, expr ast.Expr, data *ast.DataType, empty bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				if empty {
					dst.Code("null == " + name + " ? null : temp is num ? " + t.Name + ".valueOf(temp.toInt()) : null == num.tryParse(temp.toString()) ? null : " + t.Name + ".valueOf(num.tryParse(temp.toString())!.toInt())")
				} else {
					dst.Code("null == " + name + " ? " + t.Name + ".valueOf(0) : " + t.Name + ".valueOf(temp is num ? temp.toInt() : num.tryParse(temp.toString())?.toInt() ?? 0)")
				}
			} else if ast.Data == t.Obj.Kind {
				if empty {
					dst.Code("null == " + name + " ? null : " + t.Name + ".fromMap(temp)")
				} else {
					dst.Code("null == " + name + " ? " + t.Name + ".fromMap({}) : " + t.Name + ".fromMap(temp)")
				}
			} else {
				dst.Code("map[\"" + name + "\"]")
			}
		} else {
			switch expr.(*ast.Ident).Name {
			case build.Int8, build.Int16, build.Int32, build.Int64, build.Uint8, build.Uint16, build.Uint32, build.Uint64:
				if empty {
					dst.Code("null == " + name + " ? null : (temp is num ? temp.toInt() : num.tryParse(temp.toString())?.toInt())")
				} else {
					dst.Code("null == " + name + " ? 0 : (temp is num ? temp.toInt() : num.tryParse(temp.toString())?.toInt() ?? 0)")
				}
				break
			case build.Float, build.Double:
				if empty {
					dst.Code("null == " + name + " ? null : (temp is num ? temp.toDouble() : num.tryParse(temp.toString())?.toDouble())")
				} else {
					dst.Code("null == " + name + " ? 0 : (temp is num ? temp.toDouble() : num.tryParse(temp.toString())?.toDouble() ?? 0)")
				}
				break
			case build.String:
				if empty {
					dst.Code("null == " + name + " ? null : (temp is String ? temp : temp.toString())")
				} else {
					dst.Code("null == " + name + " ? \"\" : (temp is String ? temp : temp.toString())")
				}
			case build.Date:
				if empty {
					dst.Code("null == " + name + " ? null : temp is num ? DateTime.fromMillisecondsSinceEpoch(temp.toInt()) : null == num.tryParse(temp.toString()) ? null : DateTime.fromMillisecondsSinceEpoch(num.tryParse(temp.toString())!.toInt())")
				} else {
					dst.Code("null == " + name + " ? DateTime.fromMillisecondsSinceEpoch(0) : DateTime.fromMillisecondsSinceEpoch(temp is num ? temp.toInt() : num.tryParse(temp.toString())?.toInt() ?? 0)")
				}
			case build.Bool:
				if empty {
					dst.Code("null == " + name + " ? null : (temp is bool ? temp: (temp is num ? 0 != temp : (null == num.tryParse(temp.toString()) ? null : 0 != num.tryParse(temp.toString()))))")
				} else {
					dst.Code("null == " + name + " ? false:(temp is bool ? temp: 0 != (temp is num ? temp : num.tryParse(temp.toString()) ?? 0))")
				}
			default:
				dst.Code("map[\"" + name + "\"]")
			}
		}
	case *ast.ArrayType:
		t := expr.(*ast.ArrayType)
		if empty {
			dst.Code("null == " + name + " ? null : (temp! is! List ? null : (temp as List).map((temp) => ")
			b.printJsonValue(dst, "temp", t.VType, data, empty)
			dst.Code(").toList())")
		} else {
			dst.Code("null == " + name + " ? <")
			b.printType(dst, t.VType, false)
			dst.Code(">[] : (temp! is! List ? <")
			b.printType(dst, t.VType, false)
			dst.Code(">[] : (temp as List).map((temp) => ")
			b.printJsonValue(dst, "temp", t.VType, data, empty)
			dst.Code(").toList())")
		}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printJsonValue(dst, name, t.Type(), data, t.Empty)
	}
}

func (b *Builder) printExtend(dst *Writer, extends []*ast.Ident, start bool) {
	for i, v := range extends {
		if 0 != i || start {
			dst.Code(", ")
		}
		dst.Code(build.StringToHumpName(v.Name))

	}
}
