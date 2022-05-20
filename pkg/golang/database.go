package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
	"strings"
)

func printScanData(dst io.Writer, typ *ast.DataType) error {
	name := build.StringToHumpName(typ.Name.Name)
	_, _ = dst.Write([]byte("func DbScan" + name + "(query *sql.Rows, val *" + name + ") error {\n"))
	_, _ = dst.Write([]byte("	return query.Scan("))
	isFist := true
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		_, ok := field.Tags["db"]
		if ok {
			if !isFist {
				_, _ = dst.Write([]byte(", "))
			}
			isFist = false
			_, _ = dst.Write([]byte("&val." + build.StringToHumpName(field.Name.Name)))
		}
		return nil
	})
	if nil != err {
		return err
	}
	_, _ = dst.Write([]byte(")\n"))
	_, _ = dst.Write([]byte("}\n\n"))
	return nil
}

func printGetData(dst io.Writer, typ *ast.DataType) error {
	name := build.StringToHumpName(typ.Name.Name)
	dbName := typ.Tags["db"].Value.Value
	if 0 != len(dbName) {
		dbName = dbName[1 : len(dbName)-1]
	} else {
		dbName = typ.Name.Name
	}

	_, _ = dst.Write([]byte("func DbGet" + name + "(db *sql.DB, id int64) (*" + name + ", error) {\n"))
	_, _ = dst.Write([]byte("	query, err := db.Query(`SELECT "))
	isFist := true
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		db, ok := field.Tags["db"]
		if ok {
			if !isFist {
				_, _ = dst.Write([]byte(", "))
			}
			isFist = false
			fieldName := db.Value.Value
			fieldName = fieldName[1 : len(fieldName)-1]
			if 0 >= len(fieldName) {
				fieldName = field.Name.Name
			}
			_, _ = dst.Write([]byte(build.StringToUnderlineName(fieldName)))
		}
		return nil
	})
	if nil != err {
		return err
	}

	_, _ = dst.Write([]byte(" FROM " + dbName + " WHERE id = ?`, id)\n"))
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

	return nil
}

func printInsertData(dst io.Writer, typ *ast.DataType) error {
	name := build.StringToHumpName(typ.Name.Name)
	dbName := typ.Tags["db"].Value.Value
	if 0 != len(dbName) {
		dbName = dbName[1 : len(dbName)-1]
	} else {
		dbName = typ.Name.Name
	}

	_, _ = dst.Write([]byte("func DbInsert" + name + "(db *sql.DB, val *UserInfo) (int, error) {\n"))
	_, _ = dst.Write([]byte("\tif nil == val {\n"))
	_, _ = dst.Write([]byte("\t	return 0, nil\n"))
	_, _ = dst.Write([]byte("\t}\n"))
	_, _ = dst.Write([]byte("	result, err := db.Exec(`INSERT INTO user_info("))
	isFist := true
	value := strings.Builder{}
	param := strings.Builder{}
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		db, ok := field.Tags["db"]
		if ok {
			if !isFist {
				_, _ = dst.Write([]byte(", "))
				_, _ = value.Write([]byte(", "))
			}
			isFist = false
			fieldName := db.Value.Value
			fieldName = fieldName[1 : len(fieldName)-1]
			if 0 >= len(fieldName) {
				fieldName = field.Name.Name
			}
			_, _ = dst.Write([]byte(build.StringToUnderlineName(fieldName)))
			_, _ = value.Write([]byte("?"))

			_, _ = param.Write([]byte(", "))
			_, _ = param.Write([]byte("&val." + build.StringToHumpName(field.Name.Name)))
		}
		return nil
	})
	if nil != err {
		return err
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
	return nil
}

func printInsertListData(dst io.Writer, typ *ast.DataType) error {
	name := build.StringToHumpName(typ.Name.Name)
	dbName := typ.Tags["db"].Value.Value
	if 0 != len(dbName) {
		dbName = dbName[1 : len(dbName)-1]
	} else {
		dbName = typ.Name.Name
	}

	_, _ = dst.Write([]byte("func DbInsertList" + name + "(db *sql.DB, val []*UserInfo) (int, error) {\n"))
	_, _ = dst.Write([]byte("	if nil == val || 0 == len(val) {\n"))
	_, _ = dst.Write([]byte("		return 0, nil\n"))
	_, _ = dst.Write([]byte("	}\n"))
	_, _ = dst.Write([]byte("	value := strings.Builder{}\n"))
	_, _ = dst.Write([]byte("	var param []interface{}\n"))
	_, _ = dst.Write([]byte("	value.Write([]byte(`INSERT INTO user_info("))
	isFist := true
	value := strings.Builder{}
	param := strings.Builder{}
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		db, ok := field.Tags["db"]
		if ok {
			if !isFist {
				_, _ = dst.Write([]byte(", "))
				_, _ = value.Write([]byte(", "))
				_, _ = param.Write([]byte(", "))
			}
			isFist = false
			fieldName := db.Value.Value
			fieldName = fieldName[1 : len(fieldName)-1]
			if 0 >= len(fieldName) {
				fieldName = field.Name.Name
			}
			_, _ = dst.Write([]byte(build.StringToUnderlineName(fieldName)))
			_, _ = value.Write([]byte("?"))

			_, _ = param.Write([]byte("&item." + build.StringToHumpName(field.Name.Name)))
		}
		return nil
	})
	if nil != err {
		return err
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
	return nil
}

func printUpdateData(dst io.Writer, typ *ast.DataType) error {
	name := build.StringToHumpName(typ.Name.Name)
	dbName := typ.Tags["db"].Value.Value
	if 0 != len(dbName) {
		dbName = dbName[1 : len(dbName)-1]
	} else {
		dbName = typ.Name.Name
	}

	_, _ = dst.Write([]byte("func DbUpdate" + name + "(db *sql.DB, val *UserInfo) (int, error) {\n"))
	_, _ = dst.Write([]byte("	result, err := db.Exec(`UPDATE  user_info SET "))
	isFist := true
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		db, ok := field.Tags["db"]
		if ok {
			if !isFist {
				_, _ = dst.Write([]byte(", "))
			}
			isFist = false
			fieldName := db.Value.Value
			fieldName = fieldName[1 : len(fieldName)-1]
			if 0 >= len(fieldName) {
				fieldName = field.Name.Name
			}
			_, _ = dst.Write([]byte(build.StringToUnderlineName(fieldName)))
			_, _ = dst.Write([]byte(" = ?"))
		}
		return nil
	})
	if nil != err {
		return err
	}
	_, _ = dst.Write([]byte(" WHERE id = ?`, val.UserId"))
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		_, ok := field.Tags["db"]
		if ok {
			_, _ = dst.Write([]byte(", &val." + build.StringToHumpName(field.Name.Name)))
		}
		return nil
	})
	if nil != err {
		return err
	}
	_, _ = dst.Write([]byte(" )\n"))
	_, _ = dst.Write([]byte("	if err != nil {\n"))
	_, _ = dst.Write([]byte("		return 0, err\n"))
	_, _ = dst.Write([]byte("	}\n"))

	_, _ = dst.Write([]byte("	count, err := result.RowsAffected()\n"))
	_, _ = dst.Write([]byte("	return int(count), err\n"))
	_, _ = dst.Write([]byte("}\n\n"))
	return nil
}

func printDeleteData(dst io.Writer, typ *ast.DataType) {
	name := build.StringToHumpName(typ.Name.Name)
	dbName := typ.Tags["db"].Value.Value
	if 0 != len(dbName) {
		dbName = dbName[1 : len(dbName)-1]
	} else {
		dbName = typ.Name.Name
	}
	_, _ = dst.Write([]byte("func DbDel" + name + "(db *sql.DB, id int64) (int, error) {\n"))
	_, _ = dst.Write([]byte("	result, err := db.Exec(`DELETE FROM  " + build.StringToUnderlineName(typ.Name.Name) + " WHERE id = ?`, id)\n"))
	_, _ = dst.Write([]byte("	if err != nil {\n"))
	_, _ = dst.Write([]byte("		return 0, err\n"))
	_, _ = dst.Write([]byte("	}\n\n"))

	_, _ = dst.Write([]byte("	count, err := result.RowsAffected()\n"))
	_, _ = dst.Write([]byte("	return int(count), err\n"))
	_, _ = dst.Write([]byte("}\n\n"))
}
