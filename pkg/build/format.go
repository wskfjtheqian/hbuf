package build

import (
	"hbuf/pkg/ast"
)

type Format struct {
	Null bool
	Reg  string
	Min  string
	Max  string
}

func GetFormat(tags []*ast.Tag) *Format {
	val, ok := GetTag(tags, "format")
	if !ok {
		return nil
	}
	f := &Format{
		Null: false,
	}
	if nil != val.KV {
		for _, item := range val.KV {
			if "null" == item.Name.Name {
				f.Null = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]
			} else if "reg" == item.Name.Name {
				f.Reg = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			} else if "min" == item.Name.Name {
				f.Min = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			} else if "max" == item.Name.Name {
				f.Max = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			}
		}
	}
	return f
}
