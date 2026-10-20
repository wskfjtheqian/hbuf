package goplus

import (
	"fmt"
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"regexp"
	"strconv"
	"strings"
)

// DataContext database 输出模板的 data 类型上下文。
// 各动作块由构建期预渲染,模板负责按序组装。
type DataContext struct {
	Data *wire.DataType
	// 各动作代码块,模板按序拼接
	FieldBlock         string
	ScanBlock          string
	NameBlock          string
	JoinBlock          string
	ListBlock          string
	ListAsyncBlock     string
	MapBlock           string
	CountBlock         string
	DeleteBlock        string
	RemoveBlock        string
	InsertBlock        string
	InsertBatchBlock   string
	UpdateBlock        string
	SetBlock           string
	UpdateChangeBlock  string
	ChangesFieldsBlock string
	UpdateChangesBlock string
	GetBlock           string
}

// databaseField 对齐输出用的字段描述(文本/类型/注释)。
type databaseField struct {
	text    string
	typ     string
	comment string
}

// fieldLines 按列对齐输出字段定义行,与进程内后端一致。
func fieldLines(tab int, list ...databaseField) string {
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
	var sb strings.Builder
	for _, item := range list {
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat("\t", tab))
		sb.WriteString(stringFillRight(item.text, ' ', txtLen))
		sb.WriteString(stringFillRight(item.typ, ' ', typLen+1))
		sb.WriteString("//" + item.comment)
	}
	return sb.String()
}

// dbBlock 输出代码块:prefix + 对齐行 + suffix。
// 空列表时对齐行不输出(与进程内一致,如 "const ()")。
func dbBlock(sb *strings.Builder, prefix string, list []databaseField, suffix string) {
	sb.WriteString(prefix)
	if len(list) > 0 {
		sb.WriteString(fieldLines(1, list...))
		sb.WriteString("\n")
	}
	sb.WriteString(suffix)
}

// buildDatabase 为带 db 注解的 data 类型构建 database 模板上下文。
// 与进程内后端 printDatabaseCode 行为一致:支持 table 间接引用(子类挂父表)、
// join 联表、缓存、分页、批量操作与变更字段跟踪。
func (c *Ctx) buildDatabase(imp Imports, data *wire.DataType) (*DataContext, error) {
	dc := &DataContext{Data: data}
	db, wFields, key, err := c.getDBField(data)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return dc, nil
	}
	cch := getCache(data.Tags)

	fdb := db
	fields := wFields
	fType := data
	typ := data
	// table 引用其他 data 类型时,生成对象用目标类型的表结构
	replaced := false
	if len(db.Table) > 0 {
		t, _ := c.Find(db.Table)
		if t == nil || t.Data == nil {
			return dc, nil
		}
		typ = t.Data
		db, fields, key, err = c.getDBField(typ)
		if err != nil {
			return nil, err
		}
		if db == nil {
			return dc, nil
		}
		replaced = true
	}

	if len(fdb.Table) == 0 || (len(fdb.Table) != 0 && (strings.ToLower(fdb.Get) == "self" ||
		strings.ToLower(fdb.Map) == "self" ||
		strings.ToLower(fdb.List) == "self" ||
		strings.ToLower(fdb.ListAsync) == "self")) {
		dc.FieldBlock, err = c.buildFieldBlock(fType, db, wFields)
		if err != nil {
			return nil, err
		}
		dc.ScanBlock, err = c.buildScanBlock(fType, db, wFields, key)
		if err != nil {
			return nil, err
		}
		dc.NameBlock = buildNameBlock(fType)
	}
	join := len(fdb.Join) > 0 && strings.ToLower(fdb.Join[0]) == "true"
	if join {
		dc.JoinBlock, err = c.buildJoinBlock(imp, typ, db, fType)
		if err != nil {
			return nil, err
		}
	}

	val := strings.ToLower(fdb.List)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		dc.ListBlock = c.buildListBlock(imp, typ, val, db, w, f, fType, cch, replaced)
	}

	val = strings.ToLower(fdb.ListAsync)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		dc.ListAsyncBlock = c.buildListAsyncBlock(imp, typ, val, db, w, f, fType, cch, replaced)
	}

	if 0 < len(fdb.Map) {
		ks := strings.Split(fdb.Map, ":")
		if 1 < len(ks) {
			val = strings.ToLower(ks[1])
			if ("self" == val || "parent" == val) && 0 < len(ks[0]) {
				w := wFields
				f := fields
				if "self" == val {
					f = wFields
				}
				dc.MapBlock = c.buildMapBlock(imp, val, typ, db, w, f, fType, ks[0], cch, replaced)
			}
		}
	}

	if fdb.Count != "" {
		dc.CountBlock, err = c.buildCountBlock(imp, typ, db, wFields, fType, fields, cch, fdb.Count)
		if err != nil {
			return nil, err
		}
	}

	if fdb.Del {
		w := wFields
		if !replaced {
			key.DB.Where = []string{"AND id = ?"}
			w = []*DBField{key}
		}
		dc.DeleteBlock = c.buildDeleteBlock(imp, db, w, fType, cch != nil)
	}

	if fdb.Remove {
		w := wFields
		if !replaced {
			key.DB.Where = []string{"AND id = ?"}
			w = []*DBField{key}
		}
		dc.RemoveBlock = c.buildRemoveBlock(imp, db, w, fType, cch != nil)
	}

	val = strings.ToLower(fdb.Insert)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		dc.InsertBlock, err = c.buildInsertBlock(imp, typ, val, db, w, f, fType, key, cch, replaced)
		if err != nil {
			return nil, err
		}
	}

	val = strings.ToLower(fdb.Inserts)
	if "self" == val || "parent" == val {
		f := fields
		if "self" == val {
			f = wFields
		}
		dc.InsertBatchBlock = c.buildInsertBatchBlock(imp, typ, db, f, key, cch != nil)
	}

	val = strings.ToLower(fdb.Update)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if !replaced {
			key.DB.Where = []string{"AND id = ?"}
			w = []*DBField{key}
		}
		dc.UpdateBlock, err = c.buildUpdateBlock(imp, typ, val, db, w, f, fType, cch, replaced, SetWhereNil)
		if err != nil {
			return nil, err
		}
	}

	val = strings.ToLower(fdb.Set)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if !replaced {
			key.DB.Where = []string{"AND id = ?"}
			w = []*DBField{key}
		}
		dc.SetBlock, err = c.buildUpdateBlock(imp, typ, val, db, w, f, fType, cch, replaced, SetWhereNot)
		if err != nil {
			return nil, err
		}
	}

	isTemp := false
	val = strings.ToLower(fdb.Change)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if !replaced {
			isTemp = true
			key.DB.Where = []string{"AND id = ?"}
			w = []*DBField{key}
		}
		dc.UpdateChangeBlock, err = c.buildUpdateChangeBlock(imp, typ, val, db, w, f, fType, cch, replaced)
		if err != nil {
			return nil, err
		}
	}

	val = strings.ToLower(fdb.Changes)
	if "self" == val || "parent" == val {
		f := fields
		if "self" == val {
			f = wFields
		}
		if !replaced {
			isTemp = true
			key.DB.Where = []string{"AND id = ?"}
		}
		dc.UpdateChangesBlock = c.buildUpdateChangesBlock(imp, typ, val, db, key, f, fType, cch)
	}

	if isTemp {
		dc.ChangesFieldsBlock = buildChangesFieldsBlock(fType)
	}

	val = strings.ToLower(fdb.Get)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if !replaced {
			key.DB.Where = []string{"AND id = ?"}
			w = []*DBField{key}
		}
		dc.GetBlock, err = c.buildGetBlock(imp, typ, val, db, w, f, fType, cch, replaced)
		if err != nil {
			return nil, err
		}
	}
	return dc, nil
}

// getDBField 解析 data 类型与其继承链上的 db 字段。
func (c *Ctx) getDBField(typ *wire.DataType) (*DB, []*DBField, *DBField, error) {
	db := getDB(typ.Name, typ.Tags)
	if db == nil {
		return nil, nil, nil, nil
	}

	var fields []*DBField
	var key *DBField
	err := c.enumField(typ, func(field *wire.Field, data *wire.DataType) error {
		fdb := getDB(field.Name, field.Tags)
		if fdb != nil {
			f := &DBField{Field: field, DB: fdb}
			fields = append(fields, f)
			if nil == key || fdb.Key {
				key = f
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return db, fields, key, nil
}

// buildFieldBlock 生成字段定义块(type/const/var/FieldsBy...),与进程内 printField 一致。
func (c *Ctx) buildFieldBlock(typ *wire.DataType, db *DB, fields []*DBField) (string, error) {
	join := len(db.Join) > 0 && strings.ToLower(db.Join[0]) == "true"

	uName := tpl.HumpName(typ.Name)
	lName := tpl.LowerFirstName(typ.Name)
	var sb strings.Builder
	sb.WriteString("type " + uName + "Field uint16\n\n")

	if join {
		sb.WriteString("func " + uName + "FieldBy(")
		err := c.enumField(typ, func(field *wire.Field, data *wire.DataType) error {
			db1 := getDB(field.Name, field.Tags)
			if db1 == nil {
				return nil
			}
			tableName := ""
			if field.Type != nil && field.Type.Kind == "ref" {
				tableName = field.Type.Name
			}
			table, _ := c.Find(tableName)
			if table == nil || table.Data == nil {
				return nil
			}

			sb.WriteString("_" + tpl.LowerFirstName(field.Name) + " []" + tpl.HumpName(tableName) + "Field, ")
			return nil
		})
		if err != nil {
			return "", err
		}

		sb.WriteString(") []" + uName + "Field {\n")
		sb.WriteString("\tvar result []" + uName + "Field\n")
		var before string
		err = c.enumField(typ, func(field *wire.Field, data *wire.DataType) error {
			db1 := getDB(field.Name, field.Tags)
			if db1 == nil {
				return nil
			}
			tableName := ""
			if field.Type != nil && field.Type.Kind == "ref" {
				tableName = field.Type.Name
			}
			table, _ := c.Find(tableName)
			if table == nil || table.Data == nil {
				return nil
			}

			sb.WriteString("\tfor _, field := range _" + tpl.LowerFirstName(field.Name) + " {\n")
			sb.WriteString("\t\tresult = append(result, " + uName + "Field(field)")
			if len(before) > 0 {
				sb.WriteString(" + " + uName + "Field(" + tpl.LowerFirstName(before) + "FieldCount)")
			}
			sb.WriteString(")\n")
			sb.WriteString("\t}\n")
			before = tableName
			return nil
		})
		if err != nil {
			return "", err
		}

		sb.WriteString("\treturn result\n")
		sb.WriteString("}\n\n")
		return sb.String(), nil
	}
	sb.WriteString("const " + lName + "FieldCount uint = " + strconv.Itoa(len(fields)) + "\n\n")
	sb.WriteString("func (f " + uName + "Field) Name() string {\n")
	sb.WriteString("\treturn " + lName + "FieldNames[f]\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (f " + uName + "Field) DbName() string {\n")
	sb.WriteString("\treturn " + lName + "FieldDbNames[f]\n")
	sb.WriteString("}\n\n")

	list := make([]databaseField, len(fields))
	for i, field := range fields {
		item := databaseField{
			text:    uName + "Field_" + tpl.HumpName(field.Field.Name) + " ",
			comment: strings.ReplaceAll(field.Field.Doc, "\n", ""),
		}
		if 0 == i {
			item.typ = uName + "Field = iota"
		}
		list[i] = item
	}

	dbBlock(&sb, "const (", list, ")\n\n")

	sb.WriteString("var " + lName + "FieldDbNames = [")
	sb.WriteString(lName + "FieldCount]string{")
	list = make([]databaseField, len(fields))
	for i, field := range fields {
		item := databaseField{
			text:    "\"" + field.DB.Name + "\",",
			comment: strings.ReplaceAll(field.Field.Doc, "\n", ""),
		}
		list[i] = item
	}
	dbBlock(&sb, "", list, "}\n\n")

	sb.WriteString("var " + lName + "FieldNames = [")
	sb.WriteString(lName + "FieldCount]string{")
	list = make([]databaseField, len(fields))
	for i, field := range fields {
		item := databaseField{
			text:    "\"" + field.Field.Name + "\",",
			comment: strings.ReplaceAll(field.Field.Doc, "\n", ""),
		}
		list[i] = item
	}
	dbBlock(&sb, "", list, "}\n\n")

	sb.WriteString("var " + lName + "FieldDbGets = [")
	sb.WriteString(lName + "FieldCount]string{")
	for i, field := range fields {
		get := field.DB.Get
		if len(get) == 0 {
			get = field.DB.Name
		}

		item := databaseField{
			text:    "\"" + get + "\",",
			comment: strings.ReplaceAll(field.Field.Doc, "\n", ""),
		}
		list[i] = item
	}
	dbBlock(&sb, "", list, "}\n\n")

	sb.WriteString("func " + uName + "FieldsByDbName(fields ...string) []" + uName + "Field {\n")
	sb.WriteString("\tlist := make([]" + uName + "Field, 0, " + lName + "FieldCount)\n")
	sb.WriteString("\tfor _, item := range fields {\n")
	sb.WriteString("\t\tswitch item {\n")
	for _, field := range fields {
		sb.WriteString("\t\tcase \"" + field.DB.Name + "\":\n")
		sb.WriteString("\t\t\tlist = append(list, " + uName + "Field_" + tpl.HumpName(field.Field.Name) + ")\n")
	}
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn list\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func " + uName + "FieldsByName(fields ...string) []" + uName + "Field {\n")
	sb.WriteString("\tlist := make([]" + uName + "Field, 0, " + lName + "FieldCount)\n")
	sb.WriteString("\tfor _, item := range fields {\n")
	sb.WriteString("\t\tswitch item {\n")
	for _, field := range fields {
		sb.WriteString("\t\tcase \"" + field.Field.Name + "\":\n")
		sb.WriteString("\t\t\tlist = append(list, " + uName + "Field_" + tpl.HumpName(field.Field.Name) + ")\n")
	}
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn list\n")
	sb.WriteString("}\n\n")

	return sb.String(), nil
}

// buildScanBlock 生成 DbScanColumns/DbScanNames,与进程内 printScanData 一致。
func (c *Ctx) buildScanBlock(typ *wire.DataType, db *DB, fields []*DBField, key *DBField) (string, error) {
	join := len(db.Join) > 0 && strings.ToLower(db.Join[0]) == "true"

	name := tpl.HumpName(typ.Name)
	_, scan, _ := c.getItemAndValue(fields, "self")
	var sb strings.Builder
	sb.WriteString("func (val *" + name + ") DbScanColumns(columns ..." + name + "Field) []any {\n")
	if join {
		sb.WriteString("\tvar result []any\n")
		var before string
		err := c.enumField(typ, func(field *wire.Field, data *wire.DataType) error {
			db1 := getDB(field.Name, field.Tags)
			if db1 == nil {
				return nil
			}
			tableName := ""
			if field.Type != nil && field.Type.Kind == "ref" {
				tableName = field.Type.Name
			}
			table, _ := c.Find(tableName)
			if table == nil || table.Data == nil {
				return nil
			}

			fName := tpl.HumpName(field.Name)
			sb.WriteString("\tresult = append(result, val." + fName + ".DbScanColumns(utl.Slice(columns, func(i int, v " + name + "Field) " + tableName + "Field {\n")
			sb.WriteString("\t\treturn " + tableName + "Field(v)")
			if len(before) > 0 {
				sb.WriteString(" - " + tableName + "Field(" + tpl.LowerFirstName(before) + "FieldCount)")
			}
			sb.WriteString("\n")
			sb.WriteString("\t})...)...)\n")
			before = tableName
			return nil
		})
		if err != nil {
			return "", err
		}
	} else {
		sb.WriteString("\tif len(columns) == 0 {\n")
		sb.WriteString("\t\treturn []any{" + scan.String() + "}\n")
		sb.WriteString("\t}\n")

		sb.WriteString("\tresult := make([]any, 0, len(columns))\n")
		sb.WriteString("\tfor _, field := range columns {\n")
		sb.WriteString("\t\tswitch field {\n")
		for _, field := range fields {
			fieldName := tpl.HumpName(field.Field.Name)
			sb.WriteString("\t\tcase " + name + "Field_" + fieldName + ":\n")
			sb.WriteString("\t\t\tresult = append(result, " + c.converter(field, "val") + ")\n")
		}
		sb.WriteString("\t\tdefault:\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n")
	}
	sb.WriteString("\treturn result\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (val *" + name + ") DbScanNames(columns ..." + name + "Field) []string {\n")
	if join {
		sb.WriteString("\tvar result []string\n")
		var before string
		err := c.enumField(typ, func(field *wire.Field, data *wire.DataType) error {
			db1 := getDB(field.Name, field.Tags)
			if db1 == nil {
				return nil
			}
			tableName := ""
			if field.Type != nil && field.Type.Kind == "ref" {
				tableName = field.Type.Name
			}
			if tableName == "" {
				return nil
			}

			as := db1.Name
			if len(as) == 0 {
				as = tpl.UnderlineName(field.Name)
			}

			fName := tpl.HumpName(field.Name)
			sb.WriteString("\tfor _, name := range val." + fName + ".DbScanNames(utl.Slice(columns, func(i int, v " + name + "Field) " + tableName + "Field {\n")
			sb.WriteString("\t\treturn " + tableName + "Field(v)")
			if len(before) > 0 {
				sb.WriteString(" - " + tableName + "Field(" + tpl.LowerFirstName(before) + "FieldCount)")
			}
			sb.WriteString("\n")
			sb.WriteString("\t})...) {\n")
			sb.WriteString("\t\tresult = append(result, \"" + as + ".\"+name)\n")
			sb.WriteString("\t}\n")
			before = tableName
			return nil
		})
		if err != nil {
			return "", err
		}
	} else {
		sb.WriteString("\tif len(columns) == 0 {\n")
		sb.WriteString("\t\treturn " + tpl.LowerFirstName(typ.Name) + "FieldDbGets[:]\n")
		sb.WriteString("\t}\n")

		sb.WriteString("\tresult := make([]string, 0, len(columns))\n")
		sb.WriteString("\tfor _, field := range columns {\n")
		sb.WriteString("\t\tif int(field) < len(" + tpl.LowerFirstName(typ.Name) + "FieldNames) {\n")
		sb.WriteString("\t\t\tresult = append(result, " + tpl.LowerFirstName(typ.Name) + "FieldDbGets[field])\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n")
	}
	sb.WriteString("\treturn result\n")
	sb.WriteString("}\n\n")
	return sb.String(), nil
}

// buildNameBlock 生成 DbName 函数,与进程内 printNameData 一致。
func buildNameBlock(typ *wire.DataType) string {
	name := tpl.HumpName(typ.Name)
	return "func (val *" + name + ") DbName() string {\n" +
		"\treturn `" + tpl.UnderlineName(typ.Name) + "`\n" +
		"}\n\n"
}

// getItemAndValue 生成扫描列名列表、扫描取值与占位符。
func (c *Ctx) getItemAndValue(fields []*DBField, key string) ([]string, strings.Builder, strings.Builder) {
	var item []string
	scan := strings.Builder{}
	ques := strings.Builder{}
	isFist := true
	for _, field := range fields {
		get := ""
		if field.DB.Get == "-" {
			continue
		} else if 0 < len(field.DB.Get) {
			get = strings.ReplaceAll(field.DB.Get, "?", tpl.UnderlineName(field.DB.Name))
		} else if "self" == key {
			get = tpl.UnderlineName(field.DB.Name)
		} else {
			continue
		}
		if !isFist {
			scan.WriteString(", ")
			ques.WriteString(", ")
		}
		isFist = false
		item = append(item, get)
		scan.WriteString(c.converter(field, "val"))
		ques.WriteString("?")
	}
	return item, scan, ques
}

// converter 字段的数据库读写转换表达式(converter 注解支持 json/string)。
func (c *Ctx) converter(field *DBField, name string) string {
	fName := tpl.HumpName(field.Field.Name)
	if "json" == field.DB.Converter {
		return "db.NewJson(&" + name + "." + fName + ")"
	}
	if "string" == field.DB.Converter {
		return "db.NewString(&" + name + "." + fName + ")"
	}
	return "&" + name + "." + fName
}

func (c *Ctx) getLimit(fields []*DBField) (*DBField, bool) {
	for _, field := range fields {
		if 0 < len(field.DB.Limit) {
			return field, true
		}
	}
	return nil, false
}

func (c *Ctx) getOffset(fields []*DBField) (*DBField, bool) {
	for _, field := range fields {
		if 0 < len(field.DB.Offset) {
			return field, true
		}
	}
	return nil, false
}

// getKey 按名称在字段列表中查找 key 字段,返回 Go 类型与字段名(小写开头)。
func (c *Ctx) getKey(imp Imports, fields []*DBField, name string) (string, string, bool) {
	for _, field := range fields {
		if field.Field.Name == name {
			return c.goType(field.Field.Type, false, imp), tpl.LowerFirstName(field.Field.Name), true
		}
	}
	return "", "", false
}

// paramWhere 生成条件/分组/排序/分页代码块,与进程内 getParamWhere 一致。
func (c *Ctx) paramWhere(imp Imports, fields []*DBField, page, orderBy, groupBy bool) (string, error) {
	var sb strings.Builder

	for _, field := range fields {
		text := field.DB.Where
		fieldName := tpl.HumpName(field.Field.Name)
		for i, item := range text {
			if 1 == len(text) || !isArray(field.Field.Type) {
				if isNil(field.Field.Type) {
					sb.WriteString("\tif nil != g." + fieldName + " {\n")
					line, err := c.paramCode(imp, item, field, fields, "", "\t\ts", "g")
					if err != nil {
						return "", err
					}
					sb.WriteString(line)
					sb.WriteString("\t}\n")
				} else {
					line, err := c.paramCode(imp, item, field, fields, "", "\ts", "g")
					if err != nil {
						return "", err
					}
					sb.WriteString(line)
				}
			} else {
				if isNil(field.Field.Type) {
					sb.WriteString("\tif nil != g." + fieldName + " && " + strconv.Itoa(i) + " < len(g." + fieldName + ")")
					if field.Field.Type.Element != nil && field.Field.Type.Element.Empty {
						sb.WriteString(" && nil != g." + fieldName + "[" + strconv.Itoa(i) + "]")
					}
					sb.WriteString(" {\n")
					line, err := c.paramCode(imp, item, field, fields, "["+strconv.Itoa(i)+"]", "\t\ts", "g")
					if err != nil {
						return "", err
					}
					sb.WriteString(line)
					sb.WriteString("\t}\n")
				} else {
					line, err := c.paramCode(imp, item, field, fields, "["+strconv.Itoa(i)+"]", "\ts", "g")
					if err != nil {
						return "", err
					}
					sb.WriteString(line)
				}
			}
		}
	}

	if groupBy {
		isFist := true
		for _, field := range fields {
			group := field.DB.Group
			if 0 < len(group) {
				if isNil(field.Field.Type) {
					sb.WriteString("\tif nil != g." + tpl.HumpName(field.Field.Name) + " {\n")
					// 与进程内一致:if 行末尾的悬空 tab 使后续行多一级缩进
					sb.WriteString("\t")
				}
				if isFist {
					sb.WriteString("\ts.T(\"GROUP BY \")")
				} else {
					sb.WriteString("\ts.T(\", \")")
				}
				line, err := c.paramCode(imp, group, field, fields, "", "", "g")
				if err != nil {
					return "", err
				}
				sb.WriteString(line)
				if isNil(field.Field.Type) {
					sb.WriteString("\t}\n")
				}
				isFist = false
			}
		}
	}

	if orderBy {
		temp := false
		for _, field := range fields {
			if 0 < len(field.DB.Order) {
				temp = true
				break
			}
		}
		if temp {
			sb.WriteString("\torder := db.NewBuilder()\n")
			for _, field := range fields {
				order := field.DB.Order
				if 0 < len(order) {
					if !isArray(field.Field.Type) {
						tab := 0
						if isNil(field.Field.Type) {
							tab = 1
							sb.WriteString("\tif nil != g." + tpl.HumpName(field.Field.Name) + " {\n")
						}
						tb := strings.Repeat("\t", tab+1)
						sb.WriteString(tb + "if !order.IsEmpty() {\n")
						sb.WriteString(tb + "\torder.T(\", \")\n")
						sb.WriteString(tb + "}\n")
						sb.WriteString(tb + "order")
						line, err := c.paramCode(imp, order, field, fields, "", "", "g")
						if err != nil {
							return "", err
						}
						sb.WriteString(line)
						if tab > 0 {
							sb.WriteString("\t}\n")
						}
					} else {
						sb.WriteString("\tfor i := 0; i < len(g." + tpl.HumpName(field.Field.Name) + "); i += 2 {\n")
						sb.WriteString("\t\tif strings.ToUpper(g." + tpl.HumpName(field.Field.Name) + "[i+1]) == \"DESC\" || strings.ToUpper(g." + tpl.HumpName(field.Field.Name) + "[i+1]) == \"ASC\" {\n")
						sb.WriteString("\t\t\tif !order.IsEmpty() {\n")
						sb.WriteString("\t\t\t\torder.T(\", \")\n")
						sb.WriteString("\t\t\t}\n")
						sb.WriteString("\t\t\torder.T(g." + tpl.HumpName(field.Field.Name) + "[i]).T(\" \").T(strings.ToUpper(g." + tpl.HumpName(field.Field.Name) + "[i+1]))\n")
						sb.WriteString("\t\t}\n")
						sb.WriteString("\t}\n")
					}
				}
			}
			sb.WriteString("\tif !order.IsEmpty() {\n")
			sb.WriteString("\t\ts.T(\"ORDER BY \").Join(order)\n")
			sb.WriteString("\t}\n")
		}
	}

	if page {
		if limit, ok := c.getLimit(fields); ok {
			limitName := tpl.HumpName(limit.Field.Name)
			if offset, ok := c.getOffset(fields); ok {
				offsetName := tpl.HumpName(offset.Field.Name)
				if isNil(offset.Field.Type) || isNil(limit.Field.Type) {
					if !isNil(offset.Field.Type) {
						sb.WriteString("\tif nil != g." + limitName + " {\n")
					} else if !isNil(limit.Field.Type) {
						sb.WriteString("\tif nil != g." + offsetName + " {\n")
					} else {
						sb.WriteString("\tif nil != g." + offsetName + " && nil != g." + limitName + " {\n")
					}
					sb.WriteString("\t\ts.T(\" LIMIT " + offset.DB.Offset + ", " + limit.DB.Limit + "\")")
					sb.WriteString(".P(g." + offsetName + ", g." + limitName + ")\n")
					sb.WriteString("\t}\n")
				} else {
					sb.WriteString("\ts.T(\" LIMIT " + offset.DB.Offset + ", " + limit.DB.Limit + "\")")
					sb.WriteString(".P(g." + offsetName + ", g." + limitName + ")\n")
				}
			} else {
				if isNil(limit.Field.Type) {
					sb.WriteString("\tif nil != g." + limitName + " {\n")
					sb.WriteString("\t\ts.T(\" LIMIT " + limit.DB.Limit + "\")")
					sb.WriteString(".P(g." + limitName + ")\n")
					sb.WriteString("\t}\n")
				} else {
					sb.WriteString("\ts.T(\" LIMIT " + limit.DB.Limit + "\")")
					sb.WriteString(".P(g." + limitName + ")\n")
				}
			}
		}
	}
	return sb.String(), nil
}

func findField(fields []*DBField, name string) *DBField {
	for _, item := range fields {
		if item.Field.Name == name {
			return item
		}
	}
	return nil
}

var paramRex = regexp.MustCompile(`(\?{\w+})|(\${\w+})|\$|\?`)
var countRex = regexp.MustCompile(`(\${\w+})`)

// paramCode 把带占位符的 SQL 片段渲染成 db.Builder 调用链(一行)。
// 支持 ${field} 列名替换、?{field} 参数绑定、$/? 简写与数组展开。
func (c *Ctx) paramCode(imp Imports, text string, self *DBField, fields []*DBField, array, tab, object string) (string, error) {
	match := paramRex.FindAllStringSubmatchIndex(text, -1)
	var sb strings.Builder
	sb.WriteString(tab)
	if nil != match {
		var index = 0
		for _, item := range match {
			if 0 < item[0] {
				sb.WriteString(".T(\"")
				sb.WriteString(text[index:item[0]])
				sb.WriteString("\")")
			}
			t := text[item[0]:item[1]]
			if t == "$" || (2 < len(t) && "${" == t[0:2]) {
				field := self
				if t != "$" {
					field = findField(fields, t[2:len(t)-1])
				}

				sb.WriteString(".T(")
				if isNil(field.Field.Type) {
					sb.WriteString("*" + object + ".")
				} else {
					sb.WriteString(object + ".")
				}
				sb.WriteString(tpl.HumpName(field.Field.Name))
				sb.WriteString(")")
			} else if t == "?" || (2 < len(t) && "?{" == t[0:2]) {
				field := self
				temp := array
				if t != "?" {
					field = findField(fields, t[2:len(t)-1])
					temp = ""
				}
				if nil == field {
					return "", fmt.Errorf("invalid name: %s", t[2:len(t)-1])
				}
				if 0 == len(field.DB.Converter) && isArray(field.Field.Type) && 0 == len(temp) {
					sb.WriteString(".L(\",\", ")
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hutl", "utl")
					sb.WriteString("utl.ToAnyList(" + object + "." + tpl.HumpName(field.Field.Name) + ")...")
				} else {
					sb.WriteString(".V(")
					sb.WriteString(c.converter(field, object))
					sb.WriteString(temp)
				}
				sb.WriteString(")")
			}
			index = item[1]
		}
		if index < len(text) {
			sb.WriteString(".T(\"")
			sb.WriteString(text[index:])
			sb.WriteString("\")")
		}
	} else {
		sb.WriteString(".T(\"")
		sb.WriteString(text)
		sb.WriteString("\")")
	}
	sb.WriteString("\n")
	return sb.String(), nil
}
