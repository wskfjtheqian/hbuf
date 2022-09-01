package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
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

func (b *Builder) printCacheCode(dst *build.Writer, typ *ast.DataType) error {
	c := getCache(typ.Name.Name, typ.Tags)
	if nil == c {
		return nil
	}

	dst.Import("context", "")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/db", "")
	dst.Import("strings", "")

	dbs, fields, key, err := b.getDBField(typ)
	if 0 == len(dbs) || nil != err {
		return nil
	}

	fDbs := dbs
	fFields := fields
	fType := typ
	if !dbs[0].Table && 0 < len(dbs[0].Name) {
		table := b.build.GetDataType(b.getFile(typ.Name), dbs[0].Name)
		if nil == table {
			return nil
		}
		typ = table.Decl.(*ast.TypeSpec).Type.(*ast.DataType)
		dbs, fields, _, err = b.getDBField(typ)
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

	if fDbs[0].Find {
		b.printCacheFindData(dst, typ, dbs[0], fields, fType, fFields)
	}

	if fDbs[0].Count {
		b.printCacheCountData(dst, typ, dbs[0], fields, fType, fFields)
	}

	if fDbs[0].Get {
		if typ == fType {
			key.Dbs[0].Where = "AND id = ?"
			fFields = []*build.DBField{key}
		}
		b.printCacheGetData(dst, typ, dbs[0], fields, fType, fFields)
	}
	b.printCacheDelData(dst, typ, dbs[0], fields, fType, fFields)
	return nil
}

func (b *Builder) printCacheDelData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField) {
	fName := build.StringToHumpName(fType.Name.Name)

	dst.Code("func CacheDel" + fName + "(ctx context.Context) error {\n")
	dst.Code("\tkey := strings.Builder{}\n")
	dst.Code("\tkey.WriteString(\"db/cache/" + db.Name + "/*\")\n")
	dst.Code("\treturn cache.CacheDel(ctx, key.String())\n")
	dst.Code("}\n")
	dst.Code("\n")

}

func (b *Builder) printCacheGetData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	dst.Import("crypto/md5", "")
	dst.Import("encoding/hex", "")

	dst.Code("func cacheGetGet" + fName + "(ctx context.Context,s *db.Sql) (*" + dName + ", error) {\n")
	dst.Code("\tdata := md5.Sum([]byte(s.ToText()))\n")
	dst.Code("\treturn cache.CacheGet(ctx, \"db/cache/" + db.Name + "/get_" + build.StringToUnderlineName(fType.Name.Name) + "/\"+hex.EncodeToString(data[:]), &" + dName + "{})\n")
	dst.Code("}\n")
	dst.Code("\n")

	dst.Code("func cacheSetGet" + fName + "(ctx context.Context, val *" + dName + ", s *db.Sql) error {\n")
	dst.Code("\tdata := md5.Sum([]byte(s.ToText()))\n")
	dst.Code("\treturn cache.CacheSet(ctx, \"db/cache/" + db.Name + "/get_" + build.StringToUnderlineName(fType.Name.Name) + "/\"+hex.EncodeToString(data[:]), val)\n")
	dst.Code("}\n")
	dst.Code("\n")

}

func (b *Builder) printCacheFindData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	dst.Import("crypto/md5", "")
	dst.Import("encoding/hex", "")

	dst.Code("func cacheGetFind" + fName + "(ctx context.Context, s *db.Sql) (*[]" + dName + ", error) {\n")
	dst.Code("\tdata := md5.Sum([]byte(s.ToText()))\n")
	dst.Code("\treturn cache.CacheGet(ctx, \"db/cache/" + db.Name + "/find_" + build.StringToUnderlineName(fType.Name.Name) + "/\"+hex.EncodeToString(data[:]), &[]" + dName + "{})\n")
	dst.Code("}\n")
	dst.Code("\n")

	dst.Code("func cacheSetFind" + fName + "(ctx context.Context, val *[]" + dName + ", s *db.Sql) error {\n")
	dst.Code("\tdata := md5.Sum([]byte(s.ToText()))\n")
	dst.Code("\treturn cache.CacheSet(ctx, \"db/cache/" + db.Name + "/find_" + build.StringToUnderlineName(fType.Name.Name) + "/\"+hex.EncodeToString(data[:]), val)\n")
	dst.Code("}\n")
	dst.Code("\n")

}

func (b *Builder) printCacheCountData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField) {
	fName := build.StringToHumpName(fType.Name.Name)
	dst.Import("crypto/md5", "")
	dst.Import("encoding/hex", "")

	dst.Code("func cacheGetCount" + fName + "(ctx context.Context, s *db.Sql) (*int64, error) {\n")
	dst.Code("\tdata := md5.Sum([]byte(s.ToText()))\n")
	dst.Code("\tvar count int64\n")
	dst.Code("\treturn cache.CacheGet(ctx, \"db/cache/" + db.Name + "/count_" + build.StringToUnderlineName(fType.Name.Name) + "/\"+hex.EncodeToString(data[:]), &count)\n")
	dst.Code("}\n")
	dst.Code("\n")

	dst.Code("func cacheSetCount" + fName + "(ctx context.Context, val *int64, s *db.Sql) error {\n")
	dst.Code("\tdata := md5.Sum([]byte(s.ToText()))\n")
	dst.Code("\treturn cache.CacheSet(ctx, \"db/cache/" + db.Name + "/count_" + build.StringToUnderlineName(fType.Name.Name) + "/\"+hex.EncodeToString(data[:]), val)\n")
	dst.Code("}\n")
	dst.Code("\n")

}
