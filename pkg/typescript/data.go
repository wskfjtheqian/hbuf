package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printDataCode(dst *build.Writer, typ *ast.DataType) error {
	return b.printData(dst, typ)
}

func (b *Builder) printData(dst *build.Writer, typ *ast.DataType) error {
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("///" + typ.Doc.Text())
	}
	uName := build.StringToHumpName(typ.Name.Name)
	dst.Import("hbuf_ts", "Data")
	dst.Code("export class " + uName + " implements Data")
	if nil != typ.Extends {
		b.printExtend(dst, typ.Extends, true)
	}
	dst.Code(" {\n")
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Tab(1).Code("///" + field.Doc.Text())
		}
		dst.Tab(1).Code("")
		dst.Code(build.StringToFirstLower(field.Name.Name) + ": ")
		b.printType(dst, field.Type, false, false)
		dst.Code(" = ")
		b.printDefault(dst, field.Type, false)
		dst.Code(";\n\n")
		return nil
	})
	if err != nil {
		return err
	}

	dst.Tab(1).Code("public static fromMap(_json: Record<string, any>, _tag: string): " + uName + "{\n")
	dst.Tab(2).Code("const ret = new " + uName + "()\n")
	dst.Tab(2).Code("let _temp:any\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Tab(2).Code("ret." + build.StringToFirstLower(field.Name.Name) + " = ")
		jsonName := build.StringToUnderlineName(field.Name.Name)
		b.printFormMap(dst, "(_temp = _json[\""+jsonName+"\"])", "_temp", field.Type, data, false, false)
		dst.Code("\n")
		return nil
	})
	if err != nil {
		return err
	}

	dst.Tab(2).Code("return ret\n")
	dst.Tab(1).Code("}\n\n")

	dst.Code("\n")
	dst.Tab(1).Code("public toMap(_tag: string): Record<string, any> {\n")
	dst.Tab(2).Code("return {\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Tab(3).Code("\"" + build.StringToUnderlineName(field.Name.Name))
		dst.Code("\": ")
		b.printToMap(dst, "this.", build.StringToFirstLower(field.Name.Name), field.Type, data, false, false)
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return err
	}
	dst.Tab(2).Code("};\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public static fromData(_data: ArrayBuffer): " + uName + " {\n")
	dst.Tab(2).Code("const ret = new " + uName + "()\n")
	dst.Tab(2).Code("return ret\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public toData(): ArrayBuffer {\n")
	dst.Tab(2).Code("return new ArrayBuffer(0)\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public clone(): ").Code(uName).Code(" {\n")
	dst.Tab(2).Code("const ret = new ").Code(uName).Code("()\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Tab(2).Code("ret.").Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(" = ")
		b.printCopy(dst, "this.", build.StringToFirstLower(field.Name.Name), field.Type, data, true, false)
		dst.Code("\n")
		return nil
	})
	if err != nil {
		return err
	}
	dst.Tab(2).Code("return ret\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public $getChangeField(val: ").Code(uName).Code(", _tag: string, ignore: string[] = []): {\n")
	dst.Tab(2).Code("change : ").Code(uName).Code(",\n")
	dst.Tab(2).Code("fields: string[],\n")
	dst.Tab(1).Code("} {\n")
	dst.Tab(2).Code("const ret: ")
	dst.Code("{ change : ").Code(uName).Code(" , fields: string[] } = {\n")
	//dst.Tab(3).Code("change: this,\n")
	dst.Tab(3).Code("change: new ").Code(uName).Code("(),\n")
	dst.Tab(3).Code("fields: [],\n")
	dst.Tab(2).Code("}\n")

	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		lName := build.StringToFirstLower(field.Name.Name)
		dst.Tab(2).Code("if (ignore.includes(\"").Code(lName).Code("\") || ")
		if build.IsEnum(field.Type) {
			dst.Code("this.").Code(lName).Code("?.value != val.").Code(lName).Code("?.value) {\n")
		} else if build.IsDate(field.Type) {
			dst.Code("this.").Code(lName).Code("?.getTime() != val.").Code(lName).Code("?.getTime()) {\n")
		} else if build.IsDecimal(field.Type) {
			dst.Code("this.").Code(lName).Code("?.toString() != val.").Code(lName).Code("?.toString()) {\n")
		} else {
			dst.Code("this.").Code(lName).Code(" != val.").Code(lName).Code(") {\n")
		}
		dst.Tab(3).Code("ret.fields.push(\"").Code(field.Name.Name).Code("\")\n")
		dst.Tab(3).Code("ret.change.").Code(lName)
		dst.Code(" = this.").Code(lName).Code("\n")
		dst.Tab(2).Code("}\n")
		return nil
	})
	if err != nil {
		return err
	}

	dst.Tab(2).Code("return ret\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("private static $fieldMaps: Record<string, string> = {\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		lName := build.StringToFirstLower(field.Name.Name)
		dst.Tab(2).Code("\"").Code(lName).Code("\": \"").Code(field.Name.Name).Code("\",\n")
		return nil
	})
	if err != nil {
		return err
	}
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public static $convertField(fields: string[]): string[] {\n")
	dst.Tab(2).Code("const list = new Set<string>()\n")
	dst.Tab(2).Code("for (const field of fields) {\n")
	dst.Tab(3).Code("if (").Code(uName).Code(".$fieldMaps[field] != null) {\n")
	dst.Tab(4).Code("list.add(").Code(uName).Code(".$fieldMaps[field])\n")
	dst.Tab(3).Code("}\n")
	dst.Tab(2).Code("}\n")
	dst.Tab(2).Code("return Array.from(list)\n")
	dst.Tab(1).Code("}\n\n")
	dst.Code("}\n\n")
	return nil
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
						dst.Code(self).Code(name).Code(" == null ? null : new Decimal(").Code(self).Code(name).Code(")")
					} else {
						dst.Code("new Decimal(").Code(self).Code(name).Code(")")
					}
				}
			case build.Int64, build.Uint64:
				if isRecordKey {
					dst.Code(self).Code(name)
				} else {
					if empty {
						dst.Code(self).Code(name).Code(" == null ? null : BigInt(").Code(self).Code(name).Code(")")
					} else {
						dst.Code("BigInt(").Code(self).Code(name).Code(")")
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
			case build.Bytes:
				if isRecordKey {
					dst.Code(self).Code(name)
				} else {
					if empty {
						dst.Code(self).Code(name).Code(" == null ? null : new Uint8Array(").Code(self).Code(name).Code("!)")
					} else {
						dst.Code("new Uint8Array(").Code(self).Code(name).Code("!)")
					}
				}

			default:
				dst.Code(self).Code(name)
			}
		}
	case *ast.ArrayType:
		t := expr.(*ast.ArrayType)
		empty = t.IsEmpty()
		dst.Import("hbuf_ts", "convertArray")
		if empty {
			dst.Code(self).Code(name).Code(" == null ? null : ")
			dst.Code("(convertArray(").Code(self).Code(name).Code(", (item) => ")
			b.printCopy(dst, "", "item", t.VType, data, empty, false)
			dst.Code("))")
		} else {
			dst.Code(self).Code(name).Code(" == null ? [] : ")
			dst.Code("(convertArray(").Code(self).Code(name).Code(", (item) => ")
			b.printCopy(dst, "", "item", t.VType, data, empty, false)
			dst.Code("))!")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		dst.Import("hbuf_ts", "convertRecord", "RecordEntry")
		if empty {
			dst.Code(self).Code(name).Code(" == null ? null : ")
			dst.Code("(convertRecord(").Code(self).Code(name).Code(", (key, value) => new RecordEntry(")
			b.printCopy(dst, "", "key", t.Key, data, empty, true)
			dst.Code(",")
			b.printCopy(dst, "", "value", t.VType, data, empty, false)
			dst.Code(")))")
		} else {
			dst.Code(self).Code(name).Code(" == null ? {} : ")
			dst.Code("(convertRecord(").Code(self).Code(name).Code(", (key, value) => new RecordEntry(")
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
			b.getPackage(dst, t, "", false, false)
			if ast.Enum == t.Obj.Kind {
				if isRecordKey {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : Number(" + v + ").valueOf()")
					} else {
						dst.Code("null == ").Code(name).Code(" ? 0 : (Number(" + v + ").valueOf() ?? 0)")
					}
				} else {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : " + t.Name + ".valueOf(Number(" + v + ").valueOf())")
					} else {
						dst.Code("null == ").Code(name).Code(" ? " + t.Name + ".valueOf(0) : " + t.Name + ".valueOf(Number(" + v + ").valueOf())")
					}
				}
			} else if ast.Data == t.Obj.Kind {
				if empty {
					dst.Code("null == ").Code(name).Code(" ? null : " + t.Name + ".fromMap(" + v + ", _tag)")
				} else {
					dst.Code("null == ").Code(name).Code(" ? " + t.Name + ".fromMap({}, _tag) : " + t.Name + ".fromMap(" + v + ", _tag)")
				}
			} else {
				dst.Code("map[\"").Code(name).Code("\"]")
			}
		} else {
			switch build.BaseType(expr.(*ast.Ident).Name) {
			case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32:
				if empty {
					dst.Code("null == ").Code(name).Code(" ? null : Number(" + v + ").valueOf()")
				} else {
					dst.Code("null == ").Code(name).Code(" ? 0 : (Number(" + v + ").valueOf() ?? 0)")
				}
			case build.Int64, build.Uint64:
				if isRecordKey {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : " + v + ".toString()")
					} else {
						dst.Code("null == ").Code(name).Code(" ? \"\" : " + v + ".toString()")
					}
				} else {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : BigInt(" + v + " as string).valueOf()")
					} else {
						dst.Code("null == ").Code(name).Code(" ? BigInt(0) : (BigInt(" + v + " as string).valueOf() ?? BigInt(0))")
					}
				}
			case build.Float, build.Double:
				if empty {
					dst.Code("null == ").Code(name).Code(" ? null : Number(" + v + ").valueOf()")
				} else {
					dst.Code("null == ").Code(name).Code(" ? 0 : (Number(" + v + ").valueOf() ?? 0)")
				}
			case build.String:
				if empty {
					dst.Code("null == ").Code(name).Code(" ? null : " + v + ".toString()")
				} else {
					dst.Code("null == ").Code(name).Code(" ? \"\" : " + v + ".toString()")
				}
			case build.Date:
				if isRecordKey {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : Number(" + v + ").valueOf()")
					} else {
						dst.Code("null == ").Code(name).Code(" ? 0 : (Number(" + v + ").valueOf() ?? 0)")
					}
				} else {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : new Date(Number(" + v + ").valueOf())")
					} else {
						dst.Code("null == ").Code(name).Code(" ? new Date(0): new Date(Number(" + v + ").valueOf())")
					}
				}
			case build.Bool:
				if isRecordKey {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : " + v + ".toString()")
					} else {
						dst.Code("null == ").Code(name).Code(" ? \"\" : " + v + ".toString()")
					}
				} else {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : (\"true\" === " + v + " ? true : Boolean(" + v + "))")
					} else {
						dst.Code("null == ").Code(name).Code(" ? false : (\"true\" === " + v + " ? true : Boolean(" + v + "))")
					}
				}
			case build.Decimal:
				if isRecordKey {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : " + v + ".toString()")
					} else {
						dst.Code("null == ").Code(name).Code(" ? \"\" : " + v + ".toString()")
					}
				} else {
					dst.Import("decimal.js", "Decimal")
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : new Decimal(" + v + " as string )")
					} else {
						dst.Code("null == ").Code(name).Code(" ? new Decimal(0) : new Decimal(" + v + " as string ) ")
					}
				}
			case build.Bytes:
				if isRecordKey {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : " + v + ".toString()")
					} else {
						dst.Code("null == ").Code(name).Code(" ? \"\" : " + v + ".toString()")
					}
				} else {
					if empty {
						dst.Code("null == ").Code(name).Code(" ? null : new Uint8Array([...atob(").Code(name).Code(")].map(c => c.charCodeAt(0)))")
					} else {
						dst.Code("null == ").Code(name).Code(" ? new Uint8Array() : new Uint8Array([...atob(").Code(name).Code(")].map(c => c.charCodeAt(0)))")
					}
				}
			default:
				dst.Code("map[\"").Code(name).Code("\"]")
			}
		}
	case *ast.ArrayType:
		t := expr.(*ast.ArrayType)
		empty = t.IsEmpty()
		dst.Import("hbuf_ts", "convertArray", "isArray")
		if empty {
			dst.Code("null == ").Code(name).Code(" ? null : (")
			dst.Code("!isArray(" + v + ") ? null : ")
			dst.Code("(convertArray(" + v + ", (item) => ")
			b.printFormMap(dst, "item", "item", t.VType, data, empty, false)
			dst.Code(")))")
		} else {
			dst.Code("null == ").Code(name).Code(" ? [] : (")
			dst.Code("!isArray(" + v + ") ? [] : ")
			dst.Code("(convertArray(" + v + ", (item) => ")
			b.printFormMap(dst, "item", "item", t.VType, data, empty, false)
			dst.Code("))!)")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		dst.Import("hbuf_ts", "convertArray", "isRecord", "RecordEntry")
		if empty {
			dst.Code("null == ").Code(name).Code(" ? null : (")
			dst.Code("!isRecord(" + v + ") ? null : ")
			dst.Code("(convertRecord(" + v + ", (key, value) => new RecordEntry(")
			b.printFormMap(dst, "key", "key", t.Key, data, empty, true)
			dst.Code(",")
			b.printFormMap(dst, "value", "value", t.VType, data, empty, false)
			dst.Code("))))")
		} else {
			dst.Code("null == ").Code(name).Code(" ? {} : (")
			dst.Code("!isRecord(" + v + ") ? {} : ")
			dst.Code("(convertRecord(" + v + ", (key, value) => new RecordEntry(")
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

func (b *Builder) printToMap(dst *build.Writer, key string, name string, expr ast.Expr, data *ast.DataType, empty bool, isRecordKey bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				dst.Code(key)
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
				dst.Code(key)
				if empty {
					dst.Code(name + "?.toMap(_tag)")
				} else {
					dst.Code(name + ".toMap(_tag)")
				}
			} else {
				dst.Code(name)
			}
		} else {
			switch build.BaseType(expr.(*ast.Ident).Name) {
			case build.Int8, build.Int16, build.Int32, build.Uint32, build.Uint8, build.Uint16, build.Float, build.Double, build.String, build.Bool:
				dst.Code(key)
				dst.Code(name)
			case build.Uint64, build.Int64:
				dst.Code(key)
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
				dst.Code(key)
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
				dst.Code(key)
				if empty {
					dst.Code(name + "?.toString()")
				} else {
					dst.Code(name + ".toString()")
				}

			case build.Bytes:
				if empty {
					dst.Code(key).Code(name).Code("==null?null:")
				}
				dst.Code("btoa(String.fromCharCode(...").Code(key).Code(name).Code("))")
			default:
				dst.Code(name)
			}
		}
	case *ast.ArrayType:
		t := expr.(*ast.ArrayType)
		empty = t.IsEmpty()
		dst.Import("hbuf_ts", "convertArray")
		if empty {
			dst.Code("convertArray(" + key + name + ",(e) => ")
			b.printToMap(dst, "", "e", t.VType, data, empty, false)
			dst.Code(")")
		} else {
			dst.Code("convertArray(" + key + name + ",(e) => ")
			b.printToMap(dst, "", "e", t.VType, data, empty, false)
			dst.Code(")")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		dst.Import("hbuf_ts", "convertRecord", "RecordEntry")
		if empty {
			dst.Code("convertRecord(" + key + name + ", (key, value) => new RecordEntry(")
			b.printToMap(dst, "", "key", t.Key, data, empty, true)
			dst.Code(",")
			b.printToMap(dst, "", "value", t.VType, data, empty, false)
			dst.Code("))")
		} else {
			dst.Code("convertRecord(" + key + name + ", (key, value) => new RecordEntry(")
			b.printToMap(dst, "", "key", t.Key, data, empty, true)
			dst.Code(",")
			b.printToMap(dst, "", "value", t.VType, data, empty, false)
			dst.Code("))")
		}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printToMap(dst, key, name, t.Type(), data, t.Empty, isRecordKey)
	}
}

func (b *Builder) printExtend(dst *build.Writer, extends []*ast.Extends, start bool) {
	for i, v := range extends {
		if 0 != i || start {
			dst.Code(", ")
		}

		b.getPackage(dst, v.Name, "", true, false)
		dst.Code(build.StringToHumpName(v.Name.Name))
	}
}
