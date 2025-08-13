package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
)

// 创建表单代码
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
	table      string
	format     string
	digit      int
	index      *int
	width      float64
	height     float64
	maxLine    int
	extensions []string
	clip       bool
	maxCount   int
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
		maxCount:   1,
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
				form.index = &atoi
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
				atoi, err := strconv.ParseInt(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10, 64)
				if err != nil {
					println(err.Error())
					return nil
				}
				form.maxLine = int(atoi)
			} else if "maxCount" == item.Name.Name {
				atoi, err := strconv.ParseInt(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10, 64)
				if err != nil {
					println(err.Error())
					return nil
				}
				form.maxCount = int(atoi)
			} else if "clip" == item.Name.Name {
				form.clip = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]

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
	dst.Tab(1).Code("final List<" + name + "> list;\n\n")
	dst.Tab(1).Code("final Set<" + name + "> select;\n\n")
	dst.Tab(1).Code("final Function(BuildContext context,  Set<" + name + "> select)? onSelect;\n\n")
	dst.Tab(1).Code("final TablesBuild<" + name + "> tables = TablesBuild();\n\n")
	dst.Tab(1).Code("Table" + name + build.StringToHumpName(u.suffix) + "Build (this.list,{this.select = const {}, this.onSelect});\n\n")

	dst.Tab(1).Code("Widget build(BuildContext context, {Function(Table" + name + build.StringToHumpName(u.suffix) + "Build ui)? builder}) {\n")
	dst.Tab(2).Code("tables.rowBuilder = (context, y) {\n")
	dst.Tab(3).Code("return TablesRow<" + name + ">(data: list[y]);\n")
	dst.Tab(2).Code("};\n")
	dst.Tab(2).Code("tables.rowCount = list.length;\n")

	dst.Tab(2).Code("if(null != onSelect){\n")
	dst.Tab(3).Code("tables.columns[\"select_checkbox\"] = TablesColumn(\n")
	dst.Tab(4).Code("index: -1,\n")
	dst.Tab(4).Code("headerBuilder: (context) {\n")
	dst.Tab(5).Code("var length = list.where((element) => select.contains(element)).length;\n")
	dst.Tab(6).Code("return TablesCell(\n")
	dst.Tab(7).Code("child: Checkbox(\n")
	dst.Tab(7).Code("value: list.length == length ? true : (0 == length ? false : null),\n")
	dst.Tab(7).Code("tristate: true,\n")
	dst.Tab(7).Code("onChanged: (val) {\n")
	dst.Tab(8).Code("if (null == val) {\n")
	dst.Tab(9).Code("select.clear();\n")
	dst.Tab(8).Code("} else {\n")
	dst.Tab(9).Code("for (var item in list) {\n")
	dst.Tab(10).Code("select.add(item);\n")
	dst.Tab(9).Code("}\n")
	dst.Tab(8).Code("}\n")
	dst.Tab(8).Code("onSelect?.call(context, select);\n")
	dst.Tab(7).Code("},\n")
	dst.Tab(6).Code("),\n")
	dst.Tab(5).Code(");\n")
	dst.Tab(4).Code("},\n")
	dst.Tab(4).Code("cellBuilder: (context, x, y, " + name + " data) {\n")
	dst.Tab(5).Code("return TablesCell(\n")
	dst.Tab(6).Code("child: Checkbox(\n")
	dst.Tab(7).Code("value: select.contains(data),\n")
	dst.Tab(7).Code("onChanged: (val) {\n")
	dst.Tab(8).Code("if (select.contains(data)) {\n")
	dst.Tab(9).Code("select.remove(data);\n")
	dst.Tab(8).Code("} else {\n")
	dst.Tab(9).Code("select.add(data);\n")
	dst.Tab(8).Code("}\n")
	dst.Tab(8).Code("onSelect?.call(context, select);\n")
	dst.Tab(7).Code("},\n")
	dst.Tab(6).Code("),\n")
	dst.Tab(5).Code(");\n")
	dst.Tab(4).Code("},\n")
	dst.Tab(3).Code(");\n")
	dst.Tab(2).Code("}\n\n")

	i := 0
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {

		table := b.getUI(field.Tags)
		if nil == table || 0 == len(table.table) {
			return nil
		}

		isArray := build.IsArray(field.Type)
		isNull := build.IsNil(field.Type)
		i++
		index := i
		if nil != table.index {
			index = *table.index
		}
		dst.Tab(2).Code("tables.columns[\"")
		fieldName := build.StringToFirstLower(field.Name.Name)
		dst.Code(fieldName)
		dst.Code("\"] = TablesColumn(\n")
		dst.Tab(4).Code("index: " + strconv.Itoa(index) + ",\n")
		dst.Tab(4).Code("headerBuilder: (context) {\n")
		dst.Tab(5).Code("return TablesCell(child: Text(" + name + "Localizations.of(context)." + fieldName + "));\n")
		dst.Tab(4).Code("},\n")
		dst.Tab(4).Code("cellBuilder: (context, x, y, data) {\n")
		if "image" == table.table {
			dst.Tab(5).Code("return TablesCell(\n")
			dst.Tab(6).Code("child: ((data." + fieldName)
			if isNull {
				dst.Code("?")
			}
			dst.Code(".isEmpty ")
			if isNull {
				dst.Code("?? false")
			}
			dst.Code(") ? \"\" : data." + fieldName)
			if isArray {
				if isNull {
					dst.Code("?")
				}
				dst.Code(".first")
				dst.Code("??\"\"")
			} else if isNull {
				dst.Code("??\"\"")
			}
			dst.Code(").startsWith(\"http\")\n")
			dst.Tab(8).Code("? Image.network(\n")
			dst.Tab(10).Code("data." + fieldName)
			if isArray {
				if isNull {
					dst.Code("!")
				}
				dst.Code(".first")
				dst.Code("!")
			} else if isNull {
				dst.Code("!")
			}
			dst.Code(",\n")
			dst.Tab(10).Code("fit: BoxFit.contain,\n")
			dst.Tab(9).Code(")\n")
			dst.Tab(8).Code(": const SizedBox(),\n")
			dst.Tab(5).Code(");\n")
		} else {
			dst.Tab(4).Code("return TablesCell(\n")
			dst.Tab(5).Code("child: Tooltip(\n")
			dst.Tab(6).Code("message: data." + fieldName)
			if isArray {
				if isNull {
					dst.Code("?")
				}
				dst.Code(".map((e) => e")
				b.printToString(dst, field.Type, false, table.digit, table.format, "??\"\"")
				dst.Code(").join(',')")
				if isNull {
					dst.Code(" ?? ''")
				}
			} else {
				b.printToString(dst, field.Type, false, table.digit, table.format, "??\"\"")
			}
			dst.Code(",\n")
			dst.Tab(6).Code("child: Text(\n")
			dst.Tab(7).Code("data." + fieldName)
			if isArray {
				if isNull {
					dst.Code("?")
				}
				dst.Code(".map((e) => e")
				b.printToString(dst, field.Type, false, table.digit, table.format, "??\"\"")
				dst.Code(").join(',')")
				if isNull {
					dst.Code(" ?? ''")
				}
			} else {
				b.printToString(dst, field.Type, false, table.digit, table.format, "??\"\"")
			}
			dst.Code(",\n")
			dst.Tab(7).Code("maxLines: 1,\n")
			dst.Tab(7).Code("overflow: TextOverflow.ellipsis,\n")
			dst.Tab(6).Code("),\n")
			dst.Tab(5).Code("),\n")
			dst.Tab(4).Code(");\n")
		}
		dst.Tab(4).Code("},\n")
		dst.Tab(3).Code(");\n")
		lang.Add(fieldName, field.Tags)
		return nil
	})
	if err != nil {
		return
	}

	dst.Tab(2).Code("builder?.call(this);\n")
	dst.Tab(2).Code("return tables.build(context);\n")
	dst.Tab(1).Code("}\n")
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
				dst.Code(".toStringAsFixed(" + strconv.Itoa(digit) + ")")
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
	default:
		if empty {
			dst.Code("?.toString()")
		} else {
			dst.Code(".toString()")
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
	dst.Tab(1).Code("final " + name + " info;\n\n")
	dst.Tab(1).Code("final bool readOnly;\n\n")
	dst.Tab(1).Code("final Map<double, int> sizes;\n\n")
	dst.Tab(1).Code("final EdgeInsetsGeometry padding;\n\n")

	dst.Tab(1).Code("Form" + name + build.StringToHumpName(u.suffix) + "Build (\n")
	dst.Tab(2).Code("this.info, {\n")
	dst.Tab(2).Code("this.readOnly = false,\n")
	dst.Tab(2).Code("this.sizes = const {},\n")
	dst.Tab(2).Code("this.padding = const EdgeInsets.only(),\n")
	dst.Tab(1).Code("});\n\n")

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
		isArray := build.IsArray(field.Type)
		isNil := build.IsNil(field.Type)

		verify, err := build.GetVerify(field.Tags, dst.File, b.GetDataType)
		if err != nil {
			return err
		}
		if "text" == form.form {
			dst.Tab(1).Code("final TextFormBuild " + fieldName + " = TextFormBuild();\n\n")
			if isArray {
				setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
				if isNil {
					setValue.Code("?")
				}
				setValue.Code(".map((e) => e")
				b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
				setValue.Code(").join(\",\");\n")
				if !form.onlyRead {
					setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
					setValue.Code("val?.split(\",\").map((e) =>e")
					b.printFormString(setValue, "val", field.Type, false, form.digit, form.format)
					setValue.Code(").toList()")
					if !isNil {
						setValue.Code(" ?? [] ")
					}
					setValue.Code(";\n")
				}
			} else {
				setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
				b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
				setValue.Code(";\n")
				if !form.onlyRead {
					setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
					setValue.Code("\"\" == val ? null : ")
					b.printFormString(setValue, "val", field.Type, false, form.digit, form.format)
					setValue.Code(";\n")
				}
			}

			setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
			setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
			setValue.Code("\t\t" + fieldName + ".maxLines = " + strconv.Itoa(form.maxLine) + ";\n")
			setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
			if nil != verify {
				b.getPackage(dst, typ.Name, "verify")
				setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val!);\n")
			}
			setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")

			fields.Code("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		} else if "click" == form.form {
			dst.Tab(1).Code("final ClickFormBuild<")
			b.printType(dst, field.Type, false)
			dst.Code("> " + fieldName + " = ClickFormBuild();\n\n")

			setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName + ";\n")
			if !form.onlyRead {
				setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val")
				if !build.IsNil(field.Type) {
					setValue.Code("!")
				}
				setValue.Code(";\n")
			}
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
			dst.Tab(1).Code("final FileFormBuild " + fieldName + " = FileFormBuild();\n\n")
			setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
			b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
			setValue.Code(";\n")
			if !form.onlyRead {
				setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
				setValue.Code("\"\" == val ? null : ")
				b.printFormString(setValue, "val", field.Type, false, form.digit, form.format)
				setValue.Code(";\n")
			}
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
			dst.Tab(1).Code("final ImageFormBuild " + fieldName + " =  ImageFormBuild();\n\n")

			if isArray {
				if build.IsNil(field.Type) {
					setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName + "?.map((e) => ImageFormImage(e))?.toList() ?? [];\n")
					if !form.onlyRead {
						setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!.map((e) => e.url).toList();\n")
					}
				} else {
					setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName + ".map((e) => ImageFormImage(e)).toList();\n")
					if !form.onlyRead {
						setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!.map((e) => e.url).toList();\n")
					}
				}
			} else {
				if build.IsNil(field.Type) {
					setValue.Code("\t\t" + fieldName + ".initialValue = [if (info." + fieldName + "?.startsWith(\"http\") ?? false) ImageFormImage(info." + fieldName + "!)];\n")
					if !form.onlyRead {
						setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ((val?.isEmpty ?? true) ? null : val!.first.url);\n")
					}
				} else {
					setValue.Code("\t\t" + fieldName + ".initialValue = [ImageFormImage(info." + fieldName + ")];\n")
					if !form.onlyRead {
						setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!.first.url;\n")
					}
				}
			}

			setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
			setValue.Code("\t\t" + fieldName + ".maxCount = " + strconv.Itoa(form.maxCount) + ";\n")
			setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
			setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
			if form.clip {
				setValue.Code("\t\t" + fieldName + ".clip = true;\n")
			} else {
				setValue.Code("\t\t" + fieldName + ".clip = false;\n")
			}
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
			dst.Tab(1).Code("final MenuFormBuild<")
			b.printType(dst, field.Type, false)
			dst.Code("> " + fieldName + " = MenuFormBuild();\n\n")
			setValue.Code("\t\t" + fieldName + ".value = info." + fieldName + ";\n")
			if !form.onlyRead {
				setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val")
				if !build.IsNil(field.Type) {
					setValue.Code("!")
				}
				setValue.Code(";\n")
			}
			setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
			setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
			setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
			setValue.Code("\t\t" + fieldName + ".toNull = true;\n")
			setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n")
			if nil != verify {
				b.getPackage(dst, typ.Name, "verify")
				setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val?.toString());\n")
			}
			setValue.Code("\t\t" + fieldName + ".items = [\n")
			b.printMenuItem(setValue, field.Type, false)
			setValue.Code("\t\t];\n\n")

			fields.Code("\t\t\t" + fieldName + ".build(context),\n")
			lang.Add(fieldName, field.Tags)
		} else if "date" == form.form {
			dst.Tab(1).Code("final DatetimeFormBuild " + fieldName + " =  DatetimeFormBuild();\n\n")
			setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
			b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
			setValue.Code(";\n")
			if !form.onlyRead {
				setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
				setValue.Code("\"\" == val ? null : ")
				setValue.Code("val ;\n")
			}
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
		} else if "switch" == form.form {
			dst.Tab(1).Code("final SwitchFormBuild " + fieldName + " =  SwitchFormBuild();\n\n")
			setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
			//b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
			setValue.Code(";\n")
			if !form.onlyRead {
				setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
				if !build.IsNil(field.Type) {
					setValue.Code("val ?? false ;\n")
				} else {
					setValue.Code("val ;\n")
				}

			}
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
		}

		return nil
	})
	if err != nil {
		return
	}

	dst.Tab(1).Code("List<Widget> build(BuildContext context, {Function(BuildContext context, Form" + name + build.StringToHumpName(u.suffix) + "Build ui)? builder}) {\n")
	dst.Code(setValue.String())
	dst.Tab(2).Code("builder?.call(context, this);\n")
	dst.Tab(2).Code("return <Widget>[\n")
	dst.Code(fields.String())
	dst.Tab(2).Code("];\n")
	dst.Tab(1).Code("}\n")
	dst.Code("}\n\n")

	dst.ImportByWriter(setValue)
}

func (b *Builder) printMenuItem(dst *build.Writer, expr ast.Expr, empty bool) {
	switch expr.(type) {
	case *ast.EnumType:
		t := expr.(*ast.EnumType)
		name := build.StringToHumpName(t.Name.Name)
		dst.Tab(3).Code("for (var item in " + name + ".values)\n")
		dst.Tab(4).Code("DropdownMenuItem<" + name + ">(\n")
		dst.Tab(5).Code("value: item,\n")
		dst.Tab(5).Code("child: Text(item.toText(context)),\n")
		dst.Tab(4).Code("),\n")
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
