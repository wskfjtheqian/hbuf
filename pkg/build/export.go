package build

import (
	"hbuf/pkg/ast"
)

type Export struct {
	Filter []string
	Width  uint32
}

func GetExport(tags []*ast.Tag) (*Export, error) {
	val, ok := GetTag(tags, "export")
	if !ok {
		return nil, nil
	}

	v := &Export{
		Filter: make([]string, 0),
	}
	for _, item := range val.KV {
		if "filter" == item.Name.Name {
			for _, i := range item.Values {
				filter := i.Value[1 : len(i.Value)-1]
				if 0 == len(filter) {
					return nil, NewError(i.Pos()+1, "Not set filter")
				}
				v.Filter = append(v.Filter, filter)
			}
		}
	}
	return v, nil
}
