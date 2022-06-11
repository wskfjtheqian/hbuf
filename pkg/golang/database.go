package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
	"sort"
	"strconv"
	"strings"
)

type DB struct {
	index int
	name  string
	key   bool
	types string
}

type DBField struct {
	dbs   []*DB
	field *ast.Field
}

func getDB(n string, tag map[string]*ast.Tag) []*DB {
	var dbs []*DB
	for _, value := range tag {
		if 0 == strings.Index(value.Name.Name, "db") {
			var index int64 = 0
			if "db" != value.Name.Name {
				var err error
				index, err = strconv.ParseInt(value.Name.Name[2:], 10, 32)
				if nil != err {
					continue
				}
			}

			val := value.Value.Value
			val = val[1 : len(val)-1]
			var name string
			var types string
			var key bool = false

			arr := strings.Split(val, ";")
			for _, a := range arr {
				if 0 == strings.Index(a, "name:") {
					name = a[len("name:"):]
				} else if 0 == strings.Index(a, "key:") {
					key = "key" == a[len("key:"):]
				} else if 0 == strings.Index(a, "type:") {
					types = a[len("type:"):]
				}
			}
			if "" == name {
				name = build.StringToUnderlineName(n)
			}
			db := DB{
				index: int(index),
				name:  name,
				key:   key,
				types: types,
			}
			dbs = append(dbs, &db)
		}
	}

	sort.Slice(dbs, func(i, j int) bool {
		return dbs[i].index > dbs[j].index
	})
	return dbs
}

func printScanData(dst io.Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)
	_, _ = dst.Write([]byte("func DbScan" + name + "(query *sql.Rows, val *" + name + ") error {\n"))
	_, _ = dst.Write([]byte("	return query.Scan("))
	isFist := true
	for _, field := range fields {
		if !isFist {
			_, _ = dst.Write([]byte(", "))
		}
		isFist = false
		_, _ = dst.Write([]byte("&val." + build.StringToHumpName(field.field.Name.Name)))
	}

	_, _ = dst.Write([]byte(")\n"))
	_, _ = dst.Write([]byte("}\n\n"))

}

func printGetData(dst io.Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)

	_, _ = dst.Write([]byte("func DbGet" + name + "(db *sql.DB, " + build.StringToFirstLower(key.field.Name.Name) + " "))
	printType(dst, key.field.Type, false)
	_, _ = dst.Write([]byte(") (*" + name + ", error) {\n"))
	_, _ = dst.Write([]byte("	query, err := db.Query(`SELECT "))

	isFist := true
	for _, field := range fields {
		if !isFist {
			_, _ = dst.Write([]byte(", "))
		}
		isFist = false
		_, _ = dst.Write([]byte(field.dbs[0].name))
	}
	_, _ = dst.Write([]byte(" FROM " + db.name + " WHERE " + key.dbs[0].name + " = ?`, " + build.StringToFirstLower(key.field.Name.Name) + ")\n"))
	_, _ = dst.Write([]byte("	if err != nil {\n"))
	_, _ = dst.Write([]byte("		return nil, err\n"))
	_, _ = dst.Write([]byte("	}\n"))
	_, _ = dst.Write([]byte("	defer query.Close()\n"))

	_, _ = dst.Write([]byte("	if !query.Next() {\n"))
	_, _ = dst.Write([]byte("		return nil, nil\n"))
	_, _ = dst.Write([]byte("	}\n"))

	_, _ = dst.Write([]byte("	var val " + name + "\n"))
	_, _ = dst.Write([]byte("	err = DbScan" + name + "(query, &val)\n"))
	_, _ = dst.Write([]byte("	if err != nil {\n"))
	_, _ = dst.Write([]byte("		return nil, err\n"))
	_, _ = dst.Write([]byte("	}\n"))
	_, _ = dst.Write([]byte("	return &val, nil\n"))
	_, _ = dst.Write([]byte("}\n\n"))

}

func printInsertData(dst io.Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)

	_, _ = dst.Write([]byte("func DbInsert" + name + "(db *sql.DB, val *" + name + ") (int, error) {\n"))
	_, _ = dst.Write([]byte("\tif nil == val {\n"))
	_, _ = dst.Write([]byte("\t	return 0, nil\n"))
	_, _ = dst.Write([]byte("\t}\n"))
	_, _ = dst.Write([]byte("	result, err := db.Exec(`INSERT INTO " + db.name + "("))
	isFist := true
	value := strings.Builder{}
	param := strings.Builder{}
	for _, field := range fields {
		if !isFist {
			_, _ = dst.Write([]byte(", "))
			_, _ = value.Write([]byte(", "))
		}
		isFist = false
		_, _ = dst.Write([]byte(build.StringToUnderlineName(field.dbs[0].name)))
		_, _ = value.Write([]byte("?"))

		_, _ = param.Write([]byte(", "))
		_, _ = param.Write([]byte("&val." + build.StringToHumpName(field.field.Name.Name)))
	}
	_, _ = dst.Write([]byte(") VALUES("))
	_, _ = dst.Write([]byte(value.String()))
	_, _ = dst.Write([]byte(")`"))
	_, _ = dst.Write([]byte(param.String()))
	_, _ = dst.Write([]byte(")\n"))
	_, _ = dst.Write([]byte("	if err != nil {\n"))
	_, _ = dst.Write([]byte("		return 0, err\n"))
	_, _ = dst.Write([]byte("	}\n"))
	_, _ = dst.Write([]byte("	count, err := result.RowsAffected()\n"))
	_, _ = dst.Write([]byte("	return int(count), err\n"))
	_, _ = dst.Write([]byte("}\n\n"))
}

func printInsertListData(dst io.Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)
	_, _ = dst.Write([]byte("func DbInsertList" + name + "(db *sql.DB, val []*" + name + ") (int, error) {\n"))
	_, _ = dst.Write([]byte("	if nil == val || 0 == len(val) {\n"))
	_, _ = dst.Write([]byte("		return 0, nil\n"))
	_, _ = dst.Write([]byte("	}\n"))
	_, _ = dst.Write([]byte("	value := strings.Builder{}\n"))
	_, _ = dst.Write([]byte("	var param []interface{}\n"))
	_, _ = dst.Write([]byte("	value.Write([]byte(`INSERT INTO " + db.name + "("))
	isFist := true
	value := strings.Builder{}
	param := strings.Builder{}
	for _, field := range fields {
		if !isFist {
			_, _ = dst.Write([]byte(", "))
			_, _ = value.Write([]byte(", "))
			_, _ = param.Write([]byte(", "))
		}
		isFist = false
		_, _ = dst.Write([]byte(build.StringToUnderlineName(field.dbs[0].name)))
		_, _ = value.Write([]byte("?"))

		_, _ = param.Write([]byte("&item." + build.StringToHumpName(field.field.Name.Name)))
	}
	_, _ = dst.Write([]byte(") VALUES)`))\n"))
	_, _ = dst.Write([]byte("	for i, item := range val {\n"))
	_, _ = dst.Write([]byte("		if 0 != i {\n"))
	_, _ = dst.Write([]byte("			value.Write([]byte(\",\"))\n"))
	_, _ = dst.Write([]byte("		}\n"))
	_, _ = dst.Write([]byte("		value.Write([]byte(\"(" + value.String() + ")\"))\n"))
	_, _ = dst.Write([]byte("		param = append(param, " + param.String() + ")\n"))
	_, _ = dst.Write([]byte("	}\n"))
	_, _ = dst.Write([]byte("	result, err := db.Exec(value.String(), param...)\n"))
	_, _ = dst.Write([]byte("	if err != nil {\n"))
	_, _ = dst.Write([]byte("		return 0, err\n"))
	_, _ = dst.Write([]byte("	}\n"))
	_, _ = dst.Write([]byte("	count, err := result.RowsAffected()\n"))
	_, _ = dst.Write([]byte("	return int(count), err\n\n"))
	_, _ = dst.Write([]byte("}\n\n"))
}

func printUpdateData(dst io.Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)

	_, _ = dst.Write([]byte("func DbUpdate" + name + "(db *sql.DB, val *" + name + ") (int, error) {\n"))
	_, _ = dst.Write([]byte("	result, err := db.Exec(`UPDATE  " + db.name + " SET "))
	isFist := true
	values := strings.Builder{}
	for _, field := range fields {
		if !isFist {
			_, _ = dst.Write([]byte(", "))
		}
		isFist = false
		_, _ = dst.Write([]byte(field.dbs[0].name + " = ?"))
		_, _ = values.Write([]byte(", &val." + build.StringToHumpName(field.field.Name.Name)))
	}

	_, _ = dst.Write([]byte(" WHERE " + key.dbs[0].name + " = ?`"))
	_, _ = dst.Write([]byte(values.String()))
	_, _ = dst.Write([]byte(", &val." + build.StringToHumpName(key.field.Name.Name) + " )\n"))
	_, _ = dst.Write([]byte("	if err != nil {\n"))
	_, _ = dst.Write([]byte("		return 0, err\n"))
	_, _ = dst.Write([]byte("	}\n"))

	_, _ = dst.Write([]byte("	count, err := result.RowsAffected()\n"))
	_, _ = dst.Write([]byte("	return int(count), err\n"))
	_, _ = dst.Write([]byte("}\n\n"))
}

func printDeleteData(dst io.Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)

	_, _ = dst.Write([]byte("func DbDel" + name + "(db *sql.DB, " + build.StringToFirstLower(key.field.Name.Name) + " "))
	printType(dst, key.field.Type, false)
	_, _ = dst.Write([]byte(") (int, error) {\n"))

	_, _ = dst.Write([]byte("	result, err := db.Exec(`DELETE FROM  " + db.name + " WHERE " + key.dbs[0].name + " = ?`, " + build.StringToFirstLower(key.field.Name.Name) + ")\n"))
	_, _ = dst.Write([]byte("	if err != nil {\n"))
	_, _ = dst.Write([]byte("		return 0, err\n"))
	_, _ = dst.Write([]byte("	}\n\n"))

	_, _ = dst.Write([]byte("	count, err := result.RowsAffected()\n"))
	_, _ = dst.Write([]byte("	return int(count), err\n"))
	_, _ = dst.Write([]byte("}\n\n"))
}

func printDatabase(dst io.Writer, typ *ast.DataType) error {
	dbs := getDB(typ.Name.Name, typ.Tags)
	if 0 == len(dbs) {
		return nil
	}

	var fields []*DBField
	var key *DBField
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dbs := getDB(field.Name.Name, field.Tags)
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
		return err
	}

	if 0 < len(dbs) {
		printScanData(dst, typ, dbs[0], fields, key)
		printGetData(dst, typ, dbs[0], fields, key)
		printDeleteData(dst, typ, dbs[0], fields, key)
		printUpdateData(dst, typ, dbs[0], fields, key)
		printInsertData(dst, typ, dbs[0], fields, key)
		printInsertListData(dst, typ, dbs[0], fields, key)
	}
	return nil
}
