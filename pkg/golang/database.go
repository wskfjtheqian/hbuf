package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"regexp"
	"strconv"
	"strings"
)

type cache struct {
	min int
	max int
}

func getCache(name string, tags []*ast.Tag) *cache {
	val, ok := build.GetTag(tags, "cache")
	if !ok {
		return nil
	}

	c := &cache{
		min: 2 * 60 * 60,
		max: 3 * 60 * 60,
	}
	if nil != val.KV {
		for _, item := range val.KV {
			if "min" == item.Name.Name {
				val, err := strconv.Atoi(item.Value.Value[1 : len(item.Value.Value)-1])
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				c.min = val
			} else if "max" == item.Name.Name {
				val, err := strconv.Atoi(item.Value.Value[1 : len(item.Value.Value)-1])
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				c.max = val
			}
		}
	}
	return c
}

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
	if 0 < len(dbs[0].Table) && "true" != strings.ToLower(dbs[0].Table) {
		table := b.build.GetDataType(b.getFile(typ.Name), dbs[0].Table)
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

	if 0 == len(dbs[0].Table) {
		return nil
	}
	if "true" == strings.ToLower(fDbs[0].Table) {
		b.printScanData(dst, typ, dbs[0], fields, key)
	}

	if fDbs[0].List {
		b.printListData(dst, typ, dbs[0], fields, fType, fFields, c)
	}

	if 0 < len(fDbs[0].Map) {
		b.printMapData(dst, typ, dbs[0], fields, fType, fFields, fDbs[0].Map, c)
	}

	if fDbs[0].Count {
		b.printCountData(dst, typ, dbs[0], fields, fType, fFields, c)
	}

	if fDbs[0].Del {
		if typ == fType {
			key.Dbs[0].Where = "AND id = ?"
			fFields = []*build.DBField{key}
		}
		b.printDeleteData(dst, typ, dbs[0], fields, fType, fFields, nil != c)
	}

	if fDbs[0].Remove {
		if typ == fType {
			key.Dbs[0].Where = "AND id = ?"
			fFields = []*build.DBField{key}
		}
		b.printRemoveData(dst, typ, dbs[0], fields, fType, fFields, nil != c)
	}

	if fDbs[0].Insert {
		b.printInsertData(dst, typ, dbs[0], fields, key, nil != c)
	}

	if fDbs[0].Inserts {
		b.printInsertListData(dst, typ, dbs[0], fields, key, nil != c)
	}

	update := strings.ToLower(fDbs[0].Update)
	if "all" == update || "is" == update {
		if typ == fType {
			key.Dbs[0].Where = "AND id = ?"
			fFields = []*build.DBField{key}
		}
		b.printUpdateData(dst, typ, update, dbs[0], fFields, fType, c)
	}

	set := strings.ToLower(fDbs[0].Set)
	if "all" == set || "is" == set {
		if typ == fType {
			key.Dbs[0].Where = "AND id = ?"
			fFields = []*build.DBField{key}
		}
		b.printSetData(dst, typ, set, dbs[0], fFields, fType, c)
	}

	if fDbs[0].Get {
		if typ == fType {
			key.Dbs[0].Where = "AND id = ?"
			fFields = []*build.DBField{key}
		}
		b.printGetData(dst, typ, dbs[0], fields, fType, fFields, c)
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
	dst.Code("\treturn `" + item.String() + "`,\n")
	dst.Code("\t\t[]any{" + scan.String() + "}\n")
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

func (b *Builder) getParamWhere(dst *build.Writer, fields []*build.DBField, page bool, orderBy bool) *build.Writer {
	where := build.NewWriter()
	where.Packages = dst.Packages

	for _, field := range fields {
		text := field.Dbs[0].Where
		if 0 < len(text) {
			if build.IsNil(field.Field.Type) {
				where.Code("\tif nil != g." + build.StringToHumpName(field.Field.Name.Name) + " {\n")
				b.printWhere(where, text, field, "\t")
				where.Code("\t}\n")
			} else {
				b.printWhere(where, text, field, "")
			}
		}
	}

	if orderBy {
		isOrderFist := true
		for _, field := range fields {
			order := field.Dbs[0].Order
			if 0 < len(order) {
				where.Code("\tif \"ASC\" == g." + build.StringToHumpName(field.Field.Name.Name) + " || \"DESC\" == g." + build.StringToHumpName(field.Field.Name.Name))
				if build.IsNil(field.Field.Type) {
					where.Code("&& nil != g." + build.StringToHumpName(field.Field.Name.Name))
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
				where.Code("\", \"$\", g.")
				where.Code(build.StringToHumpName(field.Field.Name.Name))
				where.Code("))\n")

				where.Code("\t}\n")

			}
		}
	}

	if page {
		if limit, ok := b.getLimit(fields); ok {
			if offset, ok := b.getOffset(fields); ok {
				where.Code("\ts.T(\" LIMIT " + offset.Dbs[0].Offset + ", " + limit.Dbs[0].Limit + "\")\n")
				where.Code("\ts.P(g." + build.StringToHumpName(offset.Field.Name.Name) + ", g." + build.StringToHumpName(limit.Field.Name.Name) + ")\n")
			} else {
				where.Code("\ts.T(\" LIMIT " + limit.Dbs[0].Limit + "\")\n")
				where.Code("\ts.P(g." + build.StringToHumpName(limit.Field.Name.Name) + ")\n")
			}
		}
	}
	return where
}

func (b *Builder) printWhere(where *build.Writer, text string, field *build.DBField, s string) {
	count := strings.Count(text, "?")
	if build.IsArray(field.Field.Type) {
		temp := "\" + hbuf.ToQuestions(g." + build.StringToHumpName(field.Field.Name.Name) + ", \",\") + \""
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
			where.Code("hbuf.ToAnyList(g." + build.StringToHumpName(field.Field.Name.Name) + ")...")
		} else {
			where.Code("g.")
			where.Code(build.StringToHumpName(field.Field.Name.Name))
		}
	}
	where.Code(")\n")
}

func (b *Builder) printListData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	w := b.getParamWhere(dst, fFields, true, true)
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields)
	dst.Code("func (g " + fName + ")DbList(ctx context.Context) ([]" + dName + ", error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\tret := make([]" + dName + ", 0)\n")
	if nil != c {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
		dst.Code("\tlist, key, err := cache.DbGet(ctx, \"" + db.Name + "\", s, &ret)\n")
		dst.Code("\tif list != nil {\n")
		dst.Code("\t\treturn *list, nil\n")
		dst.Code("\t}\n")
	}

	dst.Code("\tquery, err := s.Query(ctx)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn nil, err\n")
	dst.Code("\t}\n")
	dst.Code("\tdefer query.Close()\n")

	dst.Code("\n")
	dst.Code("\tfor query.Next() {\n")
	dst.Code("\t\tvar val " + dName + "\n")
	dst.Code("\t\terr = query.Scan(" + scan.String() + ")\n")
	dst.Code("\t\tif err != nil {\n")
	dst.Code("\t\t\treturn nil, err\n")
	dst.Code("\t\t}\n")
	dst.Code("\t\tret = append(ret, val)\n")
	dst.Code("\t}\n")
	if nil != c {
		dst.Import("math/rand", "")
		dst.Import("time", "")
		dst.Code("\t_ = cache.Set(ctx, key, &ret, ")
		if 0 < c.min {
			dst.Code("time.Duration(rand.Intn(" + strconv.Itoa(c.max) + "-" + strconv.Itoa(c.min) + ")+" + strconv.Itoa(c.min) + ")*time.Second")
		} else {
			dst.Code("0")
		}
		dst.Code(")\n")
	}
	dst.Code("\treturn ret, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printMapData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, keyName string, c *cache) {
	kType, KName, ok := b.getKey(dst, fields, keyName)
	if !ok {
		return
	}
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	w := b.getParamWhere(dst, fFields, true, true)
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields)
	dst.Code("func (g " + fName + ")DbMap(ctx context.Context) (map[" + kType + "]" + dName + ", error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\tret := make(map[" + kType + "]" + dName + ")\n")
	if nil != c {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
		dst.Code("\tlist, key, err := cache.DbGet(ctx, \"" + db.Name + "\", s, &ret)\n")
		dst.Code("\tif list != nil {\n")
		dst.Code("\t\treturn *list, nil\n")
		dst.Code("\t}\n")
	}

	dst.Code("\tquery, err := s.Query(ctx)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn nil, err\n")
	dst.Code("\t}\n")
	dst.Code("\tdefer query.Close()\n")

	dst.Code("\n")
	dst.Code("\tfor query.Next() {\n")
	dst.Code("\t\tvar val " + dName + "\n")
	dst.Code("\t\terr = query.Scan(" + scan.String() + ")\n")
	dst.Code("\t\tif err != nil {\n")
	dst.Code("\t\t\treturn nil, err\n")
	dst.Code("\t\t}\n")
	dst.Code("\t\tret[val." + build.StringToHumpName(KName) + "] = val\n")
	dst.Code("\t}\n")
	if nil != c {
		dst.Import("math/rand", "")
		dst.Import("time", "")
		dst.Code("\t_ = cache.Set(ctx, key, &ret, ")
		if 0 < c.min {
			dst.Code("time.Duration(rand.Intn(" + strconv.Itoa(c.max) + "-" + strconv.Itoa(c.min) + ")+" + strconv.Itoa(c.min) + ")*time.Second")
		} else {
			dst.Code("0")
		}
		dst.Code(")\n")
	}
	dst.Code("\treturn ret, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printCountData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, fFields, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ")DbCount(ctx context.Context) (int64, error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT COUNT(1) FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\tvar count int64\n")
	if nil != c {
		dst.Code("\tc, key, err := cache.DbGet(ctx, \"" + db.Name + "\", s, &count)\n")
		dst.Code("\tif c != nil {\n")
		dst.Code("\t\treturn *c, nil\n")
		dst.Code("\t}\n")
	}

	dst.Code("\tquery, err := s.Query(ctx)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, err\n")
	dst.Code("\t}\n")
	dst.Code("\tdefer query.Close()\n")
	dst.Code("\n")
	dst.Code("\tif !query.Next() {\n")
	dst.Code("\t\treturn 0, nil\n")
	dst.Code("\t}\n")
	dst.Code("\terr = query.Scan(&count)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, err\n")
	dst.Code("\t}\n")
	if nil != c {
		dst.Import("math/rand", "")
		dst.Import("time", "")
		dst.Code("\t_ = cache.Set(ctx, key, &count, ")
		if 0 < c.min {
			dst.Code("time.Duration(rand.Intn(" + strconv.Itoa(c.max) + "-" + strconv.Itoa(c.min) + ")+" + strconv.Itoa(c.min) + ")*time.Second")
		} else {
			dst.Code("0")
		}
		dst.Code(")\n")
	}
	dst.Code("\treturn count, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printDeleteData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, fFields, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ")DbDel(ctx context.Context) (int, error) {\n")
	if isCache {
		dst.Code("\terr := cache.DbDel(ctx, \"" + db.Name + "\")\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, err\n")
		dst.Code("\t}\n")
	}

	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE " + db.Name + " SET del_time = NOW() WHERE 1 = 1\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\tresult, err := s.Exec(ctx)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, err\n")
	dst.Code("\t}\n")

	dst.Code("\tcount, err := result.RowsAffected()\n")
	dst.Code("\treturn int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printRemoveData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, fFields, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ")DbRemove(ctx context.Context) (int, error) {\n")
	if isCache {
		dst.Code("\terr := cache.DbDel(ctx, \"" + db.Name + "\")\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, err\n")
		dst.Code("\t}\n")
	}

	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"DELETE FROM " + db.Name + " WHERE 1 = 1\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\tresult, err := s.Exec(ctx)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, err\n")
	dst.Code("\t}\n")

	dst.Code("\tcount, err := result.RowsAffected()\n")
	dst.Code("\treturn int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func (g " + name + ")DbInsert(ctx context.Context) (int, error) {\n")
	if isCache {
		dst.Code("\terr := cache.DbDel(ctx, \"" + db.Name + "\")\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, err\n")
		dst.Code("\t}\n")
	}
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"INSERT INTO " + db.Name + " \")\n")
	dst.Code("\ts.T(\"SET " + key.Dbs[0].Name + " = \").V(&g." + build.StringToHumpName(key.Field.Name.Name) + ")\n")
	for _, field := range fields {
		if field == key {
			continue
		}
		fName := build.StringToHumpName(field.Field.Name.Name)
		if build.IsNil(field.Field.Type) {
			dst.Code("\tif nil != g." + fName + " {\n")
			dst.Code("\t")
		}
		dst.Code("\ts.T(\", " + field.Dbs[0].Name + " =\").V(&g." + fName + ")\n")
		if build.IsNil(field.Field.Type) {
			dst.Code("\t}\n")
		}
	}

	dst.Code("\tresult, err := s.Exec(ctx)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, err\n")
	dst.Code("\t}\n")
	dst.Code("\tcount, err := result.RowsAffected()\n")
	dst.Code("\treturn int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertListData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func (g " + name + ")DbInsertList(ctx context.Context, val []*" + name + ") (int, error) {\n")
	dst.Code("\tif nil == val || 0 == len(val) {\n")
	dst.Code("\t\treturn 0, nil\n")
	dst.Code("\t}\n")
	if isCache {
		dst.Code("\terr := cache.DbDel(ctx, \"" + db.Name + "\")\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, err\n")
		dst.Code("\t}\n")
	}
	dst.Code("\ts := db.NewSql()\n")
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
	dst.Code("\tfor i, val := range val {\n")
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

	dst.Code("\tresult, err := s.Exec(ctx)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, err\n")
	dst.Code("\t}\n")
	dst.Code("\tcount, err := result.RowsAffected()\n")
	dst.Code("\treturn int(count), err\n\n")
	dst.Code("}\n\n")
}

func (b *Builder) printUpdateData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, fields []*build.DBField, fType *ast.DataType, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	//dName := build.StringToHumpName(typ.Name.Name)

	w := b.getParamWhere(dst, fields, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ")DbUpdate(ctx context.Context) (int, error) {\n")
	if nil != c {
		dst.Code("\terr := cache.DbDel(ctx, \"" + db.Name + "\")\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, err\n")
		dst.Code("\t}\n")
	}
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE " + db.Name + " \")\n")
	i := 0
	for _, field := range fields {
		set := ""
		if "all" == key {
			set = field.Dbs[0].Name + " = ?"
		} else if "is" == key {
			if 0 < len(field.Dbs[0].Set) {
				set = field.Dbs[0].Name + " = " + field.Dbs[0].Set
			}
		}
		if 0 == len(set) {
			continue
		}
		name := build.StringToHumpName(field.Field.Name.Name)
		if build.IsNil(field.Field.Type) {
			dst.Code("\tif nil != g." + name + " {\n")
			dst.Code("\t")
		}
		if 0 == i {
			dst.Code("\ts.T(\"SET " + set + " \").P(&g." + name + ")\n")
		} else {
			dst.Code("\ts.T(\", " + set + " \").P(&g." + name + ")\n")
		}
		i++
		if build.IsNil(field.Field.Type) {
			dst.Code("\t}\n")
		}
	}

	dst.Code("\ts.T(\"WHERE 1 == 1 \")\n")
	dst.Code(w.String())
	dst.Code("\n")

	dst.Code("\tresult, err := s.Exec(ctx)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, err\n")
	dst.Code("\t}\n")
	dst.Code("\tcount, err := result.RowsAffected()\n")
	dst.Code("\treturn int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printSetData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, fields []*build.DBField, fType *ast.DataType, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	//dName := build.StringToHumpName(typ.Name.Name)

	w := b.getParamWhere(dst, fields, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ")DbSet(ctx context.Context) (int, error) {\n")
	if nil != c {
		dst.Code("\terr := cache.DbDel(ctx, \"" + db.Name + "\")\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, err\n")
		dst.Code("\t}\n")
	}
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE " + db.Name + " \")\n")
	i := 0
	for _, field := range fields {
		set := ""
		if "all" == key {
			set = field.Dbs[0].Name + " = ?"
		} else if "is" == key {
			if 0 < len(field.Dbs[0].Set) {
				set = field.Dbs[0].Name + " = " + field.Dbs[0].Set
			}
		}
		if 0 == len(set) {
			continue
		}
		name := build.StringToHumpName(field.Field.Name.Name)
		if 0 == i {
			dst.Code("\ts.T(\"SET " + set + " \").P(&g." + name + ")\n")
		} else {
			dst.Code("\ts.T(\", " + set + " \").P(&g." + name + ")\n")
		}
		i++
	}

	dst.Code("\ts.T(\"WHERE 1 == 1 \")\n")
	dst.Code(w.String())
	dst.Code("\n")

	dst.Code("\tresult, err := s.Exec(ctx)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn 0, err\n")
	dst.Code("\t}\n")
	dst.Code("\tcount, err := result.RowsAffected()\n")
	dst.Code("\treturn int(count), err\n")
	dst.Code("}\n\n")
}

func (b *Builder) printGetData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	w := b.getParamWhere(dst, fFields, false, false)
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields)

	dst.Code("func (g " + fName + ")DbGet(ctx context.Context) (*" + dName + ", error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM " + db.Name + " WHERE del_time IS NULL\")\n")
	dst.Code(w.GetCode().String())
	dst.Code("\ts.T(\" LIMIT 1\")\n")
	dst.Code("\tvar val " + dName + "\n")

	if nil != c {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
		dst.Code("\tcv, key, err := cache.DbGet(ctx, \"" + db.Name + "\", s, &val)\n")
		dst.Code("\tif cv != nil {\n")
		dst.Code("\t\treturn cv, nil\n")
		dst.Code("\t}\n")
	}
	dst.Code("\tquery, err := s.Query(ctx)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn nil, err\n")
	dst.Code("\t}\n")
	dst.Code("\tdefer query.Close()\n")

	dst.Code("\tif !query.Next() {\n")
	dst.Code("\t\treturn nil, nil\n")
	dst.Code("\t}\n")

	dst.Code("\terr = query.Scan(" + scan.String() + ")\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn nil, err\n")
	dst.Code("\t}\n")
	if nil != c {
		dst.Import("math/rand", "")
		dst.Import("time", "")
		dst.Code("\t_ = cache.Set(ctx, key, &val, ")
		if 0 < c.min {
			dst.Code("time.Duration(rand.Intn(" + strconv.Itoa(c.max) + "-" + strconv.Itoa(c.min) + ")+" + strconv.Itoa(c.min) + ")*time.Second")
		} else {
			dst.Code("0")
		}
		dst.Code(")\n")
	}
	dst.Code("\treturn &val, nil\n")
	dst.Code("}\n\n")

}
