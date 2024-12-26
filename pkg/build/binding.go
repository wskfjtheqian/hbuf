package build

import (
	"hbuf/pkg/ast"
	"strings"
)

type Binding struct {
	Server *ast.ServerType
	Method *ast.FuncType
}

func GetBinding(tags []*ast.Tag, file *ast.File, getType func(file *ast.File, name string) *ast.Object) (*Binding, error) {
	val, ok := GetTag(tags, "bind")
	if !ok {
		return nil, nil
	}

	if nil != val.KV {
		for _, item := range val.KV {
			if "value" == item.Name.Name {
				value := item.Values[0].Value[1 : len(item.Values[0].Value)-1]

				if 0 == len(value) {
					return nil, NewError(item.Pos()+1, "Not set value")
				}
				temp := strings.Split(value, ".")
				if 0 == len(temp) {
					return nil, NewError(item.Pos()+1, "Not find server method:"+value)
				}

				object := getType(file, temp[0])
				if nil == object {
					return nil, NewError(item.Pos()+1, "Not find server object: "+value)
				}

				if _, ok := object.Decl.(*ast.TypeSpec); !ok {
					return nil, NewError(item.Pos()+1, "Not a valid server type: "+value)
				}

				st, ok := object.Decl.(*ast.TypeSpec).Type.(*ast.ServerType)
				if !ok {
					return nil, NewError(item.Pos()+1, "Not a valid server type: "+value)
				}

				ei := getServerMethod(st, temp[1])
				if nil == ei {
					return nil, NewError(item.Pos()+1, "Not a valid enumeration field: "+value)
				}

				return &Binding{
					Server: st,
					Method: ei,
				}, nil
			}
		}
	}
	return nil, nil
}

func getServerMethod(em *ast.ServerType, name string) *ast.FuncType {
	for _, item := range em.Methods {
		if item.Name.Name == name {
			return item
		}
	}
	return nil
}
