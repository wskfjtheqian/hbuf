package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printDataCode(dst *build.Writer, typ *ast.DataType) {
	dst.Import("dart:typed_data", "")
	dst.Import("package:hbuf_dart/hbuf_dart.dart", "")

	b.printData(dst, typ)
	b.printDataEntity(dst, typ)

}
func (b *Builder) printData(dst *build.Writer, typ *ast.DataType) {
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
		dst.Code("  }")
	} else {
		dst.Code("  ")
	}
	dst.Code("){\n")
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

	isParam = false
	dst.Code("  " + build.StringToHumpName(typ.Name.Name) + " copyWith(")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if !isParam {
			dst.Code("{\n")
			isParam = true
		}
		dst.Code("    ")
		b.printType(dst, field.Type, true)
		dst.Code("? " + build.StringToFirstLower(field.Name.Name))
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	if isParam {
		dst.Code("  }")
	} else {
		dst.Code("  ")
	}
	dst.Code(");\n\n")

	dst.Code("  @override\n")
	dst.Code("  " + build.StringToHumpName(typ.Name.Name) + " copy();\n")

	dst.Code("}\n\n")
}

func (b *Builder) printDataEntity(dst *build.Writer, typ *ast.DataType) {
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
		b.printFormMap(dst, "(temp = map[\""+jsonName+"\"])", field.Type, data, false)
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
		dst.Code("\":")
		b.printToMap(dst, build.StringToFirstLower(field.Name.Name), field.Type, data, false)
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("    };\n")
	dst.Code("  }\n\n")

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
	dst.Code("  }\n\n")

	isParam = false
	dst.Code("  _" + build.StringToHumpName(typ.Name.Name) + " copyWith(")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if !isParam {
			dst.Code("{\n")
			isParam = true
		}
		dst.Code("    ")
		b.printType(dst, field.Type, true)
		dst.Code("? " + build.StringToFirstLower(field.Name.Name))
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	if isParam {
		dst.Code("  }")
	} else {
		dst.Code("  ")
	}
	dst.Code(") {\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + "(\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("      ")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(": ")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		dst.Code("?? this.")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("    );\n")
	dst.Code("  }\n\n")

	dst.Code("  @override\n")
	dst.Code("  bool operator ==(Object other) =>\n")
	dst.Code("      identical(this, other) ||\n")
	dst.Code("      other is _" + build.StringToHumpName(typ.Name.Name) + " &&\n")
	dst.Code("          runtimeType == other.runtimeType ")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("&& \n          ")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(" == other.")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		return nil
	})
	if err != nil {
		return
	}
	dst.Code(";\n\n")

	dst.Code("  @override\n")
	dst.Code("  int get hashCode => 0 ")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code(" ^ ")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(".hashCode")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code(";\n\n")

	dst.Code("  @override\n")
	dst.Code("  " + build.StringToHumpName(typ.Name.Name) + " copy(){\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + "(\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("      ")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(": ")
		b.printCopy(dst, build.StringToFirstLower(field.Name.Name), field.Type, data, true)
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("    );\n")
	dst.Code("  }\n")

	dst.Code("}\n\n")
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
			switch expr.(*ast.Ident).Name {
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

func (b *Builder) printFormMap(dst *build.Writer, name string, expr ast.Expr, data *ast.DataType, empty bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				if empty {
					dst.Code("null == " + name + " ? null :  " + name + "  is num ? " + t.Name + ".valueOf( " + name + " .toInt()) : null == num.tryParse( " + name + " .toString()) ? null : " + t.Name + ".valueOf(num.tryParse( " + name + " .toString())!.toInt())")
				} else {
					dst.Code("null == " + name + " ? " + t.Name + ".valueOf(0) : " + t.Name + ".valueOf( " + name + "  is num ?  " + name + " .toInt() : num.tryParse( " + name + " .toString())?.toInt() ?? 0)")
				}
			} else if ast.Data == t.Obj.Kind {
				if empty {
					dst.Code("null == " + name + " ? null : " + t.Name + ".fromMap( " + name + " )")
				} else {
					dst.Code("null == " + name + " ? " + t.Name + ".fromMap({}) : " + t.Name + ".fromMap( " + name + " )")
				}
			} else {
				dst.Code("map[\"" + name + "\"]")
			}
		} else {
			switch expr.(*ast.Ident).Name {
			case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint64:
				if empty {
					dst.Code("null == " + name + " ? null : ( " + name + "  is num ?  " + name + " .toInt() : num.tryParse( " + name + " .toString())?.toInt())")
				} else {
					dst.Code("null == " + name + " ? 0 : ( " + name + "  is num ?  " + name + " .toInt() : num.tryParse( " + name + " .toString())?.toInt() ?? 0)")
				}
			case build.Int64, build.Uint32:
				if empty {
					dst.Code("null == " + name + " ? null : Int64.parseInt( " + name + " .toString())")
				} else {
					dst.Code("null == " + name + " ? Int64.ZERO : Int64.parseInt( " + name + " .toString())")
				}
			case build.Float, build.Double:
				if empty {
					dst.Code("null == " + name + " ? null : ( " + name + "  is num ?  " + name + " .toDouble() : num.tryParse( " + name + " .toString())?.toDouble())")
				} else {
					dst.Code("null == " + name + " ? 0 : ( " + name + "  is num ?  " + name + " .toDouble() : num.tryParse( " + name + " .toString())?.toDouble() ?? 0)")
				}
			case build.String:
				if empty {
					dst.Code("null == " + name + " ? null : ( " + name + "  is String ?  " + name + "  :  " + name + " .toString())")
				} else {
					dst.Code("null == " + name + " ? \"\" : ( " + name + "  is String ?  " + name + "  :  " + name + " .toString())")
				}
			case build.Date:
				if empty {
					dst.Code("null == " + name + " ? null :  " + name + "  is num ? DateTime.fromMillisecondsSinceEpoch( " + name + " .toInt()) : null == num.tryParse( " + name + " .toString()) ? null : DateTime.fromMillisecondsSinceEpoch(num.tryParse( " + name + " .toString())!.toInt())")
				} else {
					dst.Code("null == " + name + " ? DateTime.fromMillisecondsSinceEpoch(0) : DateTime.fromMillisecondsSinceEpoch( " + name + "  is num ?  " + name + " .toInt() : num.tryParse( " + name + " .toString())?.toInt() ?? 0)")
				}
			case build.Bool:
				if empty {
					dst.Code("null == " + name + " ? null : ( " + name + "  is bool ?  " + name + " : ( " + name + "  is num ? 0 !=  " + name + "  : (null == num.tryParse( " + name + " .toString()) ? null : 0 != num.tryParse( " + name + " .toString()))))")
				} else {
					dst.Code("null == " + name + " ? false:( " + name + "  is bool ?  " + name + " : 0 != ( " + name + "  is num ?  " + name + "  : num.tryParse( " + name + " .toString()) ?? 0))")
				}
			case build.Decimal:
				if empty {
					dst.Code("null == " + name + " ? null : Decimal.tryParse( " + name + " .toString())")
				} else {
					dst.Code("null == " + name + " ? Decimal.zero : (Decimal.tryParse( " + name + " .toString()) ?? Decimal.zero)")
				}
			default:
				dst.Code("map[\"" + name + "\"]")
			}
		}
	case *ast.ArrayType:
		t := expr.(*ast.ArrayType)
		empty = t.IsEmpty()
		if empty {
			dst.Code("null == " + name + " ? null : (temp! is! List ? null : (temp as List).map((temp) => ")
			b.printFormMap(dst, "temp", t.VType, data, empty)
			dst.Code(").toList())")
		} else {
			dst.Code("null == " + name + " ? <")
			b.printType(dst, t.VType, false)
			dst.Code(">[] : (temp! is! List ? <")
			b.printType(dst, t.VType, false)
			dst.Code(">[] : (temp as List).map((temp) => ")
			b.printFormMap(dst, "temp", t.VType, data, empty)
			dst.Code(").toList())")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		if empty {
			dst.Code("null == " + name + " ? null : (temp! is! Map ? null : (temp as Map).map((key,value) => MapEntry(")
			b.printFormMap(dst, "key", t.Key, data, empty)
			dst.Code(",")
			b.printFormMap(dst, "value", t.VType, data, empty)
			dst.Code(")))")
		} else {
			dst.Code("null == " + name + " ? <")
			b.printType(dst, t.Key, false)
			dst.Code(",")
			b.printType(dst, t.VType, false)
			dst.Code(">{}: (temp! is! Map ?<")
			b.printType(dst, t.Key, false)
			dst.Code(",")
			b.printType(dst, t.VType, false)
			dst.Code(">{} : (temp as Map).map((key,value) => MapEntry(")
			b.printFormMap(dst, "key", t.Key, data, empty)
			dst.Code(",")
			b.printFormMap(dst, "value", t.VType, data, empty)
			dst.Code(")))")
		}

	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printFormMap(dst, name, t.Type(), data, t.Empty)
	}
}

func (b *Builder) printToMap(dst *build.Writer, name string, expr ast.Expr, data *ast.DataType, empty bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				if empty {
					dst.Code(name + "?.value")
				} else {
					dst.Code(name + ".value")
				}
			} else if ast.Data == t.Obj.Kind {
				if empty {
					dst.Code(name + "?.toMap()")
				} else {
					dst.Code(name + ".toMap()")
				}
			} else {
				dst.Code(name)
			}
		} else {
			switch expr.(*ast.Ident).Name {
			case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32, build.Float, build.Double, build.String, build.Bool:
				dst.Code(name)
			case build.Uint64, build.Int64:
				if empty {
					dst.Code(name + "?.toString()")
				} else {
					dst.Code(name + ".toString()")
				}
			case build.Date:
				if empty {
					dst.Code(name + "?.millisecondsSinceEpoch")
				} else {
					dst.Code(name + ".millisecondsSinceEpoch")
				}
			case build.Decimal:
				if empty {
					dst.Code(name + "?.toString()")
				} else {
					dst.Code(name + ".toString()")
				}
			default:
				dst.Code(name)
			}
		}
	case *ast.ArrayType:
		t := expr.(*ast.ArrayType)
		empty = t.IsEmpty()
		if empty {
			dst.Code(name + "?.map((e) => ")
			b.printToMap(dst, "e", t.VType, data, empty)
			dst.Code(")?.toList()")
		} else {
			dst.Code(name + ".map((e) => ")
			b.printToMap(dst, "e", t.VType, data, empty)
			dst.Code(").toList()")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		if empty {
			dst.Code(name + "?.map((key,value) => MapEntry(")
			b.printToMap(dst, "key", t.Key, data, empty)
			dst.Code(",")
			b.printToMap(dst, "value", t.Key, data, empty)
			dst.Code(")")
		} else {
			dst.Code(name + ".map((key,value) => MapEntry(")
			b.printToMap(dst, "key", t.Key, data, empty)
			dst.Code(",")
			b.printToMap(dst, "value", t.Key, data, empty)
			dst.Code("))")
		}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printToMap(dst, name, t.Type(), data, t.Empty)
	}
}

func (b *Builder) printExtend(dst *build.Writer, extends []*ast.Ident, start bool) {
	for i, v := range extends {
		if 0 != i || start {
			dst.Code(", ")
		}
		dst.Code(build.StringToHumpName(v.Name))

	}
}
