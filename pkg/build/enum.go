package build

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/scanner"
)

type UiEnum struct {
	Typ   string
	Class []string
}

func GetUiEnum(tags []*ast.Tag) *UiEnum {
	val, ok := GetTag(tags, "ui")
	if !ok {
		return nil
	}
	f := &UiEnum{}
	if nil != val.KV {
		for _, item := range val.KV {
			if "class" == item.Name.Name {
				for _, value := range item.Values {
					f.Class = append(f.Class, value.Value[1:len(value.Value)-1])
				}
			} else if "type" == item.Name.Name {
				f.Typ = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
			}
		}
	}
	return f
}

func (b *Builder) checkEnum(file *ast.File, enum *ast.EnumType, index int) error {
	name := enum.Name.Name
	if _, ok := _keys[BaseType(name)]; ok {
		return scanner.Error{
			Pos: b.fset.Position(enum.Name.Pos()),
			Msg: "Invalid Name: " + name,
		}
	}

	if b.checkDuplicateType(file, index, name) {
		return scanner.Error{
			Pos: b.fset.Position(enum.Name.Pos()),
			Msg: "Duplicate type: " + name,
		}
	}

	err := b.checkEnumItem(file, enum)
	if err != nil {
		return err
	}

	enum.Name.Obj.Data = file
	return nil
}

func (b *Builder) checkEnumItem(file *ast.File, enum *ast.EnumType) error {
	for index, item := range enum.Items {
		err := b.checkTags(item.Tags)
		if err != nil {
			return err
		}
		if _, ok := _keys[BaseType(item.Name.Name)]; ok {
			return scanner.Error{
				Pos: b.fset.Position(enum.Name.Pos()),
				Msg: "Invalid Name: " + item.Name.Name,
			}
		}
		if b.checkEnumDuplicateItem(enum, index, item.Name.Name) {
			return scanner.Error{
				Pos: b.fset.Position(item.Name.Pos()),
				Msg: "Duplicate item: " + item.Name.Name,
			}
		}
		if b.checkEnumDuplicateValue(enum, index, item.Id.Value) {
			return scanner.Error{
				Pos: b.fset.Position(item.Id.Pos()),
				Msg: "Duplicate item: " + item.Id.Value,
			}
		}
	}
	return nil
}

func (b *Builder) checkEnumDuplicateItem(enum *ast.EnumType, index int, name string) bool {
	for i := index + 1; i < len(enum.Items); i++ {
		s := enum.Items[i]
		if s.Name.Name == name {
			return true
		}
	}
	return false
}

func (b *Builder) checkEnumDuplicateValue(enum *ast.EnumType, index int, id string) bool {
	for i := index + 1; i < len(enum.Items); i++ {
		s := enum.Items[i]
		if s.Id.Value == id {
			return true
		}
	}
	return false
}
