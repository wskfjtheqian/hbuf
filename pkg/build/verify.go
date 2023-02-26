package build

import (
	"hbuf/pkg/ast"
	"strings"
)

type VerifyEnum struct {
	Name string
	Enum *ast.EnumType
	Item *ast.EnumItem
}

type Verify struct {
	format []*VerifyEnum
}

func (v *Verify) GetFormat() []*VerifyEnum {
	return v.format
}

func GetVerify(tags []*ast.Tag, file *ast.File, getType func(file *ast.File, name string) *ast.Object) (*Verify, error) {
	val, ok := GetTag(tags, "verify")
	if !ok {
		return nil, nil
	}

	v := &Verify{
		format: make([]*VerifyEnum, 0),
	}
	for _, item := range val.KV {
		if "format" == item.Name.Name {
			for _, i := range item.Values {
				format := i.Value[1 : len(i.Value)-1]
				if 0 == len(format) {
					//TODO 错误处理
					return nil, nil
				}
				temp := strings.Split(format, ".")
				if 0 == len(temp) {
					//TODO 错误处理
					return nil, nil
				}

				object := getType(file, temp[0])
				if nil == object {
					//TODO 错误处理
					return nil, nil
				}
				switch object.Decl.(type) {
				case *ast.TypeSpec:
					break
				default:
					//TODO 错误处理
					return nil, nil
				}

				switch object.Decl.(*ast.TypeSpec).Type.(type) {
				case *ast.EnumType:
					break
				default:
					//TODO 错误处理
					return nil, nil
				}
				em := object.Decl.(*ast.TypeSpec).Type.(*ast.EnumType)
				ei := getEnumItem(em, temp[1])
				if nil == ei {
					//TODO 错误处理
					return nil, nil
				}
				v.format = append(v.format, &VerifyEnum{
					Enum: em,
					Item: ei,
					Name: item.Name.Name,
				})
			}
		}
	}
	return v, nil
}

func getEnumItem(em *ast.EnumType, name string) *ast.EnumItem {
	for _, item := range em.Items {
		if item.Name.Name == name {
			return item
		}
	}
	return nil
}
