package goplus

import (
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"strconv"
	"strings"
)

// buildInsertBlock 生成 DbInsert,与进程内 printInsertData 一致。
// replaced 等价于进程内 typ != fType,此时 val 视为 "parent"。
func (c *Ctx) buildInsertBlock(imp Imports, typ *wire.DataType, val string, db *DB, wFields []*DBField, fields []*DBField, fType *wire.DataType, key *DBField, cch *cache, replaced bool) (string, error) {
	dbImport(imp)

	fName := tpl.HumpName(fType.Name)
	if replaced {
		val = "parent"
	}
	// 与进程内一致:计算 where 仅为登记 import,输出不引用
	_, _ = c.paramWhere(imp, wFields, false, false, false)

	var sb strings.Builder
	sb.WriteString("func (g " + fName + ") DbInsert(ctx context.Context, action db.InsertAction) (int64, int64, error) {\n")
	sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
	sb.WriteString("\ts := db.NewBuilder()\n")
	sb.WriteString("\tv := db.NewBuilder()\n")
	sb.WriteString("\ts.T(action.String()).T(schema).T(\".\").T(table).T(\" (\").Del(\",\")\n")
	sb.WriteString("\tv.T(\"VALUES(\").Del(\",\")\n")

	for _, field := range fields {
		set := ""
		if 0 < len(field.DB.Set) {
			set = field.DB.Set
		} else if "self" == val {
			set = "?"
		} else {
			continue
		}

		tag := 0
		name := tpl.HumpName(field.Field.Name)
		if isNil(field.Field.Type) && !field.DB.Force {
			sb.WriteString("\tif nil != g." + name + " {\n")
			tag = 1
		}

		sb.WriteString(strings.Repeat("\t", tag+1) + "s.T(\",\").T(\"" + field.DB.Name + "\")\n")
		sb.WriteString(strings.Repeat("\t", tag+1) + "v.T(\",\")")
		line, err := c.paramCode(imp, set, field, fields, "", "", "g")
		if err != nil {
			return "", err
		}
		sb.WriteString(line)
		if tag > 0 {
			sb.WriteString("\t}\n")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\ts.T(\") \")\n")
	sb.WriteString("\tv.T(\") \")\n")
	sb.WriteString("\ts.Join(v)\n")
	if nil != cch {
		sb.WriteString("\t_ = db.ClearCache(ctx, schema, table)\n")
	}
	sb.WriteString("\treturn s.Exec(ctx)\n")
	sb.WriteString("}\n\n")
	return sb.String(), nil
}

// buildInsertBatchBlock 生成 DbInsertBatch,与进程内 printInsertBatchData 一致。
func (c *Ctx) buildInsertBatchBlock(imp Imports, typ *wire.DataType, db *DB, fields []*DBField, key *DBField, isCache bool) string {
	dbImport(imp)

	name := tpl.HumpName(typ.Name)
	var sb strings.Builder
	sb.WriteString("func (g " + name + ") DbInsertBatch(ctx context.Context, action db.InsertAction, list ...*" + name + ") (int64, int64, error) {\n")
	sb.WriteString("\tif nil == list || 0 == len(list) {\n")
	sb.WriteString("\t\treturn 0, 0, nil\n")
	sb.WriteString("\t}\n\n")

	sb.WriteString("\tvar total int64\n")
	sb.WriteString("\tvar lastId int64\n")
	sb.WriteString("\tlimit := int(db.BatchLimit(ctx))\n")
	sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n\n")

	sb.WriteString("\tfor j := 0; j < len(list); j += limit {\n")
	sb.WriteString("\t\ts := db.NewBuilder()\n")
	sb.WriteString("\t\ts.T(action.String()).T(schema).T(\".\").T(table).T(\" (")
	isFist := true
	for _, field := range fields {
		if !isFist {
			sb.WriteString(", ")
		}
		isFist = false
		sb.WriteString(field.DB.Name)
	}
	sb.WriteString(") VALUES\")\n")
	sb.WriteString("\t\tend := min(j+limit, len(list))\n")
	sb.WriteString("\t\tfor i, val := range list[j:end] {\n")
	sb.WriteString("\t\t\tif 0 != i {\n")
	sb.WriteString("\t\t\t\ts.T(\",\")\n")
	sb.WriteString("\t\t\t}\n")
	sb.WriteString("\t\t\ts.T(\"(\").L(\",\", ")
	isFist = true
	for _, field := range fields {
		if !isFist {
			sb.WriteString(", ")
		}
		isFist = false
		sb.WriteString(c.converter(field, "val"))
	}
	sb.WriteString(").T(\")\")\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t\tcount, id, err := s.Exec(ctx)\n")
	sb.WriteString("\t\tif err != nil {\n")
	sb.WriteString("\t\t\treturn 0, 0, err\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t\ttotal += count\n")
	sb.WriteString("\t\tlastId = id\n")

	sb.WriteString("\t}\n")

	if isCache {
		sb.WriteString("\t_ = db.ClearCache(ctx, schema, table)\n")
	}
	sb.WriteString("\treturn total, lastId, nil\n")
	sb.WriteString("}\n\n")
	return sb.String()
}

// buildUpdateBlock 生成 DbUpdate(where=SetWhereNil)与 DbSet(where=SetWhereNot),
// 与进程内 printUpdateData/printSetData 一致。
func (c *Ctx) buildUpdateBlock(imp Imports, typ *wire.DataType, key string, db *DB, wFields []*DBField, fields []*DBField, fType *wire.DataType, cch *cache, replaced bool, where SetWhere) (string, error) {
	dbImport(imp)

	fName := tpl.HumpName(fType.Name)
	object := "g"
	w, err := c.paramWhere(imp, wFields, false, false, false)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	method := "DbUpdate"
	if where == SetWhereNot {
		method = "DbSet"
	}
	sb.WriteString("func (g " + fName + ") " + method + "(ctx context.Context")
	if key == "parent" {
		object = "d"
		sb.WriteString(", d *" + tpl.HumpName(typ.Name))
	}
	sb.WriteString(") (int64, int64, error) {\n")
	sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
	sb.WriteString("\ts := db.NewBuilder()\n")
	sb.WriteString("\ts.T(\"UPDATE \").T(schema).T(\".\").T(table).T(\" SET \").Del(\",\")\n")

	allSet := !replaced || key == "parent"
	set, err := c.setCode(imp, typ, fields, allSet, where, object)
	if err != nil {
		return "", err
	}
	sb.WriteString(set)

	sb.WriteString("\ts.T(\"WHERE 1 = 1 \")\n")
	sb.WriteString(w)
	sb.WriteString("\n")
	if nil != cch {
		sb.WriteString("\t_ = db.ClearCache(ctx, schema, table)\n")
	}
	sb.WriteString("\treturn s.Exec(ctx)\n")
	sb.WriteString("}\n\n")
	return sb.String(), nil
}

// buildUpdateChangeBlock 生成 DbUpdateChange,与进程内 printUpdateChange 一致。
func (c *Ctx) buildUpdateChangeBlock(imp Imports, typ *wire.DataType, key string, db *DB, wFields []*DBField, fields []*DBField, fType *wire.DataType, cch *cache, replaced bool) (string, error) {
	dbImport(imp)

	uName := tpl.HumpName(fType.Name)
	w, err := c.paramWhere(imp, wFields, false, false, false)
	if err != nil {
		return "", err
	}

	object := "g"
	var sb strings.Builder
	sb.WriteString("func (g " + uName + ") DbUpdateChange(ctx context.Context")
	if key == "parent" {
		object = "d"
		sb.WriteString(", d *" + tpl.HumpName(typ.Name))
	}
	sb.WriteString(") (int64, int64, error) {\n")
	sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
	sb.WriteString("\ts := db.NewBuilder()\n")
	sb.WriteString("\ts.T(\"UPDATE \").T(schema).T(\".\").T(table).T(\" SET \").Del(\",\")\n")

	sb.WriteString("\tisChange := false\n")
	allSet := !replaced || key == "parent"
	set, err := c.setCode(imp, typ, fields, allSet, SetWhereChange, object)
	if err != nil {
		return "", err
	}
	sb.WriteString(set)

	sb.WriteString("\ts.T(\"WHERE 1 = 1 \")\n")
	sb.WriteString(w)
	sb.WriteString("\n")
	sb.WriteString("\tif !isChange {\n")
	sb.WriteString("\t\treturn 0, 0, nil\n")
	sb.WriteString("\t}\n")
	if nil != cch {
		sb.WriteString("\t_ = db.ClearCache(ctx, schema, table)\n")
	}
	sb.WriteString("\treturn s.Exec(ctx)\n")
	sb.WriteString("}\n\n")
	return sb.String(), nil
}

// buildChangesFieldsBlock 生成变更字段管理函数,与进程内 printChangesFields 一致。
func buildChangesFieldsBlock(fType *wire.DataType) string {
	uName := tpl.HumpName(fType.Name)
	lName := tpl.LowerFirstName(fType.Name)
	var sb strings.Builder
	sb.WriteString("func (g *" + uName + ") DbSetChangeFields(fields ..." + uName + "Field) {\n")
	sb.WriteString("\tg.changeFields = make([]bool, " + lName + "FieldCount)\n")
	sb.WriteString("\tfor _, item := range fields {\n")
	sb.WriteString("\t\tg.changeFields[item] = true\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (g *" + uName + ") DbAllChangeFields() {\n")
	sb.WriteString("\tg.changeFields = make([]bool, " + lName + "FieldCount)\n")
	sb.WriteString("\tfor i := 0; i < int(" + lName + "FieldCount); i++ {\n")
	sb.WriteString("\t\tg.changeFields[i] = true\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (g *" + uName + ") DbClearChangeFields() {\n")
	sb.WriteString("\tg.changeFields = make([]bool, " + lName + "FieldCount)\n")
	sb.WriteString("}\n\n")
	return sb.String()
}

// buildUpdateChangesBlock 生成 DbUpdateChanges,与进程内 printUpdateChanges 一致。
func (c *Ctx) buildUpdateChangesBlock(imp Imports, typ *wire.DataType, val string, db *DB, key *DBField, fields []*DBField, fType *wire.DataType, cch *cache) string {
	dbImport(imp)

	uName := tpl.HumpName(fType.Name)
	var sb strings.Builder
	sb.WriteString("func (g " + uName + ") DbUpdateChanges(ctx context.Context")
	if val == "parent" {
		sb.WriteString(", list ..." + tpl.HumpName(typ.Name))
	} else {
		sb.WriteString(", list ..." + uName)
	}
	sb.WriteString(") (int64, int64, error) {\n")

	lName := tpl.LowerFirstName(fType.Name)
	sb.WriteString("\tbuilder := make(map[" + uName + "Field]*db.Builder, " + lName + "FieldCount)\n")
	sb.WriteString("\tids := make([]any, 0, len(list))\n")
	sb.WriteString("\tfor i := 0; i < int(" + lName + "FieldCount); i++ {\n")
	sb.WriteString("\t\tbuilder[" + uName + "Field(i)] = db.NewBuilder()\n")
	sb.WriteString("\t}\n")

	sb.WriteString("\tfor _, item := range list {\n")
	sb.WriteString("\tisChange := false\n")
	keyName := tpl.HumpName(key.Field.Name)
	for _, field := range fields {
		if field == key {
			continue
		}
		name := tpl.HumpName(field.Field.Name)
		sb.WriteString("\t\tif item.changeFields[" + uName + "Field_" + name + "] {\n")
		sb.WriteString("\t\t\tisChange = false\n")
		sb.WriteString("\t\t\tbuilder[" + uName + "Field_" + name + "].T(\"WHEN\").V(&item." + keyName + ").T(\"THEN\").V(&item." + name + ")\n")
		sb.WriteString("\t\t}\n")
	}

	sb.WriteString("\t\tif !isChange {\n")
	sb.WriteString("\t\t\tids = append(ids, &item." + keyName + ")\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t}\n")

	sb.WriteString("\tif len(builder)== 0 {\n")
	sb.WriteString("\t\treturn 0, 0, nil\n")
	sb.WriteString("\t}\n")

	sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
	sb.WriteString("\ts := db.NewBuilder()\n")
	sb.WriteString("\ts.T(\"UPDATE \").T(schema).T(\".\").T(table).T(\" SET \").Del(\",\")\n")

	sb.WriteString("\tfor i := 0; i < int(" + lName + "FieldCount); i++ {\n")
	sb.WriteString("\t\tif val,ok:= builder[" + uName + "Field(i)];ok && !val.IsEmpty() {\n")
	sb.WriteString("\t\t\ts.T(\", \").T(" + uName + "Field(i).DbName()).T(\"= CASE id\").Join(val).T(\"ELSE\").T(" + uName + "Field(i).DbName()).T(\"END\")\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t}\n")

	sb.WriteString("\ts.T(\"WHERE 1 = 1 \")\n")
	sb.WriteString("\ts.T(\"AND id IN(\").L(\",\", ids...).T(\")\")\n")
	if nil != cch {
		sb.WriteString("\t_ = db.ClearCache(ctx, schema, table)\n")
	}
	sb.WriteString("\treturn s.Exec(ctx)\n")
	sb.WriteString("}\n\n")
	return sb.String()
}

// SetWhere SET 语句生成条件(与进程内 SetWhere 一致)。
type SetWhere int

const (
	SetWhereNot    SetWhere = 0
	SetWhereNil    SetWhere = 1
	SetWhereChange SetWhere = 2
)

// setCode 生成 SET 字段赋值代码块,与进程内 printSet 一致。
func (c *Ctx) setCode(imp Imports, typ *wire.DataType, fields []*DBField, allSet bool, where SetWhere, object string) (string, error) {
	uName := tpl.HumpName(typ.Name)
	var sb strings.Builder
	for _, field := range fields {
		set := ""
		if 0 < len(field.DB.Set) {
			set = field.DB.Name + " = " + field.DB.Set
		} else if allSet {
			set = field.DB.Name + " = ?"
		} else {
			continue
		}

		isWhere := false
		name := tpl.HumpName(field.Field.Name)
		if where == SetWhereChange {
			isWhere = true
			sb.WriteString("\tif " + object + ".changeFields[" + uName + "Field_" + name + "] {\n")
			sb.WriteString("\t\tisChange = true\n")
			// 与进程内一致:if 块末尾的悬空 tab 使后续行多一级缩进
			sb.WriteString("\t")
		} else if where == SetWhereNil && isNil(field.Field.Type) && !field.DB.Force {
			isWhere = true
			sb.WriteString("\tif nil != " + object + "." + name + " {\n")
			// 与进程内一致:if 块末尾的悬空 tab 使后续行多一级缩进
			sb.WriteString("\t")
		}

		sb.WriteString("\ts.T(\",\")")
		line, err := c.paramCode(imp, set, field, fields, "", "", object)
		if err != nil {
			return "", err
		}
		sb.WriteString(line)

		if isWhere {
			sb.WriteString("\t}\n")
		}
	}
	return sb.String(), nil
}

// buildGetBlock 生成 DbGet,与进程内 printGetData 一致。
func (c *Ctx) buildGetBlock(imp Imports, typ *wire.DataType, key string, db *DB, wFields []*DBField, fields []*DBField, fType *wire.DataType, cch *cache, replaced bool) (string, error) {
	dbImport(imp)

	fName := tpl.HumpName(fType.Name)
	dName := tpl.HumpName(typ.Name)
	if !replaced {
		key = "self"
	} else if "self" == key {
		dName = fName
	} else {
		key = "self"
	}

	w, err := c.paramWhere(imp, wFields, false, true, true)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("func (g " + fName + ") DbGet(ctx context.Context, columns ..." + dName + "Field) (*" + dName + ", error) {\n")
	sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
	sb.WriteString("\ts := db.NewBuilder()\n")
	imp.Import("strings", "")
	sb.WriteString("\tvar val *" + dName + "\n")

	sb.WriteString("\ts.T(\"SELECT \").T(strings.Join(val.DbScanNames(columns...), \", \")).T(\" FROM \").T(schema).T(\".\").T(table).T(\" WHERE is_deleted = 0\")\n")
	sb.WriteString(w)
	sb.WriteString("\ts.T(\" LIMIT 1\")\n")

	tab := 0
	if nil != cch {
		tab = 1
		imp.Import("math/rand", "")
		imp.Import("time", "")

		sb.WriteString("\treturn db.SaveCache(ctx, schema, table, s, time.Duration(rand.Intn(" + strconv.Itoa(cch.max))
		sb.WriteString("-" + strconv.Itoa(cch.min) + ")+" + strconv.Itoa(cch.min) + ")*time.Second,")
		sb.WriteString(" func(ctx context.Context) (*" + dName + ", error) {\n")
	}
	imp.Import("database/sql", "")
	sb.WriteString(strings.Repeat("\t", tab+1) + "_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "val = &" + dName + "{}\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "return false, rows.Scan(val.DbScanColumns(columns...)...)\n")
	sb.WriteString(strings.Repeat("\t", tab+1) + "})\n")
	sb.WriteString(strings.Repeat("\t", tab+1) + "return val, err\n")
	if nil != cch {
		sb.WriteString("\t})\n")
	}
	sb.WriteString("}\n\n")
	return sb.String(), nil
}
