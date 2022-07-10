package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"sort"
	"strconv"
	"strings"
)

type DB struct {
	index   int
	name    string
	key     bool
	typ     string
	insert  bool
	inserts bool
	update  bool
	set     bool
	del     bool
	get     bool
	find    bool
	count   bool
	table   bool
	where   string
	offset  string
	limit   string
}

type DBField struct {
	dbs   []*DB
	field *ast.Field
}

func (b *Builder) getDB(n string, tag []*ast.Tag) []*DB {
	var dbs []*DB
	for _, val := range tag {
		if 0 == strings.Index(val.Name.Name, "db") {
			var index int64 = 0
			if "db" != val.Name.Name {
				var err error
				index, err = strconv.ParseInt(val.Name.Name[2:], 10, 32)
				if nil != err {
					continue
				}
			}

			db := DB{
				index: int(index),
			}
			if nil != val.KV {
				for _, item := range val.KV {
					if "name" == item.Name.Name {
						db.name = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "key" == item.Name.Name {
						db.key = "key" == item.Value.Value[1:len(item.Value.Value)-1]
					} else if "typ" == item.Name.Name {
						db.typ = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "where" == item.Name.Name {
						db.where = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "offset" == item.Name.Name {
						db.offset = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "limit" == item.Name.Name {
						db.limit = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "insert" == item.Name.Name {
						db.insert = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "inserts" == item.Name.Name {
						db.inserts = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "update" == item.Name.Name {
						db.update = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "del" == item.Name.Name {
						db.del = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "get" == item.Name.Name {
						db.get = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "find" == item.Name.Name {
						db.find = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "table" == item.Name.Name {
						db.table = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "set" == item.Name.Name {
						db.set = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "count" == item.Name.Name {
						db.count = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					}
				}
			}
			if "" == db.name {
				db.name = build.StringToUnderlineName(n)
			}
			dbs = append(dbs, &db)
		}
	}

	sort.Slice(dbs, func(i, j int) bool {
		return dbs[i].index > dbs[j].index
	})
	return dbs
}

func (b *Builder) printDatabaseCode(dst *Writer, typ *ast.DataType) error {
	dst.Import("database/sql")

	dbs, fields, key, err := b.getDBField(typ)
	if 0 == len(dbs) || nil != err {
		return nil
	}

	fDbs := dbs
	fFields := fields
	fType := typ
	if !dbs[0].table && 0 < len(dbs[0].name) {
		table := b.build.GetDataType(b.getFile(typ.Name), dbs[0].name)
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

	if !dbs[0].table {
		return nil
	}
	if fDbs[0].table {
		b.printScanData(dst, typ, dbs[0], fields, key)
	}

	if fDbs[0].find {
		b.printFindData(dst, typ, dbs[0], fields, fType, fFields)
	}

	if fDbs[0].count {
		b.printCountData(dst, typ, dbs[0], fields, fType, fFields)
	}

	if fDbs[0].del {
		b.printDeleteData(dst, typ, dbs[0], fields, key)
	}

	if fDbs[0].insert {
		b.printInsertData(dst, typ, dbs[0], fields, key)
	}

	if fDbs[0].inserts {
		b.printInsertListData(dst, typ, dbs[0], fields, key)
	}

	if fDbs[0].update {
		b.printUpdateData(dst, typ, dbs[0], fields, key)
	}

	if fDbs[0].set {
		b.printSetData(dst, typ, dbs[0], fields, key)
	}

	if fDbs[0].get {
		if typ == fType {
			key.dbs[0].where = "AND id = ?"
			fFields = []*DBField{key}
		}
		b.printGetData(dst, typ, dbs[0], fields, fType, fFields)
	}
	return nil
}

func (b *Builder) getDBField(typ *ast.DataType) ([]*DB, []*DBField, *DBField, error) {
	dbs := b.getDB(typ.Name.Name, typ.Tags)
	if 0 == len(dbs) {
		return nil, nil, nil, nil
	}

	var fields []*DBField
	var key *DBField
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dbs := b.getDB(field.Name.Name, field.Tags)
		if 0 < len(dbs) {
			f := DBField{
				field: field,
				dbs:   dbs,
			}
			fields = append(fields, &f)
			if nil == key || dbs[0].key {
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

func (b *Builder) printScanData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)
	item, scan, _ := b.getItemAndValue(fields)
	dst.Code("func DbScan" + name + "(val *" + name + ") (string, []interface{}) {\n")
	dst.Code("	return `" + item.String() + "`,\n")
	dst.Code("		[]interface{}{" + scan.String() + "}\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) getItemAndValue(fields []*DBField) (strings.Builder, strings.Builder, strings.Builder) {
	item := strings.Builder{}
	scan := strings.Builder{}
	ques := strings.Builder{}
	isFist := true
	for _, field := range fields {
		if !isFist {
			item.WriteString(", ")
			scan.WriteString(", ")
			ques.WriteString(", ")
		}
		isFist = false
		item.WriteString(build.StringToUnderlineName(field.dbs[0].name))
		scan.WriteString("&val." + build.StringToHumpName(field.field.Name.Name))
		ques.WriteString("?")
	}
	return item, scan, ques

}

func (b *Builder) getParamWhere(dst *Writer, fields []*DBField) (*Writer, *Writer) {
	param := NewWriter(dst.pack)
	where := NewWriter(dst.pack)

	isFist := true
	for _, field := range fields {
		text := field.dbs[0].where
		if 0 < len(text) {
			if !isFist {
				param.Code(", ")
			}
			isFist = false
			param.Code(build.StringToFirstLower(field.field.Name.Name))
			param.Code(" ")
			b.printType(param, field.field.Type, false)

			if build.IsNil(field.field.Type) {
				where.Code("\tif nil != " + build.StringToFirstLower(field.field.Name.Name) + " {\n")
				where.Code("\t\tsql.WriteString(\" " + text + "\")\n")
				where.Code("\t\tparam = append(param")

				count := strings.Count(text, "?")
				for i := 0; i < count; i++ {
					where.Code(", " + build.StringToFirstLower(field.field.Name.Name))
				}
				where.Code(")\n")
				where.Code("\t}\n")
			} else {
				where.Code("\tsql.WriteString(\" " + text + "\")\n")
				where.Code("\tparam = append(param")

				count := strings.Count(text, "?")
				for i := 0; i < count; i++ {
					where.Code(", " + build.StringToFirstLower(field.field.Name.Name))
				}
				where.Code(")\n")
			}
		}
	}
	return param, where
}

func (b *Builder) printGetData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, fType *ast.DataType, fFields []*DBField) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	p, w := b.getParamWhere(dst, fFields)
	dst.AddImports(p.GetImports())
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields)

	dst.Code("func DbGet" + fName + "(db *sql.DB, " + p.code.String() + ") (*" + dName + ", error) {\n")
	dst.Code("\tsql := strings.Builder{}\n")
	dst.Code("\tvar param []interface{}\n")
	dst.Code("\tsql.WriteString(\"SELECT " + item.String() + " FROM " + db.name + " WHERE 1 = 1\")\n")
	dst.Code(w.code.String())

	dst.Code("	query, err := db.Query(sql.String(), param...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return nil, err\n")
	dst.Code("	}\n")
	dst.Code("	defer query.Close()\n")

	dst.Code("	if !query.Next() {\n")
	dst.Code("		return nil, nil\n")
	dst.Code("	}\n")

	dst.Code("	var val " + dName + "\n")
	dst.Code("	err = query.Scan(" + scan.String() + ")\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return nil, err\n")
	dst.Code("	}\n")
	dst.Code("	return &val, nil\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	dst.Import("strings")
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func DbInsert" + name + "(db *sql.DB, val *" + name + ") (int, error) {\n")
	dst.Code("	if nil == val {\n")
	dst.Code("		return 0, nil\n")
	dst.Code("	}\n")
	dst.Code("	value := strings.Builder{}\n")
	dst.Code("	ques := strings.Builder{}\n")
	dst.Code("	var param []interface{}\n\n")
	for _, field := range fields {
		fName := build.StringToHumpName(field.field.Name.Name)
		if build.IsNil(field.field.Type) {
			dst.Code("	if nil != val." + fName + " {\n")
			dst.Code("		value.WriteString(\"," + field.dbs[0].name + "\")\n")
			dst.Code("		ques.WriteString(\",?\")\n")
			dst.Code("		param = append(param, &val." + fName + ")\n\n")
			dst.Code("	}\n")
		} else {
			dst.Code("	value.WriteString(\"," + field.dbs[0].name + "\")\n")
			dst.Code("	ques.WriteString(\",?\")\n")
			dst.Code("	param = append(param, &val." + fName + ")\n\n")
		}
	}
	dst.Code("	valText := value.String()\n")
	dst.Code("	if len(valText) > 0 {\n")
	dst.Code("		valText = valText[1:]\n")
	dst.Code("	}\n")
	dst.Code("	quesText := ques.String()\n")
	dst.Code("	if len(quesText) > 0 {\n")
	dst.Code("		quesText = quesText[1:]\n")
	dst.Code("	}\n\n")
	dst.Code("	result, err := db.Exec(\"INSERT INTO " + db.name + "(\"+valText+\") VALUES(\"+quesText+\")\", param...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")
	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertListData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	dst.Import("strings")
	name := build.StringToHumpName(typ.Name.Name)
	item, scan, ques := b.getItemAndValue(fields)
	dst.Code("func DbInsertList" + name + "(db *sql.DB, val []*" + name + ") (int, error) {\n")
	dst.Code("	if nil == val || 0 == len(val) {\n")
	dst.Code("		return 0, nil\n")
	dst.Code("	}\n")
	dst.Code("	value := strings.Builder{}\n")
	dst.Code("	var param []interface{}\n")
	dst.Code("	value.Write([]byte(`INSERT INTO " + db.name + "(")
	dst.Code(item.String())
	dst.Code(") VALUES(`))\n")
	dst.Code("	for i, val := range val {\n")
	dst.Code("		if 0 != i {\n")
	dst.Code("			value.WriteString(\",\")\n")
	dst.Code("		}\n")
	dst.Code("		value.WriteString(\"(" + ques.String() + ")\")\n")
	dst.Code("		param = append(param, " + scan.String() + ")\n")
	dst.Code("	}\n")
	dst.Code("	result, err := db.Exec(value.String(), param...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")
	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n\n")
	dst.Code("}\n\n")
}

func (b *Builder) printUpdateData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func DbUpdate" + name + "(db *sql.DB, val *" + name + ") (int, error) {\n")
	dst.Code("	value := strings.Builder{}\n")
	dst.Code("	var param []interface{}\n\n")
	for _, field := range fields {
		if field == key {
			continue
		}
		fName := build.StringToHumpName(field.field.Name.Name)
		if build.IsNil(field.field.Type) {
			dst.Code("	if nil != val." + fName + " {\n")
			dst.Code("		value.WriteString(\"," + field.dbs[0].name + "= ? \")\n")
			dst.Code("		param = append(param, &val." + fName + ")\n")
			dst.Code("	}\n\n")
		} else {
			dst.Code("	value.WriteString(\"," + field.dbs[0].name + " = ?\")\n")
			dst.Code("	param = append(param, &val." + fName + ")\n\n")
		}
	}

	dst.Code("	valText := value.String()\n")
	dst.Code("	if 0 == len(valText) {\n")
	dst.Code("		return 0, nil\n")
	dst.Code("	}\n")
	dst.Code("	valText = valText[1:]\n")
	dst.Code("	param = append(param, &val." + build.StringToHumpName(key.field.Name.Name) + ")\n")
	dst.Code("	result, err := db.Exec(\"UPDATE  admin_info SET \"+valText+\" WHERE id = ?\", param...)\n")

	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")

	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printDeleteData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbDel" + name + "(db *sql.DB, " + build.StringToFirstLower(key.field.Name.Name) + " ")
	b.printType(dst, key.field.Type, false)
	dst.Code(") (int, error) {\n")

	dst.Code("	result, err := db.Exec(`DELETE FROM  " + db.name + " WHERE " + key.dbs[0].name + " = ?`, " + build.StringToFirstLower(key.field.Name.Name) + ")\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n\n")

	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printFindData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, fType *ast.DataType, fFields []*DBField) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	p, w := b.getParamWhere(dst, fFields)
	dst.AddImports(p.GetImports())
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields)

	dst.Code("func DbFind" + fName + "(db *sql.DB, " + p.code.String() + ") ([]" + dName + ", error) {\n")
	dst.Code("\tsql := strings.Builder{}\n")
	dst.Code("\tvar param []interface{}\n")
	dst.Code("\tsql.WriteString(\"SELECT " + item.String() + " FROM " + db.name + " WHERE 1 = 1\")\n")
	dst.Code(w.code.String())

	dst.Code("	query, err := db.Query(sql.String(), param...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return nil, err\n")
	dst.Code("	}\n")
	dst.Code("	defer query.Close()\n")
	dst.Code("	ret := make([]" + dName + ", 0)\n")
	dst.Code("\n")
	dst.Code("	for query.Next() {\n")
	dst.Code("		var val " + dName + "\n")
	dst.Code("		err = query.Scan(" + scan.String() + ")\n")
	dst.Code("		if err != nil {\n")
	dst.Code("			return nil, err\n")
	dst.Code("		}\n")
	dst.Code("		ret = append(ret, val)\n")
	dst.Code("	}\n")
	dst.Code("	return ret, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printCountData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, fType *ast.DataType, fFields []*DBField) {
	fName := build.StringToHumpName(fType.Name.Name)

	p, w := b.getParamWhere(dst, fFields)
	dst.AddImports(p.GetImports())
	dst.AddImports(w.GetImports())

	dst.Code("func DbCount" + fName + "(db *sql.DB, " + p.code.String() + ") (int64, error) {\n")
	dst.Code("\tsql := strings.Builder{}\n")
	dst.Code("\tvar param []interface{}\n")
	dst.Code("\tsql.WriteString(\"SELECT COUNT(1) FROM " + db.name + " WHERE 1 = 1\")\n")
	dst.Code(w.code.String())

	dst.Code("	query, err := db.Query(sql.String(), param...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")
	dst.Code("	defer query.Close()\n")
	dst.Code("\n")
	dst.Code("	if !query.Next() {\n")
	dst.Code("	  return 0, nil\n")
	dst.Code("	}\n")
	dst.Code("	var count int64\n")
	dst.Code("	err = query.Scan(&count)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")
	dst.Code("	return count, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printSetData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbSet" + name + "(db *sql.DB, val *" + name + ") (int, error) {\n")
	dst.Code("	result, err := db.Exec(`UPDATE  " + db.name + " SET ")
	isFist := true
	values := strings.Builder{}
	for _, field := range fields {
		if field == key {
			continue
		}
		if !isFist {
			dst.Code(", ")
		}
		isFist = false
		dst.Code(field.dbs[0].name + " = ?")
		_, _ = values.Write([]byte(", &val." + build.StringToHumpName(field.field.Name.Name)))
	}

	dst.Code(" WHERE " + key.dbs[0].name + " = ?`")
	dst.Code(values.String())
	dst.Code(", &val." + build.StringToHumpName(key.field.Name.Name) + " )\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")
	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}
