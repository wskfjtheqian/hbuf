package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
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
	_, _ = dst.Write([]byte("}\n"))

	return nil
}

func printInsertData(dst io.Writer, typ *ast.DataType) {

}

func printUpdateData(dst io.Writer, typ *ast.DataType) {

}

func printDeleteData(dst io.Writer, typ *ast.DataType) {

}
