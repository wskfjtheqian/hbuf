package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
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

		dst.Code("String? verify" + dName + "_" + fName + "(BuildContext context, ")
		b.printType(dst, field.Type, false)
		dst.Code(" val) {\n")

		for _, val := range verify.GetFormat() {
			f := build.GetFormat(val.Item.Tags)
			if nil == f {
				continue
			}
			b.getPackage(dst, val.Enum.Name, "enum")
			if !f.IsNull && build.IsNil(field.Type) {
				dst.Code("\tif (null == val) {\n")
				dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
				dst.Code("\t}\n")
			}
			if 0 < len(f.Reg) {
				dst.Code("\tif (")
				if build.IsNil(field.Type) {
					dst.Code("null != val && ")
				}
				dst.Code("!RegExp(\"" + f.Reg + "\").hasMatch(val)) {\n")
				dst.Code("\t\treturn " + build.StringToHumpName(val.Enum.Name.Name) + "." + build.StringToAllUpper(val.Item.Name.Name) + ".toText(context);\n")
				dst.Code("\t}\n")
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
