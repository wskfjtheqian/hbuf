package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/scanner"
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
		b.printNameData(dst, typ)
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
		w := wFields
		f := fields
		if "self" == val {
			f = wFields
		}
		b.printInsertData(dst, typ, val, dbs[0], w, f, fType, key, c)
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

func (b *Builder) printNameData(dst *build.Writer, typ *ast.DataType) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func (val *" + name + ") DbName() string {\n")
	dst.Code("\treturn `" + build.StringToUnderlineName(typ.Name.Name) + "`\n")
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
		scan.WriteString(b.converter(field, "val"))
		ques.WriteString("?")
	}
	return item, scan, ques

}

func (b *Builder) getParamWhere(dst *build.Writer, fields []*build.DBField, page, orderBy, groupBy bool) *build.Writer {
	where := build.NewWriter()
	where.Packages = dst.Packages

	for _, field := range fields {
		text := field.Dbs[0].Where
		fieldName := build.StringToHumpName(field.Field.Name.Name)
		for i, item := range text {
			if 1 == len(text) || !build.IsArray(field.Field.Type) {
				if build.IsNil(field.Field.Type) {
					where.Code("\tif nil != g." + fieldName + " {\n")
					_ = b.printParam(where, item, field, fields, "", "\t\ts")
					where.Code("\t}\n")
				} else {
					_ = b.printParam(where, item, field, fields, "", "\ts")
				}
			} else {
				if build.IsNil(field.Field.Type) {
					where.Code("\tif nil != g." + fieldName + " && " + strconv.Itoa(i) + " < len(g." + fieldName + ")")
					array := field.Field.Type.(*ast.ArrayType)
					if array.VType.Empty {
						where.Code(" && nil != g." + fieldName + "[" + strconv.Itoa(i) + "]")
					}
					where.Code(" {\n")
					_ = b.printParam(where, item, field, fields, "["+strconv.Itoa(i)+"]", "\t\ts")
					where.Code("\t}\n")
				} else {
					_ = b.printParam(where, item, field, fields, "["+strconv.Itoa(i)+"]", "\ts")
				}
			}
		}
	}

	if groupBy {
		isFist := true
		for _, field := range fields {
			group := field.Dbs[0].Group
			if 0 < len(group) {
				if build.IsNil(field.Field.Type) {
					where.Code("\tif nil != g." + build.StringToHumpName(field.Field.Name.Name) + " {\t")
				}
				if isFist {
					where.Code("\ts.T(\" GROUP BY \")")
				} else {
					where.Code("\ts.T(\", \")")
				}
				_ = b.printParam(where, group, field, fields, "", "")
				if build.IsNil(field.Field.Type) {
					where.Code("\t}\n")
				}
			}
		}
	}

	if orderBy {

		for _, field := range fields {
			order := field.Dbs[0].Order
			if 0 < len(order) {
				where.Code("\tif ")
				if build.IsNil(field.Field.Type) {
					where.Code("nil != g." + build.StringToHumpName(field.Field.Name.Name))
					where.Code(" && (\"ASC\" == *g." + build.StringToHumpName(field.Field.Name.Name) + " || \"DESC\" == *g." + build.StringToHumpName(field.Field.Name.Name) + ")")
				} else {
					where.Code("\"ASC\" == g." + build.StringToHumpName(field.Field.Name.Name) + " || \"DESC\" == g." + build.StringToHumpName(field.Field.Name.Name))
				}

				where.Code(" {\n")
				where.Code("\t\ts.T(\" ORDER BY \")")

				_ = b.printParam(where, order, field, fields, "", "")

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

func (b *Builder) findField(fields []*build.DBField, name string) *build.DBField {
	for _, item := range fields {
		if item.Field.Name.Name == name {
			return item
		}
	}
	return nil
}

var paramRex = regexp.MustCompile(`(\?{\w+})|(\${\w+})|\$|\?`)

func (b *Builder) printParam(buf *build.Writer, text string, self *build.DBField, fields []*build.DBField, array, tab string) error {
	match := paramRex.FindAllStringSubmatchIndex(text, -1)
	buf.Code(tab)
	if nil != match {
		var index = 0
		for _, item := range match {
			if 0 < item[0] {
				buf.Code(".T(\"")
				buf.Code(text[index:item[0]])
				buf.Code("\")")
			}
			t := text[item[0]:item[1]]
			if t == "$" || (2 < len(t) && "${" == t[0:2]) {
				field := self
				if t != "$" {
					field = b.findField(fields, t[2:len(t)-1])
				}

				buf.Code(".T(")
				if build.IsNil(field.Field.Type) {
					buf.Code("*g.")
				} else {
					buf.Code("g.")
				}
				buf.Code(build.StringToHumpName(field.Field.Name.Name))
				buf.Code(")")

			} else if t == "?" || (2 < len(t) && "?{" == t[0:2]) {
				field := self
				temp := array
				if t != "?" {
					field = b.findField(fields, t[2:len(t)-1])
					temp = ""
				}
				if nil == field {
					return scanner.Error{
						//Pos: b.Position(data.Name.Pos()),
						Msg: "Invalid name: " + t[2:len(t)-1],
					}
				}
				if 0 == len(field.Dbs[0].Converter) && build.IsArray(field.Field.Type) && 0 == len(temp) {
					buf.Code(".L(\",\", ")
					buf.Import("github.com/wskfjtheqian/hbuf_golang/pkg/utils", "utl")
					buf.Code("utl.ToAnyList(g." + build.StringToHumpName(field.Field.Name.Name) + ")...")
				} else {
					buf.Code(".V(")
					buf.Code(b.converter(field, "g"))
					buf.Code(temp)
				}
				buf.Code(")")
			}
			index = item[1]
		}
		if index < len(text) {
			buf.Code(".T(\"")
			buf.Code(text[index:])
			buf.Code("\")")
		}
	} else {
		buf.Code(".T(\"")
		buf.Code(text)
		buf.Code("\")")
	}
	buf.Code("\n")
	return nil
}

func (b *Builder) printListData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)
	if typ != fType {
		key = "self"
	} else if "self" == key {
		dName = fName
	}
	w := b.getParamWhere(dst, wFields, true, true, true)
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields, key)
	dst.Code("func (g " + fName + ") DbList(ctx context.Context) ([]" + dName + ", error) {\n")
	dst.Tab(1).Code("tableName := db.GET(ctx).Table(\"").Code(db.Name).Code("\")\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM \").T(tableName)")
	if db.Fake {
		dst.Code(".T(\" WHERE delete_time IS  NULL\")\n")
	} else {
		dst.Code(".T(\" WHERE 1 = 1\")\n")
	}
	dst.Code(w.GetCode().String())

	dst.Code("\tret := make([]" + dName + ", 0)\n")
	if nil != c {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
		dst.Code("\tlist, key, _ := cache.DbGet(ctx, \"\", tableName, s, &ret)\n")
		dst.Code("\tif list != nil {\n")
		dst.Code("\t\treturn *list, nil\n")
		dst.Code("\t}\n")
	}
	dst.Import("database/sql", "")
	dst.Code("\t_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Code("\t\tvar val " + dName + "\n")
	dst.Code("\t\terr := rows.Scan(" + scan.String() + ")\n")
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
		dst.Code("\t_, _ = cache.DbSet(ctx, key, tableName, s, &ret, ")
		if 0 < c.min {
			dst.Code("time.Duration(rand.Intn(" + strconv.Itoa(c.max) + "-" + strconv.Itoa(c.min) + ")+" + strconv.Itoa(c.min) + ")*time.Second")
		} else {
			dst.Code("0")
		}
		dst.Code(")\n")
		dst.Code("\tdefer cache.DbUnlock(ctx, tableName)\n")
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
	w := b.getParamWhere(dst, wFields, true, true, true)
	dst.AddImports(w.GetImports())
	dst.AddImports(kType.GetImports())
	dst.AddImports(KName.GetImports())

	item, scan, _ := b.getItemAndValue(fields, key)
	dst.Code("func (g " + fName + ") DbMap(ctx context.Context) (map[" + kType.String() + "]" + dName + ", error) {\n")
	dst.Tab(1).Code("tableName := db.GET(ctx).Table(\"").Code(db.Name).Code("\")\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM \").T(tableName)")
	if db.Fake {
		dst.Code(".T(\" WHERE delete_time IS  NULL\")\n")
	} else {
		dst.Code(".T(\" WHERE 1 = 1\")\n")
	}
	dst.Code(w.GetCode().String())

	dst.Code("\tret := make(map[" + kType.String() + "]" + dName + ")\n")
	if nil != c {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
		dst.Code("\tlist, key, _ := cache.DbGet(ctx, \"\", tableName, s, &ret)\n")
		dst.Code("\tif list != nil {\n")
		dst.Code("\t\treturn *list, nil\n")
		dst.Code("\t}\n")
	}
	dst.Import("database/sql", "")
	dst.Code("\t_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Code("\t\tvar val " + dName + "\n")
	dst.Code("\t\terr := rows.Scan(" + scan.String() + ")\n")
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
		dst.Code("\t_, _ = cache.DbSet(ctx, key, tableName, s, &ret, ")
		if 0 < c.min {
			dst.Code("time.Duration(rand.Intn(" + strconv.Itoa(c.max) + "-" + strconv.Itoa(c.min) + ")+" + strconv.Itoa(c.min) + ")*time.Second")
		} else {
			dst.Code("0")
		}
		dst.Code(")\n")
		dst.Code("\tdefer cache.DbUnlock(ctx, tableName)\n")
	}
	dst.Code("\treturn ret, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printCountData(dst *build.Writer, typ *ast.DataType, db *build.DB, wFields []*build.DBField, fType *ast.DataType, fFields []*build.DBField, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, wFields, false, false, true)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbCount(ctx context.Context) (int64, error) {\n")
	dst.Tab(1).Code("tableName := db.GET(ctx).Table(\"").Code(db.Name).Code("\")\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT COUNT(1) FROM \").T(tableName)")
	if db.Fake {
		dst.Code(".T(\" WHERE delete_time IS  NULL\")\n")
	} else {
		dst.Code(".T(\" WHERE 1 = 1\")\n")
	}
	dst.Code(w.GetCode().String())

	dst.Code("\tvar count int64\n")
	if nil != c {
		dst.Code("\tc, key, _ := cache.DbGet(ctx, \"\", tableName, s, &count)\n")
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
		dst.Code("\t_, _ = cache.DbSet(ctx, key, tableName, s, &count, ")
		if 0 < c.min {
			dst.Code("time.Duration(rand.Intn(" + strconv.Itoa(c.max) + "-" + strconv.Itoa(c.min) + ")+" + strconv.Itoa(c.min) + ")*time.Second")
		} else {
			dst.Code("0")
		}
		dst.Code(")\n")
		dst.Code("\tdefer cache.DbUnlock(ctx, tableName)\n")
	}
	dst.Code("\treturn count, nil\n")
	dst.Code("}\n")
	dst.Code("\n")
}

func (b *Builder) printDeleteData(dst *build.Writer, db *build.DB, wFields []*build.DBField, fType *ast.DataType, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, wFields, false, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbDel(ctx context.Context) (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.GET(ctx).Table(\"").Code(db.Name).Code("\")\n")
	if isCache {
		dst.Code("\terr := cache.DbDel(ctx, tableName)\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, 0, err\n")
		dst.Code("\t}\n")
		dst.Code("\tdefer cache.DbUnlock(ctx, tableName)\n")
	}

	dst.Code("\ts := db.NewSql()\n")
	if db.Fake {
		dst.Code("\ts.T(\"UPDATE \").T(tableName).T(\" SET delete_time = NOW() WHERE 1 = 1\")\n")
	} else {
		dst.Code("\ts.T(\"DELETE FROM \").T(tableName).T(\" WHERE 1 = 1\")\n")
	}
	dst.Code(w.GetCode().String())

	dst.Code("\treturn s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printRemoveData(dst *build.Writer, db *build.DB, wFields []*build.DBField, fType *ast.DataType, isCache bool) {
	fName := build.StringToHumpName(fType.Name.Name)

	w := b.getParamWhere(dst, wFields, false, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbRemove(ctx context.Context) (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.GET(ctx).Table(\"").Code(db.Name).Code("\")\n")
	if isCache {
		dst.Code("\terr := cache.DbDel(ctx, tableName)\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, 0, err\n")
		dst.Code("\t}\n")
		dst.Code("\tdefer cache.DbUnlock(ctx, tableName)\n")
	}

	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"DELETE FROM \").T(tableName).T(\" WHERE 1 = 1\")\n")
	dst.Code(w.GetCode().String())

	dst.Code("\treturn s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertData(dst *build.Writer, typ *ast.DataType, val string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, key *build.DBField, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	if typ != fType {
		val = "parent"
	}
	w := b.getParamWhere(dst, wFields, false, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbInsert(ctx context.Context) (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.GET(ctx).Table(\"").Code(db.Name).Code("\")\n")
	if nil != c {
		dst.Code("\terr := cache.DbDel(ctx, tableName)\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, 0, err\n")
		dst.Code("\t}\n")
		dst.Code("\tdefer cache.DbUnlock(ctx, tableName)\n")
	}
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"INSERT INTO \").T(tableName).T(\" SET \").Del(\",\")\n")

	set := b.printSet(fields, val, false)
	dst.AddImports(set.GetImports())
	dst.Code(set.String())

	dst.Code("\treturn s.Exec(ctx)\n")
	dst.Code("}\n\n")
}

func (b *Builder) printInsertListData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, key *build.DBField, isCache bool) {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Code("func (g " + name + ") DbInsertList(ctx context.Context, val []*" + name + ") (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.GET(ctx).Table(\"").Code(db.Name).Code("\")\n")
	dst.Code("\tif nil == val || 0 == len(val) {\n")
	dst.Code("\t\treturn 0, 0, nil\n")
	dst.Code("\t}\n")
	if isCache {
		dst.Code("\terr := cache.DbDel(ctx, tableName)\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, 0, err\n")
		dst.Code("\t}\n")
		dst.Code("\tdefer cache.DbUnlock(ctx, tableName)\n")
	}
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"INSERT INTO \").T(tableName).T(\" (")
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
		dst.Code(b.converter(field, "val"))
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
	w := b.getParamWhere(dst, fields, false, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbUpdate(ctx context.Context) (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.GET(ctx).Table(\"").Code(db.Name).Code("\")\n")
	if nil != c {
		dst.Code("\terr := cache.DbDel(ctx, tableName)\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, 0, err\n")
		dst.Code("\t}\n")
		dst.Code("\tdefer cache.DbUnlock(ctx, tableName)\n")
	}
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE \").T(tableName).T(\" SET \").Del(\",\")\n")

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
	w := b.getParamWhere(dst, wFields, false, false, false)
	dst.AddImports(w.GetImports())

	dst.Code("func (g " + fName + ") DbSet(ctx context.Context) (int64, int64, error) {\n")
	dst.Tab(1).Code("tableName := db.GET(ctx).Table(\"").Code(db.Name).Code("\")\n")
	if nil != c {
		dst.Code("\terr := cache.DbDel(ctx, tableName)\n")
		dst.Code("\tif err != nil {\n")
		dst.Code("\t\treturn 0, 0, err\n")
		dst.Code("\t}\n")
		dst.Code("\tdefer cache.DbUnlock(ctx, tableName)\n")
	}
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"UPDATE \").T(tableName).T(\" SET \").Del(\",\")\n")

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
		if isNil && build.IsNil(field.Field.Type) && !field.Dbs[0].Force {
			name := build.StringToHumpName(field.Field.Name.Name)
			dst.Code("\tif nil != g.")
			dst.Code(name)
			dst.Code(" {\n")
			dst.Code("\t")
		}

		dst.Code("\ts.T(\",\")")
		_ = b.printParam(dst, set, field, fields, "", "")

		if isNil && build.IsNil(field.Field.Type) && !field.Dbs[0].Force {
			dst.Code("\t}\n")
		}
	}
	return dst

}

func (b *Builder) printGetData(dst *build.Writer, typ *ast.DataType, key string, db *build.DB, wFields []*build.DBField, fields []*build.DBField, fType *ast.DataType, c *cache) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)
	if typ == fType {
		key = "self"
	} else if "self" == key {
		dName = fName
	} else {
		key = "self"
	}

	w := b.getParamWhere(dst, wFields, false, true, true)
	dst.AddImports(w.GetImports())

	item, scan, _ := b.getItemAndValue(fields, key)

	dst.Code("func (g " + fName + ") DbGet(ctx context.Context) (*" + dName + ", error) {\n")
	dst.Tab(1).Code("tableName := db.GET(ctx).Table(\"").Code(db.Name).Code("\")\n")
	dst.Code("\ts := db.NewSql()\n")
	dst.Code("\ts.T(\"SELECT " + item.String() + " FROM \").T(tableName)")
	if db.Fake {
		dst.Code(".T(\" WHERE delete_time IS  NULL\")\n")
	} else {
		dst.Code(".T(\" WHERE 1 = 1\")\n")
	}
	dst.Code(w.GetCode().String())
	dst.Code("\ts.T(\" LIMIT 1\")\n")
	dst.Code("\tvar val *" + dName + "\n")

	if nil != c {
		dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
		dst.Code("\tcv, key, _ := cache.DbGet(ctx, \"\", tableName, s, &val)\n")
		dst.Code("\tif cv != nil {\n")
		dst.Code("\t\treturn *cv, nil\n")
		dst.Code("\t}\n")
	}
	dst.Import("database/sql", "")
	dst.Code("\t_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {\n")
	dst.Code("\t\tval = &" + dName + "{}\n")
	dst.Code("\t\treturn false, rows.Scan(" + scan.String() + ")\n")
	dst.Code("\t})\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t\treturn nil, err\n")
	dst.Code("\t}\n")
	if nil != c {
		dst.Import("math/rand", "")
		dst.Import("time", "")
		dst.Code("\t_, _ = cache.DbSet(ctx, key, tableName, s, &val, ")
		if 0 < c.min {
			dst.Code("time.Duration(rand.Intn(" + strconv.Itoa(c.max) + "-" + strconv.Itoa(c.min) + ")+" + strconv.Itoa(c.min) + ")*time.Second")
		} else {
			dst.Code("0")
		}
		dst.Code(")\n")
		dst.Code("\tdefer cache.DbUnlock(ctx, tableName)\n")
	}
	dst.Code("\treturn val, nil\n")
	dst.Code("}\n\n")

}
