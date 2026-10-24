package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printDataCode(dst *build.Writer, typ *ast.DataType) error {
	return b.printData(dst, typ)
}

func (b *Builder) printData(dst *build.Writer, typ *ast.DataType) error {
	// desc 单例（NewSyncDescriptor 包裹，延迟构建，支持循环引用）
	err := b.printDataDescriptor(dst, typ)
	if err != nil {
		return err
	}

	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("///" + typ.Doc.Text())
	}
	uName := className(typ.Name.Name)
	dst.Import("hbuf_ts", "type HData", "type Descriptor", "Marshal", "Unmarshal")
	dst.Code("export class " + uName + " implements HData")
	if nil != typ.Extends {
		b.printExtend(dst, typ.Extends, true)
	}
	dst.Code(" {\n")

	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Tab(1).Code("///" + field.Doc.Text())
		}
		dst.Tab(1).Code("")
		dst.Code(build.StringToFirstLower(field.Name.Name) + ": ")
		b.printType(dst, field.Type, false, false, false)
		dst.Code(" = ")
		b.printDefault(dst, field.Type, false)
		dst.Code(";\n\n")
		return nil
	})
	if err != nil {
		return err
	}

	dName := descName(typ.Name.Name)

	dst.Tab(1).Code("descriptors(): Descriptor {\n")
	dst.Tab(2).Code("return " + dName + "\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public static fromMap(_json: Record<string, any>): " + uName + "{\n")
	dst.Tab(2).Code("return ").Code(dName).Code(".fromJson(_json) as " + uName + "\n")
	dst.Tab(1).Code("}\n\n")

	dst.Code("\n")
	dst.Tab(1).Code("public toMap(): Record<string, any> {\n")
	dst.Tab(2).Code("return ").Code(dName).Code(".toJson(this) as Record<string, any> \n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public static fromData(_data: ArrayBuffer): " + uName + " {\n")
	dst.Tab(2).Code("const ret = new " + uName + "()\n")
	dst.Tab(2).Code("Unmarshal(new Uint8Array(_data), ret)\n")
	dst.Tab(2).Code("return ret\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public toData(): ArrayBuffer {\n")
	dst.Tab(2).Code("return Marshal(this).slice().buffer\n")
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

	dst.Tab(1).Code("public $getChangeField(val: ").Code(uName).Code(", ignore: string[] = []): {\n")
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
			dst.Code("this.").Code(lName).Code(" != val.").Code(lName).Code(") {\n")
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
				dst.Code(self).Code(name)
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
					// date 恒为 Date | null，判空后克隆
					dst.Code(self).Code(name).Code(" == null ? null : new Date(").Code(self).Code(name).Code("!.getTime())")
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
			dst.Code("(convertArray(").Code(self).Code(name).Code(", (item: any) => ")
			b.printCopy(dst, "", "item", t.VType, data, empty, false)
			dst.Code("))")
		} else {
			dst.Code(self).Code(name).Code(" == null ? [] : ")
			dst.Code("(convertArray(").Code(self).Code(name).Code(", (item: any) => ")
			b.printCopy(dst, "", "item", t.VType, data, empty, false)
			dst.Code("))!")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		if empty {
			dst.Code(self).Code(name).Code(" == null ? null : ")
		} else {
			dst.Code(self).Code(name).Code(" == null ? new Map() : ")
		}
		dst.Code("(new Map(Array.from(").Code(self).Code(name).Code(".entries()).map(([key, value]) => [")
		b.printCopy(dst, "", "key", t.Key, data, empty, true)
		dst.Code(", ")
		b.printCopy(dst, "", "value", t.VType, data, empty, false)
		dst.Code("])))")

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
			b.getPackage(dst, t, "", false, false, "", "", nil)
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
					dst.Code("null == ").Code(name).Code(" ? null : " + t.Name + ".fromMap(" + v + ")")
				} else {
					dst.Code("null == ").Code(name).Code(" ? " + t.Name + ".fromMap({}) : " + t.Name + ".fromMap(" + v + ")")
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
						dst.Code("null == ").Code(name).Code(" ? null : BigInt(" + v + " as string)")
					} else {
						dst.Code("null == ").Code(name).Code(" ? BigInt(0) : (BigInt(" + v + " as string) ?? BigInt(0))")
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
					// date 恒为 Date | null，JSON null 保持 null
					dst.Code("null == ").Code(name).Code(" ? null : new Date(Number(" + v + ").valueOf())")
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
				if empty {
					dst.Code("null == ").Code(name).Code(" ? null : " + v + ".toString()")
				} else {
					dst.Code("null == ").Code(name).Code(" ? \"\" : " + v + ".toString()")
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
			dst.Code("(convertArray(" + v + ", (item: any) => ")
			b.printFormMap(dst, "item", "item", t.VType, data, empty, false)
			dst.Code(")))")
		} else {
			dst.Code("null == ").Code(name).Code(" ? [] : (")
			dst.Code("!isArray(" + v + ") ? [] : ")
			dst.Code("(convertArray(" + v + ", (item: any) => ")
			b.printFormMap(dst, "item", "item", t.VType, data, empty, false)
			dst.Code("))!)")
		}
	case *ast.MapType:
		t := expr.(*ast.MapType)
		empty = t.IsEmpty()
		dst.Import("hbuf_ts", "isRecord")
		if empty {
			dst.Code("null == ").Code(name).Code(" ? null : (")
			dst.Code("!isRecord(" + v + ") ? null : ")
		} else {
			dst.Code("null == ").Code(name).Code(" ? new Map() : (")
			dst.Code("!isRecord(" + v + ") ? new Map() : ")
		}
		dst.Code("(new Map(Object.keys(" + v + ").map((_key) => [")
		b.printFormMap(dst, "_key", "_key", t.Key, data, empty, true)
		dst.Code(", ")
		b.printFormMap(dst, v+"[_key]", v+"[_key]", t.VType, data, empty, false)
		dst.Code("]))))")

	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printFormMap(dst, name, v, t.Type(), data, t.Empty, isRecordKey)
	}
}

func (b *Builder) printExtend(dst *build.Writer, extends []*ast.Extends, start bool) {
	for i, v := range extends {
		if 0 != i || start {
			dst.Code(", ")
		}

		b.getPackage(dst, v.Name, "", true, false, "", "", nil)
		dst.Code(build.StringToHumpName(v.Name.Name))
	}
}
