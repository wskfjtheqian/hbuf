package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
	"time"
)

func (b *Builder) printVerifyCode(dst *build.Writer, data *ast.DataType) error {
	dst.Import("package:flutter/material.dart", "")

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

	err = b.printVerifyDataCode(dst, data)
	if err != nil {
		return err
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

		dst.Code("String? verify" + dName + "_" + fName + "(BuildContext context, String? text) {\n")
		//first := true
		for i, val := range verify.GetFormat() {
			f := build.GetFormat(val.Item.Tags)
			if nil == f {
				continue
			}
			b.getPackage(dst, val.Enum.Name, "enum")
			if build.IsNil(field.Type) && 0 == i {
				if !f.Null {
					dst.Code("\tif (text?.isEmpty ?? true) {\n")
					dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
					dst.Code("\t}\n")
				} else {
					dst.Code("\tif (text?.isEmpty ?? true) {\n")
					dst.Code("\t\treturn null;\n")
					dst.Code("\t}\n")
				}
			}
			if build.IsEnum(field.Type) {
				//dst.Code("\tif 0 < len(i.Get" + fName + "().ToName()) {\n")
				//dst.Code("\t\treturn &hbuf.Result{Code: int(" + pack + "), Msg: " + pack + ".ToName()}\n")
				//dst.Code("\t}\n")
			} else if build.IsMap(field.Type) {

			} else if build.IsArray(field.Type) {

			} else {
				t := build.GetBaseType(field.Type)
				switch t {
				case build.Int8, build.Int16, build.Int32:
					b.verifyNum(dst, val, f, "-?[1-9]\\\\d*")
				case build.Uint8, build.Uint16, build.Uint32:
					b.verifyNum(dst, val, f, "[1-9]\\\\d*")
				case build.Float, build.Double:
					b.verifyNum(dst, val, f, "-?[1-9]\\\\d*.\\\\d*|0.\\\\d*[1-9]\\\\d*")
				case build.Int64:
					b.verify(dst, val, f, "-?[1-9]\\\\d*")
				case build.Uint64:
					b.verify(dst, val, f, "[1-9]\\\\d*")
				case build.Date:
					dst.Code("\tDateTime? val = DateTime.tryParse(text!);\n")
					dst.Code("\tif (null == val) {\n")
					dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
					dst.Code("\t}\n")
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Code("\tif (")
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
						dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
						dst.Code("\t}\n")
					}
				case build.Decimal:
					dst.Code("\tif (!RegExp(\"-?[1-9]\\\\d*.\\\\d*|0.\\\\d*[1-9]\\\\d*\").hasMatch(text)) {\n")
					dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
					dst.Code("\t}\n")
					dst.Code("\tDecimal? val = Decimal.tryParse(text!);\n")
					dst.Code("\tif (null == val) {\n")
					dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
					dst.Code("\t}\n")
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Code("\tif (")
						if 0 < len(f.Min) {
							dst.Code("1 == val.compareTo(Decimal.fromInt(" + f.Max + "))")
						}
						if 0 < len(f.Max) {
							if 0 < len(f.Min) {
								dst.Code(" || ")
							}
							dst.Code("-1 == val.compareTo(Decimal.fromInt(" + f.Max + "))")
						}
						dst.Code(") {\n")
						dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
						dst.Code("\t}\n")
					}
				case build.String:
					dst.Code("\tif (!RegExp(\"" + f.Reg + "\").hasMatch(text)) {\n")
					dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
					dst.Code("\t}\n")
				}
			}
		}
		dst.Code("\treturn null;\n")
		dst.Code("}\n\n")
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) verify(dst *build.Writer, val *build.VerifyEnum, f *build.Format, reg string) {
	dst.Import("package:fixnum/fixnum.dart", "")
	dst.Code("\tif (!RegExp(\"" + reg + "\").hasMatch(text)) {\n")
	dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
	dst.Code("\t}\n")
	dst.Code("\tInt64? val = Int64.tryParse(text!);\n")
	dst.Code("\tif (null == val) {\n")
	dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
	dst.Code("\t}\n")
	if 0 < len(f.Min) || 0 < len(f.Max) {
		dst.Code("\tif (")
		if 0 < len(f.Min) {
			dst.Code("1 == val.compareTo(Int64(" + f.Max + "))")
		}
		if 0 < len(f.Max) {
			if 0 < len(f.Min) {
				dst.Code(" || ")
			}
			dst.Code("-1 == val.compareTo(Int64(" + f.Max + "))")
		}
		dst.Code(") {\n")
		dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
		dst.Code("\t}\n")
	}
}

func (b *Builder) verifyNum(dst *build.Writer, val *build.VerifyEnum, f *build.Format, reg string) {
	dst.Code("\tif (!RegExp(\"" + reg + "\").hasMatch(text)) {\n")
	dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
	dst.Code("\t}\n")
	dst.Code("\tnum? val = num.tryParse(text!);\n")
	dst.Code("\tif (null == val) {\n")
	dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
	dst.Code("\t}\n")
	if 0 < len(f.Min) || 0 < len(f.Max) {
		dst.Code("\tif (")
		if 0 < len(f.Min) {
			dst.Code(f.Min + " > val")
		}
		if 0 < len(f.Max) {
			if 0 < len(f.Min) {
				dst.Code(" || ")
			}
			dst.Code(f.Max + " < val")
		}
		dst.Code(") {\n")
		dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
		dst.Code("\t}\n")
	}
}

func (b *Builder) printVerifyDataCode(dst *build.Writer, data *ast.DataType) error {
	dName := build.StringToHumpName(data.Name.Name)
	b.getPackage(dst, data.Name, "data")

	dst.Code("extension Verify" + dName + " on " + dName + " {\n")
	dst.Code("\tString? verify(BuildContext context) {\n")
	isErr := true
	err := build.EnumField(data, func(field *ast.Field, data *ast.DataType) error {
		fName := build.StringToHumpName(field.Name.Name)
		_, ok := build.GetTag(field.Tags, "verify")
		if ok {
			if isErr {
				dst.Code("\t\tString? err;\n")
				isErr = false
			}
			dst.Code("\t\terr = verify" + dName + "_" + fName + "(context, " + build.StringToFirstLower(field.Name.Name) + ");\n")
			dst.Code("\t\tif (err != null) {\n")
			dst.Code("\t\t\treturn err;\n")
			dst.Code("\t\t}\n")
		}
		return nil
	})
	if err != nil {
		return err
	}
	dst.Code("\t\treturn null;\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")

	return nil
}
