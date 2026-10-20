package goplus

import (
	"fmt"
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"strconv"
	"strings"
)

// dbImport 登记 database 输出通用 import。
func dbImport(imp Imports) {
	imp.Import("context", "")
	imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hsql", "db")
}

// buildJoinBlock 生成 DbJoin/DbWhere,与进程内 printJoin 一致。
func (c *Ctx) buildJoinBlock(imp Imports, typ *wire.DataType, db *DB, fType *wire.DataType) (string, error) {
	dbImport(imp)

	fName := tpl.HumpName(fType.Name)
	var sb strings.Builder
	sb.WriteString("func (g " + fName + ") DbJoin(ctx context.Context) *db.Builder {\n")
	sb.WriteString("\ts := db.NewBuilder()\n")
	first := true
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
		ref := table.Data
		dName := tpl.UnderlineName(ref.Name)
		if first {
			sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + dName + "\")\n")
			first = false
		} else {
			sb.WriteString("\tschema, table = db.SchemaTable(ctx, \"" + dName + "\")\n")
		}
		sb.WriteString("\ts")
		if len(db1.Join) > 0 {
			sb.WriteString(".T(\"" + db1.Join[0] + " JOIN\")")
		}
		name := db1.Name
		if len(name) == 0 {
			name = tpl.UnderlineName(field.Name)
		}
		sb.WriteString(".T(schema).T(\".\").T(table)")
		sb.WriteString(".T(\"AS " + name + "\")")
		if len(db1.Join) > 1 {
			sb.WriteString(".T(\"" + db1.Join[1] + "\")")
		}
		sb.WriteString("\n")
		return nil
	})
	if err != nil {
		return "", err
	}

	sb.WriteString("\treturn s\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (g " + fName + ") DbWhere() *db.Builder {\n")
	sb.WriteString("\ts := db.NewBuilder()\n")
	index := 0
	err = c.enumField(typ, func(field *wire.Field, data *wire.DataType) error {
		db1 := getDB(field.Name, field.Tags)
		if db1 == nil {
			return nil
		}
		name := db1.Name
		if len(name) == 0 {
			name = tpl.UnderlineName(field.Name)
		}

		if index == 0 {
			sb.WriteString("\ts.T(\"" + name + ".is_deleted = 0\")\n")
		} else {
			sb.WriteString("\ts.T(\"AND " + name + ".is_deleted = 0\")\n")
		}
		index++
		return nil
	})
	if err != nil {
		return "", err
	}
	sb.WriteString("\treturn s\n")
	sb.WriteString("}\n\n")
	return sb.String(), nil
}

// buildListBlock 生成 DbList,与进程内 printListData 一致。
// replaced 等价于进程内 typ != fType(table 间接引用)。
func (c *Ctx) buildListBlock(imp Imports, typ *wire.DataType, key string, db *DB, wFields []*DBField, fields []*DBField, fType *wire.DataType, cch *cache, replaced bool) string {
	dbImport(imp)

	fName := tpl.HumpName(fType.Name)
	dName := tpl.HumpName(typ.Name)
	if replaced {
		key = "self"
	} else if "self" == key {
		dName = fName
	}
	w, _ := c.paramWhere(imp, wFields, true, true, true)

	var sb strings.Builder
	sb.WriteString("func (g " + fName + ") DbList(ctx context.Context, columns ..." + dName + "Field) ([]" + dName + ", error) {\n")
	sb.WriteString("\tvar val " + dName + "\n")
	sb.WriteString("\ts := db.NewBuilder()\n")

	imp.Import("strings", "")
	sb.WriteString("\ts.T(\"SELECT \").T(strings.Join(val.DbScanNames(columns...), \", \")).T(\" FROM \")\n")
	join := len(db.Join) > 0 && strings.ToLower(db.Join[0]) == "true"
	if join {
		sb.WriteString("\ts.Join(val.DbJoin(ctx))\n")
		sb.WriteString("\ts.T(\"WHERE\").Join(val.DbWhere())\n")
	} else {
		sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
		sb.WriteString("\ts.T(schema).T(\".\").T(table)\n")
		sb.WriteString("\ts.T(\"WHERE is_deleted = 0\")\n")
	}

	sb.WriteString(w)

	tab := 0
	if nil != cch && !join {
		tab = 1
		imp.Import("math/rand", "")
		imp.Import("time", "")

		sb.WriteString("\treturn db.SaveCache(ctx, schema, table, s, time.Duration(rand.Intn(" + strconv.Itoa(cch.max))
		sb.WriteString("-" + strconv.Itoa(cch.min) + ")+" + strconv.Itoa(cch.min) + ")*time.Second, ")
		sb.WriteString("func(ctx context.Context) ([]" + dName + ", error) {\n")
	}
	imp.Import("database/sql", "")
	sb.WriteString(strings.Repeat("\t", tab+1) + "var ret []" + dName + "\n")
	sb.WriteString(strings.Repeat("\t", tab+1) + "_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "var val " + dName + "\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "err := rows.Scan(val.DbScanColumns(columns...)...)\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "if err == nil {\n")
	sb.WriteString(strings.Repeat("\t", tab+3) + "ret = append(ret, val)\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "}\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "return true, err\n")
	sb.WriteString(strings.Repeat("\t", tab+1) + "})\n")

	sb.WriteString(strings.Repeat("\t", tab+1) + "return ret, err\n")
	if nil != cch && !join {
		sb.WriteString("\t})\n")
	}
	sb.WriteString("}\n\n")
	return sb.String()
}

// buildListAsyncBlock 生成 DbListAsync,与进程内 printListAsyncData 一致。
func (c *Ctx) buildListAsyncBlock(imp Imports, typ *wire.DataType, key string, db *DB, wFields []*DBField, fields []*DBField, fType *wire.DataType, cch *cache, replaced bool) string {
	dbImport(imp)

	fName := tpl.HumpName(fType.Name)
	dName := tpl.HumpName(typ.Name)
	if replaced {
		key = "self"
	} else if "self" == key {
		dName = fName
	}
	w, _ := c.paramWhere(imp, wFields, true, true, true)

	var sb strings.Builder
	sb.WriteString("func (g " + fName + ") DbListAsync(ctx context.Context, fn func(ctx context.Context, ret *" + dName + ") (bool, error), columns ..." + dName + "Field) error {\n")
	sb.WriteString("\tvar val " + dName + "\n")
	sb.WriteString("\ts := db.NewBuilder()\n")

	imp.Import("strings", "")
	sb.WriteString("\ts.T(\"SELECT \").T(strings.Join(val.DbScanNames(columns...), \", \")).T(\" FROM \")\n")
	join := len(db.Join) > 0 && strings.ToLower(db.Join[0]) == "true"
	if join {
		sb.WriteString("\ts.Join(val.DbJoin(ctx))\n")
		sb.WriteString("\ts.T(\"WHERE\").Join(val.DbWhere())\n")
	} else {
		sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
		sb.WriteString("\ts.T(schema).T(\".\").T(table)\n")
		sb.WriteString("\ts.T(\"WHERE is_deleted = 0\")\n")
	}
	sb.WriteString(w)

	imp.Import("database/sql", "")
	sb.WriteString("\t_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	sb.WriteString("\t\tvar val " + dName + "\n")
	sb.WriteString("\t\terr := rows.Scan(val.DbScanColumns(columns...)...)\n")
	sb.WriteString("\t\tif err != nil {\n")
	sb.WriteString("\t\t\treturn false, err\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t\treturn fn(ctx, &val)\n")
	sb.WriteString("\t})\n")

	sb.WriteString("\treturn err\n")
	sb.WriteString("}\n\n")
	return sb.String()
}

// buildMapBlock 生成 DbMap,与进程内 printMapData 一致。
func (c *Ctx) buildMapBlock(imp Imports, key string, typ *wire.DataType, db *DB, wFields []*DBField, fields []*DBField, fType *wire.DataType, keyName string, cch *cache, replaced bool) string {
	dbImport(imp)

	kType, kName, ok := c.getKey(imp, fields, keyName)
	if !ok {
		return ""
	}
	fName := tpl.HumpName(fType.Name)
	dName := tpl.HumpName(typ.Name)
	if replaced {
		key = "self"
	} else if "self" == key {
		dName = fName
	}
	w, _ := c.paramWhere(imp, wFields, true, true, true)

	var sb strings.Builder
	sb.WriteString("func (g " + fName + ") DbMap(ctx context.Context, columns ..." + dName + "Field) (map[" + kType + "]*" + dName + ", error) {\n")
	sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
	sb.WriteString("\ts := db.NewBuilder()\n")

	imp.Import("strings", "")
	sb.WriteString("\tvar val " + dName + "\n")
	sb.WriteString("\ts.T(\"SELECT \").T(strings.Join(val.DbScanNames(columns...), \", \")).T(\" FROM \").T(schema).T(\".\").T(table).T(\" WHERE is_deleted = 0\")\n")
	sb.WriteString(w)

	tab := 0
	if nil != cch {
		tab = 1
		imp.Import("math/rand", "")
		imp.Import("time", "")

		sb.WriteString("\treturn db.SaveCache(ctx, schema, table, s, time.Duration(rand.Intn(" + strconv.Itoa(cch.max))
		sb.WriteString("-" + strconv.Itoa(cch.min) + ")+" + strconv.Itoa(cch.min) + ")*time.Second,")
		sb.WriteString(" func(ctx context.Context) (map[" + kType + "]*" + dName + ", error) {\n")
	}
	imp.Import("database/sql", "")
	sb.WriteString(strings.Repeat("\t", tab+1) + "ret := make(map[" + kType + "]*" + dName + ")\n")
	sb.WriteString(strings.Repeat("\t", tab+1) + "_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "var val " + dName + "\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "err := rows.Scan(val.DbScanColumns(columns...)...)\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "if err == nil {\n")
	sb.WriteString(strings.Repeat("\t", tab+3) + "ret[val.Get" + tpl.HumpName(kName) + "()] = &val\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "}\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "return true, err\n")
	sb.WriteString(strings.Repeat("\t", tab+1) + "})\n")
	sb.WriteString(strings.Repeat("\t", tab+1) + "return ret, err\n")
	if nil != cch {
		sb.WriteString("\t})\n")
	}
	sb.WriteString("}\n\n")
	return sb.String()
}

// countCode 渲染 count 注解中的 ${field} 占位符,与进程内 printCount 一致。
func (c *Ctx) countCode(text string, fields []*DBField) (string, error) {
	match := countRex.FindAllStringSubmatchIndex(text, -1)
	var sb strings.Builder
	if nil != match {
		var index = 0
		for _, item := range match {
			if 0 < item[0] {
				sb.WriteString(".T(\"")
				sb.WriteString(text[index:item[0]])
				sb.WriteString("\")")
			}
			t := text[item[0]:item[1]]
			if 2 < len(t) && "${" == t[0:2] {
				field := findField(fields, t[2:len(t)-1])
				if nil == field {
					return "", fmt.Errorf("not find field: %s", t[2:len(t)-1])
				}
				sb.WriteString(".T(g.Get" + tpl.HumpName(field.Field.Name) + "())")
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
	return sb.String(), nil
}

// buildCountBlock 生成 DbCount,与进程内 printCountData 一致。
func (c *Ctx) buildCountBlock(imp Imports, typ *wire.DataType, db *DB, wFields []*DBField, fType *wire.DataType, fFields []*DBField, cch *cache, count string) (string, error) {
	dbImport(imp)

	fName := tpl.HumpName(fType.Name)

	w, _ := c.paramWhere(imp, wFields, false, false, true)

	var sb strings.Builder
	sb.WriteString("func (g " + fName + ") DbCount(ctx context.Context) (int64, error) {\n")
	join := len(db.Join) > 0 && strings.ToLower(db.Join[0]) == "true"
	if join {
		dName := tpl.HumpName(typ.Name)
		sb.WriteString("\tvar v " + dName + "\n")
	} else {
		sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
	}

	sb.WriteString("\ts := db.NewBuilder()\n")
	sb.WriteString("\ts.T(\"SELECT COUNT(\")")

	code, err := c.countCode(count, wFields)
	if err != nil {
		return "", err
	}
	sb.WriteString(code)

	sb.WriteString(".T(\") FROM \")")
	if join {
		sb.WriteString(".Join(v.DbJoin(ctx)).T(\"WHERE\").Join(v.DbWhere())\n")
	} else {
		sb.WriteString(".T(schema).T(\".\").T(table).T(\" WHERE is_deleted = 0\")\n")
	}
	sb.WriteString(w)

	tab := 0
	if nil != cch && !join {
		tab = 1
		imp.Import("math/rand", "")
		imp.Import("time", "")

		sb.WriteString("\treturn db.SaveCache(ctx, schema, table, s, time.Duration(rand.Intn(" + strconv.Itoa(cch.max))
		sb.WriteString("-" + strconv.Itoa(cch.min) + ")+" + strconv.Itoa(cch.min) + ")*time.Second,")
		sb.WriteString(" func(ctx context.Context) (int64, error) {\n")
	}
	imp.Import("database/sql", "")
	sb.WriteString(strings.Repeat("\t", tab+1) + "var val int64\n")
	sb.WriteString(strings.Repeat("\t", tab+1) + "_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	sb.WriteString(strings.Repeat("\t", tab+2) + "return false, rows.Scan(&val)\n")
	sb.WriteString(strings.Repeat("\t", tab+1) + "})\n")
	sb.WriteString(strings.Repeat("\t", tab+1) + "return val, err\n")
	if nil != cch && !join {
		sb.WriteString("\t})\n")
	}
	sb.WriteString("}\n\n")
	return sb.String(), nil
}

// buildDeleteBlock 生成 DbDel(软删除),与进程内 printDeleteData 一致。
func (c *Ctx) buildDeleteBlock(imp Imports, db *DB, wFields []*DBField, fType *wire.DataType, isCache bool) string {
	dbImport(imp)

	fName := tpl.HumpName(fType.Name)

	w, _ := c.paramWhere(imp, wFields, false, false, false)

	var sb strings.Builder
	sb.WriteString("func (g " + fName + ") DbDel(ctx context.Context) (int64, int64, error) {\n")
	sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
	sb.WriteString("\ts := db.NewBuilder()\n")
	imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/htime", "")
	sb.WriteString("\ts.T(\"UPDATE \").T(schema).T(\".\").T(table).T(\" SET is_deleted = 1, delete_time = \").V(htime.NowTime()).T(\" WHERE is_deleted = 0\")\n")
	sb.WriteString(w)

	if isCache {
		sb.WriteString("\t_ = db.ClearCache(ctx, schema, table)\n")
	}
	sb.WriteString("\treturn s.Exec(ctx)\n")
	sb.WriteString("}\n\n")
	return sb.String()
}

// buildRemoveBlock 生成 DbRemove(物理删除),与进程内 printRemoveData 一致。
func (c *Ctx) buildRemoveBlock(imp Imports, db *DB, wFields []*DBField, fType *wire.DataType, isCache bool) string {
	dbImport(imp)

	fName := tpl.HumpName(fType.Name)

	w, _ := c.paramWhere(imp, wFields, false, false, false)

	var sb strings.Builder
	sb.WriteString("func (g " + fName + ") DbRemove(ctx context.Context) (int64, int64, error) {\n")
	sb.WriteString("\tschema, table := db.SchemaTable(ctx, \"" + db.Name + "\")\n")
	sb.WriteString("\ts := db.NewBuilder()\n")
	sb.WriteString("\ts.T(\"DELETE FROM \").T(schema).T(\".\").T(table).T(\" WHERE 1 = 1\")\n")
	sb.WriteString(w)

	if isCache {
		sb.WriteString("\t_ = db.ClearCache(ctx, schema, table)\n")
	}
	sb.WriteString("\treturn s.Exec(ctx)\n")
	sb.WriteString("}\n\n")
	return sb.String()
}
