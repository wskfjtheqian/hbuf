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
	dst.Code("func DbScan" + name + "(val *" + name + ") (string, []interface{}) {\n")
	dst.Code("	return `" + item.String() + "`,\n")
	dst.Code("		[]interface{}{" + scan.String() + "}\n")
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

func (b *Builder) getParamWhere(dst *build.Writer, fields []*build.DBField, page bool) (*build.Writer, *build.Writer, *build.Writer, *build.Writer) {
	param := build.NewWriter()
	param.Packages = dst.Packages
	where := build.NewWriter()
	where.Packages = dst.Packages
	caches := build.NewWriter()
	caches.Packages = dst.Packages
	texts := build.NewWriter()
	texts.Import("fmt", "")
	texts.Packages = dst.Packages

	isFist := true
	for _, field := range fields {
		text := field.Dbs[0].Where
		if 0 < len(text) {
			if !isFist {
				caches.Code(", ")
				param.Code(", ")

			}
			isFist = false
			caches.Code(build.StringToFirstLower(field.Field.Name.Name))
			param.Code(build.StringToFirstLower(field.Field.Name.Name))
			param.Code(" ")
			b.printType(param, field.Field.Type, false)
			texts.Code("\tkey.WriteString(\"&" + build.StringToUnderlineName(field.Field.Name.Name) + "=\")\n")
			texts.Code("\tkey.WriteString(fmt.Sprint(" + build.StringToFirstLower(field.Field.Name.Name) + "))\n")

			if build.IsNil(field.Field.Type) {
				where.Code("\tif nil != " + build.StringToFirstLower(field.Field.Name.Name) + " {\n")
				b.printWhere(where, text, field)
				where.Code("\t}\n")
			} else {
				b.printWhere(where, text, field)
			}
		}
	}
	if page {
		if limit, ok := b.getLimit(fields); ok {
			if !isFist {
				param.Code(", ")
				caches.Code(", ")
			}
			isFist = false
			caches.Code(build.StringToFirstLower(limit.Field.Name.Name))
			param.Code(build.StringToFirstLower(limit.Field.Name.Name))
			param.Code(" ")
			b.printType(param, limit.Field.Type, false)
			texts.Code("\tkey.WriteString(\"&" + build.StringToUnderlineName(limit.Field.Name.Name) + "=\")\n")
			texts.Code("\tkey.WriteString(fmt.Sprint(" + build.StringToFirstLower(limit.Field.Name.Name) + "))\n")

			if offset, ok := b.getOffset(fields); ok {
				if !isFist {
					caches.Code(", ")
					param.Code(", ")
				}
				isFist = false
				caches.Code(build.StringToFirstLower(offset.Field.Name.Name))
				param.Code(build.StringToFirstLower(offset.Field.Name.Name))
				param.Code(" ")
				b.printType(param, offset.Field.Type, false)
				texts.Code("\tkey.WriteString(\"&" + build.StringToUnderlineName(offset.Field.Name.Name) + "=\")\n")
				texts.Code("\tkey.WriteString(fmt.Sprint(" + build.StringToFirstLower(offset.Field.Name.Name) + "))\n")

				where.Code("\tsql.WriteString(\" LIMIT " + offset.Dbs[0].Offset + ", " + limit.Dbs[0].Limit + "\")\n")
				where.Code("\tparam = append(param," + build.StringToFirstLower(offset.Field.Name.Name) + ", " + build.StringToFirstLower(limit.Field.Name.Name) + " )\n")
			} else {
				where.Code("\tsql.WriteString(\" LIMIT " + limit.Dbs[0].Limit + "\")\n")
				where.Code("\tparam = append(param," + build.StringToFirstLower(limit.Field.Name.Name) + " )\n")
			}
		}
	}
	return param, where, caches, texts
}

func (b *Builder) printWhere(where *build.Writer, text string, field *build.DBField) {
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
	where.Code("\tsql.WriteString(\" " + text + "\")\n")
	where.Code("\tparam = append(param")

	for i := 0; i < count; i++ {
		if build.IsArray(field.Field.Type) {
			where.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
			where.Code(", hbuf.ToAnyList(" + build.StringToFirstLower(field.Field.Name.Name) + ")...")
		} else {
			where.Code(", " + build.StringToFirstLower(field.Field.Name.Name))
		}
	}
	where.Code(")\n")
}

func (b *Builder) printGetData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	p, w, c, _ := b.getParamWhere(dst, fFields, false)
	dst.AddImports(p.GetImports())
	dst.AddImports(w.GetImports())
	dst.AddImports(c.GetImports())

	item, scan, _ := b.getItemAndValue(fields)

	dst.Code("func DbGet" + fName + "(ctx context.Context, " + p.GetCode().String() + ") (*" + dName + ", error) {\n")
	if isCache {
		dst.Code("\tcv, err := CacheGetGet" + fName + "(ctx," + c.GetCode().String() + ")\n")
		dst.Code("\tif cv != nil {\n")
		dst.Code("\t	return cv, nil\n")
		dst.Code("\t}\n")
	}
	dst.Code("\tsql := strings.Builder{}\n")
	dst.Code("\tvar param []interface{}\n")
	dst.Code("\tsql.WriteString(\"SELECT " + item.String() + " FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	w.Code("\tsql.WriteString(\" LIMIT 1\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("	query, err := db.GET(ctx).Query(sql.String(), param...)\n")
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
	if isCache {
		dst.Code("	_ = CacheSetGet" + fName + "(ctx, &val, " + c.GetCode().String() + ")\n")
	}
	dst.Code("	return &val, nil\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	dst.Import("strings", "")
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func DbInsert" + name + "(ctx context.Context, val *" + name + ") (int, error) {\n")
	dst.Code("	if nil == val {\n")
	dst.Code("		return 0, nil\n")
	dst.Code("	}\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("	value := strings.Builder{}\n")
	dst.Code("	ques := strings.Builder{}\n")
	dst.Code("	var param []interface{}\n\n")
	for _, field := range fields {
		fName := build.StringToHumpName(field.Field.Name.Name)
		if build.IsNil(field.Field.Type) {
			dst.Code("	if nil != val." + fName + " {\n")
			dst.Code("		value.WriteString(\"," + field.Dbs[0].Name + "\")\n")
			dst.Code("		ques.WriteString(\",?\")\n")
			dst.Code("		param = append(param, &val." + fName + ")\n\n")
			dst.Code("	}\n")
		} else {
			dst.Code("	value.WriteString(\"," + field.Dbs[0].Name + "\")\n")
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
	dst.Code("	result, err := db.GET(ctx).Exec(\"INSERT INTO " + db.Name + "(\"+valText+\") VALUES(\"+quesText+\")\", param...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")
	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertListData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	dst.Import("strings", "")
	name := build.StringToHumpName(typ.Name.Name)
	item, scan, ques := b.getItemAndValue(fields)
	dst.Code("func DbInsertList" + name + "(ctx context.Context, val []*" + name + ") (int, error) {\n")
	dst.Code("	if nil == val || 0 == len(val) {\n")
	dst.Code("		return 0, nil\n")
	dst.Code("	}\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("	value := strings.Builder{}\n")
	dst.Code("	var param []interface{}\n")
	dst.Code("	value.Write([]byte(`INSERT INTO " + db.Name + "(")
	dst.Code(item.String())
	dst.Code(") VALUES(`))\n")
	dst.Code("	for i, val := range val {\n")
	dst.Code("		if 0 != i {\n")
	dst.Code("			value.WriteString(\",\")\n")
	dst.Code("		}\n")
	dst.Code("		value.WriteString(\"(" + ques.String() + ")\")\n")
	dst.Code("		param = append(param, " + scan.String() + ")\n")
	dst.Code("	}\n")
	dst.Code("	result, err := db.GET(ctx).Exec(value.String(), param...)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")
	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n\n")
	dst.Code("}\n\n")
}

func (b *Builder) printUpdateData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func DbUpdate" + name + "(ctx context.Context, val *" + name + ") (int, error) {\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("	value := strings.Builder{}\n")
	dst.Code("	var param []interface{}\n\n")
	for _, field := range fields {
		if field == key {
			continue
		}
		fName := build.StringToHumpName(field.Field.Name.Name)
		if build.IsNil(field.Field.Type) {
			dst.Code("	if nil != val." + fName + " {\n")
			dst.Code("		value.WriteString(\"," + field.Dbs[0].Name + "= ? \")\n")
			dst.Code("		param = append(param, &val." + fName + ")\n")
			dst.Code("	}\n\n")
		} else {
			dst.Code("	value.WriteString(\"," + field.Dbs[0].Name + " = ?\")\n")
			dst.Code("	param = append(param, &val." + fName + ")\n\n")
		}
	}

	dst.Code("	valText := value.String()\n")
	dst.Code("	if 0 == len(valText) {\n")
	dst.Code("		return 0, nil\n")
	dst.Code("	}\n")
	dst.Code("	valText = valText[1:]\n")
	dst.Code("	param = append(param, &val." + build.StringToHumpName(key.Field.Name.Name) + ")\n")
	dst.Code("	result, err := db.GET(ctx).Exec(\"UPDATE " + db.Name + " SET \"+valText+\" WHERE del_time IS  NULL AND id = ?\", param...)\n")

	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")

	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printRemoveData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbDel" + name + "(ctx context.Context, " + build.StringToFirstLower(key.Field.Name.Name) + " ")
	b.printType(dst, key.Field.Type, false)
	dst.Code(") (int, error) {\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("	result, err := db.GET(ctx).Exec(`DELETE FROM  " + db.Name + " WHERE " + key.Dbs[0].Name + " = ?`, " + build.StringToFirstLower(key.Field.Name.Name) + ")\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n\n")

	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printDeleteData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbDel" + name + "(ctx context.Context, " + build.StringToFirstLower(key.Field.Name.Name) + " ")
	b.printType(dst, key.Field.Type, false)
	dst.Code(") (int, error) {\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("	result, err := db.GET(ctx).Exec(`UPDATE " + db.Name + " SET del_time = NOW() WHERE " + key.Dbs[0].Name + " = ?`, " + build.StringToFirstLower(key.Field.Name.Name) + ")\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n\n")

	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printFindData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	p, w, c, _ := b.getParamWhere(dst, fFields, true)
	dst.AddImports(p.GetImports())
	dst.AddImports(w.GetImports())
	dst.AddImports(c.GetImports())

	item, scan, _ := b.getItemAndValue(fields)

	dst.Code("func DbFind" + fName + "(ctx context.Context, " + p.GetCode().String() + ") ([]" + dName + ", error) {\n")
	if isCache {
		dst.Code("\tlist, err := CacheGetFind" + fName + "(ctx," + c.GetCode().String() + ")\n")
		dst.Code("\tif list != nil {\n")
		dst.Code("\t	return *list, nil\n")
		dst.Code("\t}\n")
	}
	dst.Code("\tsql := strings.Builder{}\n")
	dst.Code("\tvar param []interface{}\n")
	dst.Code("\tsql.WriteString(\"SELECT " + item.String() + " FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("	query, err := db.GET(ctx).Query(sql.String(), param...)\n")
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
	if isCache {
		dst.Code("	_ = CacheSetFind" + fName + "(ctx, &ret, " + c.GetCode().String() + ")\n")
	}
	dst.Code("	return ret, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printCountData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)

	p, w, c, _ := b.getParamWhere(dst, fFields, false)
	dst.AddImports(p.GetImports())
	dst.AddImports(w.GetImports())
	dst.AddImports(c.GetImports())

	dst.Code("func DbCount" + fName + "(ctx context.Context, " + p.GetCode().String() + ") (int64, error) {\n")
	if isCache {
		dst.Code("\tc, err := CacheGetCount" + fName + "(ctx," + c.GetCode().String() + ")\n")
		dst.Code("\tif c != nil {\n")
		dst.Code("\t	return *c, nil\n")
		dst.Code("\t}\n")
	}

	dst.Code("\tsql := strings.Builder{}\n")
	dst.Code("\tvar param []interface{}\n")
	dst.Code("\tsql.WriteString(\"SELECT COUNT(1) FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("	query, err := db.GET(ctx).Query(sql.String(), param...)\n")
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
	if isCache {
		dst.Code("	_ = CacheSetCount" + fName + "(ctx, &count, " + c.GetCode().String() + ")\n")
	}
	dst.Code("	return count, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printSetData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)

	dst.Code("func DbSet" + name + "(ctx context.Context, val *" + name + ") (int, error) {\n")
	if isCache {
		dst.Code("\t_ = CacheDel" + name + "(ctx)\n")
	}
	dst.Code("	result, err := db.GET(ctx).Exec(`UPDATE  " + db.Name + " SET ")
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
		dst.Code(field.Dbs[0].Name + " = ?")
		_, _ = values.Write([]byte(", &val." + build.StringToHumpName(field.Field.Name.Name)))
	}

	dst.Code(" WHERE del_time IS  NULL AND " + key.Dbs[0].Name + " = ?`")
	dst.Code(values.String())
	dst.Code(", &val." + build.StringToHumpName(key.Field.Name.Name) + " )\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return 0, err\n")
	dst.Code("	}\n")
	dst.Code("	count, err := result.RowsAffected()\n")
	dst.Code("	return int(count), err\n")
	dst.Code("}\n\n")
}
