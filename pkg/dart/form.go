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

func getUiForm(tags map[string]*ast.Tag) *uiForm {
	val, ok := tags["uiForm"]
	if !ok {
		return nil
	}
	text := val.Value.Value[1 : len(val.Value.Value)-1]
	list := strings.Split(text, ";")
	form := uiForm{}
	if 0 < len(list) {
		for _, item := range list {
			if 0 == strings.Index(item, "name=") {
				form.suffix = item[len("name="):]
			}
		}
	}
	return &form
}

type uiText struct {
	onlyRead bool
}

func getUiText(tags map[string]*ast.Tag) *uiText {
	val, ok := tags["uiText"]
	if !ok {
		return nil
	}
	text := val.Value.Value[1 : len(val.Value.Value)-1]
	list := strings.Split(text, ";")
	form := uiText{}
	if 0 < len(list) {
		for _, item := range list {
			if 0 == strings.Index(item, "onlyRead=") {
				form.onlyRead = "true" == item[len("onlyRead="):]
			}
		}
	}
	return &form
}

type uiMenu struct {
	onlyRead bool
}

func getUiMenu(tags map[string]*ast.Tag) *uiMenu {
	val, ok := tags["uiMenu"]
	if !ok {
		return nil
	}
	text := val.Value.Value[1 : len(val.Value.Value)-1]
	list := strings.Split(text, ";")
	form := uiMenu{}
	for _, item := range list {
		if 0 == strings.Index(item, "onlyRead=") {
			form.onlyRead = "true" == item[len("onlyRead="):]
		}
	}
	return &form
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
