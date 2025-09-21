package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
	"strings"
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
		dst.Import("element-plus", "type {LocaleContext}")
		dst.Code("export const verify").Code(dName).Code("_").Code(fName).Code(" = (locale: LocaleContext) => (rule: any, value: any, callback: any): any => {\n")
		dst.Tab(1).Code("value = '' + value\n")
		isNull := build.IsNil(field.Type)
		for i, val := range verify.GetFormat() {
			f := build.GetFormat(val.Item.Tags)
			if nil == f {
				continue
			}
			pName := b.getPackage(dst, val.Enum.Name, "enum")
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
					b.verifyNum(dst, pName, val, f, "-?[0-9]\\\\d*.\\\\d*|0.\\\\d*[0-9]\\\\d*", field.Type, "", "")
				case build.Int64:
					b.verifyNum(dst, pName, val, f, "[0-9]\\\\d*", field.Type, "-9223372036854775808", "9223372036854775808")
				case build.Uint64:
					b.verifyNum(dst, pName, val, f, "[0-9]\\\\d*", field.Type, "0", "18446744073709551615615")
				case build.Date:
					dst.Tab(1).Code("DateTime? val = DateTime.tryParse(value!);\n")
					dst.Tab(1).Code("if (null == val) {\n")
					b.printVerifyError(dst, pName, val)
					dst.Tab(1).Code("}\n")
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Tab(1).Code("if (")
						if 0 < len(f.Min) {
							parse, err := time.Parse("2006-01-02T15:04:05Z", f.Min)
							if err != nil {
								return err
							}
							dst.Code(strconv.FormatInt(parse.UnixMilli(), 10) + " > val.millisecondsSinceEpoch")
						}
						if 0 < len(f.Max) {
							if 0 < len(f.Min) {
								dst.Code(" || ")
							}
							parse, err := time.Parse("2006-01-02T15:04:05Z", f.Max)
							if err != nil {
								return err
							}
							dst.Code(strconv.FormatInt(parse.UnixMilli(), 10) + " < val.millisecondsSinceEpoch")
						}
						dst.Code(") {\n")
						b.printVerifyError(dst, pName, val)
						b.printVerifyError(dst, pName, val)
						dst.Tab(1).Code("}\n")
					}
				case build.Decimal:
					if 0 < len(f.Reg) {
						dst.Tab(1).Code("if (!new RegExp(\"" + strings.ReplaceAll(f.Reg, "$", "\\$") + "\").test(value!)) {\n")
						b.printVerifyError(dst, pName, val)
						dst.Tab(1).Code("}\n")
					}
					if i == len(verify.GetFormat())-1 {
						dst.Tab(1).Code("if (!new RegExp(\"-?[0-9]\\\\d*.\\\\d*|0.\\\\d*[0-9]\\\\d*\").test(value")
						if build.IsNil(field.Type) {
							dst.Code("!")
						}
						dst.Code(")) {\n")
						b.printVerifyError(dst, pName, val)
						dst.Tab(1).Code("}\n")

						dst.Import("decimal.js", "* as d")
						dst.Tab(1).Code("Decimal? val = Decimal.tryParse(value!);\n")
						dst.Tab(1).Code("if (null == val) {\n")
						b.printVerifyError(dst, pName, val)
						dst.Tab(1).Code("}\n")
						if 0 < len(f.Min) || 0 < len(f.Max) {
							dst.Tab(1).Code("if (")
							if 0 < len(f.Min) {
								dst.Code("1 == val.compareTo(Decimal.fromInt(" + f.Min + "))")
							}
							if 0 < len(f.Max) {
								if 0 < len(f.Min) {
									dst.Code(" || ")
								}
								dst.Code("-1 == val.compareTo(Decimal.fromInt(" + f.Max + "))")
							}
							dst.Code(") {\n")
							b.printVerifyError(dst, pName, val)
							dst.Tab(1).Code("}\n")
						}
					}
				case build.String:
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Tab(1).Code("if (")
						if 0 < len(f.Min) {
							dst.Code(f.Min + " > (value?.length ?? 0)")
						}
						if 0 < len(f.Max) {
							if 0 < len(f.Min) {
								dst.Code(" || ")
							}
							dst.Code(f.Max + " < (value?.length ?? 0)")
						}
						dst.Code(") {\n")
						b.printVerifyError(dst, pName, val)
						dst.Tab(1).Code("}\n")
					}
					if len(f.Reg) > 0 {
						dst.Tab(1).Code("if (!new RegExp(\"" + f.Reg + "\").test(value!)) {\n")
						b.printVerifyError(dst, pName, val)
						dst.Tab(1).Code("}\n")
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
	dst.Import("decimal.js", "* as d")
	dst.Tab(2).Code("const val = new d.Decimal(value!);\n")

	if 0 < len(min) {
		dst.Tab(2).Code("if (val.lessThan(new d.Decimal(\"").Code(min).Code("\")) || val.greaterThan(new d.Decimal(\"").Code(max).Code("\"))) {\n")
		dst.Tab(1)
		b.printVerifyError(dst, pName, val)
		dst.Tab(2).Code("}\n")
	}

	if 0 < len(f.Min) || 0 < len(f.Max) {
		dst.Tab(2).Code("if (")
		if 0 < len(f.Min) {
			dst.Code("new d.Decimal(").Code(f.Min).Code(").greaterThan(val)")
		}
		if 0 < len(f.Max) {
			if 0 < len(f.Min) {
				dst.Code(" || ")
			}
			dst.Code("new d.Decimal(").Code(f.Max).Code(").lessThan(val)")
		}
		dst.Code(") {\n")
		b.printVerifyError(dst, pName, val)

		dst.Tab(1).Code("}\n")
	}

	dst.Tab(1).Code("} catch (e){\n")
	b.printVerifyError(dst, pName, val)
	dst.Tab(1).Code("}\n")
}
func (b *Builder) printVerifyError(dst *build.Writer, pName string, val *build.VerifyEnum) {
	dst.Tab(2).Code("return callback(new Error(locale.t(").Code(pName).Code(".")
	dst.Code(build.StringToHumpName(val.Enum.Name.Name)).Code(".")
	dst.Code(build.StringToAllUpper(val.Item.Name.Name)).Code(".toString())))\n")
}
