package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
	"strings"
)

// 创建表单代码
func (b *Builder) printFormCode(dst *build.Writer, expr ast.Expr) {

	switch expr.(type) {
	case *ast.DataType:
		dst.Import("vue", "defineComponent", "type PropType")
		typ := expr.(*ast.DataType)
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
	form         []string
	table        []string
	suffix       string
	onlyRead     bool
	format       []string
	defaultValue string
	index        *int
	width        float64
	height       float64
	extensions   []string
	clip         bool
	unlink       bool
	mode         string
	sortable     bool
	maxLine      *int64
	maxLen       *int64
	minLen       *int64
	step         *float64
	min          *float64
	max          *float64

	outType string
	outSize []int
	fit     string
	typ     string
	limit   int
}

func (b *Builder) getUI(tags []*ast.Tag) *ui {
	val, ok := build.GetTag(tags, "ui")
	if !ok {
		return nil
	}
	form := ui{
		extensions: []string{},
	}
	if nil != val.KV {
		for _, item := range val.KV {
			if "form" == item.Name.Name {
				for _, value := range item.Values {
					form.form = append(form.form, value.Value[1:len(value.Value)-1])
				}
			} else if "table" == item.Name.Name {
				for _, value := range item.Values {
					form.table = append(form.table, value.Value[1:len(value.Value)-1])
				}
			} else if "onlyRead" == item.Name.Name {
				form.onlyRead = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]
			} else if "outType" == item.Name.Name {
				form.outType = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			} else if "fit" == item.Name.Name {
				form.fit = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			} else if "type" == item.Name.Name {
				form.typ = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			} else if "outSize" == item.Name.Name {
				for _, value := range item.Values {
					atoi, err := strconv.Atoi(value.Value[1 : len(value.Value)-1])
					if err != nil {
						//TODO 添加错误处理
						return nil
					}
					form.outSize = append(form.outSize, atoi)
				}
			} else if "limit" == item.Name.Name {
				atoi, err := strconv.Atoi(item.Values[0].Value[1 : len(item.Values[0].Value)-1])
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.limit = atoi

			} else if "index" == item.Name.Name {
				atoi, err := strconv.Atoi(item.Values[0].Value[1 : len(item.Values[0].Value)-1])
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.index = &atoi
			} else if "format" == item.Name.Name {
				for _, value := range item.Values {
					form.format = append(form.format, value.Value[1:len(value.Value)-1])
				}
			} else if "default" == item.Name.Name {
				form.defaultValue = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
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
				form.maxLine = &atoi
			} else if "maxLen" == item.Name.Name {
				atoi, err := strconv.ParseInt(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10, 64)
				if err != nil {
					println(err.Error())
					return nil
				}
				form.maxLen = &atoi
			} else if "minLen" == item.Name.Name {
				atoi, err := strconv.ParseInt(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10, 64)
				if err != nil {
					println(err.Error())
					return nil
				}
				form.minLen = &atoi
			} else if "clip" == item.Name.Name {
				form.clip = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]
			} else if "unlink" == item.Name.Name {
				form.unlink = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]
			} else if "mode" == item.Name.Name {
				form.mode = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			} else if "sort" == item.Name.Name {
				form.sortable = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]
			} else if "extensions" == item.Name.Name {
				for _, value := range item.Values {
					form.extensions = append(form.extensions, value.Value[1:len(value.Value)-1])
				}
			} else if "min" == item.Name.Name {
				v, err := strconv.ParseFloat(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10)
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.min = &v
			} else if "max" == item.Name.Name {
				v, err := strconv.ParseFloat(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10)
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.max = &v
			} else if "step" == item.Name.Name {
				v, err := strconv.ParseFloat(item.Values[0].Value[1:len(item.Values[0].Value)-1], 10)
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				form.step = &v
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

	b.printLang(dst, typ, u)

	if len(u.form) > 0 && u.form[0] == "true" {
		b.printForm(dst, typ, u)
	}
	if len(u.table) > 0 && u.table[0] == "true" {
		if u.width == 0 {
			u.width = 160
		}
		if u.height == 0 {
			u.height = 160
		}
		b.printTable(dst, typ, u)
	}

}

func (b *Builder) printLang(dst *build.Writer, typ *ast.DataType, u *ui) error {
	name := build.StringToHumpName(typ.Name.Name)
	lang := dst.GetLang(name)

	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		fieldName := build.StringToFirstLower(field.Name.Name)
		lang.Add(fieldName, field.Tags)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) printTable(dst *build.Writer, typ *ast.DataType, u *ui) {
	name := build.StringToHumpName(typ.Name.Name)
	lang := dst.GetLang(name)

	dst.Code("export const " + name + "TableColumn = defineComponent({\n")
	dst.Tab(1).Code("name: '" + name + "TableColumn',\n")
	dst.Tab(1).Code("props: {\n")
	dst.Tab(2).Code("setting: (Function as unknown) as () => ((name: string, lang: string, list: { name: string, val: any }[]) => { name: string, val: any }[]),\n")
	dst.Tab(2).Code("position: Array<string>,\n")
	dst.Tab(2).Code("filter: (Function as unknown) as () => (item: string) => boolean,\n")
	dst.Tab(2).Code("langSuffix: {\n")
	dst.Tab(3).Code("type: String,\n")
	dst.Tab(3).Code("default: \"\",\n")
	dst.Tab(2).Code("}\n")
	dst.Tab(1).Code("},\n")
	dst.Tab(1).Code("setup(props:any) {\n")
	dst.Tab(2).Code("return (ctx: Record<string, any>) => {\n")
	dst.Tab(3).Code("const maps: Record<string, any> = {\n")
	langName := build.StringToFirstLower(name)
	//i := 0
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		table := b.getUI(field.Tags)
		if nil == table || 0 == len(table.table) {
			return nil
		}

		isEnum := build.IsEnum(field.Type)
		isArray := build.IsArray(field.Type)
		//isNull := build.IsNil(field.Type)
		isBool := build.IsBool(field.Type)
		//i++
		//index := i
		if nil != table.index {
			//index = *table.index
		}
		//dst.Code("                <el-table-column prop="adminId" label="adminId" width="140"/>\n")
		fieldName := build.StringToFirstLower(field.Name.Name)
		dst.Tab(4).Code("\"").Code(fieldName).Code("\": () =>(\n")
		fl := lang.Add(fieldName, field.Tags)

		dst.Tab(5).Code("<el-table-column prop=\"").Code(fieldName).Code("\"")

		if table.sortable {
			dst.Code(" sortable=\"custom\"")
		}
		dst.Code(" show-overflow-tooltip")
		width := table.width
		if width == 0 {
			width = u.width
		}
		dst.Code(" min-width=\"").Code(strconv.FormatFloat(width, 'g', -1, 64)).Code("\"")
		dst.Code(">\n")

		tag := ""
		custom := ""
		if len(table.table) > 0 {
			tag = table.table[0]
		}
		if len(table.table) > 1 {
			custom = table.table[1]
		}
		dst.Tab(6).Code("{{\n")
		dst.Tab(7).Code("header: (scope:any) =>  (\n")
		dst.Tab(8).Code("<span class=\"table-header\">\n")
		dst.Tab(9).Code("<span>{{ default: () => ctx.$t(\"").Code(langName).Code("Lang.").Code(fieldName).Code("\" + props.langSuffix) }}</span>\n")
		if b.langValueLen(fl) > 1 {
			dst.Import("@element-plus/icons-vue", "QuestionFilled")
			dst.Tab(9).Code("<el-tooltip  effect=\"dark\" content={ ctx.$t(\"").Code(langName).Code("Lang.").Code(fieldName).Code("1\")}>\n")
			dst.Tab(10).Code("<el-icon class=\"table-header__tip-icon\"><QuestionFilled /></el-icon>\n")
			dst.Tab(9).Code("</el-tooltip>\n")
		}
		dst.Tab(8).Code("</span>\n")
		dst.Tab(7).Code("),\n")
		dst.Tab(7).Code("default: (scope:{row: ")
		b.printType(dst, typ.Name, false, false, false)
		dst.Code(" }) => (null == scope.row.").Code(fieldName).Code(") ? \"").Code(u.defaultValue).Code("\" : (\n")

		if len(custom) > 0 {
			dst.Tab(8).Code("<").Code(custom).Code(" value={scope.row.").Code(fieldName).Code("} />\n")
		} else if "image" == tag {
			dst.Tab(8).Code("<el-popover effect=\"light\" trigger=\"hover\" placement=\"top\" width=\"auto\">\n")
			dst.Tab(10).Code("{{\n")
			dst.Tab(11).Code("default: () => <el-image style={\"width: 200px; height: 200px\"} src={scope.row.").Code(fieldName).Code("} fit=\"contain\"/>,\n")
			dst.Tab(11).Code("reference: () => <el-avatar shape=\"square\" size={60} src={scope.row.").Code(fieldName).Code("+'?w=60&w=60&f=cover'} cover style={\"margin-top: 6px\"}/>,\n")
			dst.Tab(10).Code("}}\n")
			dst.Tab(8).Code("</el-popover>\n")
		} else if "switch" == tag {
			dst.Tab(8).Code("<el-switch v-model={scope.row!.").Code(fieldName).Code("} disabled />\n")
		} else if "link" == tag {
			dst.Tab(8).Code("<el-link href={scope.row!.").Code(fieldName).Code("} target=\"_blank\">{{\n")
			dst.Tab(9).Code("default: () => scope.row.").Code(fieldName).Code("\n")
			dst.Tab(8).Code("}}</el-link>\n")
		} else if "color" == tag {
			dst.Tab(8).Code("<el-tag color={scope.row!.").Code(fieldName).Code("} effect=\"dark\">{{\n")
			dst.Tab(9).Code("default: () => scope.row.").Code(fieldName).Code("\n")
			dst.Tab(8).Code("}}</el-tag>\n")
		} else if "tag" == tag {
			if !isArray {
				dst.Tab(8).Code("<el-tag ")
				if isBool {
					dst.Code("type={scope.row.").Code(fieldName).Code(" ? \"success\" : \"danger\"}")
				} else if isEnum {
					dst.Code("type={scope.row.").Code(fieldName).Code("?.type}")
					dst.Code("class={scope.row.").Code(fieldName).Code("?.cssClass }")
				}
				dst.Code(">{{\n")
				dst.Tab(9).Code("default: () =>")
				b.printTableString(dst, "scope.row."+fieldName, field.Type, false, table.format, " || \"\"", true, false, true)
				dst.Code("\n")
				dst.Tab(8).Code("}}</el-tag>\n")
			} else {
				dst.Tab(8).Code("scope.row.").Code(fieldName).Code(".length < 2 ? (\n")
				dst.Tab(9).Code("<el-tag ")
				if isBool {
					dst.Code("type={scope.row.").Code(fieldName).Code(" ? \"success\" : \"danger\"}")
				} else if isEnum {
					dst.Code("type={scope.row.").Code(fieldName).Code("?.type}")
					dst.Code("class={scope.row.").Code(fieldName).Code("?.cssClass }")
				}
				dst.Code(">\n")
				dst.Tab(10).Code("{{")
				dst.Code("default: () =>")
				b.printTableString(dst, "scope.row."+fieldName, field.Type, false, table.format, " || \"\"", true, false, true)
				dst.Code("}}\n")
				dst.Tab(9).Code("</el-tag>\n")
				dst.Tab(8).Code(") : (\n")
				dst.Tab(9).Code("<div>\n")
				dst.Tab(10).Code("<el-tag")
				if isBool {
					dst.Code(" type={scope.row.").Code(fieldName).Code("![0] ? \"success\" : \"danger\"}")
				} else if isEnum {
					dst.Code(" type={scope.row.").Code(fieldName).Code("![0]?.type}")
					dst.Code(" class={scope.row.").Code(fieldName).Code("![0]?.cssClass }")
				}
				dst.Code(">\n")
				dst.Tab(11).Code("{{")
				dst.Code("default: () =>")
				b.printTableString(dst, "scope.row."+fieldName+"![0]", field.Type, false, table.format, " || \"\"", true, false, false)
				dst.Code("}}\n")
				dst.Tab(10).Code("</el-tag>\n")
				dst.Tab(10).Code("<el-popover width={200}>{{\n")
				dst.Tab(11).Code("reference:()=>(\n")
				dst.Tab(12).Code("<span style={{ marginLeft: '6px' }}>\n")
				dst.Tab(13).Code("<el-tag effect=\"dark\" type=\"warning\">\n")
				dst.Import("@element-plus/icons-vue", "MoreFilled")
				dst.Tab(14).Code("<el-icon><MoreFilled /></el-icon>\n")
				dst.Tab(13).Code("</el-tag>\n")
				dst.Tab(12).Code("</span>\n")
				dst.Tab(11).Code("),\n")
				dst.Tab(11).Code("default:()=>(").Code("scope.row.").Code(fieldName).Code("!.slice(1).map((item, index) => (\n")
				dst.Tab(12).Code("<el-tag")
				dst.Code(" key={index}")
				dst.Code(" style={{ marginLeft: '6px', marginBottom: '4px' }}")
				if isBool {
					dst.Code(" type={scope.row.").Code(fieldName).Code(" ? \"success\" : \"danger\"}")
				} else if isEnum {
					dst.Code(" type={scope.row.").Code(fieldName).Code("?.type}")
					dst.Code(" class={scope.row.").Code(fieldName).Code("?.cssClass }")
				}
				dst.Code(">\n")
				dst.Tab(13).Code("{{")
				dst.Code("default: () =>")
				b.printTableString(dst, "item", field.Type, false, table.format, " || \"\"", true, false, false)
				dst.Code("}}\n")
				dst.Tab(12).Code("</el-tag>\n")
				dst.Tab(11).Code(")))\n")
				dst.Tab(10).Code("}}</el-popover>\n")
				dst.Tab(9).Code("</div>\n")
				dst.Tab(8).Code(")\n")
			}

		} else {
			dst.Tab(8)
			b.printTableString(dst, "scope.row."+fieldName, field.Type, false, table.format, " || \"\"", true, false, true)
			dst.Code("\n")
		}

		dst.Tab(7).Code(")\n")
		dst.Tab(6).Code("}}\n")
		dst.Tab(5).Code("</el-table-column>\n")
		dst.Tab(4).Code("),\n")

		return nil
	})
	if err != nil {
		return
	}

	dst.Tab(3).Code("}\n")
	dst.Code("            for (const key in ctx.$slots) {\n")
	dst.Code("                maps[key] = ctx.$slots[key]\n")
	dst.Code("            }\n")
	dst.Code("\n")
	dst.Code("            let list: { name: string, val: any }[] = []\n")
	dst.Code("            for (const key in maps) {\n")
	dst.Code("                if (maps[key] && (!ctx.filter || ctx.filter(key))) {\n")
	dst.Code("                    list.push({name: key, val: maps[key]})\n")
	dst.Code("                }\n")
	dst.Code("            }\n")
	dst.Code("            list.sort((a, b) => {\n")
	dst.Code("                const indexA = props.position?.indexOf(a.name) ?? 0\n")
	dst.Code("                const indexB = props.position?.indexOf(b.name) ?? 0\n")
	dst.Code("                if (indexA < 0 && indexB < 0) {\n")
	dst.Code("                    return 0\n")
	dst.Code("                } else if (indexA < 0) {\n")
	dst.Code("                    return 1\n")
	dst.Code("                } else if (indexB < 0) {\n")
	dst.Code("                    return -1\n")
	dst.Code("                } else {\n")
	dst.Code("                    return indexA - indexB\n")
	dst.Code("                }\n")
	dst.Code("            })\n")
	dst.Code("\n")
	dst.Code("            if (props.setting) {\n")
	dst.Code("                list = props.setting(\"").Code(name).Code("\", \"").Code(langName).Code("Lang\", list)\n")
	dst.Code("            }\n")
	dst.Code("            return list.map((it) => it.val())\n")
	dst.Code("\n")

	dst.Tab(2).Code("};\n")
	dst.Tab(1).Code("}\n")
	dst.Code("});\n\n")

}

func (b *Builder) printTableString(dst *build.Writer, name string, expr ast.Expr, empty bool, format []string, val string, isDigit bool, checkIsNull bool, forList bool) {
	switch expr.(type) {
	case *ast.EnumType:
		if empty && checkIsNull {
			dst.Code("null == ").Code(name).Code(" ? \"\" : ")
		}
		dst.Code("ctx.$t(").Code(name).Code("?.toString()").Code(")")
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			b.getPackage(dst, expr, "", false, false, "", "")
			b.printTableString(dst, name, t.Obj.Decl.(*ast.TypeSpec).Type, empty, format, val, isDigit, checkIsNull, true)
		} else {
			switch build.BaseType(t.Name) {
			case build.Int8, build.Int16, build.Int32, build.Int64, build.Uint8, build.Uint16, build.Uint32, build.Uint64, build.Float, build.Double, build.Decimal:
				formatName := "fn"
				formatStr := ""
				if len(format) > 1 {
					formatName = format[1]
					formatStr = format[0]
				} else if len(format) > 0 {
					formatStr = format[0]
				}
				dst.Code("ctx.$").Code(formatName).Code("(").Code(name).Code(", \"").Code(formatStr).Code("\")")
			case build.Bool:
				if empty && checkIsNull {
					dst.Code("null == ").Code(name).Code(" ? \"\" : ")
				}
				dst.Code("ctx.$t(").Code(name).Code("?.toString())")
			case build.Date:
				formatName := "fd"
				formatStr := "YYYY/MM/DD HH:mm:ss"
				if len(format) > 1 {
					formatName = format[1]
					formatStr = format[0]
				} else if len(format) > 0 {
					formatStr = format[0]
				}
				dst.Code("ctx.$").Code(formatName).Code("(").Code(name).Code(", \"").Code(formatStr).Code("\")")
			default:
				dst.Code(name)
			}
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		if forList {
			dst.Code(name).Code("?.map((e:any)=>")
			b.printTableString(dst, "e", ar.Type(), false, format, val, isDigit, true, true)
			dst.Code(")?.join(\",\") || \"\"")
		} else {
			b.printTableString(dst, name, ar.Type(), false, format, val, isDigit, true, true)
			dst.Code(" || \"\"")
		}
	case *ast.MapType:
		dst.Code("\"\"+").Code(name)
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printTableString(dst, name, t.Type(), t.Empty, format, val, isDigit, checkIsNull, true)
		if t.Empty {
			dst.Code(val)
		}
	default:
		dst.Code("\"\"+").Code(name)
	}
}

func (b *Builder) printForm(dst *build.Writer, typ *ast.DataType, u *ui) {
	name := build.StringToHumpName(typ.Name.Name)
	lang := dst.GetLang(name)

	dst.Code("export const " + name + "FormItems = defineComponent({\n")
	dst.Tab(1).Code("name: '" + name + "FormItems',\n")
	dst.Tab(1).Code("props: {\n")
	dst.Tab(2).Code("size: String as PropType<\"large\" | \"default\" | \"small\">,\n")
	dst.Tab(2).Code("isAdd: Boolean,\n")
	dst.Tab(2).Code("position: Array<string>,\n")
	dst.Tab(2).Code("filter: (Function as unknown) as () => (item: string) => boolean,\n")
	dst.Tab(2).Code("model: ")
	b.printType(dst, typ.Name, false, false, false)
	dst.Tab(1).Code("\n")
	dst.Tab(1).Code("},\n")
	dst.Tab(1).Code("setup(props: Record<string, any>) {\n")
	dst.Import("element-plus", "useLocale")
	dst.Tab(2).Code("const _locale = useLocale()\n")
	dst.Tab(2).Code("const prop = props.prop ?? \"\"\n")
	dst.Tab(2).Code("return (ctx: Record<string, any>) => {\n")
	dst.Tab(3).Code("const model = ctx.model! as ")
	b.printType(dst, typ.Name, false, false, false)
	dst.Code("\n")
	dst.Tab(3).Code("const maps: Record<string, any> = {\n")
	langName := build.StringToFirstLower(name)

	disabled := make([]string, 0)
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		form := b.getUI(field.Tags)
		if nil == form || 0 == len(form.form) {
			return nil
		}

		isNum := build.IsNumber(field.Type)
		isEnum := build.IsEnum(field.Type)
		isArray := build.IsArray(field.Type)
		isNull := build.IsNil(field.Type)
		isMap := build.IsMap(field.Type)
		_, verify := build.GetTag(field.Tags, "verify")
		//i++
		//index := i
		if nil != form.index {
			//index = *table.index
		}

		formTag := ""
		customTag := ""
		if len(form.form) > 0 {
			formTag = form.form[0]
		}
		if len(form.form) > 1 {
			customTag = form.form[1]
		}

		fieldName := build.StringToFirstLower(field.Name.Name)
		className := build.StringToMiddleLine(field.Name.Name)

		dst.Tab(4).Code("\"").Code(fieldName).Code("\": () =>(\n")

		if "object" != formTag || len(customTag) > 0 {
			dst.Tab(5).Code("<el-form-item class=\"").Code(className).Code("\" prop=\"").Code(fieldName).Code("\"")
			dst.Code(" label={ctx.$t(\"").Code(langName).Code("Lang.").Code(fieldName).Code("\")}")
			if verify {
				b.getPackage(dst, typ.Name, "verify", false, false, "verify", "_"+build.StringToHumpName(field.Name.Name))
				dst.Code(" rules={[{validator: verify").Code(name).Code("_").Code(build.StringToHumpName(field.Name.Name))
				dst.Code("(_locale, model.").Code(fieldName).Code("), trigger: 'blur'}]}")
			}
			dst.Code(">\n")
		}

		if "date" == formTag {
			if len(customTag) == 0 {
				customTag = "el-date-picker"
			}

			dst.Tab(6).Code("<").Code(customTag).Code("\n")
			dst.Tab(7).Code("modelValue={")
			if isArray {
				dst.Import("hbuf_ts", "convertArray")
				dst.Code(" convertArray(model.").Code(fieldName).Code(", (e) => ctx.$timeToLocal(e))")
			} else {
				dst.Code("ctx.$timeToLocal(model.").Code(fieldName).Code(")")
			}
			dst.Code("}\n")

			dst.Tab(7).Code("onUpdate:modelValue={($event: (number | string | Date) | (number | string | Date)[] | null) => model.").Code(fieldName).Code(" = (")

			if isArray {
				dst.Code("($event as (Date[] | null))")
				dst.Code("?")
				dst.Code(".map((v, i)=> i == 0 ? ctx.$timeToUtc(v) as Date : ctx.$timeToUtc(")
				format := ""
				if len(form.format) > 0 {
					format = form.format[0]
				}
				switch format {
				case "YYYY":
					dst.Code("new Date(v.getFullYear(), 12, 31, 23, 59, 59, 999) as Date")
				case "YYYY/MM":
					dst.Code("new Date(v.getFullYear(), v.getMonth(), 31, 23, 59, 59, 999) as Date")
				case "YYYY/MM/DD":
					dst.Code("new Date(v.getFullYear(), v.getMonth(), v.getDate(), 23, 59, 59, 999) as Date")
				case "YYYY/MM/DD HH":
					dst.Code("new Date(v.getFullYear(), v.getMonth(), v.getDate(), v.getHours(), 59, 59, 999) as Date")
				case "YYYY/MM/DD HH:mm":
					dst.Code("new Date(v.getFullYear(), v.getMonth(), v.getDate(), v.getHours(), v.getMinutes(), 59, 999) as Date")
				default:
					dst.Code("new Date(v.getFullYear(), v.getMonth(), v.getDate(), v.getHours(), v.getMinutes(), v.getSeconds(), 999) as Date")
				}
				dst.Code("))")
			} else {
				if isNull {
					dst.Code("(!$event) ? null : ")
				}
				dst.Code("ctx.$timeToUtc($event) as Date")
			}
			dst.Code(")")
			if isNull {
				dst.Code(" ?? null")
			} else if isArray {
				dst.Code(" ?? []")
			} else {
				dst.Code("new Date()")
			}
			dst.Code("}\n")
			if isArray {
				dst.Tab(7).Code("default-time={[new Date(2000, 1, 1, 0, 0, 0, 0), new Date(2000, 1, 1, 23, 59, 59, 999)]}\n")
			}
			dst.Tab(7).Code("type=\"")
			dateType := ""
			if isArray {
				format := ""
				if len(form.format) > 0 {
					format = form.format[0]
				}
				switch format {
				case "YYYY":
					dateType = "yearrange"
				case "YYYY/MM":
					dateType = "monthrange"
				case "YYYY/MM/DD":
					dateType = "daterange"
				default:
					dateType = "datetimerange"
				}
			} else {
				format := ""
				if len(form.format) > 0 {
					format = form.format[0]
				}
				switch format {
				case "YYYY":
					dateType = "year"
				case "YYYY/MM":
					dateType = "month"
				case "YYYY/MM/DD":
					dateType = "date"
				default:
					dateType = "datetime"
				}
			}
			dst.Code(dateType).Code("\"\n")
			dst.Tab(7).Code("shortcuts={ctx.$datePackerShortcuts(\"").Code(dateType).Code("\", ctx.$t)}\n")
			dst.Tab(7).Code("size={props.size}\n")
			dst.Tab(7).Code("clearable=")
			if isNull {
				dst.Code("{true}\n")
			} else {
				dst.Code("{false}\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code(" disabled  \n")
			}
			dst.Tab(6).Code("/>\n")
		} else if "menu" == formTag {
			if len(customTag) == 0 {
				customTag = "el-select"
			}
			dst.Tab(6).Code("<").Code(customTag).Code("\n")
			dst.Tab(7).Code("modelValue={")
			b.printGetStringValue(dst, field.Type, "model."+fieldName, isNull, form.format)
			dst.Code("}\n")
			dst.Tab(7).Code("onUpdate:modelValue={($event: string[] | string | null) => model.").Code(fieldName).Code(" = ")
			b.printSetStringValue(dst, field.Type, "$event", isNull)
			dst.Code("}\n")

			dst.Tab(7).Code("style={\"min-width:180px\"}\n")
			dst.Tab(7).Code("size={props.size}\n")
			dst.Tab(7).Code("filterable\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			if isArray {
				dst.Tab(7).Code("multiple\n")
			}
			dst.Tab(7).Code(">\n")
			if build.IsBool(field.Type) {
				dst.Tab(8).Code("<el-option label={ctx.$t('true')} value={'true'}/>\n")
				dst.Tab(8).Code("<el-option label={ctx.$t('false')} value={'false'}/>\n")
			} else {
				b.printMenuItem(dst, field.Type, false, "el-option")
			}
			dst.Tab(6).Code("</").Code(customTag).Code(">\n")
		} else if "switch" == formTag {
			if len(customTag) == 0 {
				customTag = "el-switch"
			}
			dst.Tab(6).Code("<").Code(customTag).Code("\n")

			dst.Tab(7).Code("modelValue={model.").Code(fieldName).Code(" ??= false")
			dst.Code("}\n")

			dst.Tab(7).Code("onUpdate:modelValue={($event: boolean) => model.").Code(fieldName).Code(" = $event")
			dst.Code("}\n")

			if form.onlyRead {
				dst.Code(" disabled")
			}
			dst.Tab(6).Code("/>\n")
		} else if "radio" == formTag {
			if len(customTag) == 0 {
				customTag = "el-radio-group"
			}
			dst.Tab(6).Code("<").Code(customTag).Code("\n")
			dst.Tab(7).Code("modelValue={")
			b.printGetStringValue(dst, field.Type, "model."+fieldName, isNull, form.format)
			dst.Code("}\n")
			dst.Tab(7).Code("onUpdate:modelValue={($event: string[] | string | null) => model.").Code(fieldName).Code(" = ")
			b.printSetStringValue(dst, field.Type, "$event", isNull)
			dst.Code("}\n")

			dst.Tab(7).Code("size={props.size}\n")
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(7).Code(">\n")
			b.printMenuItem(dst, field.Type, false, "el-radio")
			dst.Tab(6).Code("</").Code(customTag).Code(">\n")
		} else if "radioButton" == formTag {
			if len(customTag) == 0 {
				customTag = "el-radio-group"
			}
			dst.Tab(6).Code("<").Code(customTag).Code("\n")
			dst.Tab(7).Code("modelValue={")
			b.printGetStringValue(dst, field.Type, "model."+fieldName, isNull, form.format)
			dst.Code("}\n")
			dst.Tab(7).Code("onUpdate:modelValue={($event: string[] | string | null) => model.").Code(fieldName).Code(" = ")
			b.printSetStringValue(dst, field.Type, "$event", isNull)
			dst.Code("}\n")

			dst.Tab(7).Code("size={props.size}\n")
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(7).Code(">\n")
			b.printMenuItem(dst, field.Type, false, "el-radio-button")
			dst.Tab(6).Code("</").Code(customTag).Code(">\n")
		} else if "pass" == formTag {
			if len(customTag) == 0 {
				customTag = "el-input"
			}
			dst.Tab(6).Code("<").Code(customTag).Code("\n")
			dst.Tab(7).Code("modelValue={")
			b.printGetStringValue(dst, field.Type, "model."+fieldName, isNull, form.format)
			dst.Code("}\n")
			dst.Tab(7).Code("onUpdate:modelValue={($event: string | null) => model.").Code(fieldName).Code(" = ")
			b.printSetStringValue(dst, field.Type, "$event", isNull)
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
			dst.Tab(6).Code("/>\n")
		} else if isArray {
			if len(customTag) == 0 {
				customTag = "el-input-tag"
			}
			dst.Tab(6).Code("<").Code(customTag).Code("\n")

			dst.Tab(7).Code("v-model={model.").Code(fieldName).Code("}\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(7).Code(">\n")
			dst.Tab(6).Code("</").Code(customTag).Code(">\n")
		} else if ("number" == formTag && (isNum || isEnum)) || ("text" == formTag && isNum) {
			if len(customTag) == 0 {
				customTag = "el-input-number"
			}
			suffix := ""
			if len(form.format) > 0 && len(form.format[0]) > 0 {
				suffix = (form.format[0])[len(form.format[0])-1:]
				if suffix == "0" || suffix == "." || suffix == "," || suffix == "#" {
					suffix = ""
				}
			}
			scale := 0
			if suffix == "%" {
				scale = 100
			} else if suffix == "‰" {
				scale = 1000
			} else if suffix == "‱" {
				scale = 10000
			}

			dst.Tab(6).Code("<").Code(customTag).Code("\n")
			dst.Tab(7).Code("modelValue={")
			b.printGetNumberValue(dst, field.Type, "model."+fieldName, isNull, scale)
			dst.Code("}\n")
			dst.Tab(7).Code("onUpdate:modelValue={($event: number | null) => model.").Code(fieldName).Code(" = ")
			b.printSetNumberValue(dst, field.Type, "$event", isNull, scale)
			dst.Code("}\n")

			dst.Tab(7).Code("size={props.size}\n")
			if "text" == formTag {
				dst.Tab(7).Code("controls={false}\n")
			}

			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			digit := 0
			if len(form.format) > 0 {
				format := form.format[0]
				format = format[:len(format)-len(suffix)]
				index := strings.Index(format, ".")
				if index > -1 {
					digit = len(format) - index - 1
				}
			}
			dst.Tab(7).Code("precision={").Code(strconv.Itoa(digit)).Code("}\n")
			if form.min != nil {
				dst.Tab(7).Code("min={").Code(strconv.FormatFloat(*form.min, 'f', -1, 64)).Code("}\n")
			}
			if form.max != nil {
				dst.Tab(7).Code("max={").Code(strconv.FormatFloat(*form.max, 'f', -1, 64)).Code("}\n")
			}
			if form.step != nil {
				dst.Tab(7).Code("step={").Code(strconv.FormatFloat(*form.step, 'f', -1, 64)).Code("}\n")
			}
			dst.Tab(6).Code(">\n")
			if len(suffix) > 0 {
				dst.Tab(7).Code("{{suffix:()=>'").Code(suffix).Code("' }}\n")
			}
			dst.Tab(6).Code("</").Code(customTag).Code(">\n")
		} else if "color" == formTag {
			if len(customTag) == 0 {
				customTag = "el-color-picker"
			}
			dst.Tab(6).Code("<").Code(customTag).Code("\n")

			dst.Tab(7).Code("v-model={model." + fieldName + "}\n")
			dst.Tab(7).Code("size={props.size}\n")
			dst.Tab(7).Code("show-alpha color-format=\"hex\"\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(6).Code("/>\n")
		} else if "file" == formTag {
			if len(customTag) == 0 {
				customTag = "input-file"
			}
			dst.Tab(6).Code("<").Code(customTag).Code("\n")
			dst.Tab(7).Code("modelValue={")
			b.printGetStringValue(dst, field.Type, "model."+fieldName, isNull, form.format)
			dst.Code("}\n")
			dst.Tab(7).Code("onUpdate:modelValue={($event: string | null) => model.").Code(fieldName).Code(" = ")
			b.printSetStringValue(dst, field.Type, "$event", isNull)
			dst.Code("}\n")

			dst.Tab(7).Code("size={props.size}\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if len(form.typ) > 0 {
				dst.Tab(7).Code("type=\"").Code(form.typ).Code("\"\n")
			}

			if form.limit > 0 {
				dst.Tab(7).Code("limit={").Code(strconv.Itoa(form.limit)).Code("}\n")
			}

			if len(form.outSize) > 0 {
				dst.Tab(7).Code("outWidth={").Code(strconv.Itoa(form.outSize[0])).Code("}\n")
			}

			if len(form.outSize) > 1 {
				dst.Tab(7).Code("outHeight={").Code(strconv.Itoa(form.outSize[1])).Code("}\n")
			}

			if len(form.outType) > 0 {
				dst.Tab(7).Code("outType={").Code(form.outType).Code("}\n")
			}

			if form.onlyRead {
				dst.Tab(7).Code("readonly\n")
			}

			if form.clip {
				dst.Tab(7).Code("clip\n")
			}

			if isArray {
				dst.Tab(7).Code("multiple\n")
			}

			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			dst.Tab(6).Code("/>\n")

		} else if "object" == formTag {
			if len(customTag) > 0 {
				dst.Tab(6).Code("<").Code(customTag).Code("\n")
				dst.Tab(7).Code("modelValue={")
				b.printGetStringValue(dst, field.Type, "model."+fieldName, isNull, form.format)
				dst.Code("}\n")
				dst.Tab(7).Code("onUpdate:modelValue={($event: ")
				b.printType(dst, field.Type, false, false, true)
				dst.Code(") => model.").Code(fieldName).Code(" = $event")
				dst.Code("}\n")
				dst.Tab(6).Code("/>\n")
			} else {
				dst.Tab(5).Code("<>\n")
				dst.Tab(6).Code("<el-form-item class=\"").Code(className).Code("\" prop=\"").Code(fieldName).Code("\"")
				dst.Code(" label={ctx.$t(\"").Code(langName).Code("Lang.").Code(fieldName).Code("\")}")
				dst.Code("/>\n")
				ty := field.Type.Type().(*ast.Ident)
				b.getPackage(dst, ty, "", false, true, "", "FormItems")
				dst.Tab(6).Code("<").Code(build.StringToHumpName(ty.Name)).Code("FormItems\n")
				dst.Tab(7).Code("model={model." + fieldName + "}\n")
				dst.Tab(6).Code("/>\n")
				dst.Tab(5).Code("</>\n")
			}
		} else if "text" == formTag {
			if len(customTag) == 0 {
				customTag = "el-input"
			}
			dst.Tab(6).Code("<").Code(customTag).Code("\n")
			//if fieldName == "userName" {
			//	println("userName")
			//}
			dst.Tab(7).Code("modelValue={")

			b.printGetStringValue(dst, field.Type, "model."+fieldName, isNull, form.format)
			dst.Code("}\n")
			if isMap {
				dst.Tab(7).Code("onUpdate:modelValue={($event: ")
				b.printType(dst, field.Type, false, false, false)
				dst.Code(") => model.").Code(fieldName).Code(" = ")
			} else {
				dst.Tab(7).Code("onUpdate:modelValue={($event: string | null) => model.").Code(fieldName).Code(" = ")
			}
			b.printSetStringValue(dst, field.Type, "$event", isNull)
			dst.Code("}\n")

			dst.Tab(7).Code("size={props.size}\n")
			if isNull {
				dst.Tab(7).Code("clearable\n")
			}
			if isNum {
				dst.Tab(7).Code("type={\"number\"}\n")
			}
			if !isNum && len(form.mode) > 0 {
				dst.Tab(7).Code("type={\"").Code(form.mode).Code("\"}\n")
			}

			if form.onlyRead {
				dst.Tab(7).Code("disabled\n")
			}
			if form.maxLen != nil && !isNum && form.mode == "textarea" {
				dst.Tab(7).Code("show-word-limit\n")
			}
			if form.maxLen != nil {
				dst.Tab(7).Code("maxlength={").Code(strconv.FormatInt(*form.maxLen, 10)).Code("}\n")
			}
			if form.minLen != nil {
				dst.Tab(7).Code("minlength={").Code(strconv.FormatInt(*form.minLen, 10)).Code("}\n")
			}

			if form.maxLine != nil && !isNum && form.mode == "textarea" {
				dst.Tab(7).Code("rows={").Code(strconv.FormatInt(*form.maxLine, 10)).Code("}\n")
			}

			dst.Tab(6).Code("/>\n")
		}

		lang.Add(fieldName, field.Tags)
		if "object" != formTag || len(customTag) > 0 {
			dst.Tab(5).Code("</el-form-item>\n")
		}
		dst.Tab(4).Code("),\n")

		if form.onlyRead {
			disabled = append(disabled, "\""+fieldName+"\"")
		}
		return nil
	})
	if err != nil {
		return
	}

	dst.Tab(3).Code("}\n")
	dst.Tab(3).Code("for (const key in ctx.$slots) {\n")
	dst.Tab(4).Code("maps[key] = ctx.$slots[key]\n")
	dst.Tab(3).Code("}\n")
	dst.Tab(3).Code("const list: string[] = []\n")
	dst.Tab(3).Code("for (const i in ctx.position) {\n")
	dst.Tab(4).Code("const key = ctx.position[i]\n")
	dst.Tab(4).Code("if (!list.includes(key) && maps[key]) {\n")
	dst.Tab(5).Code("list.push(key)\n")
	dst.Tab(4).Code("}\n")
	dst.Tab(3).Code("}\n")
	dst.Tab(3).Code("for (const key in maps) {\n")
	dst.Tab(4).Code("if (!list.includes(key) && maps[key] && (!ctx.filter || ctx.filter(key))) {\n")
	dst.Tab(5).Code("list.push(key)\n")
	dst.Tab(4).Code("}\n")
	dst.Tab(3).Code("}\n")
	if len(disabled) > 0 {
		dst.Tab(3).Code("const disabled = [").Code(strings.Join(disabled, ", ")).Code("]\n")
		dst.Tab(3).Code("return list.filter((it) => !props.isAdd || !disabled.includes(it)).map((it) => maps[it]())\n")
	} else {
		dst.Tab(3).Code("return list.map((it) => maps[it]())\n")
	}
	dst.Tab(2).Code("};\n")
	dst.Tab(1).Code("}\n")
	dst.Code("});\n\n")
}

func (b *Builder) printMenuItem(dst *build.Writer, expr ast.Expr, empty bool, option string) {
	switch expr.(type) {
	case *ast.EnumType:
		t := expr.(*ast.EnumType)
		b.getPackage(dst, t.Name, "", false, false, "", "")
		name := build.StringToHumpName(t.Name.Name)
		dst.Tab(6 + 1).Code("{").Code(name).Code(".values.map((val) => {\n")
		dst.Tab(7 + 1).Code("return <" + option + " key={val.name} class={ 'el-option--' + val.type + ' ' + val.cssClass}\n")
		dst.Tab(8 + 1).Code("label={ctx.$t(val.toString())}\n")
		dst.Tab(8 + 1).Code("value={val.name}\n")
		dst.Tab(7 + 1).Code("/>\n")
		dst.Tab(6 + 1).Code("})}\n")
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			b.printMenuItem(dst, t.Obj.Decl.(*ast.TypeSpec).Type, empty, option)
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		b.printMenuItem(dst, ar.VType, empty, option)
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

func (b *Builder) printGetStringValue(dst *build.Writer, expr ast.Expr, name string, isNull bool, format []string) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				dst.Code(name)
				if isNull {
					dst.Code("?")
				}
				dst.Code(".name")
			} else if ast.Data == t.Obj.Kind {
				dst.Code(name)
			}
		} else if build.BaseType(t.Name) == build.Date {
			if isNull {
				dst.Code(name).Code(" == null").Code(" ? \"\" : ")
			}
			formatName := "fd"
			formatStr := "YYYY/MM/DD HH:mm:ss"
			if len(format) > 1 {
				formatName = format[1]
				formatStr = format[0]
			} else if len(format) > 0 {
				formatStr = format[0]
			}
			dst.Code("ctx.$").Code(formatName).Code("(").Code(name).Code(", \"").Code(formatStr).Code("\")")
		} else {
			dst.Code(name)
			switch build.BaseType(t.Name) {
			case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32, build.Float, build.Double, build.Decimal:
				if isNull {
					dst.Code("?")
				}
				digit := 0
				if len(format) > 0 {
					index := strings.Index(format[0], ".")
					digit = len(format[0]) - index - 1
				}
				dst.Code(".toFixed(").Code(strconv.Itoa(digit)).Code(")")
			case build.Int64, build.Uint64:
				if isNull {
					dst.Code("?")
				}
				dst.Code(".toString()")
			case build.Bool:
				if isNull {
					dst.Code("?")
				}
				dst.Code(".toString()")
			default:
			}
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		dst.Code(name)
		if isNull {
			dst.Code("?")
		}
		dst.Code(".map((item: any)=> ")
		b.printGetStringValue(dst, ar.Type(), "item", ar.IsEmpty(), format)
		dst.Code(")")
	case *ast.MapType:
		dst.Code(name)
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printGetStringValue(dst, t.Type(), name, true, format)
	}
}

func (b *Builder) printSetStringValue(dst *build.Writer, expr ast.Expr, name string, isNull bool) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				b.getPackage(dst, t, "", false, false, "", "")
				dst.Code("((").Code(name).Code("?.length ?? 0) == 0 ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code(build.StringToHumpName(t.Name)).Code(".valueOf(0)")
				}
				dst.Code(" : (")
				dst.Code(build.StringToHumpName(t.Name)).Code(".nameOf(").Code(name).Code("! as string)")
				dst.Code("))")
			}
		} else {
			switch build.BaseType(t.Name) {
			case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32:
				dst.Code("((").Code(name).Code("?.length ?? 0) == 0 ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("0")
				}
				dst.Code(" : (Number.parseInt(").Code(name).Code("! as string)))")
			case build.Int64, build.Uint64:
				dst.Code("((").Code(name).Code("?.length ?? 0) == 0 ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("0n")
				}
				dst.Code(" : (BigInt(").Code(name).Code("! as string)))")
			case build.Float, build.Double:
				dst.Code("((").Code(name).Code("?.length ?? 0) == 0 ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("0")
				}
				dst.Code(" : (Number.parseFloat(").Code(name).Code("! as string)))")

			case build.Bool:
				dst.Code("((").Code(name).Code("?.length ?? 0) == 0 ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("0")
				}
				dst.Code(" : (\"true\" == ").Code(name).Code("))")
			case build.Date:
				dst.Code("((").Code(name).Code("?.length ?? 0) == 0 ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("new Date()")
				}
				dst.Code(" : (new Date(").Code(name).Code("! as string)))")
			case build.Decimal:
				dst.Import("decimal.js", "Decimal")
				dst.Code("((").Code(name).Code("?.length ?? 0) == 0 ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("Decimal(0)")
				}
				dst.Code(" : (new Decimal(").Code(name).Code("! as string)))")

			default:
				dst.Code("((").Code(name).Code("?.length ?? 0) == 0 ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("''")
				}
				dst.Code(" : ").Code(name).Code("! as string)")
			}
		}

	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		if isNull {
			dst.Code("((").Code(name).Code(" as (string[] | null))?.length ?? 0) == 0 ? null :")
		}
		dst.Code("((").Code(name).Code(" as (string[] | null))?")
		dst.Code(".map((item: string)=> ")
		b.printSetStringValue(dst, ar.Type(), "item", ar.IsEmpty())
		dst.Code(")")
		if isNull {
			dst.Code(" ?? null")
		} else {
			dst.Code(" ?? []")
		}
		dst.Code(" )")
	case *ast.MapType:
		dst.Code(name)
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printSetStringValue(dst, t.Type(), name, t.IsEmpty())
	}

}

func (b *Builder) printGetNumberValue(dst *build.Writer, expr ast.Expr, name string, isNull bool, scale int) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				if isNull {
					dst.Code(name).Code(" == null").Code(" ? null : ")
				}
				dst.Code(name).Code(".value")
			}
		} else if build.BaseType(t.Name) == build.Date {
			if isNull {
				dst.Code(name).Code(" == null").Code(" ? null : ")
			}
			dst.Code(name).Code("!.getTime()")
		} else {
			if isNull {
				dst.Code(name).Code(" == null").Code(" ? null : ")
			}
			switch build.BaseType(t.Name) {
			case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32:
				dst.Code(name)
			case build.Int64, build.Uint64:

				dst.Code("Number(").Code(name).Code(")")
			case build.Float, build.Double:
				dst.Code(name)
			case build.Bool:
				dst.Code("Number(").Code(name).Code(" == \"true\" ? 1 : 0")
			case build.Decimal:
				dst.Code(name).Code("?.toNumber()")
			default:
				dst.Code(name)
			}
			if scale > 0 {
				dst.Code(" * ").Code(strconv.Itoa(scale))
			}
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		dst.Code(name)
		if isNull {
			dst.Code("?")
		}
		dst.Code(".map((item: any)=> ")
		b.printGetNumberValue(dst, ar.Type(), "item", ar.IsEmpty(), scale)
		dst.Code(")")
	case *ast.MapType:
		dst.Code("null")
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printGetNumberValue(dst, t.Type(), name, t.IsEmpty(), scale)
	}
}

func (b *Builder) printSetNumberValue(dst *build.Writer, expr ast.Expr, name string, isNull bool, scale int) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				b.getPackage(dst, t, "", false, false, "", "")
				dst.Code("(").Code(name).Code(" == null ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code(build.StringToHumpName(t.Name)).Code(".valueOf(0)")
				}
				dst.Code(" : (")
				dst.Code(build.StringToHumpName(t.Name)).Code(".valueOf(").Code(name).Code("! as number)))")
			}
		} else {
			switch build.BaseType(t.Name) {
			case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32:
				dst.Code("(").Code(name).Code(" == null ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("0")
				}
				dst.Code(" : (").Code(name).Code("! as number")
				if scale > 0 {
					dst.Code(" / ").Code(strconv.Itoa(scale))
				}
				dst.Code("))")
			case build.Int64, build.Uint64:
				dst.Code("(").Code(name).Code(" == null ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("0n")
				}
				if scale > 0 {
					dst.Code(" : BigInt(").Code(name).Code("! as number / ").Code(strconv.Itoa(scale))
				} else {
					dst.Code(" : BigInt(").Code(name).Code("! as number")
				}
				dst.Code("))")
			case build.Float, build.Double:
				dst.Code("(").Code(name).Code(" == null ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("0")
				}
				dst.Code(" : (").Code(name).Code("! as number")
				if scale > 0 {
					dst.Code(" / ").Code(strconv.Itoa(scale))
				}
				dst.Code("))")
			case build.Bool:
				dst.Code("(").Code(name).Code(" == null ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("0")
				}
				dst.Code(" : (0 != ").Code(name).Code("))")
			case build.Date:
				dst.Code("(").Code(name).Code(" == null ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("new Date()")
				}
				dst.Code(" : (new Date(").Code(name).Code("! as number)))")
			case build.Decimal:
				dst.Import("decimal.js", "Decimal")
				dst.Code("(").Code(name).Code(" == null ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("Decimal(0)")
				}
				if scale > 0 {
					dst.Code(" : Decimal(").Code(name).Code("! as number / ").Code(strconv.Itoa(scale))
				} else {
					dst.Code(" : Decimal(").Code(name).Code("! as number")
				}
				dst.Code("))")
			default:
				dst.Code("(").Code(name).Code(" == null ? ")
				if isNull {
					dst.Code("null")
				} else {
					dst.Code("''")
				}
				dst.Code(" : (").Code(name).Code("! as number")
				if scale > 0 {
					dst.Code(" / ").Code(strconv.Itoa(scale))
				}
				dst.Code("))")
			}
		}

	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		dst.Code("(").Code(name).Code(" as (number[] | null))?")
		dst.Code(".map((item: number)=> ")
		b.printSetNumberValue(dst, ar.Type(), "item", ar.IsEmpty(), scale)
		dst.Code(")")
		if isNull {
			dst.Code(" ?? null")
		} else {
			dst.Code(" ?? []")
		}
	case *ast.MapType:
		if isNull {
			dst.Code("null")
		} else {
			dst.Code("{}")
		}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printSetNumberValue(dst, t.Type(), name, t.IsEmpty(), scale)
	}
}
