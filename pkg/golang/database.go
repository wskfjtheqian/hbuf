package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
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

func getDB(n string, tag []*ast.Tag) []*DB {
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
					} else if "types" == item.Name.Name {
						db.types = item.Value.Value[1 : len(item.Value.Value)-1]
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

func printScanData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func DbScan" + name + "(query *sql.Rows, val *" + name + ") error {\n")
	dst.Code("	return query.Scan(")
	isFist := true
	for _, field := range fields {
		if !isFist {
			dst.Code(", ")
		}
		isFist = false
		dst.Code("&val." + build.StringToHumpName(field.field.Name.Name))
	}

	dst.Code(")\n")
	dst.Code("}\n\n")

}

func printGetData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbGet" + name + "(db *sql.DB, " + build.StringToFirstLower(key.field.Name.Name) + " ")
	printType(dst, key.field.Type, false)
	dst.Code(") (*" + name + ", error) {\n")
	dst.Code("	query, err := db.Query(`SELECT ")

	isFist := true
	for _, field := range fields {
		if !isFist {
			dst.Code(", ")
		}
		isFist = false
		dst.Code(field.dbs[0].name)
	}
	dst.Code(" FROM " + db.name + " WHERE " + key.dbs[0].name + " = ?`, " + build.StringToFirstLower(key.field.Name.Name) + ")\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return nil, err\n")
	dst.Code("	}\n")
	dst.Code("	defer query.Close()\n")

	dst.Code("	if !query.Next() {\n")
	dst.Code("		return nil, nil\n")
	dst.Code("	}\n")

	dst.Code("	var val " + name + "\n")
	dst.Code("	err = DbScan" + name + "(query, &val)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return nil, err\n")
	dst.Code("	}\n")
	dst.Code("	return &val, nil\n")
	dst.Code("}\n\n")

}

func printInsertData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbInsert" + name + "(db *sql.DB, val *" + name + ") (int, error) {\n")
	dst.Code("\tif nil == val {\n")
	dst.Code("\t	return 0, nil\n")
	dst.Code("\t}\n")
	dst.Code("	result, err := db.Exec(`INSERT INTO " + db.name + "(")
	isFist := true
	value := strings.Builder{}
	param := strings.Builder{}
	for _, field := range fields {
		if !isFist {
			dst.Code(", ")
			_, _ = value.Write([]byte(", "))
		}
		isFist = false
		dst.Code(build.StringToUnderlineName(field.dbs[0].name))
		_, _ = value.Write([]byte("?"))

		_, _ = param.Write([]byte(", "))
		_, _ = param.Write([]byte("&val." + build.StringToHumpName(field.field.Name.Name)))
	}
	dst.Code(") VALUES(")
	dst.Code(value.String())
	dst.Code(")`")
	dst.Code(param.String())
	dst.Code(")\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")
	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func printInsertListData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func DbInsertList" + name + "(db *sql.DB, val []*" + name + ") (int, error) {\n")
	dst.Code("	if nil == val || 0 == len(val) {\n")
	dst.Code("		return 0, nil\n")
	dst.Code("	}\n")
	dst.Code("	value := strings.Builder{}\n")
	dst.Code("	var param []interface{}\n")
	dst.Code("	value.Write([]byte(`INSERT INTO " + db.name + "(")
	isFist := true
	value := strings.Builder{}
	param := strings.Builder{}
	for _, field := range fields {
		if !isFist {
			dst.Code(", ")
			_, _ = value.Write([]byte(", "))
			_, _ = param.Write([]byte(", "))
		}
		isFist = false
		dst.Code(build.StringToUnderlineName(field.dbs[0].name))
		_, _ = value.Write([]byte("?"))

		_, _ = param.Write([]byte("&item." + build.StringToHumpName(field.field.Name.Name)))
	}
	dst.Code(") VALUES(`))\n")
	dst.Code("	for i, item := range val {\n")
	dst.Code("		if 0 != i {\n")
	dst.Code("			value.Write([]byte(\",\"))\n")
	dst.Code("		}\n")
	dst.Code("		value.Write([]byte(\"(" + value.String() + ")\"))\n")
	dst.Code("		param = append(param, " + param.String() + ")\n")
	dst.Code("	}\n")
	dst.Code("	result, err := db.Exec(value.String(), param...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")
	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n\n")
	dst.Code("}\n\n")
}

func printUpdateData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbUpdate" + name + "(db *sql.DB, val *" + name + ") (int, error) {\n")
	dst.Code("	result, err := db.Exec(`UPDATE  " + db.name + " SET ")
	isFist := true
	values := strings.Builder{}
	for _, field := range fields {
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

func printDeleteData(dst *Writer, typ *ast.DataType, db *DB, fields []*DBField, key *DBField) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbDel" + name + "(db *sql.DB, " + build.StringToFirstLower(key.field.Name.Name) + " ")
	printType(dst, key.field.Type, false)
	dst.Code(") (int, error) {\n")

	dst.Code("	result, err := db.Exec(`DELETE FROM  " + db.name + " WHERE " + key.dbs[0].name + " = ?`, " + build.StringToFirstLower(key.field.Name.Name) + ")\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n\n")

	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func printDatabaseCode(dst *Writer, typ *ast.DataType) error {
	dst.Import("database/sql")
	dst.Import("strings")

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
