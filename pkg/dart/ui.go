package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
)

//创建表单代码
func (b *Builder) printFormCode(dst *Writer, expr ast.Expr) {
	dst.Import("package:flutter/material.dart")

	switch expr.(type) {
	case *ast.DataType:
		dst.Import("package:hbuf_flutter/hbuf_flutter.dart")
		typ := expr.(*ast.DataType)
		b.getPackage(dst, typ.Name, "")
		b.printDataUi(dst, typ)

	case *ast.ServerType:

	case *ast.EnumType:
		typ := expr.(*ast.EnumType)
		b.printEnumUi(dst, typ)
	}
}

func (b *Builder) printEnumUi(dst *Writer, typ *ast.EnumType) {
	_, ok := build.GetTag(typ.Tags, "ui")
	if !ok {
		return
	}

	enumName := build.StringToHumpName(typ.Name.Name)
	lang := NewLanguage(enumName)
	for _, item := range typ.Items {
		itemName := build.StringToAllUpper(item.Name.Name)
		lang.Add(itemName, item.Tags)
	}
	if lang != nil {
		lang.printLanguage(dst, b.lang)
	}
}

type ui struct {
	suffix   string
	onlyRead bool
	form     string
	table    string
	format   string
	digit    int
}

func (b *Builder) getUI(tags []*ast.Tag) *ui {
	val, ok := build.GetTag(tags, "ui")
	if !ok {
		return nil
	}
	form := ui{}
	if nil != val.KV {
		for _, item := range val.KV {
			if "onlyRead" == item.Name.Name {
				form.onlyRead = "true" == item.Value.Value[1:len(item.Value.Value)-1]
			} else if "form" == item.Name.Name {
				form.form = item.Value.Value[1 : len(item.Value.Value)-1]
			} else if "table" == item.Name.Name {
				form.table = item.Value.Value[1 : len(item.Value.Value)-1]
			} else if "digit" == item.Name.Name {
				atoi, err := strconv.Atoi(item.Value.Value[1 : len(item.Value.Value)-1])
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.digit = atoi
			} else if "table" == item.Name.Name {
				form.format = item.Value.Value[1 : len(item.Value.Value)-1]
			}
		}
	}
	return &form
}

func (b *Builder) printDataUi(dst *Writer, typ *ast.DataType) {
	u := b.getUI(typ.Tags)
	if nil == u {
		return
	}

	if u.form == "true" {
		lang := b.printForm(dst, typ, u)
		if lang != nil {
			lang.printLanguage(dst, b.lang)
		}
	}
	if u.table == "true" {
		lang := b.printTable(dst, typ, u)
		if lang != nil {
			lang.printLanguage(dst, b.lang)
		}
	}

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
	dst.Code("\t\t\treturn TablesRow<" + name + ">(data: list[y]);\n")
	dst.Code("\t\t};\n")
	dst.Code("\t\ttables.rowCount = list.length;\n")
	dst.Code("\t\ttables.columns = [\n")

	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		table := b.getUI(field.Tags)
		if nil == table || 0 == len(table.table) {
			return nil
		}
		fieldName := build.StringToFirstLower(field.Name.Name)
		dst.Code("\t\t\tTablesColumn(\n")
		dst.Code("\t\t\t\theaderBuilder: (context) {\n")
		dst.Code("\t\t\t\t\treturn TablesCell(child: Text(" + name + "Localizations.of(context)." + fieldName + "));\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tcellBuilder: (context, x, y,dynamic data) {\n")
		dst.Code("\t\t\t\t\treturn TablesCell(child: SelectableText(\"${data." + fieldName)
		b.printToString(dst, field.Type, false, table.digit, table.format)
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

func (b *Builder) printToString(dst *Writer, expr ast.Expr, empty bool, digit int, format string) {
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
			b.printToString(dst, t.Obj.Decl.(*ast.TypeSpec).Type, empty, digit, format)
		} else {
			switch t.Name {
			case build.Int8, build.Int16, build.Int32, build.Int64, build.Uint8, build.Uint16, build.Uint32, build.Uint64:
				if empty {
					dst.Code("?")
				}
				dst.Code(".toString()")
			case build.Float, build.Double:
				if empty {
					dst.Code("?")
				}
				dst.Code(".toStringAsFixed(" + strconv.Itoa(digit) + ")")
			case build.Bool:
				if empty {
					dst.Code("?")
				}
				dst.Code(".toString()")
			case build.Date:
				if empty {
					dst.Code("?")
				}
				dst.Code(".toString()")
			}
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
		b.printToString(dst, t.Type(), t.Empty, digit, format)
		if t.Empty {
			dst.Code("??\"\"")
		}
	}
}

func (b *Builder) printFormString(dst *Writer, name string, expr ast.Expr, empty bool, digit int, format string) {
	switch expr.(type) {
	case *ast.EnumType:
		t := expr.(*ast.EnumType)
		dst.Code(t.Name.Name + ".nameOf(" + name + ")")
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			b.getPackage(dst, expr, "")
			b.printFormString(dst, name, t.Obj.Decl.(*ast.TypeSpec).Type, empty, digit, format)
		} else {
			switch t.Name {
			case build.Int8, build.Int16, build.Int32, build.Int64, build.Uint8, build.Uint16, build.Uint32, build.Uint64:
				if empty {
					dst.Code(name + "==null ? null : int.tryParse(" + name + "!)?.toInt()")
				} else {
					dst.Code("int.tryParse(" + name + "!)!.toInt()")
				}

			case build.Float, build.Double:
				if empty {
					dst.Code(name + "==null ? null : int.tryParse(" + name + "!)?.toDouble()")
				} else {
					dst.Code("int.tryParse(" + name + "!)!.toDouble()")
				}
			case build.Bool:
				dst.Code("\"true\" == " + name)
			case build.Date:
				if empty {
					dst.Code(name + "==null ? null : DateTime.tryParse(" + name + "!)")
				} else {
					dst.Code("DateTime.tryParse(" + name + "!)")
				}
			default:
				if empty {
					dst.Code(name)
				} else {
					dst.Code(name + "!")
				}
			}
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
		b.printFormString(dst, name, t.Type(), t.Empty, digit, format)
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
			setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
			b.printToString(setValue, field.Type, false, form.digit, form.format)
			setValue.Code(";\n")
			setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
			b.printFormString(setValue, "val", field.Type, false, form.digit, form.format)
			setValue.Code(";\n")
			setValue.Code("\t\t" + fieldName + ".enabled = enabled;\n")
			setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")

			fields.Code("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		} else if "menu" == form.form {
			dst.Code("\tfinal MenuFormBuild<")
			b.printType(dst, field.Type, true)
			dst.Code("> " + fieldName + " = MenuFormBuild();\n\n")
			setValue.Code("\t\t" + fieldName + ".value = info." + fieldName + ";\n")
			setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val")
			if !build.IsNil(field.Type) {
				setValue.Code("!")
			}
			setValue.Code(";\n")
			setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n")
			setValue.Code("\t\tstatus.items = [\n")
			b.printMenuItem(setValue, field.Type, false)
			setValue.Code("\t\t];\n\n")

			fields.Code("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		} else if "pic" == form.form {
			lang.Add(fieldName, field.Tags)

		} else if "time" == form.form {
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
