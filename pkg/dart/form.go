package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strings"
)

//创建表单代码
func printFormCode(dst *Writer, typ *ast.DataType) {
	dst.Import("package:flutter/material.dart")
	getPackage(dst, typ.Name)
	lang := printForm(dst, typ)
	if lang != nil {
		lang.printLanguage(dst)
	}
}

type uiForm struct {
	suffix string
}

func getUiForm(tags []*ast.Tag) *uiForm {
	val, ok := build.GetTag(tags, "uiForm")
	if !ok {
		return nil
	}
	form := uiForm{}
	if nil != val.KV {
		for _, item := range val.KV {
			if "name" == item.Name.Name {
				form.suffix = item.Value.Value[:len(item.Value.Value)-1]
			}

		}
	}
	return &form
}

type uiText struct {
	onlyRead bool
}

func getUiText(tags []*ast.Tag) *uiText {
	val, ok := build.GetTag(tags, "uiText")
	if !ok {
		return nil
	}
	text := uiText{}
	if nil != val.KV {
		for _, item := range val.KV {
			if "onlyRead" == item.Name.Name {
				text.onlyRead = "true" == item.Value.Value[:len(item.Value.Value)-1]
			}

		}
	}
	return &text
}

type uiMenu struct {
	onlyRead bool
}

func getUiMenu(tags []*ast.Tag) *uiMenu {
	val, ok := build.GetTag(tags, "uiMenu")
	if !ok {
		return nil
	}
	menu := uiMenu{}
	if nil != val.KV {
		for _, item := range val.KV {
			if "onlyRead" == item.Name.Name {
				menu.onlyRead = "true" == item.Value.Value[:len(item.Value.Value)-1]
			}

		}
	}
	return &menu
}

func printForm(dst *Writer, typ *ast.DataType) *Language {
	form := getUiForm(typ.Tags)
	if nil == form {
		return nil
	}
	name := build.StringToHumpName(typ.Name.Name)
	lang := NewLanguage(name)
	dst.Code("class Form" + name + build.StringToHumpName(form.suffix) + "Build {\n")
	dst.Code("\tfinal " + name + " info;\n\n")
	dst.Code("\tfinal bool enabled;\n\n")

	dst.Code("\tForm" + name + build.StringToHumpName(form.suffix) + "Build (\n")
	dst.Code("\t\tthis.info, {\n")
	dst.Code("\t\tthis.enabled = true,\n")
	dst.Code("\t});\n\n")

	fields := strings.Builder{}
	setValue := strings.Builder{}
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		text := getUiText(field.Tags)
		fieldName := build.StringToFirstLower(field.Name.Name)
		if nil != text {
			dst.Code("\tfinal TextFormBuild " + fieldName + " = TextFormBuild();\n\n")
			setValue.WriteString("\t\t" + fieldName + ".initialValue = info." + fieldName + ";\n")
			setValue.WriteString("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!;\n")
			setValue.WriteString("\t\t" + fieldName + ".enabled = enabled;\n")
			setValue.WriteString("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")

			fields.WriteString("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
			return nil
		}

		menu := getUiMenu(field.Tags)
		if nil != menu {
			dst.Code("\tfinal MenuFormBuild<")
			printType(dst, field.Type, true)
			dst.Code("> " + fieldName + " = MenuFormBuild();\n\n")
			setValue.WriteString("\t\t" + fieldName + ".value = info." + fieldName + ";\n")
			setValue.WriteString("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!;\n")
			setValue.WriteString("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")

			fields.WriteString("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
			return nil
		}

		return nil
	})
	if err != nil {
		return nil
	}

	dst.Code("\tList<Widget> build(BuildContext context, {Function(Form" + name + build.StringToHumpName(form.suffix) + "Build form)? builder}) {\n")
	dst.Code(setValue.String())
	dst.Code("\t\tbuilder?.call(this);\n")
	dst.Code("\t\treturn <Widget>[\n")
	dst.Code(fields.String())
	dst.Code("\t\t];\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")

	return lang
}
