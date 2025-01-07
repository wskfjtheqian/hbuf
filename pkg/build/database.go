package build

import (
	"hbuf/pkg/ast"
	"sort"
	"strconv"
	"strings"
)

type DB struct {
	index     int
	Name      string
	Schema    string
	Key       bool
	Force     bool
	typ       string
	Insert    string
	Inserts   string
	Update    string
	Set       string
	Del       bool
	Get       string
	List      string
	Map       string
	Count     bool
	Table     string
	Remove    bool
	Where     []string
	Offset    string
	Limit     string
	Order     string
	Converter string
	Group     string
	Fake      bool
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
				Fake:  true,
			}
			if nil != val.KV {
				for _, item := range val.KV {
					if "name" == item.Name.Name {
						db.Name = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "schema" == item.Name.Name {
						db.Schema = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "converter" == item.Name.Name {
						db.Converter = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "order" == item.Name.Name {
						db.Order = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "key" == item.Name.Name {
						db.Key = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]
					} else if "typ" == item.Name.Name {
						db.typ = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "where" == item.Name.Name {
						where := make([]string, len(item.Values))
						for i, val := range item.Values {
							where[i] = val.Value[1 : len(item.Values[0].Value)-1]
						}
						db.Where = where
					} else if "offset" == item.Name.Name {
						db.Offset = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "limit" == item.Name.Name {
						db.Limit = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "insert" == item.Name.Name {
						db.Insert = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "inserts" == item.Name.Name {
						db.Inserts = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "update" == item.Name.Name {
						db.Update = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "group" == item.Name.Name {
						db.Group = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "del" == item.Name.Name {
						db.Del = "true" == strings.ToLower(item.Values[0].Value[1:len(item.Values[0].Value)-1])
					} else if "get" == item.Name.Name {
						db.Get = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "list" == item.Name.Name {
						db.List = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "map" == item.Name.Name {
						db.Map = strings.ToLower(item.Values[0].Value[1 : len(item.Values[0].Value)-1])
					} else if "table" == item.Name.Name {
						db.Table = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "set" == item.Name.Name {
						db.Set = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
					} else if "count" == item.Name.Name {
						db.Count = "true" == strings.ToLower(item.Values[0].Value[1:len(item.Values[0].Value)-1])
					} else if "force" == item.Name.Name {
						db.Force = "true" == strings.ToLower(item.Values[0].Value[1:len(item.Values[0].Value)-1])
					} else if "rm" == item.Name.Name {
						db.Remove = "true" == strings.ToLower(item.Values[0].Value[1:len(item.Values[0].Value)-1])
					} else if "fake" == item.Name.Name {
						db.Fake = "true" == strings.ToLower(item.Values[0].Value[1:len(item.Values[0].Value)-1])
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
