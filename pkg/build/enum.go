package build

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/scanner"
)

func (b *Builder) checkEnum(file *ast.File, enum *ast.EnumType, index int) error {
	name := enum.Name.Name
	if _, ok := _keys[name]; ok {
		return scanner.Error{
			Pos: b.fset.Position(enum.Name.Pos()),
			Msg: "Invalid name: " + name,
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
		if _, ok := _keys[item.Name.Name]; ok {
			return scanner.Error{
				Pos: b.fset.Position(enum.Name.Pos()),
				Msg: "Invalid name: " + item.Name.Name,
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
