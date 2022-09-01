package golang

import "C"
import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"regexp"
	"strings"
)

func (b *Builder) printDatabaseCode(dst *build.Writer, typ *ast.DataType) error {
	dst.Import("context", "")
	dst.Import("errors", "")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/db", "")

	dbs, fields, key, err := b.getDBField(typ)
	if 0 == len(dbs) || nil != err {
		return nil
	}

	c := getCache(typ.Name.Name, typ.Tags)

	fDbs := dbs
	fFields := fields
	fType := typ
	if !dbs[0].Table && 0 < len(dbs[0].Name) {
		table := b.build.GetDataType(b.getFile(typ.Name), dbs[0].Name)
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

	if !dbs[0].Table {
		return nil
	}
	if fDbs[0].Table {
		b.printScanData(dst, typ, dbs[0], fields, key)
	}

	if fDbs[0].Find {
		b.printFindData(dst, typ, dbs[0], fields, fType, fFields, nil != c)
	}

	if fDbs[0].Count {
		b.printCountData(dst, typ, dbs[0], fields, fType, fFields, nil != c)
	}

	if fDbs[0].Del {
		b.printDeleteData(dst, typ, dbs[0], fields, key, nil != c)
	}

	if fDbs[0].Remove {
		b.printRemoveData(dst, typ, dbs[0], fields, key, nil != c)
	}

	if fDbs[0].Insert {
		b.printInsertData(dst, typ, dbs[0], fields, key, nil != c)
	}

	if fDbs[0].Inserts {
		b.printInsertListData(dst, typ, dbs[0], fields, key, nil != c)
	}

	if fDbs[0].Update {
		b.printUpdateData(dst, typ, dbs[0], fields, key, nil != c)
	}

	if fDbs[0].Set {
		b.printSetData(dst, typ, dbs[0], fields, key, nil != c)
	}

	if fDbs[0].Get {
		if typ == fType {
			key.Dbs[0].Where = "AND id = ?"
			fFields = []*build.DBField{key}
		}
		b.printGetData(dst, typ, dbs[0], fields, fType, fFields, nil != c)
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

func (b *Builder) printScanData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField) {
	name := build.StringToHumpName(typ.Name.Name)
	item, scan, _ := b.getItemAndValue(fields)
	dst.Code("func DbScan" + name + "(val *" + name + ") (string, []any) {\n")
	dst.Code("	return `" + item.String() + "`,\n")
	dst.Code("		[]any{" + scan.String() + "}\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) getItemAndValue(fields []*build.DBField) (strings.Builder, strings.Builder, strings.Builder) {
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
		item.WriteString(build.StringToUnderlineName(field.Dbs[0].Name))
		scan.WriteString("&val." + build.StringToHumpName(field.Field.Name.Name))
		ques.WriteString("?")
	}
	return item, scan, ques

}

func (b *Builder) getParamWhere(dst *build.Writer, fields []*build.DBField, page bool) (*build.Writer, *build.Writer) {
	param := build.NewWriter()
	param.Packages = dst.Packages
	where := build.NewWriter()
	where.Packages = dst.Packages

	isFist := true
	for _, field := range fields {
		text := field.Dbs[0].Where
		if 0 < len(text) {
			if !isFist {
				param.Code(", ")
			}
			isFist = false
			param.Code(build.StringToFirstLower(field.Field.Name.Name))
			param.Code(" ")
			b.printType(param, field.Field.Type, false)
			if build.IsNil(field.Field.Type) {
				where.Code("\tif nil != " + build.StringToFirstLower(field.Field.Name.Name) + " {\n")
				b.printWhere(where, text, field, "\t")
				where.Code("\t}\n")
			} else {
				b.printWhere(where, text, field, "")
			}
		}
	}
	isOrderFist := true
	for _, field := range fields {
		order := field.Dbs[0].Order
		if 0 < len(order) {
			if !isFist {
				param.Code(", ")
			}
			isFist = false
			param.Code(build.StringToFirstLower(field.Field.Name.Name))
			param.Code(" ")
			b.printType(param, field.Field.Type, false)
			where.Code("\tif \"AES\" == " + build.StringToFirstLower(field.Field.Name.Name) + " || \"DESC\" == " + build.StringToFirstLower(field.Field.Name.Name))
			if build.IsNil(field.Field.Type) {
				where.Code("&& nil != " + build.StringToFirstLower(field.Field.Name.Name))
			}
			where.Code(" {\n")

			if isOrderFist {
				where.Code("\t\ts.T(\" ORDER BY \" + ")
			} else {
				where.Code("\t\ts.T(\", \" + ")
			}
			isOrderFist = true
			where.Import("strings", "")
			where.Code("strings.ReplaceAll(\"")
			where.Code(order)
			where.Code("\", \"$\",")
			where.Code(build.StringToFirstLower(field.Field.Name.Name))
			where.Code("))\n")

			where.Code("\t}\n")

		}
	}

	if page {
		if limit, ok := b.getLimit(fields); ok {
			if !isFist {
				param.Code(", ")
			}
			isFist = false
			param.Code(build.StringToFirstLower(limit.Field.Name.Name))
			param.Code(" ")
			b.printType(param, limit.Field.Type, false)
			if offset, ok := b.getOffset(fields); ok {
				if !isFist {
					param.Code(", ")
				}
				isFist = false
				param.Code(build.StringToFirstLower(offset.Field.Name.Name))
				param.Code(" ")
				b.printType(param, offset.Field.Type, false)
				where.Code("\ts.T(\" LIMIT " + offset.Dbs[0].Offset + ", " + limit.Dbs[0].Limit + "\")\n")
				where.Code("\ts.P(" + build.StringToFirstLower(offset.Field.Name.Name) + ", " + build.StringToFirstLower(limit.Field.Name.Name) + " )\n")
			} else {
				where.Code("\ts.T(\" LIMIT " + limit.Dbs[0].Limit + "\")\n")
				where.Code("\ts.P(param," + build.StringToFirstLower(limit.Field.Name.Name) + " )\n")
			}
		}
	}
	return param, where
}

func (b *Builder) printWhere(where *build.Writer, text string, field *build.DBField, s string) {
	count := strings.Count(text, "?")
	if build.IsArray(field.Field.Type) {
		temp := "\" + hbuf.ToQuestions(ids, \",\") + \""
		rex := regexp.MustCompile(`\?`)
		match := rex.FindAllStringSubmatchIndex(text, -1)
		if nil != match {
			var index = 0
			buf := strings.Builder{}
			for _, item := range match {
				buf.WriteString(text[index:item[0]])
				index = item[0] + 1
				buf.WriteString(temp)
			}
			if index < len(text) {
				buf.WriteString(text[index:])
			}
			text = buf.String()
		}
	}
	where.Code(s + "\ts.T(\" " + text + "\")\n")
	where.Code(s + "\ts.P(")

	for i := 0; i < count; i++ {
		if 0 != i {
			where.Code(", ")
		}
		if build.IsArray(field.Field.Type) {
			where.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
			where.Code("hbuf.ToAnyList(" + build.StringToFirstLower(field.Field.Name.Name) + ")...")
		} else {
			where.Code(build.StringToFirstLower(field.Field.Name.Name))
		}
	}
	where.Code(")\n")
}

func (b *Builder) printFindData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	p, w := b.getParamWhere(dst, fFields, true)
	dst.AddImports(p.GetImports())
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields)
	dst.Code("func DbFind" + fName + "(ctx context.Context, " + p.GetCode().String() + ") ([]" + dName + ", error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	dst.Code(w.GetCode().String())
	if isCache {
		dst.Code("\tlist, err := cacheGetFind" + fName + "(ctx, s)\n")
		dst.Code("\tif list != nil {\n")
		dst.Code("\t	return *list, nil\n")
		dst.Code("\t}\n")
	}
	dst.Code("\ts.Debug(ctx)\n")
	dst.Code("	query, err := db.GET(ctx).Query(s.ToSql(), s.ToParam()...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return nil, errors.New(err.Error() +\":\"+ s.ToText())\n")
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
	if isCache {
		dst.Code("	_ = cacheSetFind" + fName + "(ctx, &ret, s)\n")
	}
	dst.Code("	return ret, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printCountData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)

	p, w := b.getParamWhere(dst, fFields, false)
	dst.AddImports(p.GetImports())
	dst.AddImports(w.GetImports())

	dst.Code("func DbCount" + fName + "(ctx context.Context, " + p.GetCode().String() + ") (int64, error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT COUNT(1) FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	if isCache {
		dst.Code("\tc, err := cacheGetCount" + fName + "(ctx, s)\n")
		dst.Code("\tif c != nil {\n")
		dst.Code("\t	return *c, nil\n")
		dst.Code("\t}\n")
	}
	dst.Code("\ts.Debug(ctx)\n")
	dst.Code("	query, err := db.GET(ctx).Query(s.ToSql(), s.ToParam()...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, errors.New(err.Error() +\":\"+ s.ToText())\n")
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
	if isCache {
		dst.Code("	_ = cacheSetCount" + fName + "(ctx, &count, s)\n")
	}
	dst.Code("	return count, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printDeleteData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbDel" + name + "(ctx context.Context, " + build.StringToFirstLower(key.Field.Name.Name) + " ")
	b.printType(dst, key.Field.Type, false)
	dst.Code(") (int, error) {\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("\ts:=db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE " + db.Name + " SET del_time = NOW() \")\n")
	dst.Code("\ts.T(\"WHERE " + key.Dbs[0].Name + " = \").V(&" + build.StringToFirstLower(key.Field.Name.Name) + ")\n")
	dst.Code("\ts.Debug(ctx)\n")
	dst.Code("	result, err := db.GET(ctx).Exec(s.ToSql(), s.ToParam()...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, errors.New(err.Error() +\":\"+ s.ToText())\n")
	dst.Code("	}\n")

	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printRemoveData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbRemove" + name + "(ctx context.Context, " + build.StringToFirstLower(key.Field.Name.Name) + " ")
	b.printType(dst, key.Field.Type, false)
	dst.Code(") (int, error) {\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("\ts:=db.NewSql()\n")
	dst.Code("\ts.T(\"DELETE FROM " + db.Name + " \")\n")
	dst.Code("\ts.T(\"WHERE " + key.Dbs[0].Name + " = \").V(&" + build.StringToFirstLower(key.Field.Name.Name) + ")\n")
	dst.Code("\ts.Debug(ctx)\n")
	dst.Code("	result, err := db.GET(ctx).Exec(s.ToSql(), s.ToParam()...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, errors.New(err.Error() +\":\"+ s.ToText())\n")
	dst.Code("	}\n")

	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func DbInsert" + name + "(ctx context.Context, val *" + name + ") (int, error) {\n")
	dst.Code("	if nil == val {\n")
	dst.Code("		return 0, nil\n")
	dst.Code("	}\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("\ts:=db.NewSql()\n")
	dst.Code("\ts.T(\"INSERT INTO " + db.Name + " \")\n")
	dst.Code("\ts.T(\"SET " + key.Dbs[0].Name + " = \").V(&val." + build.StringToHumpName(key.Field.Name.Name) + ")\n")
	for _, field := range fields {
		if field == key {
			continue
		}
		fName := build.StringToHumpName(field.Field.Name.Name)
		if build.IsNil(field.Field.Type) {
			dst.Code("\tif nil != val." + fName + " {\n")
			dst.Code("\t")
		}
		dst.Code("\ts.T(\", " + field.Dbs[0].Name + " =\").V(&val." + fName + ")\n")
		if build.IsNil(field.Field.Type) {
			dst.Code("\t}\n")
		}
	}

	dst.Code("\ts.Debug(ctx)\n")
	dst.Code("\tresult, err := db.GET(ctx).Exec(s.ToSql(), s.ToParam()...)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, errors.New(err.Error() +\":\"+ s.ToText())\n")
	dst.Code("\t}\n")
	dst.Code("\tcount, err := result.RowsAffected()\n")
	dst.Code("\treturn int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertListData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func DbInsertList" + name + "(ctx context.Context, val []*" + name + ") (int, error) {\n")
	dst.Code("	if nil == val || 0 == len(val) {\n")
	dst.Code("		return 0, nil\n")
	dst.Code("	}\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("\ts:=db.NewSql()\n")
	dst.Code("\ts.T(\"INSERT INTO " + db.Name + " (")
	isFist := true
	for _, field := range fields {
		if !isFist {
			dst.Code(", ")
		}
		isFist = false
		dst.Code(field.Dbs[0].Name)
	}
	dst.Code(") VALUES\")\n")
	dst.Code("	for i, val := range val {\n")
	dst.Code("\t\tif 0 != i {\n")
	dst.Code("\t\t\ts.T(\",\")\n")
	dst.Code("\t\t}\n")
	dst.Code("\t\ts.T(\"(\").L(\",\", ")
	isFist = true
	for _, field := range fields {
		if !isFist {
			dst.Code(", ")
		}
		isFist = false
		dst.Code("&val." + build.StringToHumpName(field.Field.Name.Name))
	}
	dst.Code(").T(\")\")\n")
	dst.Code("\t}\n")
	dst.Code("\ts.Debug(ctx)\n")
	dst.Code("\tresult, err := db.GET(ctx).Exec(s.ToSql(), s.ToParam()...)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, errors.New(err.Error() +\":\"+ s.ToText())\n")
	dst.Code("\t}\n")
	dst.Code("\tcount, err := result.RowsAffected()\n")
	dst.Code("\treturn int(count), err\n\n")
	dst.Code("}\n\n")
}

func (b *Builder) printUpdateData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func DbUpdate" + name + "(ctx context.Context, val *" + name + ") (int, error) {\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("\ts:=db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE " + db.Name + " \")\n")
	dst.Code("\ts.T(\"SET " + key.Dbs[0].Name + " = " + key.Dbs[0].Name + "\")\n")
	for _, field := range fields {
		if field == key {
			continue
		}
		fName := build.StringToHumpName(field.Field.Name.Name)
		if build.IsNil(field.Field.Type) {
			dst.Code("\tif nil != val." + fName + " {\n")
			dst.Code("\t")
		}
		dst.Code("\ts.T(\", " + field.Dbs[0].Name + " =\").V(&val." + fName + ")\n")
		if build.IsNil(field.Field.Type) {
			dst.Code("\t}\n")
		}
	}

	dst.Code("\ts.T(\"WHERE del_time IS  NULL AND " + key.Dbs[0].Name + "\").V(&val." + build.StringToHumpName(key.Field.Name.Name) + ")\n")
	dst.Code("\ts.Debug(ctx)\n")
	dst.Code("\tresult, err := db.GET(ctx).Exec(s.ToSql(), s.ToParam()...)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, errors.New(err.Error() +\":\"+ s.ToText())\n")
	dst.Code("\t}\n")
	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printSetData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbSet" + name + "(ctx context.Context, val *" + name + ") (int, error) {\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("\ts:=db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE " + db.Name + " \")\n")
	isFist := true
	for _, field := range fields {
		if field == key {
			continue
		}
		if isFist {
			dst.Code("\ts.T(\"SET ")
		} else {
			dst.Code("\ts.T(\", ")
		}
		isFist = false
		dst.Code(field.Dbs[0].Name + " =\").V(&val." + build.StringToHumpName(field.Field.Name.Name) + ")\n")
	}

	dst.Code("\ts.T(\"WHERE del_time IS  NULL AND " + key.Dbs[0].Name + " = \").V(&val." + build.StringToHumpName(key.Field.Name.Name) + ")\n")
	dst.Code("\ts.Debug(ctx)\n")
	dst.Code("\tresult, err := db.GET(ctx).Exec(s.ToSql(), s.ToParam()...)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, errors.New(err.Error() +\":\"+ s.ToText())\n")
	dst.Code("\t}\n")
	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printGetData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	p, w := b.getParamWhere(dst, fFields, false)
	dst.AddImports(p.GetImports())
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields)

	dst.Code("func DbGet" + fName + "(ctx context.Context, " + p.GetCode().String() + ") (*" + dName + ", error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM " + db.Name + " WHERE del_time IS NULL\")\n")
	dst.Code("\ts.T(\" LIMIT 1\")\n")
	if isCache {
		dst.Code("\tcv, err := cacheGetGet" + fName + "(ctx, s)\n")
		dst.Code("\tif cv != nil {\n")
		dst.Code("\t	return cv, nil\n")
		dst.Code("\t}\n")
	}
	dst.Code("	query, err := db.GET(ctx).Query(s.ToSql(), s.ToParam()...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return nil, err\n")
	dst.Code("	}\n")
	dst.Code("	defer query.Close()\n")

	dst.Code("	if !query.Next() {\n")
	dst.Code("		return nil, nil\n")
	dst.Code("	}\n")

	dst.Code("	var val " + dName + "\n")
	dst.Code("\ts.Debug(ctx)\n")
	dst.Code("	err = query.Scan(" + scan.String() + ")\n")
	dst.Code("	if err != nil {\n")
	dst.Code("\t\treturn nil, errors.New(err.Error() +\":\"+ s.ToText())\n")
	dst.Code("	}\n")
	if isCache {
		dst.Code("	_ = cacheSetGet" + fName + "(ctx, &val, s)\n")
	}
	dst.Code("	return &val, nil\n")
	dst.Code("}\n\n")
}
