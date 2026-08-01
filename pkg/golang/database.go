package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/scanner"
	"regexp"
	"strconv"
	"strings"
)

type cache struct {
	min int
	max int
}

type databaseField struct {
	text    string
	typ     string
	comment string
}

func printDatabaseField(dst *build.Writer, tab int, list ...databaseField) {
	txtLen := 0
	typLen := 0
	for _, item := range list {
		if len(item.text) > txtLen {
			txtLen = len(item.text)
		}
		if len(item.typ) > typLen {
			typLen = len(item.typ)
		}
	}
	for i, item := range list {
		if i == 0 {
			dst.Code("\n")
		}
		dst.Tab(tab).Code(build.StringFillRight(item.text, ' ', txtLen))
		dst.Code(build.StringFillRight(item.typ, ' ', typLen+1))
		dst.Code("//").Code(item.comment).Code("\n")
	}
}

func getCache(name string, tags []*ast.Tag) *cache {
	val, ok := build.GetTag(tags, "cache")
	if !ok {
		return nil
	}

	c := &cache{
		min: 2 * 60 * 60,
		max: 3 * 60 * 60,
	}
	if nil != val.KV {
		for _, item := range val.KV {
			if "min" == item.Name.Name {
				val, err := strconv.Atoi(item.Values[0].Value[1 : len(item.Values[0].Value)-1])
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				c.min = val
			} else if "max" == item.Name.Name {
				val, err := strconv.Atoi(item.Values[0].Value[1 : len(item.Values[0].Value)-1])
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				c.max = val
			}
		}
	}
	return c
}

func (b *Builder) printDatabaseCode(dst *build.Writer, typ *ast.DataType) error {
	dbs, wFields, key, err := b.getDBField(typ)
	if 0 == len(dbs) || nil != err {
		return nil
	}

	c := getCache(typ.Name.Name, typ.Tags)

	fDbs := dbs
	fields := wFields
	fType := typ
	if 0 < len(dbs[0].Table) {
		table := b.build.GetDataType(b.getFile(typ.Name), dbs[0].Table)
		if nil == table {
			return nil
		}
		typ = table.Decl.(*ast.TypeSpec).Type.(*ast.DataType)
		dbs, fields, key, err = b.getDBField(typ)
		if nil != err {
			return nil
		}
		if len(dbs) == 0 {
			return nil
		}
	}

	if 0 == len(fDbs) {
		return nil
	}

	if 0 == len(fDbs[0].Table) || (0 != len(fDbs[0].Table) && (strings.ToLower(fDbs[0].Get) == "self" ||
		strings.ToLower(fDbs[0].Map) == "self" ||
		strings.ToLower(fDbs[0].List) == "self" ||
		strings.ToLower(fDbs[0].ListAsync) == "self")) {
		b.printField(dst, fType, wFields)
		b.printScanData(dst, fType, dbs[0], wFields, key)
		b.printNameData(dst, fType)
	}

	val := strings.ToLower(fDbs[0].List)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		b.printListData(dst, typ, val, dbs[0], w, f, fType, c)
	}

	val = strings.ToLower(fDbs[0].ListAsync)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		b.printListAsyncData(dst, typ, val, dbs[0], w, f, fType, c)
	}

	if 0 < len(fDbs[0].Map) {
		ks := strings.Split(fDbs[0].Map, ":")
		if 1 < len(ks) {
			val = strings.ToLower(ks[1])
			if ("self" == val || "parent" == val) && 0 < len(ks[0]) {
				w := wFields
				f := fields
				if "self" == val {
					f = wFields
				}
				b.printMapData(dst, val, typ, dbs[0], w, f, fType, ks[0], c)
			}
		}
	}

	if fDbs[0].Count != nil {
		b.printCountData(dst, typ, dbs[0], wFields, fType, fields, c, fDbs[0].Count)
	}

	if fDbs[0].Del {
		w := wFields
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printDeleteData(dst, dbs[0], w, fType, nil != c)
	}

	if fDbs[0].Remove {
		w := wFields
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printRemoveData(dst, dbs[0], w, fType, nil != c)
	}

	val = strings.ToLower(fDbs[0].Insert)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		b.printInsertOrReplaceData(dst, "Insert", typ, val, dbs[0], w, f, fType, key, c)
	}

	val = strings.ToLower(fDbs[0].Inserts)
	if "self" == val || "parent" == val {
		f := fields
		if "self" == val {
			f = wFields
		}
		b.printInsertOrReplaceBatchData(dst, "Insert", typ, dbs[0], f, key, nil != c)
	}

	val = strings.ToLower(fDbs[0].Replace)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		b.printInsertOrReplaceData(dst, "Replace", typ, val, dbs[0], w, f, fType, key, c)
	}

	val = strings.ToLower(fDbs[0].Replaces)
	if "self" == val || "parent" == val {
		f := fields
		if "self" == val {
			f = wFields
		}
		b.printInsertOrReplaceBatchData(dst, "Replace", typ, dbs[0], f, key, nil != c)
	}

	val = strings.ToLower(fDbs[0].Update)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printUpdateData(dst, typ, val, dbs[0], w, f, fType, c)
	}

	val = strings.ToLower(fDbs[0].Set)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printSetData(dst, typ, val, dbs[0], w, f, fType, c)
	}

	val = strings.ToLower(fDbs[0].Change)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printUpdateChange(dst, typ, val, dbs[0], w, f, fType, c)
	}

	val = strings.ToLower(fDbs[0].Get)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printGetData(dst, typ, val, dbs[0], w, f, fType, c)
	}
	return nil
}

func (b *Builder) getDBField(typ *ast.DataType) ([]*build.DB, []*build.DBField, *build.DBField, error) {
	dbs := build.GetDB(typ.Name.Name, typ.Tags)
	if 0 == len(dbs) {
		return nil, nil, nil, nil
	}

	var fields []*build.DBField
	var key *build.DBField
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dbs := build.GetDB(field.Name.Name, field.Tags)
		if 0 < len(dbs) {
			f := build.DBField{
				Field: field,
				Dbs:   dbs,
			}
			fields = append(fields, &f)
			if nil == key || dbs[0].Key {
				key = &f
			}
		}
		return nil
	})
	if nil != err {
		return nil, nil, nil, err
	}
	return dbs, fields, key, nil
}

func (b *Builder) printField(dst *build.Writer, typ *ast.DataType, fields []*build.DBField) {
	uName := build.StringToHumpName(typ.Name.Name)
	lName := build.StringToFirstLower(typ.Name.Name)
	dst.Code("const ").Code(lName).Code("FieldCount uint = ").Code(strconv.Itoa(len(fields))).Code("\n\n")

	dst.Code("type ").Code(uName).Code("Field uint16\n\n")
	dst.Code("func (f ").Code(uName).Code("Field) Name() string {\n")
	dst.Tab(1).Code("return ").Code(lName).Code("FieldNames[f]\n")
	dst.Code("}\n\n")

	dst.Code("func (f ").Code(uName).Code("Field) DbName() string {\n")
	dst.Tab(1).Code("return ").Code(lName).Code("FieldDbNames[f]\n")
	dst.Code("}\n\n")

	list := make([]databaseField, len(fields))
	for i, field := range fields {
		item := databaseField{
			text:    uName + "Field_" + build.StringToHumpName(field.Field.Name.Name) + " ",
			comment: strings.ReplaceAll(field.Field.Doc.Text(), "\n", ""),
		}
		if 0 == i {
			item.typ = uName + "Field = iota"
		}
		list[i] = item
	}

	dst.Code("const (")
	printDatabaseField(dst, 1, list...)
	dst.Code(")\n\n")

	dst.Code("var ").Code(lName).Code("FieldDbNames = [")
	dst.Code(lName).Code("FieldCount").Code("]string{")
	list = make([]databaseField, len(fields))
	for i, field := range fields {
		item := databaseField{
			text:    "\"" + field.Dbs[0].Name + "\",",
			comment: strings.ReplaceAll(field.Field.Doc.Text(), "\n", ""),
		}
		list[i] = item
	}
	printDatabaseField(dst, 1, list...)
	dst.Code("}\n\n")

	dst.Code("var ").Code(lName).Code("FieldNames = [")
	dst.Code(lName).Code("FieldCount").Code("]string{")
	list = make([]databaseField, len(fields))
	for i, field := range fields {
		item := databaseField{
			text:    "\"" + field.Field.Name.Name + "\",",
			comment: strings.ReplaceAll(field.Field.Doc.Text(), "\n", ""),
		}
		list[i] = item
	}

	printDatabaseField(dst, 1, list...)
	dst.Code("}\n\n")

	dst.Code("var ").Code(lName).Code("FieldDbGets = [")
	dst.Code(lName).Code("FieldCount").Code("]string{")
	for i, field := range fields {
		get := field.Dbs[0].Get
		if len(get) == 0 {
			get = field.Dbs[0].Name
		}

		item := databaseField{
			text:    "\"" + get + "\",",
			comment: strings.ReplaceAll(field.Field.Doc.Text(), "\n", ""),
		}
		list[i] = item
	}
	printDatabaseField(dst, 1, list...)
	dst.Code("}\n\n")

	dst.Code("func ").Code(uName).Code("FieldsByDbName(fields ...string) []").Code(uName).Code("Field {\n")
	dst.Tab(1).Code("list := make([]").Code(uName).Code("Field, 0, ").Code(lName).Code("FieldCount)\n")
	dst.Tab(1).Code("for _, item := range fields {\n")
	dst.Tab(2).Code("switch item {\n")
	for _, field := range fields {
		dst.Tab(2).Code("case \"").Code(field.Dbs[0].Name).Code("\":\n")
		dst.Tab(3).Code("list = append(list, ").Code(uName).Code("Field_").Code(build.StringToHumpName(field.Field.Name.Name)).Code(")\n")
	}
	dst.Tab(2).Code("}\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("return list\n")
	dst.Code("}\n\n")

	dst.Code("func ").Code(uName).Code("FieldsByName(fields ...string) []").Code(uName).Code("Field {\n")
	dst.Tab(1).Code("list := make([]").Code(uName).Code("Field, 0, ").Code(lName).Code("FieldCount)\n")
	dst.Tab(1).Code("for _, item := range fields {\n")
	dst.Tab(2).Code("switch item {\n")
	for _, field := range fields {
		dst.Tab(2).Code("case \"").Code(field.Field.Name.Name).Code("\":\n")
		dst.Tab(3).Code("list = append(list, ").Code(uName).Code("Field_").Code(build.StringToHumpName(field.Field.Name.Name)).Code(")\n")
	}
	dst.Tab(2).Code("}\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("return list\n")
	dst.Code("}\n\n")

}

func (b *Builder) printScanData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField) {
	name := build.StringToHumpName(typ.Name.Name)
	_, scan, _ := b.getItemAndValue(fields, "self")
	dst.Code("func (val *").Code(name).Code(") DbScanColumns(columns ...").Code(name).Code("Field) []any {\n")
	dst.Tab(1).Code("if len(columns) == 0 {\n")
	dst.Tab(2).Code("return []any{" + scan.String() + "}\n")
	dst.Tab(1).Code("}\n")

	dst.Tab(1).Code("result := make([]any, 0, len(columns))\n")
	dst.Tab(1).Code("for _, field := range columns {\n")
	dst.Tab(2).Code("switch field {\n")
	for _, field := range fields {
		fieldName := build.StringToHumpName(field.Field.Name.Name)
		dst.Tab(2).Code("case ").Code(name).Code("Field_").Code(fieldName).Code(":\n")
		dst.Tab(3).Code("result = append(result, ").Code(b.converter(field, "val")).Code(")\n")
	}

	dst.Tab(2).Code("default:\n")
	dst.Tab(2).Code("}\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("return result\n")

	dst.Code("}\n")
	dst.Code("\n")

	dst.Code("func (val *").Code(name).Code(") DbScanNames(columns ...").Code(name).Code("Field) []string {\n")
	dst.Tab(1).Code("if len(columns) == 0 {\n")
	dst.Tab(2).Code("return ").Code(build.StringToFirstLower(name)).Code("FieldDbGets[:]\n")
	dst.Tab(1).Code("}\n")

	dst.Tab(1).Code("result := make([]string, 0, len(columns))\n")
	dst.Tab(1).Code("for _, field := range columns {\n")
	dst.Tab(2).Code("if int(field) < len(").Code(build.StringToFirstLower(typ.Name.Name)).Code("FieldNames) {\n")
	dst.Tab(3).Code("result = append(result, ").Code(build.StringToFirstLower(typ.Name.Name)).Code("FieldDbGets[field])\n")

	dst.Tab(2).Code("}\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("return result\n")

	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printNameData(dst *build.Writer, typ *ast.DataType) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func (val *" + name + ") DbName() string {\n")
	dst.Tab(1).Code("return `" + build.StringToUnderlineName(typ.Name.Name) + "`\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) getItemAndValue(fields []*build.DBField, key string) ([]string, strings.Builder, strings.Builder) {
	var item []string
	scan := strings.Builder{}
	ques := strings.Builder{}
	isFist := true
	for _, field := range fields {
		get := ""
		if field.Dbs[0].Get == "-" {
			continue
		} else if 0 < len(field.Dbs[0].Get) {
			get = strings.ReplaceAll(field.Dbs[0].Get, "?", build.StringToUnderlineName(field.Dbs[0].Name))
		} else if "self" == key {
			get = build.StringToUnderlineName(field.Dbs[0].Name)
		} else {
			continue
		}
		build.StringToUnderlineName(field.Dbs[0].Name)
		if !isFist {
			scan.WriteString(", ")
			ques.WriteString(", ")
		}
		isFist = false
		item = append(item, get)
		scan.WriteString(b.converter(field, "val"))
		ques.WriteString("?")
	}
	return item, scan, ques

}

func (b *Builder) getParamWhere(dst *build.Writer, fields []*build.DBField, page, orderBy, groupBy bool, dName string) *build.Writer {
	where := build.NewWriter()
	where.Packages = dst.Packages

	for _, field := range fields {
		text := field.Dbs[0].Where
		fieldName := build.StringToHumpName(field.Field.Name.Name)
		for i, item := range text {
			if 1 == len(text) || !build.IsArray(field.Field.Type) {
				if build.IsNil(field.Field.Type) {
					where.Tab(1).Code("if nil != g." + fieldName + " {\n")
					_ = b.printParam(where, item, field, fields, "", "\t\ts", "g")
					where.Tab(1).Code("}\n")
				} else {
					_ = b.printParam(where, item, field, fields, "", "\ts", "g")
				}
			} else {
				if build.IsNil(field.Field.Type) {
					where.Tab(1).Code("if nil != g." + fieldName + " && " + strconv.Itoa(i) + " < len(g." + fieldName + ")")
					array := field.Field.Type.(*ast.ArrayType)
					if array.VType.Empty {
						where.Code(" && nil != g." + fieldName + "[" + strconv.Itoa(i) + "]")
					}
					where.Code(" {\n")
					_ = b.printParam(where, item, field, fields, "["+strconv.Itoa(i)+"]", "\t\ts", "g")
					where.Tab(1).Code("}\n")
				} else {
					_ = b.printParam(where, item, field, fields, "["+strconv.Itoa(i)+"]", "\ts", "g")
				}
			}
		}
	}

	if groupBy {
		isFist := true
		for _, field := range fields {
			group := field.Dbs[0].Group
			if 0 < len(group) {
				if build.IsNil(field.Field.Type) {
					where.Tab(1).Code("if nil != g." + build.StringToHumpName(field.Field.Name.Name) + " {\n").Tab(1)
				}
				if isFist {
					where.Tab(1).Code("s.T(\"GROUP BY \")")
				} else {
					where.Tab(1).Code("s.T(\", \")")
				}
				_ = b.printParam(where, group, field, fields, "", "", "g")
				if build.IsNil(field.Field.Type) {
					where.Tab(1).Code("}\n")
				}
				isFist = false
			}
		}
	}

	if orderBy {
		isFist := true
		for _, field := range fields {
			order := field.Dbs[0].Order
			if 0 < len(order) {
				if build.IsNil(field.Field.Type) {
					where.Tab(1).Code("if nil != g." + build.StringToHumpName(field.Field.Name.Name)).Code(" {\n").Tab(1)
				}
				if isFist {
					where.Tab(1).Code("s.T(\"ORDER BY \")")
				} else {
					where.Tab(1).Code("s.T(\", \")")
				}
				_ = b.printParam(where, order, field, fields, "", "", "g")
				if build.IsNil(field.Field.Type) {
					where.Tab(1).Code("}\n")
				}
				isFist = false
			}
		}
	}

	if page {
		if limit, ok := b.getLimit(fields); ok {
			limitName := build.StringToHumpName(limit.Field.Name.Name)
			if offset, ok := b.getOffset(fields); ok {
				offsetName := build.StringToHumpName(offset.Field.Name.Name)
				if build.IsNil(offset.Field.Type) || build.IsNil(limit.Field.Type) {
					if !build.IsNil(offset.Field.Type) {
						where.Tab(1).Code("if nil != g." + limitName + " {\n")
					} else if !build.IsNil(limit.Field.Type) {
						where.Tab(1).Code("if nil != g." + offsetName + " {\n")
					} else {
						where.Tab(1).Code("if nil != g." + offsetName + " && nil != g." + limitName + " {\n")
					}
					where.Tab(2).Code("s.T(\" LIMIT " + offset.Dbs[0].Offset + ", " + limit.Dbs[0].Limit + "\")")
					where.Code(".P(g." + offsetName + ", g." + limitName + ")\n")
					where.Tab(1).Code("}\n")
				} else {
					where.Tab(1).Code("s.T(\" LIMIT " + offset.Dbs[0].Offset + ", " + limit.Dbs[0].Limit + "\")")
					where.Code(".P(g." + offsetName + ", g." + limitName + ")\n")
				}

			} else {

				if build.IsNil(limit.Field.Type) {
					where.Tab(1).Code("if nil != g." + limitName + " {\n")
					where.Tab(2).Code("s.T(\" LIMIT " + limit.Dbs[0].Limit + "\")")
					where.Code(".P(g." + limitName + ")\n")
					where.Tab(1).Code("}\n")
				} else {
					where.Tab(1).Code("s.T(\" LIMIT " + limit.Dbs[0].Limit + "\")")
					where.Code(".P(g." + limitName + ")\n")
				}
			}
		}
	}
	return where
}

func (b *Builder) findField(fields []*build.DBField, name string) *build.DBField {
	for _, item := range fields {
		if item.Field.Name.Name == name {
			return item
		}
	}
	return nil
}

var paramRex = regexp.MustCompile(`(\?{\w+})|(\${\w+})|\$|\?`)

func (b *Builder) printParam(buf *build.Writer, text string, self *build.DBField, fields []*build.DBField, array, tab, object string) error {
	match := paramRex.FindAllStringSubmatchIndex(text, -1)
	buf.Code(tab)
	if nil != match {
		var index = 0
		for _, item := range match {
			if 0 < item[0] {
				buf.Code(".T(\"")
				buf.Code(text[index:item[0]])
				buf.Code("\")")
			}
			t := text[item[0]:item[1]]
			if t == "$" || (2 < len(t) && "${" == t[0:2]) {
				field := self
				if t != "$" {
					field = b.findField(fields, t[2:len(t)-1])
				}

				buf.Code(".T(")
				if build.IsNil(field.Field.Type) {
					buf.Code("*").Code(object).Code(".")
				} else {
					buf.Code(object).Code(".")
				}
				buf.Code(build.StringToHumpName(field.Field.Name.Name))
				buf.Code(")")

			} else if t == "?" || (2 < len(t) && "?{" == t[0:2]) {
				field := self
				temp := array
				if t != "?" {
					field = b.findField(fields, t[2:len(t)-1])
					temp = ""
				}
				if nil == field {
					return scanner.Error{
						//Pos: b.Position(data.Name.Pos()),
						Msg: "Invalid name: " + t[2:len(t)-1],
					}
				}
				if 0 == len(field.Dbs[0].Converter) && build.IsArray(field.Field.Type) && 0 == len(temp) {
					buf.Code(".L(\",\", ")
					buf.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hutl", "utl")
					buf.Code("utl.ToAnyList(").Code(object).Code("." + build.StringToHumpName(field.Field.Name.Name) + ")...")
				} else {
					buf.Code(".V(")
					buf.Code(b.converter(field, object))
					buf.Code(temp)
				}
				buf.Code(")")
			}
			index = item[1]
		}
		if index < len(text) {
			buf.Code(".T(\"")
			buf.Code(text[index:])
			buf.Code("\")")
		}
	} else {
		buf.Code(".T(\"")
		buf.Code(text)
		buf.Code("\")")
	}
	buf.Code("\n")
	return nil
}

func (b *Builder) printListData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)
	if typ != fType {
		key = "self"
	} else if "self" == key {
		dName = fName
	}
	w := b.getParamWhere(dst, wFields, true, true, true, dName)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbList(ctx context.Context, columns ...").Code(dName).Code("Field) ([]").Code(dName).Code(", error) {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")
	dst.Import("strings")
	dst.Tab(1).Code("var val ").Code(dName).Code("\n")
	dst.Tab(1).Code("s.T(\"SELECT \").T(strings.Join(val.DbScanNames(columns...), \", \")).T(\" FROM \").T(tableName).T(\" WHERE is_deleted = 0\")\n")
	dst.Code(w.GetCode().String())

	tab := 0
	if nil != c {
		tab = 1
		dst.Import("math/rand")
		dst.Import("time")

		dst.Tab(1).Code("return db.SaveCache(ctx, tableName, s, time.Duration(rand.Intn(").Code(strconv.Itoa(c.max))
		dst.Code("-").Code(strconv.Itoa(c.min)).Code(")+").Code(strconv.Itoa(c.min)).Code(")*time.Second,")
		dst.Code("func(ctx context.Context) ([]").Code(dName).Code(", error) {\n")
	}
	dst.Import("database/sql")
	dst.Tab(tab + 1).Code("var ret []").Code(dName).Code("\n")
	dst.Tab(tab + 1).Code("_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Tab(tab + 2).Code("var val ").Code(dName).Code("\n")
	dst.Tab(tab + 2).Code("err := rows.Scan(val.DbScanColumns(columns...)...)\n")
	dst.Tab(tab + 2).Code("if err == nil {\n")
	dst.Tab(tab + 3).Code("ret = append(ret, val)\n")
	dst.Tab(tab + 2).Code("}\n")
	dst.Tab(tab + 2).Code("return true, err\n")
	dst.Tab(tab + 1).Code("})\n")

	dst.Tab(tab + 1).Code("return ret, err\n")
	if nil != c {
		dst.Tab(1).Code("})\n")
	}
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printListAsyncData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)
	if typ != fType {
		key = "self"
	} else if "self" == key {
		dName = fName
	}
	w := b.getParamWhere(dst, wFields, true, true, true, dName)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbListAsync(ctx context.Context, fn func(ctx context.Context, ret *").Code(dName).Code(") (bool, error), columns ...").Code(dName).Code("Field) error {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")

	dst.Tab(1).Code("var val ").Code(dName).Code("\n")
	dst.Tab(1).Code("s.T(\"SELECT \").T(strings.Join(val.DbScanNames(columns...), \", \")).T(\" FROM \").T(tableName).T(\" WHERE is_deleted = 0\")\n")
	dst.Code(w.GetCode().String())

	tab := 0
	dst.Import("database/sql")
	dst.Tab(tab + 1).Code("_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Tab(tab + 2).Code("var val ").Code(dName).Code("\n")
	dst.Tab(tab + 2).Code("err := rows.Scan(val.DbScanColumns(columns...)...)\n")
	dst.Tab(tab + 2).Code("if err != nil {\n")
	dst.Tab(tab + 3).Code("return false, err\n")
	dst.Tab(tab + 2).Code("}\n")
	dst.Tab(tab + 2).Code("return fn(ctx, &val)\n")
	dst.Tab(tab + 1).Code("})\n")

	dst.Tab(1).Code("return err\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printMapData(dst *build.Writer, key string, typ *ast.DataType, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, keyName string, c *cache) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	kType, KName, ok := b.getKey(dst, fields, keyName)
	if !ok {
		return
	}
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)
	if typ != fType {
		key = "self"
	} else if "self" == key {
		dName = fName
	}
	w := b.getParamWhere(dst, wFields, true, true, true, dName)
	dst.AddImports(w.GetImports())
	dst.AddImports(kType.GetImports())
	dst.AddImports(KName.GetImports())

	dst.Code("func (g " + fName + ") DbMap(ctx context.Context, columns ...").Code(dName).Code("Field) (map[" + kType.String() + "]*").Code(dName).Code(", error) {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")

	dst.Tab(1).Code("var val ").Code(dName).Code("\n")
	dst.Tab(1).Code("s.T(\"SELECT \").T(strings.Join(val.DbScanNames(columns...), \", \")).T(\" FROM \").T(tableName).T(\" WHERE is_deleted = 0\")\n")
	dst.Code(w.GetCode().String())

	tab := 0
	if nil != c {
		tab = 1
		dst.Import("math/rand")
		dst.Import("time")

		dst.Tab(1).Code("return db.SaveCache(ctx, tableName, s, time.Duration(rand.Intn(").Code(strconv.Itoa(c.max))
		dst.Code("-").Code(strconv.Itoa(c.min)).Code(")+").Code(strconv.Itoa(c.min)).Code(")*time.Second,")
		dst.Code(" func(ctx context.Context) (map[").Code(kType.String()).Code("]*").Code(dName).Code(", error) {\n")
	}
	dst.Import("database/sql")
	dst.Tab(tab + 1).Code("ret := make(map[" + kType.String() + "]*").Code(dName).Code(")\n")
	dst.Tab(tab + 1).Code("_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Tab(tab + 2).Code("var val ").Code(dName).Code("\n")
	dst.Tab(tab + 2).Code("err := rows.Scan(val.DbScanColumns(columns...)...)\n")
	dst.Tab(tab + 2).Code("if err == nil {\n")
	dst.Tab(tab + 3).Code("ret[val.Get" + build.StringToHumpName(KName.String()) + "()] = &val\n")
	dst.Tab(tab + 2).Code("}\n")
	dst.Tab(tab + 2).Code("return true, err\n")
	dst.Tab(tab + 1).Code("})\n")
	dst.Tab(tab + 1).Code("return ret, err\n")
	if nil != c {
		dst.Tab(1).Code("})\n")
	}
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printCountData(dst *build.Writer, typ *ast.DataType, db *build.DB, wFields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, c *cache, count *ast.BasicLit) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, wFields, false, false, true, "")
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbCount(ctx context.Context) (int64, error) {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")
	dst.Tab(1).Code("s.T(\"SELECT COUNT(\")")
	b.printCount(dst, count, wFields, "", "")
	dst.Code(".T(\") FROM \").T(tableName).T(\" WHERE is_deleted = 0\")\n")
	dst.Code(w.GetCode().String())

	tab := 0
	if nil != c {
		tab = 1
		dst.Import("math/rand")
		dst.Import("time")

		dst.Tab(1).Code("return db.SaveCache(ctx, tableName, s, time.Duration(rand.Intn(").Code(strconv.Itoa(c.max))
		dst.Code("-").Code(strconv.Itoa(c.min)).Code(")+").Code(strconv.Itoa(c.min)).Code(")*time.Second,")
		dst.Code(" func(ctx context.Context) (int64, error) {\n")
	}
	dst.Import("database/sql")
	dst.Tab(tab + 1).Code("var val int64\n")
	dst.Tab(tab + 1).Code("_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Tab(tab + 2).Code("return false, rows.Scan(&val)\n")
	dst.Tab(tab + 1).Code("})\n")
	dst.Tab(tab + 1).Code("return val, err\n")
	if nil != c {
		dst.Tab(1).Code("})\n")
	}
	dst.Code("}\n")
	dst.Code("\n")
}

var countRex = regexp.MustCompile(`(\${\w+})`)

func (b *Builder) printCount(buf *build.Writer, count *ast.BasicLit, fields []*build.DBField, array, tab string) error {
	text := strings.Trim(count.Value, "\"")
	match := countRex.FindAllStringSubmatchIndex(text, -1)
	buf.Code(tab)
	if nil != match {
		var index = 0
		for _, item := range match {
			if 0 < item[0] {
				buf.Code(".T(\"")
				buf.Code(text[index:item[0]])
				buf.Code("\")")
			}
			t := text[item[0]:item[1]]
			if 2 < len(t) && "${" == t[0:2] {
				field := b.findField(fields, t[2:len(t)-1])
				if nil == field {

					return build.NewError(count.Pos()+1, "Not find field: "+text)
				}
				buf.Code(".T(").Code("g.Get").Code(build.StringToHumpName(field.Field.Name.Name)).Code("())")

			}
			index = item[1]
		}
		if index < len(text) {
			buf.Code(".T(\"")
			buf.Code(text[index:])
			buf.Code("\")")
		}
	} else {
		buf.Code(".T(\"")
		buf.Code(text)
		buf.Code("\")")
	}
	return nil
}

func (b *Builder) printDeleteData(dst *build.Writer, db *build.DB, wFields []*build.DBField, fType *ast.DataType, c bool) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, wFields, false, false, false, "")
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbDel(ctx context.Context) (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/htime")
	dst.Tab(1).Code("s.T(\"UPDATE \").T(tableName).T(\" SET is_deleted = 1, delete_time = \").V(htime.NowTime()).T(\" WHERE is_deleted = 0\")\n")
	dst.Code(w.GetCode().String())

	if c {
		dst.Tab(1).Code("_ = db.ClearCache(ctx, tableName)\n")
	}
	dst.Tab(1).Code("return s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printRemoveData(dst *build.Writer, db *build.DB, wFields []*build.DBField, fType *ast.DataType, c bool) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, wFields, false, false, false, "")
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbRemove(ctx context.Context) (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")
	dst.Tab(1).Code("s.T(\"DELETE FROM \").T(tableName).T(\" WHERE 1 = 1\")\n")
	dst.Code(w.GetCode().String())

	if c {
		dst.Tab(1).Code("_ = db.ClearCache(ctx, tableName)\n")
	}
	dst.Tab(1).Code("return s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertOrReplaceData(dst *build.Writer, s string, typ *ast.DataType, val string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, key *build.DBField, c *cache) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	fName := build.StringToHumpName(fType.Name.Name)
	if typ != fType {
		val = "parent"
	}
	w := b.getParamWhere(dst, wFields, false, false, false, "")
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") Db").Code(s).Code("(ctx context.Context) (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")
	dst.Tab(1).Code("v := db.NewBuilder()\n")
	dst.Tab(1).Code("s.T(\"").Code(strings.ToUpper(s)).Code(" INTO \").T(tableName).T(\" (\").Del(\",\")\n")
	dst.Tab(1).Code("v.T(\"VALUES(\").Del(\",\")\n")

	for _, field := range fields {
		set := ""
		if 0 < len(field.Dbs[0].Set) {
			set = field.Dbs[0].Set
		} else if "self" == val {
			set = "?"
		} else {
			continue
		}

		tag := 0
		name := build.StringToHumpName(field.Field.Name.Name)
		if build.IsNil(field.Field.Type) && !field.Dbs[0].Force {
			dst.Tab(1).Code("if nil != g.")
			dst.Code(name)
			dst.Code(" {\n")
			tag = 1
		}

		dst.Tab(tag + 1).Code("s.T(\",\").T(\"").Code(field.Dbs[0].Name).Code("\")\n")
		dst.Tab(tag + 1).Code("v.T(\",\")")
		_ = b.printParam(dst, set, field, fields, "", "", "g")
		if tag > 0 {
			dst.Tab(1).Code("}\n")
		}
		dst.Code("\n")
	}
	dst.Tab(1).Code("s.T(\") \")\n")
	dst.Tab(1).Code("v.T(\") \")\n")
	dst.Tab(1).Code("s.Join(v)\n")
	if nil != c {
		dst.Tab(1).Code("_ = db.ClearCache(ctx, tableName)\n")
	}
	dst.Tab(1).Code("return s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertOrReplaceBatchData(dst *build.Writer, s string, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func (g " + name + ") Db").Code(s).Code("Batch(ctx context.Context, list ...*" + name + ") (int64, int64, error) {\n")
	dst.Tab(1).Code("if nil == list || 0 == len(list) {\n")
	dst.Tab(2).Code("return 0, 0, nil\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("var total int64\n")
	dst.Tab(1).Code("var lastId int64\n")
	dst.Tab(1).Code("limit := int(db.BatchLimit(ctx))\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n\n")

	dst.Tab(1).Code("for j := 0; j < len(list); j += limit {\n")
	dst.Tab(2).Code("s := db.NewBuilder()\n")
	dst.Tab(2).Code("s.T(\"").Code(strings.ToUpper(s)).Code(" INTO \").T(tableName).T(\" (")
	isFist := true
	for _, field := range fields {
		if !isFist {
			dst.Code(", ")
		}
		isFist = false
		dst.Code(field.Dbs[0].Name)
	}
	dst.Code(") VALUES\")\n")
	dst.Tab(2).Code("end := min(j+limit, len(list))\n")
	dst.Tab(2).Code("for i, val := range list[j:end] {\n")
	dst.Tab(3).Code("if 0 != i {\n")
	dst.Tab(4).Code("s.T(\",\")\n")
	dst.Tab(3).Code("}\n")
	dst.Tab(3).Code("s.T(\"(\").L(\",\", ")
	isFist = true
	for _, field := range fields {
		if !isFist {
			dst.Code(", ")
		}
		isFist = false
		dst.Code(b.converter(field, "val"))
	}
	dst.Code(").T(\")\")\n")
	dst.Tab(2).Code("}\n")
	dst.Tab(2).Code("count, id, err := s.Exec(ctx)\n")
	dst.Tab(2).Code("if err != nil {\n")
	dst.Tab(3).Code("return 0, 0, err\n")
	dst.Tab(2).Code("}\n")
	dst.Tab(2).Code("total += count\n")
	dst.Tab(2).Code("lastId = id\n")

	dst.Tab(1).Code("}\n")

	if isCache {
		dst.Tab(1).Code("_ = db.ClearCache(ctx, tableName)\n")
	}
	dst.Tab(1).Code("return total, lastId, nil\n")
	dst.Code("}\n\n")
}

func (b *Builder) printUpdateData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	fName := build.StringToHumpName(fType.Name.Name)
	object := "g"
	w := b.getParamWhere(dst, wFields, false, false, false, "")
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbUpdate(ctx context.Context")
	if key == "parent" {
		object = "d"
		dst.Code(", d *").Code(build.StringToHumpName(typ.Name.Name))
	}
	dst.Code(") (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")
	dst.Tab(1).Code("s.T(\"UPDATE \").T(tableName).T(\" SET \").Del(\",\")\n")

	set := b.printSet(typ, fields, typ == fType || key == "parent", SetWhereNil, object)
	dst.AddImports(set.GetImports())
	dst.Code(set.String())

	dst.Tab(1).Code("s.T(\"WHERE 1 = 1 \")\n")
	dst.Code(w.String())
	dst.Code("\n")
	if nil != c {
		dst.Tab(1).Code("_ = db.ClearCache(ctx, tableName)\n")
	}
	dst.Tab(1).Code("return s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printSetData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	fName := build.StringToHumpName(fType.Name.Name)
	w := b.getParamWhere(dst, wFields, false, false, false, "")
	dst.AddImports(w.GetImports())

	object := "g"
	dst.Code("func (g " + fName + ") DbSet(ctx context.Context")
	if key == "parent" {
		object = "d"
		dst.Code(", d *").Code(build.StringToHumpName(typ.Name.Name))
	}
	dst.Code(") (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")
	dst.Tab(1).Code("s.T(\"UPDATE \").T(tableName).T(\" SET \").Del(\",\")\n")

	set := b.printSet(typ, fields, typ == fType || key == "parent", SetWhereNot, object)
	dst.AddImports(set.GetImports())
	dst.Code(set.String())

	dst.Tab(1).Code("s.T(\"WHERE 1 = 1 \")\n")
	dst.Code(w.String())
	dst.Code("\n")
	if nil != c {
		dst.Tab(1).Code("_ = db.ClearCache(ctx, tableName)\n")
	}
	dst.Tab(1).Code("return s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printUpdateChange(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	uName := build.StringToHumpName(fType.Name.Name)
	w := b.getParamWhere(dst, wFields, false, false, false, "")
	dst.AddImports(w.GetImports())

	object := "g"
	dst.Code("func (g " + uName + ") DbUpdateChange(ctx context.Context")
	if key == "parent" {
		object = "d"
		dst.Code(", d *").Code(build.StringToHumpName(typ.Name.Name))
	}
	dst.Code(") (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")
	dst.Tab(1).Code("s.T(\"UPDATE \").T(tableName).T(\" SET \").Del(\",\")\n")

	dst.Tab(1).Code("isChange := false\n")
	set := b.printSet(typ, fields, typ == fType || key == "parent", SetWhereChange, object)
	dst.AddImports(set.GetImports())
	dst.Code(set.String())

	dst.Tab(1).Code("s.T(\"WHERE 1 = 1 \")\n")
	dst.Code(w.String())
	dst.Code("\n")
	dst.Tab(1).Code("if !isChange {\n")
	dst.Tab(2).Code("return 0, 0, nil\n")
	dst.Tab(1).Code("}\n")
	if nil != c {
		dst.Tab(1).Code("_ = db.ClearCache(ctx, tableName)\n")
	}
	dst.Tab(1).Code("return s.Exec(ctx)\n")
	dst.Code("}\n\n")

	if typ == fType {
		lName := build.StringToFirstLower(fType.Name.Name)
		dst.Code("func (g *" + uName + ") DbSetChangeFields(fields ...").Code(uName).Code("Field) {\n")

		dst.Tab(1).Code("g.changeFields = make([]bool, ").Code(lName).Code("FieldCount)\n")
		dst.Tab(1).Code("for _, item := range fields {\n")
		dst.Tab(2).Code("g.changeFields[item] = true\n")
		dst.Tab(1).Code("}\n")
		dst.Code("}\n\n")

		dst.Code("func (g *" + uName + ") DbAllChangeFields() {\n")
		dst.Tab(1).Code("g.changeFields = make([]bool, ").Code(lName).Code("FieldCount)\n")
		dst.Tab(1).Code("for i := 0; i < int(").Code(lName).Code("FieldCount); i++ {\n")
		dst.Tab(2).Code("g.changeFields[i] = true\n")
		dst.Tab(1).Code("}\n")
		dst.Code("}\n\n")

		dst.Code("func (g *" + uName + ") DbClearChangeFields() {\n")
		dst.Tab(1).Code("g.changeFields = make([]bool, ").Code(lName).Code("FieldCount)\n")
		dst.Code("}\n\n")
	}
}

type SetWhere int

const (
	SetWhereNot    = 0
	SetWhereNil    = 1
	SetWhereChange = 2
)

func (b *Builder) printSet(typ *ast.DataType, fields []*build.DBField, allSet bool, where SetWhere, object string) *build.Writer {
	uName := build.StringToHumpName(typ.Name.Name)
	dst := build.NewWriter()
	for _, field := range fields {
		set := ""
		if 0 < len(field.Dbs[0].Set) {
			set = field.Dbs[0].Name + " = " + field.Dbs[0].Set
		} else if allSet {
			set = field.Dbs[0].Name + " = ?"
		} else {
			continue
		}

		isWhere := false
		name := build.StringToHumpName(field.Field.Name.Name)
		if where == SetWhereChange {
			isWhere = true
			dst.Tab(1).Code("if ").Code(object).Code(".changeFields[").Code(uName).Code("Field_").Code(name).Code("] {\n")
			dst.Tab(2).Code("isChange = true\n")
			dst.Tab(1)
		} else if where == SetWhereNil && build.IsNil(field.Field.Type) && !field.Dbs[0].Force {
			isWhere = true
			dst.Tab(1).Code("if nil != ").Code(object).Code(".")
			dst.Code(name)
			dst.Code(" {\n")
			dst.Tab(1)
		}

		dst.Tab(1).Code("s.T(\",\")")
		_ = b.printParam(dst, set, field, fields, "", "", object)

		if isWhere {
			dst.Tab(1).Code("}\n")
		}
	}
	return dst

}

func (b *Builder) printGetData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	dst.Import("context")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")

	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)
	if typ == fType {
		key = "self"
	} else if "self" == key {
		dName = fName
	} else {
		key = "self"
	}

	w := b.getParamWhere(dst, wFields, false, true, true, dName)
	dst.AddImports(w.GetImports())

	dst.Code("func (g ").Code(fName).Code(") DbGet(ctx context.Context, columns ...").Code(dName).Code("Field) (*").Code(dName).Code(", error) {\n")
	dst.Tab(1).Code("tableName := db.TableName(ctx, \"").Code(db.Name).Code("\")\n")
	dst.Tab(1).Code("s := db.NewBuilder()\n")
	dst.Import("strings")
	dst.Tab(1).Code("var val *").Code(dName).Code("\n")

	dst.Tab(1).Code("s.T(\"SELECT \").T(strings.Join(val.DbScanNames(columns...), \", \")).T(\" FROM \").T(tableName).T(\" WHERE is_deleted = 0\")\n")
	dst.Code(w.GetCode().String())
	dst.Tab(1).Code("s.T(\" LIMIT 1\")\n")

	tab := 0
	if nil != c {
		tab = 1
		dst.Import("math/rand")
		dst.Import("time")

		dst.Tab(1).Code("return db.SaveCache(ctx, tableName, s, time.Duration(rand.Intn(").Code(strconv.Itoa(c.max))
		dst.Code("-").Code(strconv.Itoa(c.min)).Code(")+").Code(strconv.Itoa(c.min)).Code(")*time.Second,")
		dst.Code(" func(ctx context.Context) (*").Code(dName).Code(", error) {\n")
	}
	dst.Import("database/sql")
	dst.Tab(tab + 1).Code("_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Tab(tab + 2).Code("val = &").Code(dName).Code("{}\n")
	dst.Tab(tab + 2).Code("return false, rows.Scan(val.DbScanColumns(columns...)...)\n")
	dst.Tab(tab + 1).Code("})\n")
	dst.Tab(tab + 1).Code("return val, err\n")
	if nil != c {
		dst.Tab(1).Code("})\n")
	}
	dst.Code("}\n")
	dst.Code("\n")

}
