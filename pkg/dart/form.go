package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strings"
)

//创建表单代码
func printFormCode(dst *Writer, typ *ast.DataType) {
	dst.Import("package:flutter/material.dart")
	dst.Import("package:hbuf_flutter/hbuf_flutter.dart")

	getPackage(dst, typ.Name)
	lang := printForm(dst, typ)
	if lang != nil {
		lang.printLanguage(dst)
	}
}

type ui struct {
	suffix   string
	typ      string
	onlyRead bool
	form     string
}

func getUI(tags []*ast.Tag) *ui {
	val, ok := build.GetTag(tags, "ui")
	if !ok {
		return nil
	}
	form := ui{}
	if nil != val.KV {
		for _, item := range val.KV {
			if "type" == item.Name.Name {
				form.typ = item.Value.Value[1 : len(item.Value.Value)-1]
			} else if "onlyRead" == item.Name.Name {
				form.onlyRead = "true" == item.Value.Value[1:len(item.Value.Value)-1]
			} else if "form" == item.Name.Name {
				form.form = item.Value.Value[1 : len(item.Value.Value)-1]
			}
		}
	}
	return &form
}

func printForm(dst *Writer, typ *ast.DataType) *Language {
	u := getUI(typ.Tags)
	if nil == u || u.typ != "form" {
		return nil
	}
	name := build.StringToHumpName(typ.Name.Name)
	lang := NewLanguage(name)
	dst.Code("class Form" + name + build.StringToHumpName(u.suffix) + "Build {\n")
	dst.Code("\tfinal " + name + " info;\n\n")
	dst.Code("\tfinal bool enabled;\n\n")

	dst.Code("\tForm" + name + build.StringToHumpName(u.suffix) + "Build (\n")
	dst.Code("\t\tthis.info, {\n")
	dst.Code("\t\tthis.enabled = true,\n")
	dst.Code("\t});\n\n")

	fields := strings.Builder{}
	setValue := strings.Builder{}
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		form := getUI(field.Tags)
		fieldName := build.StringToFirstLower(field.Name.Name)
		if nil == form {
			return nil
		}
		if "text" == form.form {
			dst.Code("\tfinal TextFormBuild " + fieldName + " = TextFormBuild();\n\n")
			setValue.WriteString("\t\t" + fieldName + ".initialValue = info." + fieldName + ";\n")
			setValue.WriteString("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!;\n")
			setValue.WriteString("\t\t" + fieldName + ".enabled = enabled;\n")
			setValue.WriteString("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")

			fields.WriteString("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		} else if "menu" == form.form {
			dst.Code("\tfinal MenuFormBuild<")
			printType(dst, field.Type, true)
			dst.Code("> " + fieldName + " = MenuFormBuild();\n\n")
			setValue.WriteString("\t\t" + fieldName + ".value = info." + fieldName + ";\n")
			setValue.WriteString("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!;\n")
			setValue.WriteString("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")

			fields.WriteString("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		}

		return nil
	})
	if err != nil {
		return nil
	}

	dst.Code("\tList<Widget> build(BuildContext context, {Function(Form" + name + build.StringToHumpName(u.suffix) + "Build form)? builder}) {\n")
	dst.Code(setValue.String())
	dst.Code("\t\tbuilder?.call(this);\n")
	dst.Code("\t\treturn <Widget>[\n")
	dst.Code(fields.String())
	dst.Code("\t\t];\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")

	return lang
}
