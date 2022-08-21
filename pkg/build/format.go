package build

import (
	"hbuf/pkg/ast"
)

type Format struct {
	IsNull bool
	Reg    string
}

func GetFormat(tags []*ast.Tag) *Format {
	val, ok := GetTag(tags, "format")
	if !ok {
		return nil
	}
	f := &Format{
		IsNull: true,
	}
	if nil != val.KV {
		for _, item := range val.KV {
			if "isNull" == item.Name.Name {
				f.IsNull = "true" == item.Value.Value[1:len(item.Value.Value)-1]
			} else if "reg" == item.Name.Name {
				f.Reg = item.Value.Value[1 : len(item.Value.Value)-1]
			}
		}
	}
	return f
}
