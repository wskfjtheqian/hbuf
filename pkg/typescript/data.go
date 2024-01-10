package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printDataCode(dst *build.Writer, typ *ast.DataType) {
	dst.Import("hbuf_ts", "h")

	b.printData(dst, typ)
}

func (b *Builder) printData(dst *build.Writer, typ *ast.DataType) {
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("///" + typ.Doc.Text())
	}
	dst.Code("export class " + build.StringToHumpName(typ.Name.Name) + " implements h.Data")
	if nil != typ.Extends {
		b.printExtend(dst, typ.Extends, true)
	}
	dst.Code(" {\n")
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("\t///" + field.Doc.Text())
		}
		dst.Code("\t")
		dst.Code(build.StringToFirstLower(field.Name.Name) + ": ")
		b.printType(dst, field.Type, false)
		dst.Code(" = ")
		b.printDefault(dst, field.Type, false)
		dst.Code(";\n\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("\tpublic static fromJson(json: Record<string, any>): " + build.StringToHumpName(typ.Name.Name) + "{\n")
	dst.Code("\t\tlet ret = new " + build.StringToHumpName(typ.Name.Name) + "()\n")
	dst.Code("\t\tlet temp:any\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("\t\tret." + build.StringToFirstLower(field.Name.Name) + " = ")
		jsonName := build.StringToUnderlineName(field.Name.Name)
		b.printFormMap(dst, "(temp = json[\""+jsonName+"\"])", "temp", field.Type, data, false)
		dst.Code("\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("\t\treturn ret\n")
	dst.Code("\t}\n\n")

	dst.Code("\n")
	dst.Code("\tpublic toJson(): Record<string, any> {\n")
	dst.Code("\t\treturn {\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("\t\t\t\"" + build.StringToUnderlineName(field.Name.Name))
		dst.Code("\": this.")
		b.printToJson(dst, build.StringToFirstLower(field.Name.Name), field.Type, data, false)
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("\t\t};\n")
	dst.Code("\t}\n\n")

	dst.Code("\tpublic static fromData(data: BinaryData): " + build.StringToHumpName(typ.Name.Name) + " {\n")
	dst.Code("\t\tlet ret = new " + build.StringToHumpName(typ.Name.Name) + "()\n")
	dst.Code("\t\treturn ret\n")
	dst.Code("\t}\n\n")

	dst.Code("\tpublic toData(): BinaryData {\n")
	dst.Code("\t\treturn new ArrayBuffer(0)\n")
	dst.Code("\t}\n\n")

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
			dst.Code(")")
		} else {
			dst.Code(name + ".map((temp) => ")
			b.printCopy(dst, "temp", t.VType, data, empty)
			dst.Code(")")
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

func (b *Builder) printFormMap(dst *build.Writer, name string, v string, expr ast.Expr, data *ast.DataType, empty bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			p := b.getPackage(dst, t, "")
			if ast.Enum == t.Obj.Kind {
				if empty {
					dst.Code("null == " + name + " ? null : " + p + "." + t.Name + ".valueOf(Number(" + v + ").valueOf())")
				} else {
					dst.Code("null == " + name + " ? " + p + "." + t.Name + ".valueOf(0) : " + p + "." + t.Name + ".valueOf(Number(" + v + ").valueOf())")
				}
			} else if ast.Data == t.Obj.Kind {
				if empty {
					dst.Code("null == " + name + " ? null : " + p + "." + t.Name + ".fromJson(" + v + ")")
				} else {
					dst.Code("null == " + name + " ? " + p + "." + t.Name + ".fromJson({}) : " + p + "." + t.Name + ".fromJson(" + v + ")")
				}
			} else {
				dst.Code("map[\"" + name + "\"]")
			}
		} else {
			switch build.BaseType(expr.(*ast.Ident).Name) {
			case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32:
				if empty {
					dst.Code("null == " + name + " ? null : Number(" + v + ").valueOf()")
				} else {
					dst.Code("null == " + name + " ? 0 : (Number(" + v + ").valueOf() || 0)")
				}
			case build.Int64, build.Uint64:
				dst.Import("long", "Long")
				if empty {
					dst.Code("null == " + name + " ? null : Long.fromString(" + v + ")")
				} else {
					dst.Code("null == " + name + " ? Long.ZERO : Long.fromString(" + v + ")")
				}
			case build.Float, build.Double:
				if empty {
					dst.Code("null == " + name + " ? null : Number(" + v + ").valueOf()")
				} else {
					dst.Code("null == " + name + " ? 0 : (Number(" + v + ").valueOf() || 0)")
				}
			case build.String:
				if empty {
					dst.Code("null == " + name + " ? null : (\"\" + " + v + ")")
				} else {
					dst.Code("null == " + name + " ? \"\" : (\"\" + " + v + ")")
				}
			case build.Date:
				if empty {
					dst.Code("null == " + name + " ? null : new Date(" + v + ")")
				} else {
					dst.Code("null == " + name + " ? new Date(0): new Date(" + v + ")")
				}
			case build.Bool:
				if empty {
					dst.Code("null == " + name + " ? null : (\"true\" === " + v + " ? true : Boolean(" + v + "))")
				} else {
					dst.Code("null == " + name + " ? false : (\"true\" === " + v + " ? true : Boolean(" + v + "))")
				}
			case build.Decimal:
				dst.Import("decimal.js", "* as d")
				if empty {
					dst.Code("null == " + name + " ? null : new d.Decimal(" + v + ")")
				} else {
					dst.Code("null == " + name + " ? new d.Decimal(0) : new d.Decimal(" + v + ") ")
				}
			default:
				dst.Code("map[\"" + name + "\"]")
			}
		}
	case *ast.ArrayType:
		t := expr.(*ast.ArrayType)
		empty = t.IsEmpty()
		if empty {
			dst.Code("null == " + name + " ? null : (")
			dst.Code("Object.is(" + v + ", \"array\") ? null : ")
			dst.Code("(h.arrayMap(" + v + ", (item) => ")
			b.printFormMap(dst, "item", "item", t.VType, data, empty)
			dst.Code(")))")
		} else {
			dst.Code("null == " + name + " ? [] : (")
			dst.Code("Object.is(" + v + ", \"array\") ? [] : ")
			dst.Code("(h.arrayMap(" + v + ", (item) => ")
			b.printFormMap(dst, "item", "item", t.VType, data, empty)
			dst.Code(")))")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		dst.Code("{}")
		//if empty {
		//	dst.Code("null == " + name + " ? null : (" + v + "! is! Map ? null : (temp as Map).map((key,value) => MapEntry(")
		//	b.printFormMap(dst, "key", "key", t.Key, data, empty)
		//	dst.Code(",")
		//	b.printFormMap(dst, "value", "value", t.VType, data, empty)
		//	dst.Code(")))")
		//} else {
		//	dst.Code("null == " + name + " ? {}: (temp! is! Map ?<")
		//	b.printType(dst, t.Key, false)
		//	dst.Code(",")
		//	b.printType(dst, t.VType, false)
		//	dst.Code(">{} : (temp as Map).map((key,value) => MapEntry(")
		//	b.printFormMap(dst, "key", "key", t.Key, data, empty)
		//	dst.Code(",")
		//	b.printFormMap(dst, "value", "value", t.VType, data, empty)
		//	dst.Code(")))")
		//}

	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printFormMap(dst, name, v, t.Type(), data, t.Empty)
	}
}

func (b *Builder) printToJson(dst *build.Writer, name string, expr ast.Expr, data *ast.DataType, empty bool) {
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
					dst.Code(name + "?.toJson()")
				} else {
					dst.Code(name + ".toJson()")
				}
			} else {
				dst.Code(name)
			}
		} else {
			switch build.BaseType(expr.(*ast.Ident).Name) {
			case build.Int8, build.Int16, build.Int32, build.Uint32, build.Uint8, build.Uint16, build.Float, build.Double, build.String, build.Bool:
				dst.Code(name)
			case build.Uint64, build.Int64:
				if empty {
					dst.Code(name + "?.toString()")
				} else {
					dst.Code(name + ".toString()")
				}
			case build.Date:
				if empty {
					dst.Code(name + "?.getTime()")
				} else {
					dst.Code(name + ".getTime()")
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
			b.printToJson(dst, "e", t.VType, data, empty)
			dst.Code(")")
		} else {
			dst.Code(name + ".map((e) => ")
			b.printToJson(dst, "e", t.VType, data, empty)
			dst.Code(")")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		if empty {
			dst.Code(name + "?.map((key,value) => MapEntry(")
			b.printToJson(dst, "key", t.Key, data, empty)
			dst.Code(",")
			b.printToJson(dst, "value", t.Key, data, empty)
			dst.Code(")")
		} else {
			dst.Code(name + ".map((key,value) => MapEntry(")
			b.printToJson(dst, "key", t.Key, data, empty)
			dst.Code(",")
			b.printToJson(dst, "value", t.Key, data, empty)
			dst.Code("))")
		}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printToJson(dst, name, t.Type(), data, t.Empty)
	}
}

func (b *Builder) printExtend(dst *build.Writer, extends []*ast.Extends, start bool) {
	for i, v := range extends {
		if 0 != i || start {
			dst.Code(", ")
		}

		dst.Code(b.getPackage(dst, v.Name, ""))
		dst.Code(".")
		dst.Code(build.StringToHumpName(v.Name.Name))

	}
}
