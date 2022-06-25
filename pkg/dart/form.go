package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strings"
)

//创建表单代码
func printFormCode(dst *Writer, typ *ast.DataType) {
	dst.Import("package:flutter/widgets.dart")

	printForm(dst, typ)
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
	for _, item := range list {
		if 0 == strings.Index(item, "name=") {
			form.suffix = item[len("name="):]
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
	for _, item := range list {
		if 0 == strings.Index(item, "onlyRead=") {
			form.onlyRead = "true" == item[len("onlyRead="):]
		}
	}
	return &form
}

func printForm(dst *Writer, typ *ast.DataType) {
	form := getUiForm(typ.Tags)
	if nil == form {
		return
	}
	name := build.StringToHumpName(typ.Name.Name)
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
			setValue.WriteString("\t\t" + fieldName + ".enabled = enabled;\n\n")

			fields.WriteString("\t\t\t" + fieldName + ".build(),\n")
		}

		return nil
	})
	if err != nil {
		return
	}

	dst.Code("\tList<FormField> build({Function(Form" + name + build.StringToHumpName(form.suffix) + "Build form)? builder}) {\n")
	dst.Code(setValue.String())
	dst.Code("\t\treturn <FormField>[\n")
	dst.Code(fields.String())
	dst.Code("\t\t];\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")

	//	final TextFormBuild userName = TextFormBuild();
	//
	//	final TextFormBuild passWord = TextFormBuild();
	//
	//	List<FormField> build({Function(AdminLoginByUserNameReqBuild form)? builder}) {
	//	userName.initialValue = info.userName;
	//	userName.onSaved = (val) => info.userName = val!;
	//	userName.enabled = enabled;
	//
	//	passWord.initialValue = info.passWord;
	//	passWord.onSaved = (val) => info.passWord = val!;
	//	passWord.enabled = enabled;
	//
	//	builder?.call(this);
	//	return <FormField>[
	//	userName.build(),
	//	passWord.build(),
	//];
	//}
	//}

}
