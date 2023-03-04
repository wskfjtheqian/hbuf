package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
)

//创建表单代码
func (b *Builder) printFormCode(dst *build.Writer, expr ast.Expr) {
	dst.Import("package:flutter/material.dart", "")

	switch expr.(type) {
	case *ast.DataType:
		dst.Import("package:hbuf_flutter/hbuf_flutter.dart", "")
		typ := expr.(*ast.DataType)
		b.getPackage(dst, typ.Name, "")
		b.printDataUi(dst, typ)

	case *ast.ServerType:

	case *ast.EnumType:
		typ := expr.(*ast.EnumType)
		b.printEnumUi(dst, typ)
	}

}

func (b *Builder) printEnumUi(dst *build.Writer, typ *ast.EnumType) {
	_, ok := build.GetTag(typ.Tags, "ui")
	if !ok {
		return
	}

	enumName := build.StringToHumpName(typ.Name.Name)
	lang := dst.GetLang(enumName)
	for _, item := range typ.Items {
		itemName := build.StringToAllUpper(item.Name.Name)
		lang.Add(itemName, item.Tags)
	}
}

type ui struct {
	suffix     string
	onlyRead   bool
	form       string
	toNull     bool
	table      string
	format     string
	digit      int
	index      int
	width      float64
	height     float64
	maxLine    int
	extensions []string
}

func (b *Builder) getUI(tags []*ast.Tag) *ui {
	val, ok := build.GetTag(tags, "ui")
	if !ok {
		return nil
	}
	form := ui{
		width:      300,
		height:     300,
		maxLine:    1,
		extensions: []string{},
	}
	if nil != val.KV {
		for _, item := range val.KV {
			if "onlyRead" == item.Name.Name {
				form.onlyRead = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]
			} else if "form" == item.Name.Name {
				form.form = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			} else if "table" == item.Name.Name {
				form.table = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			} else if "digit" == item.Name.Name {
				atoi, err := strconv.Atoi(item.Values[0].Value[1 : len(item.Values[0].Value)-1])
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.digit = atoi
			} else if "index" == item.Name.Name {
				atoi, err := strconv.Atoi(item.Values[0].Value[1 : len(item.Values[0].Value)-1])
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.index = atoi
			} else if "format" == item.Name.Name {
				form.format = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			} else if "width" == item.Name.Name {
				atoi, err := strconv.ParseFloat(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10)
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.width = atoi
			} else if "height" == item.Name.Name {
				atoi, err := strconv.ParseFloat(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10)
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.height = atoi
			} else if "maxLine" == item.Name.Name {
				atoi, err := strconv.ParseInt(item.Values[0].Value[1:len(item.Values[0].Value)-1], 64, 10)
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.maxLine = int(atoi)
			} else if "toNull" == item.Name.Name {
				form.toNull = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]
			} else if "extensions" == item.Name.Name {
				for _, value := range item.Values {
					form.extensions = append(form.extensions, value.Value[1:len(value.Value)-1])
				}
			}
		}
	}
	return &form
}

func (b *Builder) printDataUi(dst *build.Writer, typ *ast.DataType) {
	u := b.getUI(typ.Tags)
	if nil == u {
		return
	}

	if u.form == "true" {
		b.printForm(dst, typ, u)
	}
	if u.table == "true" {
		b.printTable(dst, typ, u)
	}

}

func (b *Builder) printTable(dst *build.Writer, typ *ast.DataType, u *ui) {
	name := build.StringToHumpName(typ.Name.Name)
	lang := dst.GetLang(name)
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
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {

		table := b.getUI(field.Tags)
		if nil == table || 0 == len(table.table) {
			return nil
		}

		dst.Code("\t\ttables.columns[\"")
		fieldName := build.StringToFirstLower(field.Name.Name)
		dst.Code(fieldName)
		dst.Code("\"] = TablesColumn(\n")
		dst.Code("\t\t\t\tindex: " + strconv.Itoa(table.index) + ",\n")
		dst.Code("\t\t\t\theaderBuilder: (context) {\n")
		dst.Code("\t\t\t\t\treturn TablesCell(child: Text(" + name + "Localizations.of(context)." + fieldName + "));\n")
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t\tcellBuilder: (context, x, y, data) {\n")
		if "image" == table.table {
			dst.Code("\t\t\t\t\treturn TablesCell(\n")
			dst.Code("\t\t\t\t\t\tchild: (data." + fieldName)
			if build.IsNil(field.Type) {
				dst.Code("??\"\"")
			}
			dst.Code(").startsWith(\"http\")\n")
			dst.Code("\t\t\t\t\t\t\t\t? Image.network(\n")
			dst.Code("\t\t\t\t\t\t\t\t\t\tdata." + fieldName)
			if build.IsNil(field.Type) {
				dst.Code("!")
			}
			dst.Code(",\n")
			dst.Code("\t\t\t\t\t\t\t\t\t\tfit: BoxFit.contain,\n")
			dst.Code("\t\t\t\t\t\t\t\t\t)\n")
			dst.Code("\t\t\t\t\t\t\t\t: const SizedBox(),\n")
			dst.Code("\t\t\t\t\t);\n")
		} else {
			dst.Code("\t\t\t\treturn TablesCell(\n")
			dst.Code("\t\t\t\t\tchild: Tooltip(\n")
			dst.Code("\t\t\t\t\t\tmessage: data." + fieldName)
			b.printToString(dst, field.Type, false, table.digit, table.format, "??\"\"")
			dst.Code(",\n")
			dst.Code("\t\t\t\t\t\tchild: Text(\n")
			dst.Code("\t\t\t\t\t\t\tdata." + fieldName)
			b.printToString(dst, field.Type, false, table.digit, table.format, "??\"\"")
			dst.Code(",\n")
			dst.Code("\t\t\t\t\t\t\tmaxLines: " + strconv.Itoa(table.maxLine) + ",\n")
			dst.Code("\t\t\t\t\t\t\toverflow: TextOverflow.ellipsis,\n")
			dst.Code("\t\t\t\t\t\t),\n")
			dst.Code("\t\t\t\t\t),\n")
			dst.Code("\t\t\t\t);\n")
		}
		dst.Code("\t\t\t\t},\n")
		dst.Code("\t\t\t);\n")
		lang.Add(fieldName, field.Tags)
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("\t\tbuilder?.call(this);\n")
	dst.Code("\t\treturn tables.build(context);\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")
}

func (b *Builder) printToString(dst *build.Writer, expr ast.Expr, empty bool, digit int, format string, val string) {
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
			b.printToString(dst, t.Obj.Decl.(*ast.TypeSpec).Type, empty, digit, format, val)
		} else {
			switch build.BaseType(t.Name) {
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
				dst.Import("package:hbuf_flutter/hbuf_flutter.dart", "")
				if 0 == len(format) {
					format = "yyyy/MM/dd HH:mm:ss"
				}
				dst.Code(".format(\"" + format + "\")")
			case build.Decimal:
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
		b.printToString(dst, t.Type(), t.Empty, digit, format, val)
		if t.Empty {
			dst.Code(val)
		}
	}
}

func (b *Builder) printFormString(dst *build.Writer, name string, expr ast.Expr, empty bool, digit int, format string) {
	switch expr.(type) {
	case *ast.EnumType:
		t := expr.(*ast.EnumType)
		dst.Code(t.Name.Name + ".nameOf(" + name + "!)")
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			b.getPackage(dst, expr, "")
			b.printFormString(dst, name, t.Obj.Decl.(*ast.TypeSpec).Type, empty, digit, format)
		} else {
			switch build.BaseType(t.Name) {
			case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32:
				if empty {
					dst.Code(name + "==null ? null : num.tryParse(" + name + ")?.toInt()")
				} else {
					dst.Code("num.tryParse(" + name + "!)!.toInt()")
				}

			case build.Uint64, build.Int64:
				dst.Import("package:fixnum/fixnum.dart", "")
				if empty {
					dst.Code(name + "==null ? null : Int64.parseInt(" + name + ")")
				} else {
					dst.Code("Int64.parseInt(" + name + "!)")
				}
			case build.Float, build.Double:
				if empty {
					dst.Code(name + "==null ? null : num.tryParse(" + name + ")?.toDouble()")
				} else {
					dst.Code("num.tryParse(" + name + "!)!.toDouble()")
				}
			case build.Bool:
				dst.Code("\"true\" == " + name)
			case build.Date:
				if empty {
					dst.Code(name + "==null ? null : DateTime.tryParse(" + name + ")")
				} else {
					dst.Code("DateTime.tryParse(" + name + "!) ?? DateTime.now()")
				}
			case build.Decimal:
				dst.Import("package:decimal/decimal.dart", "")
				if empty {
					dst.Code(name + "==null ? null : Decimal.fromJson(" + name + ")")
				} else {
					dst.Code("Decimal.fromJson(" + name + "!)")
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

func (b *Builder) printForm(dst *build.Writer, typ *ast.DataType, u *ui) {
	name := build.StringToHumpName(typ.Name.Name)
	lang := dst.GetLang(name)
	dst.Code("class Form" + name + build.StringToHumpName(u.suffix) + "Build {\n")
	dst.Code("\tfinal " + name + " info;\n\n")
	dst.Code("\tfinal bool readOnly;\n\n")
	dst.Code("\tfinal Map<double, int> sizes;\n\n")
	dst.Code("\tfinal EdgeInsetsGeometry padding;\n\n")

	dst.Code("\tForm" + name + build.StringToHumpName(u.suffix) + "Build (\n")
	dst.Code("\t\tthis.info, {\n")
	dst.Code("\t\tthis.readOnly = false,\n")
	dst.Code("\t\tthis.sizes = const {},\n")
	dst.Code("\t\tthis.padding = const EdgeInsets.only(),\n")
	dst.Code("\t});\n\n")

	fields := build.NewWriter()
	setValue := build.NewWriter()
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		form := b.getUI(field.Tags)
		fieldName := build.StringToFirstLower(field.Name.Name)
		if nil == form {
			return nil
		}

		onlyRead := "false"
		if form.onlyRead {
			onlyRead = "true"
		}
		verify, err := build.GetVerify(field.Tags, dst.File, b.GetDataType)
		if err != nil {
			return err
		}
		if "text" == form.form {
			dst.Code("\tfinal TextFormBuild " + fieldName + " = TextFormBuild();\n\n")
			setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
			b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
			setValue.Code(";\n")
			setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
			if form.toNull {
				setValue.Code("\"\" == val ? null : ")
			}
			b.printFormString(setValue, "val", field.Type, false, form.digit, form.format)
			setValue.Code(";\n")
			setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
			setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
			setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
			if nil != verify {
				b.getPackage(dst, typ.Name, "verify")
				setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val!);\n")
			}
			setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")

			fields.Code("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		} else if "file" == form.form {
			dst.Code("\tfinal FileFormBuild " + fieldName + " = FileFormBuild();\n\n")
			setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
			b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
			setValue.Code(";\n")
			setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
			if form.toNull {
				setValue.Code("\"\" == val ? null : ")
			}
			b.printFormString(setValue, "val", field.Type, false, form.digit, form.format)
			setValue.Code(";\n")
			setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
			setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
			setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
			if 0 < len(form.extensions) {
				setValue.Code("\t\t" + fieldName + ".extensions = <String>[")
				for i, extension := range form.extensions {
					if 0 < i {
						setValue.Code(", ")
					}
					setValue.Code("\"" + extension + "\"")
				}
				setValue.Code("];\n")
			}
			if nil != verify {
				b.getPackage(dst, typ.Name, "verify")
				setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val!);\n")
			}
			setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")

			fields.Code("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		} else if "image" == form.form {
			dst.Code("\tfinal ImageFormBuild " + fieldName + " =  ImageFormBuild();\n\n")

			if build.IsNil(field.Type) {
				setValue.Code("\t\t" + fieldName + ".initialValue = [if (info." + fieldName + "?.startsWith(\"http\") ?? false) NetworkImage(info." + fieldName + "!)];\n")
				setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ((val?.isEmpty ?? true) ? null : val!.first.url);\n")
			} else {
				setValue.Code("\t\t" + fieldName + ".initialValue = [NetworkImage(info." + fieldName + ")];\n")
				setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!.first.url;\n")
			}
			setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
			setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
			setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
			setValue.Code("\t\t" + fieldName + ".outWidth = " + strconv.FormatFloat(form.width, 'G', -1, 64) + ";\n")
			setValue.Code("\t\t" + fieldName + ".outHeight = " + strconv.FormatFloat(form.height, 'G', -1, 64) + ";\n")
			if 0 < len(form.extensions) {
				setValue.Code("\t\t" + fieldName + ".extensions = <String>[")
				for i, extension := range form.extensions {
					if 0 < i {
						setValue.Code(", ")
					}
					setValue.Code("\"" + extension + "\"")
				}
				setValue.Code("];\n")
			}
			if nil != verify {
				b.getPackage(dst, typ.Name, "verify")
				setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val!);\n")
			}
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
			setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
			setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
			setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
			setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n")
			if nil != verify {
				b.getPackage(dst, typ.Name, "verify")
				setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val!);\n")
			}
			setValue.Code("\t\t" + fieldName + ".items = [\n")
			b.printMenuItem(setValue, field.Type, false)
			setValue.Code("\t\t];\n\n")

			fields.Code("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		} else if "time" == form.form {
			lang.Add(fieldName, field.Tags)
		}

		return nil
	})
	if err != nil {
		return
	}

	dst.Code("\tList<Widget> build(BuildContext context, {Function(Form" + name + build.StringToHumpName(u.suffix) + "Build ui)? builder}) {\n")
	dst.Code(setValue.String())
	dst.Code("\t\tbuilder?.call(this);\n")
	dst.Code("\t\treturn <Widget>[\n")
	dst.Code(fields.String())
	dst.Code("\t\t];\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")

	dst.ImportByWriter(setValue)
}

func (b *Builder) printMenuItem(dst *build.Writer, expr ast.Expr, empty bool) {
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
