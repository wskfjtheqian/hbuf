package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printDataCode(dst *build.Writer, typ *ast.DataType) {
	dst.Import("hbuf_ts", "* as h")

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
		b.printType(dst, field.Type, false, false)
		dst.Code(" = ")
		b.printDefault(dst, field.Type, false)
		dst.Code(";\n\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("\tpublic static fromJson(json: Record<string, any>): " + build.StringToHumpName(typ.Name.Name) + "{\n")
	dst.Code("\t\tconst ret = new " + build.StringToHumpName(typ.Name.Name) + "()\n")
	dst.Code("\t\tlet temp:any\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("\t\tret." + build.StringToFirstLower(field.Name.Name) + " = ")
		jsonName := build.StringToUnderlineName(field.Name.Name)
		b.printFormMap(dst, "(temp = json[\""+jsonName+"\"])", "temp", field.Type, data, false, false)
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
		dst.Code("\": ")
		b.printToJson(dst, "this.", build.StringToFirstLower(field.Name.Name), field.Type, data, false, false)
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("\t\t};\n")
	dst.Code("\t}\n\n")

	dst.Code("\tpublic static fromData(data: BinaryData): " + build.StringToHumpName(typ.Name.Name) + " {\n")
	dst.Code("\t\tconst ret = new " + build.StringToHumpName(typ.Name.Name) + "()\n")
	dst.Code("\t\treturn ret\n")
	dst.Code("\t}\n\n")

	dst.Code("\tpublic toData(): BinaryData {\n")
	dst.Code("\t\treturn new ArrayBuffer(0)\n")
	dst.Code("\t}\n\n")

	dst.Code("\tpublic clone(): ").Code(build.StringToHumpName(typ.Name.Name)).Code(" {\n")
	dst.Code("\t\tconst ret = new ").Code(build.StringToHumpName(typ.Name.Name)).Code("()\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("\t\tret.").Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(" = ")
		b.printCopy(dst, "this.", build.StringToFirstLower(field.Name.Name), field.Type, data, true, false)
		dst.Code("\n")
		return nil
	})
	dst.Code("\t\treturn ret\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")
}

func (b *Builder) printCopy(dst *build.Writer, self, name string, expr ast.Expr, data *ast.DataType, empty bool, isRecordKey bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				dst.Code(self).Code(name)
			} else if ast.Data == t.Obj.Kind {
				if empty {
					dst.Code(self).Code(name).Code(" == null ? null : ").Code(self).Code(name).Code(".clone()")
				} else {
					dst.Code(self).Code(name).Code(".clone()")
				}
			} else {
				dst.Code(self).Code(name)
			}
		} else {
			switch build.BaseType(expr.(*ast.Ident).Name) {
			case build.Decimal:
				if isRecordKey {
					dst.Code(self).Code(name)
				} else {
					if empty {
						dst.Code(self).Code(name).Code(" == null ? null : new d.Decimal(").Code(self).Code(name).Code(")")
					} else {
						dst.Code("new d.Decimal(").Code(self).Code(name).Code(")")
					}
				}
			case build.Int64, build.Uint64:
				if isRecordKey {
					dst.Code(self).Code(name)
				} else {
					if empty {
						dst.Code(self).Code(name).Code(" == null ? null : Long.fromValue(").Code(self).Code(name).Code(")")
					} else {
						dst.Code("Long.fromValue(").Code(self).Code(name).Code(")")
					}
				}
			case build.Date:
				if isRecordKey {
					dst.Code(self).Code(name)
				} else {
					if empty {
						dst.Code(self).Code(name).Code(" == null ? null : new Date(").Code(self).Code(name).Code("!.getTime())")
					} else {
						dst.Code("new Date(").Code(self).Code(name).Code("!.getTime())")
					}
				}
			default:
				dst.Code(self).Code(name)
			}
		}
	case *ast.ArrayType:
		t := expr.(*ast.ArrayType)
		empty = t.IsEmpty()
		if empty {
			dst.Code(self).Code(name).Code(" == null ? null : ")
			dst.Code("(h.convertArray(").Code(self).Code(name).Code(", (item) => ")
			b.printCopy(dst, "", "item", t.VType, data, empty, false)
			dst.Code("))")
		} else {
			dst.Code(self).Code(name).Code(" == null ? [] : ")
			dst.Code("(h.convertArray(").Code(self).Code(name).Code(", (item) => ")
			b.printCopy(dst, "", "item", t.VType, data, empty, false)
			dst.Code("))!")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		if empty {
			dst.Code(self).Code(name).Code(" == null ? null : ")
			dst.Code("(h.convertRecord(").Code(self).Code(name).Code(", (key, value) => new h.RecordEntry(")
			b.printCopy(dst, "", "key", t.Key, data, empty, true)
			dst.Code(",")
			b.printCopy(dst, "", "value", t.VType, data, empty, false)
			dst.Code(")))")
		} else {
			dst.Code(self).Code(name).Code(" == null ? {} : ")
			dst.Code("(h.convertRecord(").Code(self).Code(name).Code(", (key, value) => new h.RecordEntry(")
			b.printCopy(dst, "", "key", t.Key, data, empty, true)
			dst.Code(",")
			b.printCopy(dst, "", "value", t.VType, data, empty, false)
			dst.Code(")))!")
		}

	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printCopy(dst, self, name, t.Type(), data, t.Empty, isRecordKey)
	}
}

func (b *Builder) printFormMap(dst *build.Writer, name string, v string, expr ast.Expr, data *ast.DataType, empty bool, isRecordKey bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			p := b.getPackage(dst, t, "")
			if ast.Enum == t.Obj.Kind {
				if isRecordKey {
					if empty {
						dst.Code("null == " + name + " ? null : Number(" + v + ").valueOf()")
					} else {
						dst.Code("null == " + name + " ? 0 : (Number(" + v + ").valueOf() || 0)")
					}
				} else {
					if empty {
						dst.Code("null == " + name + " ? null : " + p + "." + t.Name + ".valueOf(Number(" + v + ").valueOf())")
					} else {
						dst.Code("null == " + name + " ? " + p + "." + t.Name + ".valueOf(0) : " + p + "." + t.Name + ".valueOf(Number(" + v + ").valueOf())")
					}
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
				if isRecordKey {
					if empty {
						dst.Code("null == " + name + " ? null : " + v + ".toString()")
					} else {
						dst.Code("null == " + name + " ? \"\" : " + v + ".toString()")
					}
				} else {
					dst.Import("long", "Long")
					if empty {
						dst.Code("null == " + name + " ? null : Long.fromString(" + v + ".toString())")
					} else {
						dst.Code("null == " + name + " ? Long.ZERO : Long.fromString(" + v + ".toString())")
					}
				}
			case build.Float, build.Double:
				if empty {
					dst.Code("null == " + name + " ? null : Number(" + v + ").valueOf()")
				} else {
					dst.Code("null == " + name + " ? 0 : (Number(" + v + ").valueOf() || 0)")
				}
			case build.String:
				if empty {
					dst.Code("null == " + name + " ? null : " + v + ".toString()")
				} else {
					dst.Code("null == " + name + " ? \"\" : " + v + ".toString()")
				}
			case build.Date:
				if isRecordKey {
					if empty {
						dst.Code("null == " + name + " ? null : Number(" + v + ").valueOf()")
					} else {
						dst.Code("null == " + name + " ? 0 : (Number(" + v + ").valueOf() || 0)")
					}
				} else {
					if empty {
						dst.Code("null == " + name + " ? null : new Date(Number(" + v + ").valueOf())")
					} else {
						dst.Code("null == " + name + " ? new Date(0): new Date(Number(" + v + ").valueOf())")
					}
				}
			case build.Bool:
				if isRecordKey {
					if empty {
						dst.Code("null == " + name + " ? null : " + v + ".toString()")
					} else {
						dst.Code("null == " + name + " ? \"\" : " + v + ".toString()")
					}
				} else {
					if empty {
						dst.Code("null == " + name + " ? null : (\"true\" === " + v + " ? true : Boolean(" + v + "))")
					} else {
						dst.Code("null == " + name + " ? false : (\"true\" === " + v + " ? true : Boolean(" + v + "))")
					}
				}
			case build.Decimal:
				if isRecordKey {
					if empty {
						dst.Code("null == " + name + " ? null : " + v + ".toString()")
					} else {
						dst.Code("null == " + name + " ? \"\" : " + v + ".toString()")
					}
				} else {
					dst.Import("decimal.js", "* as d")
					if empty {
						dst.Code("null == " + name + " ? null : new d.Decimal(" + v + ")")
					} else {
						dst.Code("null == " + name + " ? new d.Decimal(0) : new d.Decimal(" + v + ") ")
					}
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
			dst.Code("!h.isArray(" + v + ") ? null : ")
			dst.Code("(h.convertArray(" + v + ", (item) => ")
			b.printFormMap(dst, "item", "item", t.VType, data, empty, false)
			dst.Code(")))")
		} else {
			dst.Code("null == " + name + " ? [] : (")
			dst.Code("!h.isArray(" + v + ") ? [] : ")
			dst.Code("(h.convertArray(" + v + ", (item) => ")
			b.printFormMap(dst, "item", "item", t.VType, data, empty, false)
			dst.Code("))!)")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		if empty {
			dst.Code("null == " + name + " ? null : (")
			dst.Code("h.isRecord(" + v + ") ? null : ")
			dst.Code("(h.convertRecord(" + v + ", (key, value) => new h.RecordEntry(")
			b.printFormMap(dst, "key", "key", t.Key, data, empty, true)
			dst.Code(",")
			b.printFormMap(dst, "value", "value", t.VType, data, empty, false)
			dst.Code("))))")
		} else {
			dst.Code("null == " + name + " ? {} : (")
			dst.Code("h.isRecord(" + v + ") ? {} : ")
			dst.Code("(h.convertRecord(" + v + ", (key, value) => new h.RecordEntry(")
			b.printFormMap(dst, "key", "key", t.Key, data, empty, true)
			dst.Code(",")
			b.printFormMap(dst, "value", "value", t.VType, data, empty, false)
			dst.Code(")))!)")
		}

	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printFormMap(dst, name, v, t.Type(), data, t.Empty, isRecordKey)
	}
}

func (b *Builder) printToJson(dst *build.Writer, key string, name string, expr ast.Expr, data *ast.DataType, empty bool, isRecordKey bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				if isRecordKey {
					dst.Code(name)
				} else {
					if empty {
						dst.Code(name + "?.value")
					} else {
						dst.Code(name + ".value")
					}
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
				if isRecordKey {
					dst.Code(name)
				} else {
					if empty {
						dst.Code(name + "?.toString()")
					} else {
						dst.Code(name + ".toString()")
					}
				}

			case build.Date:
				if isRecordKey {
					dst.Code(name)
				} else {
					if empty {
						dst.Code(name + "?.getTime()")
					} else {
						dst.Code(name + ".getTime()")
					}
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
			dst.Code("h.convertArray(" + key + name + ",(e) => ")
			b.printToJson(dst, "", "e", t.VType, data, empty, false)
			dst.Code(")")
		} else {
			dst.Code("h.convertArray(" + key + name + ",(e) => ")
			b.printToJson(dst, "", "e", t.VType, data, empty, false)
			dst.Code(")")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		if empty {
			dst.Code("h.convertRecord(" + key + name + ", (key, value) => new h.RecordEntry(")
			b.printToJson(dst, "", "key", t.Key, data, empty, true)
			dst.Code(",")
			b.printToJson(dst, "", "value", t.VType, data, empty, false)
			dst.Code("))")
		} else {
			dst.Code("h.convertRecord(" + key + name + ", (key, value) => new h.RecordEntry(")
			b.printToJson(dst, "", "key", t.Key, data, empty, true)
			dst.Code(",")
			b.printToJson(dst, "", "value", t.VType, data, empty, false)
			dst.Code("))")
		}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		dst.Code(key)
		b.printToJson(dst, "", name, t.Type(), data, t.Empty, isRecordKey)
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
