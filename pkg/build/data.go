package build

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/scanner"
)

func (b *Builder) checkData(file *ast.File, data *ast.DataType, index int) error {
	name := data.Name.Name
	if _, ok := _keys[name]; ok {
		return scanner.Error{
			Pos: b.fset.Position(data.Name.Pos()),
			Msg: "Invalid name: " + name,
		}
	}

	if b.checkDuplicateType(file, index, name) {
		return scanner.Error{
			Pos: b.fset.Position(data.Name.Pos()),
			Msg: "Duplicate type: " + name,
		}
	}

	err := b.checkDataExtends(file, data, index)
	if err != nil {
		return err
	}

	err = b.checkDataItem(file, data)
	if err != nil {
		return err
	}

	//obj := ast.NewObj(ast.Data, name)
	//obj.Decl = data
	//file.Scope.Insert(obj)
	return nil
}

func (b *Builder) checkDataExtends(file *ast.File, data *ast.DataType, index int) error {
	for i, item := range data.Extends {
		if _, ok := _keys[item.Name]; ok {
			return scanner.Error{
				Pos: b.fset.Position(data.Name.Pos()),
				Msg: "Invalid name: " + item.Name,
			}
		}
		if b.checkDataDuplicateExtends(data, i, item.Name) {
			return scanner.Error{
				Pos: b.fset.Position(item.NamePos),
				Msg: "Duplicate item: " + item.Name,
			}
		}

		obj := b.getDataExtends(file, index, item.Name)
		if nil == obj {
			return scanner.Error{
				Pos: b.fset.Position(item.NamePos),
				Msg: "Not find: " + item.Name,
			}
		}
		item.Obj = obj
	}
	return nil
}

func (b *Builder) getDataExtends(file *ast.File, index int, name string) *ast.Object {
	if obj := file.Scope.Lookup(name); nil != obj {
		switch obj.Decl.(type) {
		case *ast.TypeSpec:
			t := (obj.Decl.(*ast.TypeSpec)).Type
			switch t.(type) {
			case *ast.DataType:
				return obj
			}
		}

	}
	for _, spec := range file.Imports {
		if f, ok := b.pkg.Files[spec.Path.Value]; ok {
			if obj := f.Scope.Lookup(name); nil != obj {
				switch obj.Decl.(type) {
				case *ast.TypeSpec:
					t := (obj.Decl.(*ast.TypeSpec)).Type
					switch t.(type) {
					case *ast.DataType:
						return obj
					}
				}
			}
		}
	}
	return nil
}

func (b *Builder) getDataType(file *ast.File, name string) *ast.Object {
	if obj := file.Scope.Lookup(name); nil != obj {
		switch obj.Decl.(type) {
		case *ast.TypeSpec:
			t := (obj.Decl.(*ast.TypeSpec)).Type
			switch t.(type) {
			case *ast.DataType:
				return obj
			case *ast.EnumType:
				return obj
			}
		}
	}
	for _, spec := range file.Imports {
		if f, ok := b.pkg.Files[spec.Path.Value]; ok {
			if obj := f.Scope.Lookup(name); nil != obj {
				switch obj.Decl.(type) {
				case *ast.TypeSpec:
					t := (obj.Decl.(*ast.TypeSpec)).Type
					switch t.(type) {
					case *ast.DataType:
						return obj
					case *ast.EnumType:
						return obj
					}
				}
			}
		}
	}
	return nil
}

func (b *Builder) checkDataItemType(file *ast.File, typ ast.Type) error {
	switch typ.Type().(type) {
	case *ast.Ident:
		ident := typ.Type().(*ast.Ident)
		if _, ok := _types[ident.Name]; ok {
			return nil
		}
		obj := b.getDataType(file, ident.Name)
		if nil != obj {
			ident.Obj = obj
			return nil
		}
		return scanner.Error{
			Pos: b.fset.Position(ident.Pos()),
			Msg: "Invalid name: " + ident.Name,
		}
	}

	return scanner.Error{
		Pos: b.fset.Position(typ.Pos()),
		Msg: "Type Error",
	}
}

func (b *Builder) checkDataItem(file *ast.File, data *ast.DataType) error {
	for index, item := range data.Fields.List {
		switch item.Type.(type) {
		case *ast.VarType:
			err := b.checkDataItemType(file, item.Type)
			if err != nil {
				return err
			}
		case *ast.ArrayType:
			err := b.checkDataItemType(file, item.Type.(*ast.ArrayType).Type().(*ast.VarType))
			if err != nil {
				return err
			}
		case *ast.MapType:
			ma := item.Type.(*ast.MapType)
			err := b.checkDataMapKey(file, ma.Key.(*ast.VarType))
			if err != nil {
				return err
			}
			err = b.checkDataItemType(file, ma.Type().(*ast.VarType))
			if err != nil {
				return err
			}
		}
		if _, ok := _keys[item.Name.Name]; ok {
			return scanner.Error{
				Pos: b.fset.Position(item.Name.Pos()),
				Msg: "Invalid name: " + item.Name.Name,
			}
		}
		if b.checkDataDuplicateItem(data, index, item.Name.Name) {
			return scanner.Error{
				Pos: b.fset.Position(item.Name.Pos()),
				Msg: "Duplicate item: " + item.Name.Name,
			}
		}
		if b.checkDataDuplicateValue(data, index, item.Id.Value) {
			return scanner.Error{
				Pos: b.fset.Position(item.Id.Pos()),
				Msg: "Duplicate item: " + item.Id.Value,
			}
		}
	}
	return nil
}

func (b *Builder) checkDataDuplicateExtends(data *ast.DataType, index int, name string) bool {
	for i := index + 1; i < len(data.Extends); i++ {
		s := data.Extends[i]
		if s.Name == name {
			return true
		}
	}
	return false
}

func (b *Builder) checkDataDuplicateItem(data *ast.DataType, index int, name string) bool {
	for i := index + 1; i < len(data.Fields.List); i++ {
		s := data.Fields.List[i]
		if s.Name.Name == name {
			return true
		}
	}
	return false
}

func (b *Builder) checkDataDuplicateValue(data *ast.DataType, index int, id string) bool {
	for i := index + 1; i < len(data.Fields.List); i++ {
		s := data.Fields.List[i]
		if s.Id.Value == id {
			return true
		}
	}
	return false
}
