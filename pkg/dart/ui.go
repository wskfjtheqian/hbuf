package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

//创建表单代码
func (b *Builder) printFormCode(dst *Writer, expr ast.Expr) {
	dst.Import("package:flutter/material.dart")

	switch expr.(type) {
	case *ast.DataType:
		dst.Import("package:hbuf_flutter/hbuf_flutter.dart")
		typ := expr.(*ast.DataType)
		b.getPackage(dst, typ.Name, "")
		lang := b.printDataUi(dst, typ)
		if lang != nil {
			lang.printLanguage(dst)
		}
	case *ast.ServerType:

	case *ast.EnumType:
		typ := expr.(*ast.EnumType)
		lang := b.printEnumUi(dst, typ)
		if lang != nil {
			lang.printLanguage(dst)
		}
	}
}

func (b *Builder) printEnumUi(dst *Writer, typ *ast.EnumType) *Language {
	_, ok := build.GetTag(typ.Tags, "ui")
	if !ok {
		return nil
	}

	enumName := build.StringToHumpName(typ.Name.Name)
	lang := NewLanguage(enumName)
	for _, item := range typ.Items {
		itemName := build.StringToAllUpper(item.Name.Name)
		lang.Add(itemName, item.Tags)
	}
	return lang
}

type ui struct {
	suffix   string
	typ      string
	onlyRead bool
	form     string
	table    string
}

func (b *Builder) getUI(tags []*ast.Tag) *ui {
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
			} else if "table" == item.Name.Name {
				form.table = item.Value.Value[1 : len(item.Value.Value)-1]
			}
		}
	}
	return &form
}

func (b *Builder) printDataUi(dst *Writer, typ *ast.DataType) *Language {
	u := b.getUI(typ.Tags)
	if nil == u {
		return nil
	}

	if u.typ == "form" {
		return b.printForm(dst, typ, u)
	} else if u.typ == "table" {
		return b.printTable(dst, typ, u)
	}

	return nil
}

func (b *Builder) printTable(dst *Writer, typ *ast.DataType, u *ui) *Language {
	name := build.StringToHumpName(typ.Name.Name)
	lang := NewLanguage(name)
	dst.Code("class Table" + name + build.StringToHumpName(u.suffix) + "Build {\n")
	dst.Code("\tfinal List<" + name + "> list;\n\n")
	dst.Code("\tfinal TablesBuild<" + name + "> tables = TablesBuild();\n\n")
	dst.Code("\tTable" + name + build.StringToHumpName(u.suffix) + "Build (")
	dst.Code("\t\tthis.list);\n\n")

	dst.Code("\tWidget build(BuildContext context, {Function(Table" + name + build.StringToHumpName(u.suffix) + "Build ui)? builder}) {\n")
	dst.Code("\t\ttables.rowBuilder = (context, y) {\n")
	dst.Code("\t\t\treturn TablesRow<AdminInfo>(data: list[y]);\n")
	dst.Code("\t\t};\n")
	dst.Code("\t\ttables.rowCount = list.length;\n")
	dst.Code("\t\ttables.columns = [\n")

	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		table := b.getUI(field.Tags)
		fieldName := build.StringToFirstLower(field.Name.Name)
		if nil == table {
			return nil
		}
		dst.Code("\t\t\tTablesColumn(\n")
		dst.Code("\t\t\t\theaderBuilder: (context) {\n")
		dst.Code("\t\t\t\t\treturn TablesCell(child: Text(" + name + "Localizations.of(context)." + fieldName + "));\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tcellBuilder: (context, x, y,dynamic data) {\n")
		dst.Code("\t\t\t\t\treturn TablesCell(child: SelectableText(\"${data." + fieldName)
		b.printValue(dst, field.Type, false)
		dst.Code(" }\"));\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t),\n")
		lang.Add(fieldName, field.Tags)
		return nil
	})
	if err != nil {
		return nil
	}
	dst.Code("\t\t];\n")

	dst.Code("\t\tbuilder?.call(this);\n")
	dst.Code("\t\treturn tables.build(context);\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")
	return lang
}

func (b *Builder) printValue(dst *Writer, expr ast.Expr, empty bool) {
	switch expr.(type) {
	case *ast.EnumType:
		if empty {
			dst.Code("?")
		}
		dst.Code(".toText(context)")
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			b.getPackage(dst, expr, "")
			b.printValue(dst, t.Obj.Decl.(*ast.TypeSpec).Type, empty)
		}
	case *ast.ArrayType:
		//ar := expr.(*ast.ArrayType)
		//dst.Code("List<")
		//printType(dst, ar.VType, false)
		//dst.Code(">")
		//if ar.Empty && !notEmpty {
		//	dst.Code("?")
		//}
	case *ast.MapType:
		//ma := expr.(*ast.MapType)
		//dst.Code("Map<")
		//printType(dst, ma.Key, false)
		//dst.Code(", ")
		//printType(dst, ma.VType, false)
		//dst.Code(">")
		//if ma.Empty && !notEmpty {
		//	dst.Code("?")
		//}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printValue(dst, t.Type(), t.Empty)
		if t.Empty {
			dst.Code("??\"\"")
		}
	}
}

func (b *Builder) printForm(dst *Writer, typ *ast.DataType, u *ui) *Language {
	name := build.StringToHumpName(typ.Name.Name)
	lang := NewLanguage(name)
	dst.Code("class Form" + name + build.StringToHumpName(u.suffix) + "Build {\n")
	dst.Code("\tfinal " + name + " info;\n\n")
	dst.Code("\tfinal bool enabled;\n\n")

	dst.Code("\tForm" + name + build.StringToHumpName(u.suffix) + "Build (\n")
	dst.Code("\t\tthis.info, {\n")
	dst.Code("\t\tthis.enabled = true,\n")
	dst.Code("\t});\n\n")

	fields := NewWriter("")
	setValue := NewWriter("")
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		form := b.getUI(field.Tags)
		fieldName := build.StringToFirstLower(field.Name.Name)
		if nil == form {
			return nil
		}
		if "text" == form.form {
			dst.Code("\tfinal TextFormBuild " + fieldName + " = TextFormBuild();\n\n")
			setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName + ";\n")
			setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!;\n")
			setValue.Code("\t\t" + fieldName + ".enabled = enabled;\n")
			setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")

			fields.Code("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		} else if "menu" == form.form {
			dst.Code("\tfinal MenuFormBuild<")
			b.printType(dst, field.Type, true)
			dst.Code("> " + fieldName + " = MenuFormBuild();\n\n")
			setValue.Code("\t\t" + fieldName + ".value = info." + fieldName + ";\n")
			setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val;\n")
			setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n")
			setValue.Code("\t\tstatus.items = [\n")
			b.printMenuItem(setValue, field.Type, false)
			setValue.Code("\t\t];\n\n")

			fields.Code("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		}

		return nil
	})
	if err != nil {
		return nil
	}

	dst.Code("\tList<Widget> build(BuildContext context, {Function(Form" + name + build.StringToHumpName(u.suffix) + "Build ui)? builder}) {\n")
	dst.Code(setValue.String())
	dst.Code("\t\tbuilder?.call(this);\n")
	dst.Code("\t\treturn <Widget>[\n")
	dst.Code(fields.String())
	dst.Code("\t\t];\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")
	return lang
}

func (b *Builder) printMenuItem(dst *Writer, expr ast.Expr, empty bool) {
	switch expr.(type) {
	case *ast.EnumType:
		t := expr.(*ast.EnumType)
		name := build.StringToHumpName(t.Name.Name)
		dst.Code("\t\t\tfor (var item in " + name + ".values)\n")
		dst.Code("\t\t\t\tDropdownMenuItem<" + name + ">(\n")
		dst.Code("\t\t\t\t\tvalue: item,\n")
		dst.Code("\t\t\t\t\tchild: Text(item.toText(context)),\n")
		dst.Code("\t\t\t\t),\n")
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			b.getPackage(dst, expr, "")
			b.printMenuItem(dst, t.Obj.Decl.(*ast.TypeSpec).Type, empty)
		}
	case *ast.ArrayType:
		//ar := expr.(*ast.ArrayType)
		//dst.Code("List<")
		//printType(dst, ar.VType, false)
		//dst.Code(">")
		//if ar.Empty && !notEmpty {
		//	dst.Code("?")
		//}
	case *ast.MapType:
		//ma := expr.(*ast.MapType)
		//dst.Code("Map<")
		//printType(dst, ma.Key, false)
		//dst.Code(", ")
		//printType(dst, ma.VType, false)
		//dst.Code(">")
		//if ma.Empty && !notEmpty {
		//	dst.Code("?")
		//}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printMenuItem(dst, t.Type(), t.Empty)
	}
}
