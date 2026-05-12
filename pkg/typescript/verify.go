package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
	"time"
)

func (b *Builder) printVerifyCode(dst *build.Writer, data *ast.DataType) error {
	err := b.printVerifyFieldCode(dst, data)
	if err != nil {
		return err
	}

	verify, err := build.GetVerify(data.Tags, dst.File, b.GetDataType)
	if err != nil {
		return err
	}
	if nil == verify {
		return nil
	}

	return nil
}

func (b *Builder) printVerifyFieldCode(dst *build.Writer, data *ast.DataType) error {
	dName := build.StringToHumpName(data.Name.Name)
	err := build.EnumField(data, func(field *ast.Field, data *ast.DataType) error {
		fName := build.StringToHumpName(field.Name.Name)
		verify, err := build.GetVerify(field.Tags, dst.File, b.GetDataType)
		if err != nil {
			return err
		}
		if nil == verify {
			return nil
		}
		dst.Import("element-plus", "type {LocaleContext}", 0)
		dst.Code("export const verify").Code(dName).Code("_").Code(fName).Code(" = (locale: LocaleContext) => (rule: any, value: any, callback: any): any => {\n")
		dst.Tab(1).Code("value = '' + value\n")
		isNull := build.IsNil(field.Type)
		for i, val := range verify.GetFormat() {
			f := build.GetFormat(val.Item.Tags)
			if nil == f {
				continue
			}
			pName := b.getPackage(dst, val.Enum.Name, "enum", false)
			if isNull && 0 == i {
				dst.Tab(1).Code("if (value == '' || value == 'null' || value == 'undefined') {\n")
				if !f.Null {
					b.printVerifyError(dst, pName, val)
				} else {
					dst.Tab(2).Code("return callback();\n")
				}
				dst.Tab(1).Code("}\n")
			}
			if build.IsEnum(field.Type) {
				//dst.Tab(1).Code("if 0 < len(i.Get" + fName + "().ToName()) {\n")
				//dst.Tab(2).Code("return &hbuf.Result{Code: int(" + pack + "), Msg: " + pack + ".ToName()}\n")
				//dst.Tab(1).Code("}\n")
			} else if build.IsMap(field.Type) {

			} else if build.IsArray(field.Type) {
				for _, item := range f.Len {
					dst.Tab(1).Code("if ").Code("(!(value.length ").Code(b.getOperator(item.Op)).Code(" ").Code(item.Val).Code(")) {\n")
					b.printVerifyError(dst.Tab(1), pName, val)
					dst.Tab(1).Code("}\n")
				}
			} else {
				t := build.GetBaseType(field.Type)
				switch t {
				case build.Int8:
					b.verifyNum(dst, pName, val, f, "-?[0-9]\\\\d*", field.Type, "–128", "127")
				case build.Int16:
					b.verifyNum(dst, pName, val, f, "-?[0-9]\\\\d*", field.Type, "-32768", "32767")
				case build.Int32:
					b.verifyNum(dst, pName, val, f, "-?[0-9]\\\\d*", field.Type, "-2147483648", "2147483647")
				case build.Uint8:
					b.verifyNum(dst, pName, val, f, "[0-9]\\\\d*", field.Type, "0", "255")
				case build.Uint16:
					b.verifyNum(dst, pName, val, f, "[0-9]\\\\d*", field.Type, "0", "65535")
				case build.Uint32:
					b.verifyNum(dst, pName, val, f, "[0-9]\\\\d*", field.Type, "0", "4294967295")
				case build.Float, build.Double:
					b.verifyNum(dst, pName, val, f, "^[+-]?\\\\d+(\\\\.\\\\d+)?$", field.Type, "", "")
				case build.Int64:
					b.verifyNum(dst, pName, val, f, "[0-9]\\\\d*", field.Type, "-9223372036854775808", "9223372036854775808")
				case build.Uint64:
					b.verifyNum(dst, pName, val, f, "[0-9]\\\\d*", field.Type, "0", "18446744073709551615615")
				case build.Decimal:
					b.verifyNum(dst, pName, val, f, "^[+-]?\\\\d+(\\\\.\\\\d+)?$", field.Type, "", "")
				case build.Date:
					dst.Tab(1).Code("const val = DateTime.tryParse(value!);\n")
					dst.Tab(1).Code("if (null == val) {\n")
					b.printVerifyError(dst, pName, val)
					dst.Tab(1).Code("}\n")

					for _, item := range f.Val {
						parse, err := time.Parse("2006-01-02T15:04:05Z", item.Val)
						if err != nil {
							return err
						}

						dst.Tab(1).Code("if ").Code("(!(val.millisecondsSinceEpoch ").Code(b.getOperator(item.Op)).Code(" ").Code(strconv.FormatInt(parse.UnixMilli(), 10)).Code(")) {\n")
						b.printVerifyError(dst.Tab(1), pName, val)
						dst.Tab(1).Code("}\n")
					}
				case build.String:
					for _, item := range f.Len {
						dst.Tab(1).Code("if ").Code("(!(value.length ").Code(b.getOperator(item.Op)).Code(" ").Code(item.Val).Code(")) {\n")
						b.printVerifyError(dst.Tab(1), pName, val)
						dst.Tab(1).Code("}\n")
					}

					for _, item := range f.Val {
						if item.Op == build.OperatorMatch || item.Op == build.OperatorNotMatch {
							dst.Tab(1).Code("if (")
							if item.Op == build.OperatorMatch {
								dst.Code("!new RegExp(\"").Code(item.Val).Code("\").test(value!)")
							} else {
								dst.Code("new RegExp(\"").Code(item.Val).Code("\").test(value!)")
							}
							dst.Code(") {\n")
							b.printVerifyError(dst, pName, val)
							dst.Tab(1).Code("}\n")
						} else {
							dst.Tab(1).Code("if ").Code("(!(value").Code(b.getOperator(item.Op)).Code(" ").Code(item.Val).Code(")) {\n")
							b.printVerifyError(dst, pName, val)
							dst.Tab(1).Code("}\n")
						}
					}
				}
			}
		}
		dst.Tab(1).Code("return callback();\n")
		dst.Code("}\n\n")
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) verifyNum(dst *build.Writer, pName string, val *build.VerifyEnum, f *build.Format, reg string, field ast.Type, min, max string) {
	dst.Tab(1).Code("if (!new RegExp(\"" + reg + "\").test(value")
	if build.IsNil(field) {
		dst.Code("!")
	}

	dst.Code(")) {\n")
	b.printVerifyError(dst, pName, val)
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("try {\n")
	dst.Import("decimal.js", "* as d", 0)
	dst.Tab(2).Code("const val = new d.Decimal(value!);\n")

	if 0 < len(min) {
		dst.Tab(2).Code("if (val.lessThan(new d.Decimal(\"").Code(min).Code("\")) || val.greaterThan(new d.Decimal(\"").Code(max).Code("\"))) {\n")
		dst.Tab(1)
		b.printVerifyError(dst, pName, val)
		dst.Tab(2).Code("}\n")
	}

	for _, item := range f.Val {
		dst.Tab(2).Code("if ").Code("(")

		if build.OperatorGt == item.Op {
			dst.Code("!val.").Code("gt")
		} else if build.OperatorLt == item.Op {
			dst.Code("!val.").Code("lt")
		} else if build.OperatorGte == item.Op {
			dst.Code("!val.").Code("gte")
		} else if build.OperatorLte == item.Op {
			dst.Code("!val.").Code("lte")
		} else if build.OperatorEq == item.Op {
			dst.Code("!val.").Code("eq")
		} else if build.OperatorNeq == item.Op {
			dst.Code("val.").Code("eq")
		} else {
			//TODO
		}
		dst.Code("(").Code("new d.Decimal(").Code(item.Val).Code("))) {\n")

		b.printVerifyError(dst.Tab(1), pName, val)
		dst.Tab(2).Code("}\n")
	}

	dst.Tab(1).Code("} catch {\n")
	b.printVerifyError(dst, pName, val)
	dst.Tab(1).Code("}\n")
}

func (b *Builder) printVerifyError(dst *build.Writer, pName string, val *build.VerifyEnum) {
	dst.Tab(2).Code("return callback(new Error(locale.t(").Code(pName).Code(".")
	dst.Code(build.StringToHumpName(val.Enum.Name.Name)).Code(".")
	dst.Code(build.StringToAllUpper(val.Item.Name.Name)).Code(".toString())))\n")
}

func (b *Builder) getOperator(op build.Operator) string {
	if build.OperatorGt == op {
		return ">"
	} else if build.OperatorLt == op {
		return "<"
	} else if build.OperatorGte == op {
		return ">="
	} else if build.OperatorLte == op {
		return "<="
	} else if build.OperatorEq == op {
		return "=="
	} else if build.OperatorNeq == op {
		return "!="
	} else {
		//TODO 处理错误
	}
	return ""
}
