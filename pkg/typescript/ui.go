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
	unlink     bool
	maxCount   int
	step       *float64
	min        *float64
	max        *float64
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
			} else if "unlink" == item.Name.Name {
				form.unlink = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]
			} else if "extensions" == item.Name.Name {
				for _, value := range item.Values {
					form.extensions = append(form.extensions, value.Value[1:len(value.Value)-1])
				}
			} else if "min" == item.Name.Name {
				atoi, err := strconv.ParseFloat(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10)
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.min = &atoi
			} else if "max" == item.Name.Name {
				atoi, err := strconv.ParseFloat(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10)
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.max = &atoi
			} else if "step" == item.Name.Name {
				atoi, err := strconv.ParseFloat(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10)
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.step = &atoi
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
	dst.Code("\tprops: {\n")
	dst.Code("\t\tposition: Array<String>,\n")
	dst.Code("\t\tfilter: (Function as unknown) as () => (item: string) => boolean,\n")
	dst.Code("\t},\n")
	dst.Code("\tsetup(props:any) {\n")
	dst.Code("\t\treturn (_ctx: Record<string, any>) => {\n")
	dst.Code("\t\t\tconst maps: Record<string, any> = {\n")
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
		dst.Tab(4).Code("\"").Code(fieldName).Code("\": () =>(\n")

		dst.Code("\t\t\t\t\t<el-table-column prop=\"").Code(fieldName).Code("\"")
		dst.Code(" label={_ctx.$t(\"").Code(langName).Code("Lang.").Code(fieldName).Code("\")}")
		dst.Code(" show-overflow-tooltip")
		dst.Code(" min-width=\"").Code(strconv.FormatFloat(table.width, 'g', -1, 64)).Code("\"")
		dst.Code(">\n")
		if "image" == table.table {
			dst.Code("\t\t\t\t\t\t{{\n")
			dst.Code("\t\t\t\t\t\t\tdefault: (scope:any) => (<>\n")
			dst.Code("\t\t\t\t\t\t\t\t<el-popover effect=\"light\" trigger=\"hover\" placement=\"top\" width=\"auto\">\n")
			dst.Code("\t\t\t\t\t\t\t\t\t\t{{\n")
			dst.Code("\t\t\t\t\t\t\t\t\t\t\tdefault: () => <el-image style={\"width: 200px; height: 200px\"} src={scope.row.").Code(fieldName).Code("}/>,\n")
			dst.Code("\t\t\t\t\t\t\t\t\t\t\treference: () => <el-avatar shape=\"square\" size={200} src={scope.row.").Code(fieldName).Code("}/>,\n")
			dst.Code("\t\t\t\t\t\t\t\t\t\t}}\n")
			dst.Code("\t\t\t\t\t\t\t\t</el-popover>\n")
			dst.Code("\t\t\t\t\t\t\t</>)\n")
			dst.Code("\t\t\t\t\t\t}}\n")
		} else if "switch" == table.table {
			dst.Code("\t\t\t\t\t\t{{\n")
			dst.Code("\t\t\t\t\t\t\tdefault: (scope:any) => (<el-switch v-model={scope.row!.").Code(fieldName).Code("} disabled />)\n")
			dst.Code("\t\t\t\t\t\t}}\n")
		} else {
			dst.Code("\t\t\t\t\t\t{{\n")
			dst.Code("\t\t\t\t\t\t\tdefault: (scope:any) =>")
			b.printToString(dst, "scope.row."+fieldName, field.Type, false, table.digit, table.format, " || \"\"")
			dst.Code("\n")
			dst.Code("\t\t\t\t\t\t}}\n")
		}
		dst.Code("\t\t\t\t\t</el-table-column>\n")
		dst.Tab(4).Code("),\n")
		lang.Add(fieldName, field.Tags)
		return nil
	})
	if err != nil {
		return
	}

	dst.Tab(3).Code("}\n")
	dst.Tab(3).Code("for (const key in _ctx.$slots) {\n")
	dst.Code("                maps[key] = _ctx.$slots[key]\n")
	dst.Code("            }\n")
	dst.Code("            const list: string[] = []\n")
	dst.Code("            for (const i in _ctx.position) {\n")
	dst.Code("                const key = _ctx.position[i]\n")
	dst.Code("                if (!list.includes(key) && maps[key]) {\n")
	dst.Code("                    list.push(key)\n")
	dst.Code("                }\n")
	dst.Code("            }\n")
	dst.Code("            for (const key in maps) {\n")
	dst.Code("                if (!list.includes(key) && maps[key] && (!_ctx.filter || _ctx.filter(key))) {\n")
	dst.Code("                    list.push(key)\n")
	dst.Code("                }\n")
	dst.Code("            }\n")
	dst.Code("            return list.map((it) => maps[it]())\n")
	dst.Code("\t\t};\n")
	dst.Code("\t}\n")
	dst.Code("});\n\n")

}

func (b *Builder) printToString(dst *build.Writer, name string, expr ast.Expr, empty bool, digit int, format string, val string) {
	switch expr.(type) {
	case *ast.EnumType:
		if empty {
			dst.Code("null == ").Code(name).Code(" ? \"\" : ")
		}
		dst.Code("_ctx.$t(").Code(name).Code("?.toString()").Code(")")
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
				dst.Code("_ctx.$t(").Code(name).Code("?.toString())")
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
		if empty {
			dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
		}
		p := b.getPackage(dst, t.Name, "")
		dst.Code("(").Code(p).Code(".").Code(t.Name.Name).Code(".nameOf(").Code(name).Code("))")
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			b.getPackage(dst, expr, "")
			b.printFormString(dst, name, t.Obj.Decl.(*ast.TypeSpec).Type, empty, digit, format)
		} else {
			switch build.BaseType(t.Name) {
			case build.Int8, build.Int16, build.Int32:
				if empty {
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code("(/^[-+]?[0-9]+$/.test(").Code(name).Code(") ? Number.parseInt(" + name + ") :").Code(name).Code(")")
			case build.Uint8, build.Uint16, build.Uint32:
				if empty {
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code("(/^[+]?[0-9]+$/.test(").Code(name).Code(") ? Number.parseInt(" + name + ") :").Code(name).Code(")")
			case build.Int64:
				dst.Import("long", "Long")
				if empty {
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code("(/^[-+]?[0-9]+$/.test(").Code(name).Code(") ? Long.fromValue(" + name + ") :").Code(name).Code(")")
			case build.Uint64:
				dst.Import("long", "Long")
				if empty {
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code("(/^[+]?[0-9]+$/.test(").Code(name).Code(") ? Long.fromValue(" + name + ") :").Code(name).Code(")")
			case build.Float, build.Double:
				if empty {
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code("(/([-+]?\\d+)(\\.\\d+)?/.test(").Code(name).Code(") ? Number.parseFloat(" + name + ") :").Code(name).Code(")")
			case build.Bool:
				if empty {
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code("(\"true\" == ").Code(name).Code(")")
			case build.Date:
				if empty {
					dst.Code("(").Code(name).Code(" == null || ").Code(name).Code(".length == 0)").Code(" ? null : ")
				}
				dst.Code("Date.parse(" + name + ") ?? ").Code(name)
			case build.Decimal:
				dst.Import("decimal.js", "* as d")
				if empty {
					dst.Code(name + " == null ? null : ")
				}
				dst.Code("function () {try {return new d.Decimal(").Code(name).Code(")} catch (e) {return null}}()")
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
	dst.Code("\t\tposition: Array<String>,\n")
	dst.Code("\t\tfilter: (Function as unknown) as () => (item: string) => boolean,\n")
	dst.Code("\t\tmodel: ")
	b.printType(dst, typ.Name, false, false)
	dst.Code("\n")
	dst.Code("\t},\n")
	dst.Code("\tsetup(props: Record<string, any>) {\n")
	dst.Import("element-plus", "{useLocale}")
	dst.Tab(2).Code("const locale = useLocale()\n")
	dst.Code("\t\treturn (_ctx: Record<string, any>) => {\n")
	dst.Code("\t\t\tconst maps: Record<string, any> = {\n")
	langName := build.StringToFirstLower(name)
	//i := 0
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {

		form := b.getUI(field.Tags)
		if nil == form || 0 == len(form.form) {
			return nil
		}

		inNum := build.IsNumber(field.Type)
		isArray := build.IsArray(field.Type)
		isNull := build.IsNil(field.Type)
		_, verify := build.GetTag(field.Tags, "verify")
		//i++
		//index := i
		if nil != form.index {
			//index = *table.index
		}
		fieldName := build.StringToFirstLower(field.Name.Name)
		dst.Tab(4).Code("\"").Code(fieldName).Code("\": () =>(\n")
		dst.Tab(5).Code("<el-form-item prop=\"").Code(fieldName).Code("\"")
		dst.Code(" label={_ctx.$t(\"").Code(langName).Code("Lang.").Code(fieldName).Code("\")}")
		if verify {
			pName := b.getPackage(dst, typ.Name, "verify")
			dst.Code(" rules={[{validator: ").Code(pName).Code(".verify").Code(name).Code("_").Code(build.StringToHumpName(field.Name.Name)).Code("(locale), trigger: 'blur'}]}")
		}
		dst.Code(">\n")
		if "datetime" == form.form || "date" == form.form || "dates" == form.form || "year" == form.form || "month" == form.form {
			dst.Tab(6).Code("<el-date-picker\n")
			dst.Tab(7).Code("modelValue={")
			if isArray {
				dst.Import("hbuf_ts", "* as h")
				dst.Code(" h.convertArray(_ctx.model!.").Code(fieldName).Code(", (e) => _ctx.$timeToLocal(e))")
			} else {
				dst.Code("_ctx.$timeToLocal(_ctx.model!.").Code(fieldName).Code(")")
			}
			dst.Code("}\n")
			dst.Tab(7).Code("onUpdate:modelValue={($event: (number | string | Date) | (number | string | Date)[] | null) => _ctx.model!.").Code(fieldName).Code(" = ")
			if isNull {
				dst.Code("(!$event) ? null : ")
			}
			if isArray {
				dst.Import("hbuf_ts", "* as h")
				if "datetime" == form.form {
					dst.Code("h.convertArray($event, (e) => _ctx.$timeToUtc(e))")
				} else if "month" == form.form {
					dst.Code("h.convertArray([($event! as Date[])[0] as Date, new Date((($event! as Date[])[1] as Date).getFullYear(), (($event! as Date[])[1] as Date).getMonth() + 1, 0, 23, 59, 59, 999)], (e) => _ctx.$timeToUtc(e))")
				} else if "dates" == form.form {
					dst.Code("h.convertArray($event, (e) => _ctx.$timeToUtc(e))")
				} else if "year" == form.form {
					dst.Code("h.convertArray($event, (e) => _ctx.$timeToUtc(e))")
				} else {
					dst.Code("h.convertArray([($event! as Date[])[0] as Date, new Date((($event! as Date[])[1] as Date).setHours(23,59,59,999))], (e) => _ctx.$timeToUtc(e))")
				}
			} else {
				if "datetime" == form.form {
					dst.Code("_ctx.$timeToUtc($event)")
				} else if "month" == form.form {
					dst.Code("_ctx.$timeToUtc($event)")
				} else if "dates" == form.form {
					dst.Code("($event?.length ?? 0 == 0) ? null :_ctx.$timeToUtc($event[0])")
				} else if "year" == form.form {
					dst.Code("_ctx.$timeToUtc($event)")
				} else {
					dst.Code("_ctx.$timeToUtc($event)")
				}
			}
			dst.Code("}\n")
			dst.Tab(7).Code("type=\"")
			if isArray {
				if "datetime" == form.form {
					dst.Code("datetimerange")
				} else if "month" == form.form {
					dst.Code("monthrange")
				} else if "dates" == form.form {
					dst.Code("dates")
				} else if "year" == form.form {
					dst.Code("years")
				} else {
					dst.Code("daterange")
				}
			} else {
				if "datetime" == form.form {
					dst.Code("datetime")
				} else if "month" == form.form {
					dst.Code("month")
				} else if "dates" == form.form {
					dst.Code("dates")
				} else if "year" == form.form {
					dst.Code("year")
				} else {
					dst.Code("date")
				}
			}
			dst.Code("\"\n")
			if form.unlink {
				dst.Tab(7).Code("unlink-panels\n")
			}
			dst.Tab(7).Code("shortcuts={_ctx.$datePackerShortcuts(_ctx.$t)}\n")
			dst.Tab(7).Code("size={props.size}\n")
			dst.Tab(7).Code("clearable=")
			if isNull {
				dst.Code("{true}\n")
			} else {
				dst.Code("{false}\n")
			}
			if form.onlyRead {
				dst.Code(" disabled  \n")
			}
			dst.Tab(6).Code("/>\n")
		} else if "menu" == form.form {
			dst.Tab(6).Code("<el-select\n")
			b.printMenuModelValue(dst, field.Type, fieldName, false, false)

			dst.Tab(7).Code("style={\"width:180px\"}\n")
			dst.Tab(7).Code("size={props.size}\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code("disabled  \n")
			}
			dst.Tab(7).Code(">\n")
			b.printMenuItem(dst, field.Type, false, "el-option")
			dst.Tab(6).Code("</el-select>\n")
		} else if "switch" == form.form {
			dst.Code("\t\t\t\t\t<el-switch modelValue={_ctx.model!.").Code(fieldName).Code(" ??= false")
			dst.Code("}\n")

			dst.Tab(8).Code("onUpdate:modelValue={($event: string) => _ctx.model!.").Code(fieldName).Code(" = $event")
			dst.Code("}\n")

			if form.onlyRead {
				dst.Code(" disabled")
			}
			dst.Code("/>\n")
		} else if "radio" == form.form {
			dst.Tab(6).Code("<el-radio-group \n")
			b.printMenuModelValue(dst, field.Type, fieldName, false, false)

			dst.Tab(7).Code("size={props.size}\n")
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(7).Code(">\n")
			b.printMenuItem(dst, field.Type, false, "el-radio")
			dst.Tab(6).Code("</el-radio-group>\n")
		} else if "radioButton" == form.form {
			dst.Tab(6).Code("<el-radio-group\n")
			b.printMenuModelValue(dst, field.Type, fieldName, false, false)

			dst.Tab(7).Code("size={props.size}\n")
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(7).Code(">\n")
			b.printMenuItem(dst, field.Type, false, "el-radio-button")
			dst.Tab(6).Code("</el-radio-group>\n")
		} else if "pass" == form.form {
			dst.Tab(6).Code("<el-input\n")
			dst.Tab(7).Code("modelValue={")
			b.printToString(dst, "_ctx.model!."+fieldName, field.Type, false, form.digit, form.format, " ?? \"\"")
			dst.Code("}\n")

			dst.Tab(7).Code("onUpdate:modelValue={($event: string) => _ctx.model!.").Code(fieldName).Code(" = ")
			b.printFormString(dst, "$event", field.Type, false, form.digit, form.format)
			dst.Code("}\n")
			dst.Tab(7).Code("size={props.size}\n")
			dst.Tab(7).Code("type=\"password\"\n")
			dst.Tab(7).Code("show-password\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(7).Code("precision={").Code(strconv.Itoa(form.digit)).Code("}\n")
			dst.Tab(6).Code("/>\n")
		} else if isArray {
			dst.Tab(6).Code("<el-input-tag\n")
			dst.Tab(7).Code("v-model={_ctx.model!.").Code(fieldName).Code("}\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(7).Code(">\n")
			dst.Tab(6).Code("</el-input-tag>\n")
		} else if inNum {
			dst.Tab(6).Code("<el-input-number\n")
			dst.Tab(7).Code("v-model={_ctx.model!." + fieldName + "}\n")
			dst.Tab(7).Code("size={props.size}\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(7).Code("precision={").Code(strconv.Itoa(form.digit)).Code("}\n")
			if form.min != nil {
				dst.Tab(7).Code("min={").Code(strconv.FormatFloat(*form.min, 'f', -1, 64)).Code("}\n")
			}
			if form.max != nil {
				dst.Tab(7).Code("max={").Code(strconv.FormatFloat(*form.max, 'f', -1, 64)).Code("}\n")
			}
			if form.step != nil {
				dst.Tab(7).Code("step={").Code(strconv.FormatFloat(*form.step, 'f', -1, 64)).Code("}\n")
			}
			dst.Tab(6).Code("/>\n")
		} else {
			dst.Tab(6).Code("<el-input\n")
			dst.Tab(7).Code("modelValue={")
			b.printToString(dst, "_ctx.model!."+fieldName, field.Type, false, form.digit, form.format, "")
			dst.Code("}\n")

			dst.Tab(7).Code("onUpdate:modelValue={($event: string) => _ctx.model!.").Code(fieldName).Code(" = ")
			b.printFormString(dst, "$event", field.Type, false, form.digit, form.format)
			dst.Code("}\n")
			dst.Tab(7).Code("size={props.size}\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(7).Code("precision={").Code(strconv.Itoa(form.digit)).Code("}\n")
			dst.Tab(6).Code("/>\n")
		}
		lang.Add(fieldName, field.Tags)
		dst.Tab(5).Code("</el-form-item>\n")
		dst.Tab(4).Code("),\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Tab(3).Code("}\n")
	dst.Tab(3).Code("for (const key in _ctx.$slots) {\n")
	dst.Code("                maps[key] = _ctx.$slots[key]\n")
	dst.Code("            }\n")
	dst.Code("            const list: string[] = []\n")
	dst.Code("            for (const i in _ctx.position) {\n")
	dst.Code("                const key = _ctx.position[i]\n")
	dst.Code("                if (!list.includes(key) && maps[key]) {\n")
	dst.Code("                    list.push(key)\n")
	dst.Code("                }\n")
	dst.Code("            }\n")
	dst.Code("            for (const key in maps) {\n")
	dst.Code("                if (!list.includes(key) && maps[key] && (!_ctx.filter || _ctx.filter(key))) {\n")
	dst.Code("                    list.push(key)\n")
	dst.Code("                }\n")
	dst.Code("            }\n")
	dst.Code("            return list.map((it) => maps[it]())\n")
	dst.Code("\t\t};\n")
	dst.Code("\t}\n")
	dst.Code("});\n\n")
}

func (b *Builder) printMenuItem(dst *build.Writer, expr ast.Expr, empty bool, option string) {
	switch expr.(type) {
	case *ast.EnumType:
		t := expr.(*ast.EnumType)
		pkg := b.getPackage(dst, t.Name, "")
		name := build.StringToHumpName(t.Name.Name)
		dst.Code("\t\t\t\t\t\t{").Code(pkg).Code(".").Code(name).Code(".values.map((val) => {\n")
		dst.Code("\t\t\t\t\t\t\treturn <" + option + " key={val.value}\n")
		dst.Code("\t\t\t\t\t\t\t\tlabel={_ctx.$t(val.toString())}\n")
		dst.Code("\t\t\t\t\t\t\t\tvalue={val.value}\n")
		dst.Code("\t\t\t\t\t\t\t/>\n")
		dst.Code("\t\t\t\t\t\t})}\n")
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			b.getPackage(dst, expr, "")
			b.printMenuItem(dst, t.Obj.Decl.(*ast.TypeSpec).Type, empty, option)
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
		b.printMenuItem(dst, t.Type(), t.Empty, option)
	}
}

func (b *Builder) printMenuModelValue(dst *build.Writer, expr ast.Expr, fieldName string, isNull, isArray bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				pkg := b.getPackage(dst, t, "")
				name := build.StringToHumpName(t.Name)

				if isNull {
					dst.Tab(7).Code("modelValue={_ctx.model!.").Code(fieldName).Code("?.value}\n")
					dst.Tab(7).Code("onUpdate:modelValue={($event: ")
					if isArray {
						dst.Code("number[]")
					} else {
						dst.Code("number")
					}
					dst.Code(" | null) => !$event ? null : _ctx.model!.").Code(fieldName).Code(" = ")
				} else {
					dst.Tab(7).Code("modelValue={_ctx.model!.").Code(fieldName).Code(".value}\n")
					dst.Tab(7).Code("onUpdate:modelValue={($event: ")
					if isArray {
						dst.Code("number[]")
					} else {
						dst.Code("number")
					}
					dst.Code(") => _ctx.model!.").Code(fieldName).Code(" = \n")
				}
				if isArray {
					if isNull {
						dst.Code("$event?.map((e: number) => ").Code(pkg).Code(".").Code(name).Code(".valueOf(e))")
					} else {
						dst.Code("$event.map((e: number) => ").Code(pkg).Code(".").Code(name).Code(".valueOf(e))")
					}
				} else {
					dst.Code(pkg).Code(".").Code(name).Code(".valueOf($event)")
				}
				dst.Code("}\n")
			}
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		b.printMenuModelValue(dst, ar.Type(), fieldName, isNull, true)
	case *ast.MapType:
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printMenuModelValue(dst, t.Type(), fieldName, true, isArray)
	}
}
