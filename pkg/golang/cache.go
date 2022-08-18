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
	dst.Import("encoding/json", "")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/cache", "")
	dst.Import("math/rand", "")
	dst.Import("strconv", "")
	dst.Import("strings", "")

	b.printCacheSet(dst)
	b.printCacheGet(dst)
	b.printCacheDel(dst)

	dbs, fields, _, err := b.getDBField(typ)
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
		//b.printCacheCountData(dst, typ, dbs[0], fields, fType, fFields)
	}

	if fDbs[0].Get {
		//b.printCacheGetData(dst, typ, dbs[0], fields, fType, fFields)
	}
	return nil
}

func (b *Builder) printCacheFindData(dst *build.Writer, typ *ast.DataType, db *build.DB, fields []*build.DBField, fType *ast.DataType, fFields []*build.DBField) {
	fName := build.StringToHumpName(fType.Name.Name)
	dName := build.StringToHumpName(typ.Name.Name)

	p, w := b.getParamWhere(dst, fFields, true)
	dst.AddImports(p.GetImports())
	dst.AddImports(w.GetImports())

	//item, scan, _ := b.getItemAndValue(fields)

	dst.Code("func CacheFind" + fName + "(ctx context.Context, " + p.GetCode().String() + ") (*[]" + dName + ", error) {\n")
	dst.Code("\tkey := strings.Builder{}\n")
	dst.Code("\tkey.WriteString(\"db/cache/merchant_login/find_merchant_login/\")\n")
	dst.Code("\treturn CacheGet(ctx, key.String(), &[]MerchantLogin{})\n")
	dst.Code("}\n")
	dst.Code("\n")

	dst.Code("func CacheSetFind" + fName + "(ctx context.Context, val *[]" + dName + ", " + p.GetCode().String() + ") error {\n")
	dst.Code("\tkey := strings.Builder{}\n")
	dst.Code("\tkey.WriteString(\"db/cache/merchant_login/find_merchant_login/\")\n")
	dst.Code("\treturn CacheSet(ctx, key.String(), val)\n")
	dst.Code("}\n")
	dst.Code("\n")

}

func (b *Builder) printCacheSet(dst *build.Writer) {
	dst.Code("func CacheSet(ctx context.Context, key string, value any) error {\n")
	dst.Code("	marshal, err := json.Marshal(value)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return nil\n")
	dst.Code("	}\n")
	dst.Code("	c := cache.GET(ctx)\n")
	dst.Code("	err = c.Send(\"SET\", key, string(marshal))\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return nil\n")
	dst.Code("	}\n")
	dst.Code("	e := rand.Intn(3000-2000) + 2000\n")
	dst.Code("	err = c.Send(\"EXPIRE\", key, strconv.Itoa(e))\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return err\n")
	dst.Code("	}\n")
	dst.Code("	return c.Flush()\n")
	dst.Code("}\n\n")
}

func (b *Builder) printCacheGet(dst *build.Writer) {
	dst.Code("func CacheGet[T any](ctx context.Context, key string, value *T) (*T, error) {\n")
	dst.Code("	reply, err := cache.GET(ctx).Do(\"GET\", key)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return nil, err\n")
	dst.Code("	}\n")
	dst.Code("\terr = json.Unmarshal(reply.([]uint8), value)\n")
	dst.Code("\tif err != nil {\n")
	dst.Code("\t	return nil, err\n")
	dst.Code("\t}\n")
	dst.Code("	return value, nil\n")
	dst.Code("}\n\n")
}

func (b *Builder) printCacheDel(dst *build.Writer) {
	dst.Code("func CacheDel(ctx context.Context, key string) error {\n")
	dst.Code("	c := cache.GET(ctx)\n")
	dst.Code("	reply, err := c.Do(\"KEYS\", key)\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return err\n")
	dst.Code("	}\n")
	dst.Code("	if 0 == len(reply.([]any)) {\n")
	dst.Code("		return nil\n")
	dst.Code("	}\n")
	dst.Code("	err = c.Send(\"DEL\", reply.([]any))\n")
	dst.Code("	if err != nil {\n")
	dst.Code("		return err\n")
	dst.Code("	}\n")
	dst.Code("	return c.Flush()\n")
	dst.Code("}\n\n")
}
