package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
)

// 创建表单代码
func (b *Builder) printFormCode(dst *build.Writer, expr ast.Expr) {

	switch expr.(type) {
	case *ast.DataType:
		dst.Import("vue", "{defineComponent}")
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
		itemName := build.StringToHumpName(item.Name.Name)
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
		width:      160,
		height:     160,
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

	dst.Code("export const " + name + "TableColumn = defineComponent({\n")
	dst.Code("\tname: '" + name + "TableColumn',\n")
	dst.Code("\tprops: {hide:String},\n")
	dst.Code("\tsetup(props) {\n")
	dst.Code("\t\treturn (_ctx) => (\n")
	dst.Code("\t\t\t<>\n")
	langName := build.StringToFirstLower(name)
	//i := 0
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {

		table := b.getUI(field.Tags)
		if nil == table || 0 == len(table.table) {
			return nil
		}

		//isArray := build.IsArray(field.Type)
		//isNull := build.IsNil(field.Type)
		//i++
		//index := i
		if nil != table.index {
			//index = *table.index
		}
		//dst.Code("                <el-table-column prop="adminId" label="adminId" width="140"/>\n")
		fieldName := build.StringToFirstLower(field.Name.Name)

		dst.Code("\t\t\t\t<el-table-column prop=\"").Code(fieldName).Code("\"")
		dst.Code(" label={_ctx.$t(\"").Code(langName).Code("Lang.").Code(fieldName).Code("\")}")
		dst.Code(" show-overflow-tooltip")
		dst.Code(" min-width=\"").Code(strconv.FormatFloat(table.width, 'g', -1, 64)).Code("\"")
		dst.Code(">\n")
		if "image" == table.table {
			dst.Code("                    {{\n")
			dst.Code("                        default: (scope) => (<>\n")
			dst.Code("                            <el-popover effect=\"light\" trigger=\"hover\" placement=\"top\" width=\"auto\">\n")
			dst.Code("                                {{\n")
			dst.Code("                                    default: () => <el-image style=\"width: 200px; height: 200px\" src={scope.row.").Code(fieldName).Code("}/>,\n")
			dst.Code("                                    reference: () => <el-avatar shape=\"square\" size=\"60\" src={scope.row.").Code(fieldName).Code("}/>,\n")
			dst.Code("                                }}\n")
			dst.Code("                            </el-popover>\n")
			dst.Code("                        </>)\n")
			dst.Code("                    }}\n")
		}
		dst.Code("\t\t\t\t</el-table-column>\n")
		lang.Add(fieldName, field.Tags)
		//
		//dst.Code("\"] = TablesColumn(\n")
		//dst.Code("\t\t\t\tindex: " + strconv.Itoa(index) + ",\n")
		//dst.Code("\t\t\t\theaderBuilder: (context) {\n")
		//dst.Code("\t\t\t\t\treturn TablesCell(child: Text(" + name + "Localizations.of(context)." + fieldName + "));\n")
		//dst.Code("\t\t\t\t},\n")
		//dst.Code("\t\t\t\tcellBuilder: (context, x, y, data) {\n")
		//if "image" == table.table {
		//	dst.Code("\t\t\t\t\treturn TablesCell(\n")
		//	dst.Code("\t\t\t\t\t\tchild: ((data." + fieldName)
		//	if isNull {
		//		dst.Code("?")
		//	}
		//	dst.Code(".isEmpty ")
		//	if isNull {
		//		dst.Code("?? false")
		//	}
		//	dst.Code(") ? \"\" : data." + fieldName)
		//	if isArray {
		//		if isNull {
		//			dst.Code("?")
		//		}
		//		dst.Code(".first")
		//		dst.Code("??\"\"")
		//	} else if isNull {
		//		dst.Code("??\"\"")
		//	}
		//	dst.Code(").startsWith(\"http\")\n")
		//	dst.Code("\t\t\t\t\t\t\t\t? Image.network(\n")
		//	dst.Code("\t\t\t\t\t\t\t\t\t\tdata." + fieldName)
		//	if isArray {
		//		if isNull {
		//			dst.Code("!")
		//		}
		//		dst.Code(".first")
		//		dst.Code("!")
		//	} else if isNull {
		//		dst.Code("!")
		//	}
		//	dst.Code(",\n")
		//	dst.Code("\t\t\t\t\t\t\t\t\t\tfit: BoxFit.contain,\n")
		//	dst.Code("\t\t\t\t\t\t\t\t\t)\n")
		//	dst.Code("\t\t\t\t\t\t\t\t: const SizedBox(),\n")
		//	dst.Code("\t\t\t\t\t);\n")
		//} else {
		//	dst.Code("\t\t\t\treturn TablesCell(\n")
		//	dst.Code("\t\t\t\t\tchild: Tooltip(\n")
		//	dst.Code("\t\t\t\t\t\tmessage: data." + fieldName)
		//	if isArray {
		//		if isNull {
		//			dst.Code("?")
		//		}
		//		dst.Code(".map((e) => e")
		//		b.printToString(dst, field.Type, false, table.digit, table.format, "??\"\"")
		//		dst.Code(").join(',')")
		//		if isNull {
		//			dst.Code(" ?? ''")
		//		}
		//	} else {
		//		b.printToString(dst, field.Type, false, table.digit, table.format, "??\"\"")
		//	}
		//	dst.Code(",\n")
		//	dst.Code("\t\t\t\t\t\tchild: Text(\n")
		//	dst.Code("\t\t\t\t\t\t\tdata." + fieldName)
		//	if isArray {
		//		if isNull {
		//			dst.Code("?")
		//		}
		//		dst.Code(".map((e) => e")
		//		b.printToString(dst, field.Type, false, table.digit, table.format, "??\"\"")
		//		dst.Code(").join(',')")
		//		if isNull {
		//			dst.Code(" ?? ''")
		//		}
		//	} else {
		//		b.printToString(dst, field.Type, false, table.digit, table.format, "??\"\"")
		//	}
		//	dst.Code(",\n")
		//	dst.Code("\t\t\t\t\t\t\tmaxLines: 1,\n")
		//	dst.Code("\t\t\t\t\t\t\toverflow: TextOverflow.ellipsis,\n")
		//	dst.Code("\t\t\t\t\t\t),\n")
		//	dst.Code("\t\t\t\t\t),\n")
		//	dst.Code("\t\t\t\t);\n")
		//}
		//dst.Code("\t\t\t\t},\n")
		//dst.Code("\t\t\t);\n")
		//lang.Add(fieldName, field.Tags)
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("\t\t\t</>\n")
	dst.Code("\t\t);\n")
	dst.Code("\t}\n")
	dst.Code("});\n\n")

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

	dst.Code("export const " + name + "FormItems = defineComponent({\n")
	dst.Code("\tname: '" + name + "FormItems',\n")
	dst.Code("\tprops: {\n")
	dst.Code("\t\tsize: String,\n")
	dst.Code("\t\tmodel: ")
	b.printType(dst, typ.Name, false, false)
	dst.Code("\n")
	dst.Code("\t},\n")
	dst.Code("\tsetup(props) {\n")
	dst.Code("\t\treturn (_ctx) => (\n")
	dst.Code("\t\t\t<>\n")
	langName := build.StringToFirstLower(name)
	//i := 0
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {

		form := b.getUI(field.Tags)
		if nil == form || 0 == len(form.form) {
			return nil
		}

		//isArray := build.IsArray(field.Type)
		isNull := build.IsNil(field.Type)
		//i++
		//index := i
		if nil != form.index {
			//index = *table.index
		}
		fieldName := build.StringToFirstLower(field.Name.Name)

		dst.Code("\t\t\t\t<el-form-item prop=\"").Code(fieldName).Code("\"")
		dst.Code(" label={_ctx.$t(\"").Code(langName).Code("Lang.").Code(fieldName).Code("\")}")
		dst.Code(">\n")
		if "date" == form.form {
			dst.Code("\t\t\t\t\t<el-date-picker\n")
			dst.Code("\t\t\t\t\t\tv-model={props.model!.").Code(fieldName).Code("}\n")
			dst.Code("\t\t\t\t\t\ttype=\"daterange\"\n")
			dst.Code("\t\t\t\t\t\tunlink-panels\n")
			//dst.Code("\t\t\t\t\t:shortcuts=\"shortcuts\"\n")
			dst.Code("\t\t\t\t\t\tsize={props.size}\n")
			dst.Code("\t\t\t\t\t/>\n")
			lang.Add(fieldName, field.Tags)
		} else if "menu" == form.form {
			dst.Code("\t\t\t\t\t<el-select v-model={props.model!.").Code(fieldName).Code("}\n")
			dst.Code("\t\t\t\t\t\tplaceholder=\"Select\"\n")
			dst.Code("\t\t\t\t\t\tstyle=\"width:180px\"\n")
			dst.Code("\t\t\t\t\t\tsize={props.size}\n")
			dst.Code("\t\t\t\t\t\t>\n")
			b.printMenuItem(dst, field.Type, false)
			dst.Code("\t\t\t\t\t</el-select>\n")
			lang.Add(fieldName, field.Tags)
		} else {
			dst.Code("\t\t\t\t\t<el-input v-model={props.model!.").Code(fieldName).Code("}")
			if isNull {
				dst.Code(" clearable")
			}
			dst.Code(" size={props.size}")
			dst.Code("/>\n")
			lang.Add(fieldName, field.Tags)
		}

		dst.Code("\t\t\t\t</el-form-item>\n")

		return nil
	})
	if err != nil {
		return
	}

	dst.Code("\t\t\t</>\n")
	dst.Code("\t\t);\n")
	dst.Code("\t}\n")
	dst.Code("});\n\n")

	//name := build.StringToHumpName(typ.Name.Name)
	//lang := dst.GetLang(name)
	//dst.Code("class Form" + name + build.StringToHumpName(u.suffix) + "Build {\n")
	//dst.Code("\tfinal " + name + " info;\n\n")
	//dst.Code("\tfinal bool readOnly;\n\n")
	//dst.Code("\tfinal Map<double, int> sizes;\n\n")
	//dst.Code("\tfinal EdgeInsetsGeometry padding;\n\n")
	//
	//dst.Code("\tForm" + name + build.StringToHumpName(u.suffix) + "Build (\n")
	//dst.Code("\t\tthis.info, {\n")
	//dst.Code("\t\tthis.readOnly = false,\n")
	//dst.Code("\t\tthis.sizes = const {},\n")
	//dst.Code("\t\tthis.padding = const EdgeInsets.only(),\n")
	//dst.Code("\t});\n\n")
	//
	//fields := build.NewWriter()
	//setValue := build.NewWriter()
	//err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
	//	form := b.getUI(field.Tags)
	//	fieldName := build.StringToFirstLower(field.Name.Name)
	//	if nil == form {
	//		return nil
	//	}
	//
	//	onlyRead := "false"
	//	if form.onlyRead {
	//		onlyRead = "true"
	//	}
	//	isArray := build.IsArray(field.Type)
	//	isNil := build.IsNil(field.Type)
	//
	//	verify, err := build.GetVerify(field.Tags, dst.File, b.GetDataType)
	//	if err != nil {
	//		return err
	//	}
	//	if "text" == form.form {
	//		dst.Code("\tfinal TextFormBuild " + fieldName + " = TextFormBuild();\n\n")
	//		if isArray {
	//			setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
	//			if isNil {
	//				setValue.Code("?")
	//			}
	//			setValue.Code(".map((e) => e")
	//			b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
	//			setValue.Code(").join(\",\");\n")
	//			if !form.onlyRead {
	//				setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
	//				setValue.Code("val?.split(\",\").map((e) =>e")
	//				b.printFormString(setValue, "val", field.Type, false, form.digit, form.format)
	//				setValue.Code(").toList()")
	//				if !isNil {
	//					setValue.Code(" ?? [] ")
	//				}
	//				setValue.Code(";\n")
	//			}
	//		} else {
	//			setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
	//			b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
	//			setValue.Code(";\n")
	//			if !form.onlyRead {
	//				setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
	//				if form.toNull {
	//					setValue.Code("\"\" == val ? null : ")
	//				}
	//				b.printFormString(setValue, "val", field.Type, false, form.digit, form.format)
	//				setValue.Code(";\n")
	//			}
	//		}
	//
	//		setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
	//		setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
	//		setValue.Code("\t\t" + fieldName + ".maxLines = " + strconv.Itoa(form.maxLine) + ";\n")
	//		setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
	//		if nil != verify {
	//			b.getPackage(dst, typ.Name, "verify")
	//			setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val!);\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")
	//
	//		fields.Code("\t\t\t" + fieldName + ".build(context),\n")
	//		lang.Add(fieldName, field.Tags)
	//	} else if "click" == form.form {
	//		dst.Code("\tfinal ClickFormBuild<")
	//		b.printType(dst, field.Type, false)
	//		dst.Code("> " + fieldName + " = ClickFormBuild();\n\n")
	//
	//		setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName + ";\n")
	//		if !form.onlyRead {
	//			setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val")
	//			if !build.IsNil(field.Type) {
	//				setValue.Code("!")
	//			}
	//			setValue.Code(";\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
	//		setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
	//		setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
	//		if nil != verify {
	//			b.getPackage(dst, typ.Name, "verify")
	//			setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val!);\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")
	//
	//		fields.Code("\t\t\t" + fieldName + ".build(context),\n")
	//		lang.Add(fieldName, field.Tags)
	//	} else if "file" == form.form {
	//		dst.Code("\tfinal FileFormBuild " + fieldName + " = FileFormBuild();\n\n")
	//		setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
	//		b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
	//		setValue.Code(";\n")
	//		if !form.onlyRead {
	//			setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
	//			if form.toNull {
	//				setValue.Code("\"\" == val ? null : ")
	//			}
	//			b.printFormString(setValue, "val", field.Type, false, form.digit, form.format)
	//			setValue.Code(";\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
	//		setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
	//		setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
	//		if 0 < len(form.extensions) {
	//			setValue.Code("\t\t" + fieldName + ".extensions = <String>[")
	//			for i, extension := range form.extensions {
	//				if 0 < i {
	//					setValue.Code(", ")
	//				}
	//				setValue.Code("\"" + extension + "\"")
	//			}
	//			setValue.Code("];\n")
	//		}
	//		if nil != verify {
	//			b.getPackage(dst, typ.Name, "verify")
	//			setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val!);\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")
	//
	//		fields.Code("\t\t\t" + fieldName + ".build(context),\n")
	//		lang.Add(fieldName, field.Tags)
	//	} else if "image" == form.form {
	//		dst.Code("\tfinal ImageFormBuild " + fieldName + " =  ImageFormBuild();\n\n")
	//
	//		if isArray {
	//			if build.IsNil(field.Type) {
	//				setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName + "?.map((e) => ImageFormImage(e))?.toList() ?? [];\n")
	//				if !form.onlyRead {
	//					setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!.map((e) => e.url).toList();\n")
	//				}
	//			} else {
	//				setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName + ".map((e) => ImageFormImage(e)).toList();\n")
	//				if !form.onlyRead {
	//					setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!.map((e) => e.url).toList();\n")
	//				}
	//			}
	//		} else {
	//			if build.IsNil(field.Type) {
	//				setValue.Code("\t\t" + fieldName + ".initialValue = [if (info." + fieldName + "?.startsWith(\"http\") ?? false) ImageFormImage(info." + fieldName + "!)];\n")
	//				if !form.onlyRead {
	//					setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ((val?.isEmpty ?? true) ? null : val!.first.url);\n")
	//				}
	//			} else {
	//				setValue.Code("\t\t" + fieldName + ".initialValue = [ImageFormImage(info." + fieldName + ")];\n")
	//				if !form.onlyRead {
	//					setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val!.first.url;\n")
	//				}
	//			}
	//		}
	//
	//		setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
	//		setValue.Code("\t\t" + fieldName + ".maxCount = " + strconv.Itoa(form.maxCount) + ";\n")
	//		setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
	//		setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
	//		if form.clip {
	//			setValue.Code("\t\t" + fieldName + ".clip = true;\n")
	//		} else {
	//			setValue.Code("\t\t" + fieldName + ".clip = false;\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".outWidth = " + strconv.FormatFloat(form.width, 'G', -1, 64) + ";\n")
	//		setValue.Code("\t\t" + fieldName + ".outHeight = " + strconv.FormatFloat(form.height, 'G', -1, 64) + ";\n")
	//		if 0 < len(form.extensions) {
	//			setValue.Code("\t\t" + fieldName + ".extensions = <String>[")
	//			for i, extension := range form.extensions {
	//				if 0 < i {
	//					setValue.Code(", ")
	//				}
	//				setValue.Code("\"" + extension + "\"")
	//			}
	//			setValue.Code("];\n")
	//		}
	//		if nil != verify {
	//			b.getPackage(dst, typ.Name, "verify")
	//			setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val!);\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")
	//		fields.Code("\t\t\t" + fieldName + ".build(context),\n")
	//		lang.Add(fieldName, field.Tags)
	//
	//	} else if "menu" == form.form {
	//		dst.Code("\tfinal MenuFormBuild<")
	//		b.printType(dst, field.Type, false)
	//		dst.Code("> " + fieldName + " = MenuFormBuild();\n\n")
	//		setValue.Code("\t\t" + fieldName + ".value = info." + fieldName + ";\n")
	//		if !form.onlyRead {
	//			setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = val")
	//			if !build.IsNil(field.Type) {
	//				setValue.Code("!")
	//			}
	//			setValue.Code(";\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
	//		setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
	//		setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
	//		if form.toNull {
	//			setValue.Code("\t\t" + fieldName + ".toNull = true;\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n")
	//		if nil != verify {
	//			b.getPackage(dst, typ.Name, "verify")
	//			setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val?.toString());\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".items = [\n")
	//		b.printMenuItem(setValue, field.Type, false)
	//		setValue.Code("\t\t];\n\n")
	//
	//		fields.Code("\t\t\t" + fieldName + ".build(context),\n")
	//		lang.Add(fieldName, field.Tags)

	//	} else if "switch" == form.form {
	//		dst.Code("\tfinal SwitchFormBuild " + fieldName + " =  SwitchFormBuild();\n\n")
	//		setValue.Code("\t\t" + fieldName + ".initialValue = info." + fieldName)
	//		//b.printToString(setValue, field.Type, false, form.digit, form.format, "??\"\"")
	//		setValue.Code(";\n")
	//		if !form.onlyRead {
	//			setValue.Code("\t\t" + fieldName + ".onSaved = (val) => info." + fieldName + " = ")
	//			if !build.IsNil(field.Type) {
	//				setValue.Code("val ?? false ;\n")
	//			} else {
	//				setValue.Code("val ;\n")
	//			}
	//
	//		}
	//		setValue.Code("\t\t" + fieldName + ".readOnly = readOnly || " + onlyRead + ";\n")
	//		setValue.Code("\t\t" + fieldName + ".widthSizes = sizes;\n")
	//		setValue.Code("\t\t" + fieldName + ".padding = padding;\n")
	//		if nil != verify {
	//			b.getPackage(dst, typ.Name, "verify")
	//			setValue.Code("\t\t" + fieldName + ".validator = (val) => verify" + name + "_" + build.StringToHumpName(fieldName) + "(context, val!);\n")
	//		}
	//		setValue.Code("\t\t" + fieldName + ".decoration = InputDecoration(labelText: " + name + "Localizations.of(context)." + fieldName + ");\n\n")
	//
	//		fields.Code("\t\t\t" + fieldName + ".build(context),\n")
	//		lang.Add(fieldName, field.Tags)
	//	}
	//
	//	return nil
	//})
	//if err != nil {
	//	return
	//}
	//
	//dst.Code("\tList<Widget> build(BuildContext context, {Function(BuildContext context, Form" + name + build.StringToHumpName(u.suffix) + "Build ui)? builder}) {\n")
	//dst.Code(setValue.String())
	//dst.Code("\t\tbuilder?.call(context, this);\n")
	//dst.Code("\t\treturn <Widget>[\n")
	//dst.Code(fields.String())
	//dst.Code("\t\t];\n")
	//dst.Code("\t}\n")
	//dst.Code("}\n\n")
	//
	//dst.ImportByWriter(setValue)
}

func (b *Builder) printMenuItem(dst *build.Writer, expr ast.Expr, empty bool) {
	switch expr.(type) {
	case *ast.EnumType:
		t := expr.(*ast.EnumType)
		pkg := b.getPackage(dst, t.Name, "")
		name := build.StringToHumpName(t.Name.Name)
		dst.Code("\t\t\t\t\t\t{").Code(pkg).Code(".").Code(name).Code(".values.map((val) => {\n")
		dst.Code("\t\t\t\t\t\t\treturn <el-option key={val.value}\n")
		dst.Code("\t\t\t\t\t\t\t\tlabel={_ctx.$t(val.toString())}\n")
		dst.Code("\t\t\t\t\t\t\t\tvalue={val}\n")
		dst.Code("\t\t\t\t\t\t\t/>\n")
		dst.Code("\t\t\t\t\t\t})}\n")
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
