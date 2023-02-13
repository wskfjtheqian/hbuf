package build

import (
	"hbuf/pkg/ast"
	"sort"
	"strconv"
	"strings"
)

type DB struct {
	index   int
	Name    string
	Key     bool
	typ     string
	Insert  bool
	Inserts bool
	Update  bool
	Set     bool
	Del     bool
	Get     bool
	List    bool
	Map     string
	Count   bool
	Table   bool
	Remove  bool
	Where   string
	Offset  string
	Limit   string
	Order   string
}

type DBField struct {
	Dbs   []*DB
	Field *ast.Field
}

func GetDB(n string, tag []*ast.Tag) []*DB {
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
						db.Name = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "order" == item.Name.Name {
						db.Order = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "key" == item.Name.Name {
						db.Key = "key" == item.Value.Value[1:len(item.Value.Value)-1]
					} else if "typ" == item.Name.Name {
						db.typ = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "where" == item.Name.Name {
						db.Where = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "offset" == item.Name.Name {
						db.Offset = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "limit" == item.Name.Name {
						db.Limit = item.Value.Value[1 : len(item.Value.Value)-1]
					} else if "insert" == item.Name.Name {
						db.Insert = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "inserts" == item.Name.Name {
						db.Inserts = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "update" == item.Name.Name {
						db.Update = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "del" == item.Name.Name {
						db.Del = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "get" == item.Name.Name {
						db.Get = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "list" == item.Name.Name {
						db.List = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "map" == item.Name.Name {
						db.Map = strings.ToLower(item.Value.Value[1 : len(item.Value.Value)-1])
					} else if "table" == item.Name.Name {
						db.Table = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "set" == item.Name.Name {
						db.Set = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "count" == item.Name.Name {
						db.Count = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					} else if "rm" == item.Name.Name {
						db.Remove = "true" == strings.ToLower(item.Value.Value[1:len(item.Value.Value)-1])
					}
				}
			}
			if "" == db.Name {
				db.Name = StringToUnderlineName(n)
			}
			dbs = append(dbs, &db)
		}
	}

	sort.Slice(dbs, func(i, j int) bool {
		return dbs[i].index > dbs[j].index
	})
	return dbs
}
