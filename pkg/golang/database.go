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
				val, err := strconv.Atoi(item.Values[0].Value[1 : len(item.Values[0].Value)-1])
				if err != nil {
					//TODO 添加错误处理
					return nil
				}
				c.min = val
			} else if "max" == item.Name.Name {
				val, err := strconv.Atoi(item.Values[0].Value[1 : len(item.Values[0].Value)-1])
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

	dbs, wFields, key, err := b.getDBField(typ)
	if 0 == len(dbs) || nil != err {
		return nil
	}

	c := getCache(typ.Name.Name, typ.Tags)

	fDbs := dbs
	fields := wFields
	fType := typ
	if 0 < len(dbs[0].Table) {
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

	if 0 == len(fDbs[0].Table) {
		b.printScanData(dst, typ, dbs[0], wFields, key)
	}

	val := strings.ToLower(fDbs[0].List)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		b.printListData(dst, typ, val, dbs[0], w, f, fType, c)
	}

	if 0 < len(fDbs[0].Map) {
		ks := strings.Split(fDbs[0].Map, ":")
		if 1 < len(ks) {
			val = strings.ToLower(ks[1])
			if ("self" == val || "parent" == val) && 0 < len(ks[0]) {
				w := wFields
				f := fields
				if "self" == val {
					f = wFields
				}
				b.printMapData(dst, val, typ, dbs[0], w, f, fType, ks[0], c)
			}
		}
	}

	if fDbs[0].Count {
		b.printCountData(dst, typ, dbs[0], wFields, fType, fields, c)
	}

	if fDbs[0].Del {
		w := wFields
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printDeleteData(dst, dbs[0], w, fType, nil != c)
	}

	if fDbs[0].Remove {
		w := wFields
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printRemoveData(dst, dbs[0], w, fType, nil != c)
	}

	val = strings.ToLower(fDbs[0].Insert)
	if "self" == val || "parent" == val {
		f := fields
		if "self" == val {
			f = wFields
		}
		b.printInsertData(dst, typ, dbs[0], f, key, nil != c)
	}

	val = strings.ToLower(fDbs[0].Inserts)
	if "self" == val || "parent" == val {
		f := fields
		if "self" == val {
			f = wFields
		}
		b.printInsertListData(dst, typ, dbs[0], f, key, nil != c)
	}

	val = strings.ToLower(fDbs[0].Update)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printUpdateData(dst, typ, val, dbs[0], w, f, fType, c)
	}

	val = strings.ToLower(fDbs[0].Set)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printSetData(dst, typ, val, dbs[0], w, f, fType, c)
	}

	val = strings.ToLower(fDbs[0].Get)
	if "self" == val || "parent" == val {
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		if typ == fType {
			key.Dbs[0].Where = []string{"AND id = ?"}
			w = []*build.DBField{key}
		}
		b.printGetData(dst, typ, val, dbs[0], w, f, fType, c)
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
	item, scan, _ := b.getItemAndValue(fields, "self")
	dst.Code("func (val *" + name + ") DbScan() (string, []any) {\n")
	dst.Code("\treturn `" + item.String() + "`,\n")
	dst.Code("\t\t[]any{" + scan.String() + "}\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) getItemAndValue(fields []*build.DBField, key string) (strings.Builder, strings.Builder, strings.Builder) {
	item := strings.Builder{}
	scan := strings.Builder{}
	ques := strings.Builder{}
	isFist := true
	for _, field := range fields {
		get := ""
		if 0 < len(field.Dbs[0].Get) {
			get = strings.ReplaceAll(field.Dbs[0].Get, "?", build.StringToUnderlineName(field.Dbs[0].Name))
		} else if "self" == key {
			get = build.StringToUnderlineName(field.Dbs[0].Name)
		} else {
			continue
		}
		build.StringToUnderlineName(field.Dbs[0].Name)
		if !isFist {
			item.WriteString(", ")
			scan.WriteString(", ")
			ques.WriteString(", ")
		}
		isFist = false
		item.WriteString(get)
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
		fieldName := build.StringToHumpName(field.Field.Name.Name)
		for i, item := range text {
			if 1 == len(text) || !build.IsArray(field.Field.Type) {
				if build.IsNil(field.Field.Type) {
					where.Code("\tif nil != g." + fieldName + " {\n")
					b.printWhere(where, item, fieldName, "\t", build.IsArray(field.Field.Type))
					where.Code("\t}\n")
				} else {
					b.printWhere(where, item, fieldName, "", build.IsArray(field.Field.Type))
				}
			} else {
				if build.IsNil(field.Field.Type) {
					where.Code("\tif nil != g." + fieldName + " && " + strconv.Itoa(i) + " < len(g." + fieldName + ")")
					array := field.Field.Type.(*ast.ArrayType)
					if array.VType.Empty {
						where.Code(" && nil != g." + fieldName + "[" + strconv.Itoa(i) + "]")
					}
					where.Code(" {\n")
					b.printWhere(where, item, fieldName+"["+strconv.Itoa(i)+"]", "\t", false)
					where.Code("\t}\n")
				} else {
					b.printWhere(where, item, fieldName+"["+strconv.Itoa(i)+"]", "", false)
				}
			}
		}
	}

	if orderBy {
		isOrderFist := true
		for _, field := range fields {
			order := field.Dbs[0].Order
			if 0 < len(order) {
				where.Code("\tif ")
				if build.IsNil(field.Field.Type) {
					where.Code("nil != g." + build.StringToHumpName(field.Field.Name.Name))
					where.Code("&& \"ASC\" == *g." + build.StringToHumpName(field.Field.Name.Name) + " || \"DESC\" == *g." + build.StringToHumpName(field.Field.Name.Name))
				} else {
					where.Code("\"ASC\" == g." + build.StringToHumpName(field.Field.Name.Name) + " || \"DESC\" == g." + build.StringToHumpName(field.Field.Name.Name))
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
				if build.IsNil(field.Field.Type) {
					where.Code("\", \"$\", *g.")
				} else {
					where.Code("\", \"$\", g.")
				}
				where.Code(build.StringToHumpName(field.Field.Name.Name))
				where.Code("))\n")

				where.Code("\t}\n")

			}
		}
	}

	if page {
		if limit, ok := b.getLimit(fields); ok {
			if offset, ok := b.getOffset(fields); ok {
				where.Code("\ts.T(\" LIMIT " + offset.Dbs[0].Offset + ", " + limit.Dbs[0].Limit + "\")")
				where.Code(".P(g." + build.StringToHumpName(offset.Field.Name.Name) + ", g." + build.StringToHumpName(limit.Field.Name.Name) + ")\n")
			} else {
				where.Code("\ts.T(\" LIMIT " + limit.Dbs[0].Limit + "\")")
				where.Code(".P(g." + build.StringToHumpName(limit.Field.Name.Name) + ")\n")
			}
		}
	}
	return where
}

var quesRex = regexp.MustCompile(`\?`)
var paramRex = regexp.MustCompile(`\{age\}`)
var strRex = regexp.MustCompile(`(\${\w+})`)
var allRex = regexp.MustCompile(`(\${\w+})|(\{\w+})|\?`)

func (b *Builder) printWhere(where *build.Writer, text, fieldName, s string, isArray bool) {
	count := strings.Count(text, "?")
	if isArray {
		where.Import("github.com/wskfjtheqian/hbuf_golang/pkg/utils", "utl")
		temp := "\" + utl.ToQuestions(g." + fieldName + ", \",\") + \""

		match := quesRex.FindAllStringSubmatchIndex(text, -1)
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

	where.Code(s + "\ts.T(\" " + text + "\")")
	if 0 < count {
		where.Code(".P(")
		for i := 0; i < count; i++ {
			if 0 != i {
				where.Code(", ")
			}
			if isArray {
				where.Import("github.com/wskfjtheqian/hbuf_golang/pkg/utils", "utl")
				where.Code("utl.ToAnyList(g." + fieldName + ")...")
			} else {
				where.Code("g.")
				where.Code(fieldName)
			}
		}
		where.Code(")")
	}
	where.Code("\n")
}

func (b *Builder) printListData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)
	if typ != fType {
		key = "self"
	} else if "self" == key {
		dName = fName
	}
	w := b.getParamWhere(dst, wFields, true, true)
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields, key)
	dst.Code("func (g " + fName + ") DbList(ctx context.Context) ([]" + dName + ", error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\tret := make([]" + dName + ", 0)\n")
	if nil != c {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
		dst.Code("\tlist, key, _ := cache.DbGet(ctx, \"" + db.Name + "\", s, &ret)\n")
		dst.Code("\tif list != nil {\n")
		dst.Code("\t\treturn *list, nil\n")
		dst.Code("\t}\n")
	}
	dst.Import("database/sql", "")
	dst.Code("\t_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Code("\t\tvar val " + dName + "\n")
	dst.Code("\t\terr:= rows.Scan(" + scan.String() + ")\n")
	dst.Code("\t\tif err == nil {\n")
	dst.Code("\t\t\tret = append(ret, val)\n")
	dst.Code("\t\t}\n")
	dst.Code("\t\treturn true, err\n")
	dst.Code("\t})\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn nil, err\n")
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

func (b *Builder) printMapData(dst *build.Writer, key string, typ *ast.DataType, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, keyName string, c *cache) {
	kType, KName, ok := b.getKey(dst, fields, keyName)
	if !ok {
		return
	}
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)
	if typ != fType {
		key = "self"
	} else if "self" == key {
		dName = fName
	}
	w := b.getParamWhere(dst, wFields, true, true)
	dst.AddImports(w.GetImports())
	dst.AddImports(kType.GetImports())
	dst.AddImports(KName.GetImports())

	item, scan, _ := b.getItemAndValue(fields, key)
	dst.Code("func (g " + fName + ") DbMap(ctx context.Context) (map[" + kType.String() + "]" + dName + ", error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\tret := make(map[" + kType.String() + "]" + dName + ")\n")
	if nil != c {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
		dst.Code("\tlist, key, _ := cache.DbGet(ctx, \"" + db.Name + "\", s, &ret)\n")
		dst.Code("\tif list != nil {\n")
		dst.Code("\t\treturn *list, nil\n")
		dst.Code("\t}\n")
	}
	dst.Import("database/sql", "")
	dst.Code("\t_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Code("\t\tvar val " + dName + "\n")
	dst.Code("\t\terr:= rows.Scan(" + scan.String() + ")\n")
	dst.Code("\t\tif err == nil {\n")
	dst.Code("\t\t\tret[val.Get" + build.StringToHumpName(KName.String()) + "()] = val\n")
	dst.Code("\t\t}\n")
	dst.Code("\t\treturn true, err\n")
	dst.Code("\t})\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn nil, err\n")
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

func (b *Builder) printCountData(dst *build.Writer, typ *ast.DataType, db *build.DB, wFields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, wFields, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbCount(ctx context.Context) (int64, error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT COUNT(1) FROM " + db.Name + " WHERE del_time IS  NULL\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\tvar count int64\n")
	if nil != c {
		dst.Code("\tc, key, _ := cache.DbGet(ctx, \"" + db.Name + "\", s, &count)\n")
		dst.Code("\tif c != nil {\n")
		dst.Code("\t\treturn *c, nil\n")
		dst.Code("\t}\n")
	}
	dst.Import("database/sql", "")
	dst.Code("\t_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Code("\t\treturn false, rows.Scan(&count)\n")
	dst.Code("\t})\n")
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

func (b *Builder) printDeleteData(dst *build.Writer, db *build.DB, wFields []*build.DBField, fType *ast.DataType, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, wFields, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbDel(ctx context.Context) (int64, error) {\n")
	if isCache {
		dst.Code("\terr := cache.DbDel(ctx, \"" + db.Name + "\")\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, err\n")
		dst.Code("\t}\n")
	}

	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE " + db.Name + " SET del_time = NOW() WHERE 1 = 1\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\treturn s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printRemoveData(dst *build.Writer, db *build.DB, wFields []*build.DBField, fType *ast.DataType, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, wFields, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbRemove(ctx context.Context) (int64, error) {\n")
	if isCache {
		dst.Code("\terr := cache.DbDel(ctx, \"" + db.Name + "\")\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, err\n")
		dst.Code("\t}\n")
	}

	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"DELETE FROM " + db.Name + " WHERE 1 = 1\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\treturn s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func (g " + name + ") DbInsert(ctx context.Context) (int64, error) {\n")
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

	dst.Code("\treturn s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertListData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func (g " + name + ") DbInsertList(ctx context.Context, val []*" + name + ") (int64, error) {\n")
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

	dst.Code("\treturn s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printUpdateData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	if typ != fType {
		key = "parent"
	}
	w := b.getParamWhere(dst, fields, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbUpdate(ctx context.Context) (int64, error) {\n")
	if nil != c {
		dst.Code("\terr := cache.DbDel(ctx, \"" + db.Name + "\")\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, err\n")
		dst.Code("\t}\n")
	}
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE " + db.Name + " SET id = id\")\n")

	set := b.printSet(fields, key, true)
	dst.AddImports(set.GetImports())
	dst.Code(set.String())

	dst.Code("\ts.T(\"WHERE 1 = 1 \")\n")
	dst.Code(w.String())
	dst.Code("\n")

	dst.Code("\treturn s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printSetData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	if typ != fType {
		key = "parent"
	}
	w := b.getParamWhere(dst, wFields, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbSet(ctx context.Context) (int64, error) {\n")
	if nil != c {
		dst.Code("\terr := cache.DbDel(ctx, \"" + db.Name + "\")\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, err\n")
		dst.Code("\t}\n")
	}
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE " + db.Name + " SET id = id\")\n")

	set := b.printSet(fields, key, false)
	dst.AddImports(set.GetImports())
	dst.Code(set.String())

	dst.Code("\ts.T(\"WHERE 1 = 1 \")\n")
	dst.Code(w.String())
	dst.Code("\n")

	dst.Code("\treturn s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printSet(fields []*build.DBField, key string, isNil bool) *build.Writer {
	dst := build.NewWriter()
	for _, field := range fields {
		set := ""
		if 0 < len(field.Dbs[0].Set) {
			set = field.Dbs[0].Name + " = " + field.Dbs[0].Set
		} else if "self" == key {
			set = field.Dbs[0].Name + " = ?"
		} else {
			continue
		}

		name := build.StringToHumpName(field.Field.Name.Name)
		if isNil && build.IsNil(field.Field.Type) {
			dst.Code("\tif nil != g.")
			dst.Code(name)
			dst.Code(" {\n")
			dst.Code("\t")
		}
		dst.Code("\ts.T(\", " + set + " \")")
		count := strings.Count(set, "?")
		if 0 < count {
			dst.Code(".P(")
			for i := 0; i < count; i++ {
				if 0 != i {
					dst.Code(", ")
				}
				dst.Code("g.")
				dst.Code(name)
			}
			dst.Code(")")
		}
		dst.Code("\n")

		if isNil && build.IsNil(field.Field.Type) {
			dst.Code("\t}\n")
		}
	}
	return dst
}

func (b *Builder) printGetData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)
	if typ != fType {
		key = "self"
	} else if "self" == key {
		dName = fName
	}

	w := b.getParamWhere(dst, wFields, false, false)
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields, key)

	dst.Code("func (g " + fName + ") DbGet(ctx context.Context) (*" + dName + ", error) {\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM " + db.Name + " WHERE del_time IS NULL\")\n")
	dst.Code(w.GetCode().String())
	dst.Code("\ts.T(\" LIMIT 1\")\n")
	dst.Code("\tvar val " + dName + "\n")

	if nil != c {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
		dst.Code("\tcv, key, _ := cache.DbGet(ctx, \"" + db.Name + "\", s, &val)\n")
		dst.Code("\tif cv != nil {\n")
		dst.Code("\t\treturn cv, nil\n")
		dst.Code("\t}\n")
	}
	dst.Import("database/sql", "")
	dst.Code("\tcount, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Code("\t\treturn false, rows.Scan(" + scan.String() + ")\n")
	dst.Code("\t})\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn nil, err\n")
	dst.Code("\t}\n")
	dst.Code("\tif 0 == count {\n")
	dst.Code("\t\treturn nil, nil\n")
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
