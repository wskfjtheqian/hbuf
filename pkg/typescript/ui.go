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
	dst.Code("\tsetup(props:any) {\n")
	dst.Code("\t\treturn (_ctx: Record<string, any>) => (\n")
	dst.Code("\t\t\t<>\n")
	langName := build.StringToFirstLower(name)
	//i := 0
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {

		table := b.getUI(field.Tags)
		if nil == table || 0 == len(table.table) {
			return nil
		}

		//isEnum := build.IsEnum(field.Type)
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
			dst.Code("\t\t\t\t\t{{\n")
			dst.Code("\t\t\t\t\t\tdefault: (scope:any) => (<>\n")
			dst.Code("\t\t\t\t\t\t\t<el-popover effect=\"light\" trigger=\"hover\" placement=\"top\" width=\"auto\">\n")
			dst.Code("\t\t\t\t\t\t\t\t{{\n")
			dst.Code("\t\t\t\t\t\t\t\t\tdefault: () => <el-image style=\"width: 200px; height: 200px\" src={scope.row.").Code(fieldName).Code("}/>,\n")
			dst.Code("\t\t\t\t\t\t\t\t\treference: () => <el-avatar shape=\"square\" size=\"60\" src={scope.row.").Code(fieldName).Code("}/>,\n")
			dst.Code("\t\t\t\t\t\t\t\t}}\n")
			dst.Code("\t\t\t\t\t\t\t</el-popover>\n")
			dst.Code("\t\t\t\t\t\t</>)\n")
			dst.Code("\t\t\t\t\t}}\n")
		} else if "switch" == table.table {
			dst.Code("\t\t\t\t\t{{\n")
			dst.Code("\t\t\t\t\t\tdefault: (scope:any) => (<el-switch v-model={scope.row!.").Code(fieldName).Code("} disabled />)\n")
			dst.Code("\t\t\t\t\t}}\n")
		} else {
			dst.Code("\t\t\t\t\t{{\n")
			dst.Code("\t\t\t\t\t\tdefault: (scope:any) =>")
			b.printToString(dst, "scope.row."+fieldName, field.Type, false, table.digit, table.format, " || \"\"")
			dst.Code("\n")
			dst.Code("\t\t\t\t\t}}\n")
		}
		dst.Code("\t\t\t\t</el-table-column>\n")
		lang.Add(fieldName, field.Tags)
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

func (b *Builder) printToString(dst *build.Writer, name string, expr ast.Expr, empty bool, digit int, format string, val string) {
	switch expr.(type) {
	case *ast.EnumType:
		if empty {
			dst.Code("null == ").Code(name).Code(" ? \"\" : ")
		}
		dst.Code("_ctx.$t(").Code(name).Code("!.toString()").Code(")")
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			b.getPackage(dst, expr, "")
			b.printToString(dst, name, t.Obj.Decl.(*ast.TypeSpec).Type, empty, digit, format, val)
		} else {
			switch build.BaseType(t.Name) {
			case build.Int8, build.Int16, build.Int32, build.Int64, build.Uint8, build.Uint16, build.Uint32, build.Uint64:
				if empty {
					dst.Code("null == ").Code(name).Code(" ? \"\" : ")
				}
				dst.Code(name).Code("!.toString()")
			case build.Float, build.Double:
				if empty {
					dst.Code("null == ").Code(name).Code(" ? \"\" : ")
				}
				dst.Code(name).Code("!.toFixed(" + strconv.Itoa(digit) + ")")
			case build.Bool:
				if empty {
					dst.Code("null == ").Code(name).Code(" ? \"\" : ")
				}
				dst.Code(name).Code("!.toString()")
			case build.Date:
				if empty {
					dst.Code("null == ").Code(name).Code(" ? \"\" : ")
				}
				dst.Import("hbuf_ts", "* as h")
				if 0 == len(format) {
					format = "yyyy/MM/dd HH:mm:ss"
				}
				dst.Code("_ctx.$fd(").Code(name).Code(",\"").Code(format).Code("\")")
			case build.Decimal:
				if empty {
					dst.Code("null == ").Code(name).Code(" ? \"\" : ")
				}
				dst.Code(name).Code("!.toFixed(" + strconv.Itoa(digit) + ")")
			default:
				dst.Code(name)
			}
		}
	case *ast.ArrayType:
		dst.Code("\"\"+").Code(name)
	case *ast.MapType:
		dst.Code("\"\"+").Code(name)
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printToString(dst, name, t.Type(), t.Empty, digit, format, val)
		if t.Empty {
			dst.Code(val)
		}
	default:
		dst.Code("\"\"+").Code(name)
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
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code("(/^[0-9-]+$/.test(").Code(name).Code(") ? Number.parseInt(" + name + ") :").Code(name).Code(")")
			case build.Uint64, build.Int64:
				dst.Import("long", "Long")
				if empty {
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code("(/^[0-9-]+$/.test(").Code(name).Code(") ? Long.fromValue(" + name + ") :").Code(name).Code(")")
			case build.Float, build.Double:
				if empty {
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code("(/([-+]?\\d+)(\\.\\d+)?/.test(").Code(name).Code(") ? Number.parseFloat(" + name + ") :").Code(name).Code(")")
			case build.Bool:
				dst.Code("\"true\" == " + name)
			case build.Date:
				if empty {
					dst.Code(name + "==null ? null : DateTime.tryParse(" + name + ")")
				} else {
					dst.Code("DateTime.tryParse(" + name + "!) ?? DateTime.now()")
				}
			case build.Decimal:
				dst.Import("decimal.js", "* as d")
				if empty {
					dst.Code(name + " == null ? null : ")
				}
				dst.Code("function () {try {return new d.Decimal(").Code(name).Code("!)} catch (e) {return ").Code(name).Code("}}()")
			default:
				if empty {
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code(name)
			}
		}
	case *ast.ArrayType:
		dst.Code(name)
		//ar := expr.(*ast.ArrayType)
		//dst.Code("List<")
		//printType(dst, ar.VType, false)
		//dst.Code(">")
		//if ar.Empty && !notEmpty {
		//	dst.Code("?")
		//}
	case *ast.MapType:
		dst.Code(name)
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
	dst.Code("\tsetup(props: Record<string, any>) {\n")
	dst.Import("element-plus", "{useLocale}")
	dst.Tab(2).Code("const locale = useLocale()\n")
	dst.Code("\t\treturn (_ctx: Record<string, any>) => (\n")
	dst.Code("\t\t\t<>\n")
	langName := build.StringToFirstLower(name)
	//i := 0
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {

		form := b.getUI(field.Tags)
		if nil == form || 0 == len(form.form) {
			return nil
		}

		isArray := build.IsArray(field.Type)
		isNull := build.IsNil(field.Type)
		_, verify := build.GetTag(field.Tags, "verify")
		//i++
		//index := i
		if nil != form.index {
			//index = *table.index
		}
		fieldName := build.StringToFirstLower(field.Name.Name)

		dst.Code("\t\t\t\t<el-form-item prop=\"").Code(fieldName).Code("\"")
		dst.Code(" label={_ctx.$t(\"").Code(langName).Code("Lang.").Code(fieldName).Code("\")}")
		if verify {
			pName := b.getPackage(dst, typ.Name, "verify")
			dst.Code(" rules={[{validator: ").Code(pName).Code(".verify").Code(name).Code("_").Code(build.StringToHumpName(field.Name.Name)).Code("(locale), trigger: 'blur'}]}")
		}
		dst.Code(">\n")
		if "date" == form.form {
			dst.Code("\t\t\t\t\t<el-date-picker\n")
			dst.Code("\t\t\t\t\t\tv-model={props.model!.").Code(fieldName).Code("}\n")
			dst.Code("\t\t\t\t\t\ttype=\"daterange\"\n")
			dst.Code("\t\t\t\t\t\tunlink-panels\n")
			dst.Code("\t\t\t\t\t\tsize={props.size}\n")
			if isNull {
				dst.Code("\t\t\t\t\t\tclearable\n")
			}
			if form.onlyRead {
				dst.Code(" disabled  \n")
			}
			dst.Code("\t\t\t\t\t/>\n")
		} else if "menu" == form.form {
			dst.Code("\t\t\t\t\t<el-select v-model={props.model!.").Code(fieldName).Code("}\n")
			dst.Code("\t\t\t\t\t\tstyle=\"width:180px\"\n")
			dst.Code("\t\t\t\t\t\tsize={props.size}\n")
			if isNull {
				dst.Code("\t\t\t\t\t\tclearable\n")
			}
			if form.onlyRead {
				dst.Code("\t\t\t\t\t\tdisabled  \n")
			}
			dst.Code("\t\t\t\t\t\t>\n")
			b.printMenuItem(dst, field.Type, false)
			dst.Code("\t\t\t\t\t</el-select>\n")
		} else if "switch" == form.form {
			dst.Code("\t\t\t\t\t<el-switch v-model={props.model!.").Code(fieldName).Code("}")

			if form.onlyRead {
				dst.Code(" disabled")
			}
			dst.Code("/>\n")
			//} else if isNumber {
			//	dst.Code("\t\t\t\t\t<el-input-number\n")
			//	dst.Code("\t\t\t\t\t\tv-model={props.model!.").Code(fieldName).Code("}\n")
			//	dst.Code("\t\t\t\t\t\tsize={props.size}\n")
			//	dst.Code("\t\t\t\t\t\tcontrols-position=\"right\"\n")
			//	dst.Code("\t\t\t\t\t\tprecision=\"").Code(strconv.Itoa(form.digit)).Code("\"\n")
			//	if isNull {
			//		dst.Code("\t\t\t\t\t\tclearable\n")
			//	}
			//	if form.onlyRead {
			//		dst.Code(" disabled\n")
			//	}
			//	dst.Code("\t\t\t\t\t/>\n")
		} else if "pass" == form.form {
			dst.Code("\t\t\t\t\t<el-input v-model={props.model!.").Code(fieldName).Code("}")
			dst.Code(" size={props.size}")
			dst.Code(" type=\"password\"")
			dst.Code(" show-password")
			if isNull {
				dst.Code(" clearable")
			}
			if form.onlyRead {
				dst.Code(" disabled")
			}
			dst.Code("/>\n")
		} else if isArray {
			dst.Code("\t\t\t\t\t<el-select\n")
			dst.Code("\t\t\t\t\t\tv-model={props.model!.").Code(fieldName).Code("}\n")
			dst.Code("\t\t\t\t\t\tmultiple\n")
			dst.Code("\t\t\t\t\t\tfilterable\n")
			dst.Code("\t\t\t\t\t\tallow-create\n")
			dst.Code("\t\t\t\t\t\tdefault-first-option\n")
			dst.Code("\t\t\t\t\t\treserve-keyword={false}\n")
			if isNull {
				dst.Code("\t\t\t\t\t\tclearable\n")
			}
			if form.onlyRead {
				dst.Code(" disabled\n")
			}
			dst.Code("\t\t\t\t\t>\n")
			dst.Code("\t\t\t\t\t</el-select>\n")
		} else {
			dst.Tab(5).Code("<el-input ")
			dst.Code("modelValue={")
			b.printToString(dst, "props.model!."+fieldName, field.Type, false, form.digit, form.format, " ?? \"\"")
			dst.Code("}\n")

			dst.Tab(7).Code("onUpdate:modelValue={$event=> props.model!.").Code(fieldName).Code(" = ")
			b.printFormString(dst, "$event", field.Type, false, form.digit, form.format)
			dst.Code("}\n")
			dst.Tab(7).Code("size={props.size}\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(7).Code("precision=\"").Code(strconv.Itoa(form.digit)).Code("\"\n")
			dst.Tab(5).Code("/>\n")
		}
		lang.Add(fieldName, field.Tags)
		dst.Tab(4).Code("</el-form-item>\n")

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
