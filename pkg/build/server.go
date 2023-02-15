package build

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/scanner"
)

func (b *Builder) checkServer(file *ast.File, server *ast.ServerType, index int) error {
	err := b.checkTags(server.Tags)
	if err != nil {
		return err
	}
	name := server.Name.Name
	if _, ok := _keys[BaseType(name)]; ok {
		return scanner.Error{
			Pos: b.fset.Position(server.Name.Pos()),
			Msg: "Invalid name: " + name,
		}
	}

	if b.checkDuplicateType(file, index, name) {
		return scanner.Error{
			Pos: b.fset.Position(server.Name.Pos()),
			Msg: "Duplicate type: " + name,
		}
	}

	err = b.checkServerExtends(file, server, index)
	if err != nil {
		return err
	}

	err = b.checkServerItem(file, server)
	if err != nil {
		return err
	}

	server.Name.Obj.Data = file
	return nil
}

func (b *Builder) checkServerItem(file *ast.File, server *ast.ServerType) error {
	for index, item := range server.Methods {
		err := b.checkTags(item.Tags)
		if err != nil {
			return err
		}

		err = b.checkServerItemType(file, item.Result)
		if err != nil {
			return err
		}

		if _, ok := _keys[BaseType(item.Name.Name)]; ok {
			return scanner.Error{
				Pos: b.fset.Position(item.Name.Pos()),
				Msg: "Invalid name: " + item.Name.Name,
			}
		}

		err = b.checkServerItemType(file, item.Param)
		if err != nil {
			return err
		}

		if b.checkServerDuplicateItem(server, index, item.Name.Name) {
			return scanner.Error{
				Pos: b.fset.Position(item.Name.Pos()),
				Msg: "Duplicate item: " + item.Name.Name,
			}
		}

		if _, ok := _keys[BaseType(item.ParamName.Name)]; ok {
			return scanner.Error{
				Pos: b.fset.Position(item.ParamName.Pos()),
				Msg: "Invalid name: " + item.ParamName.Name,
			}
		}

		//if b.checkDataDuplicateValue(server, index, item.Id.Value) {
		//	return scanner.Error{
		//		Pos: b.fset.Position(item.Id.Pos()),
		//		Msg: "Duplicate item: " + item.Id.Value,
		//	}
		//}
	}
	return nil
}

func (b *Builder) checkServerItemType(file *ast.File, result *ast.VarType) error {
	if result.IsEmpty() {
		return scanner.Error{
			Pos: b.fset.Position(result.TypeExpr.End()),
			Msg: "Map key cannot be empty",
		}
	}

	ident := result.TypeExpr.(*ast.Ident)
	if obj := file.Scope.Lookup(ident.Name); nil != obj {
		switch obj.Decl.(*ast.TypeSpec).Type.(type) {
		case *ast.DataType:
			if _, ok := _types[BaseType(ident.Name)]; ok {
				return nil
			}
			obj := b.GetDataType(file, ident.Name)
			if nil != obj {
				ident.Obj = obj
				return nil
			}
			return nil
		}
	}

	for _, spec := range file.Imports {
		if f, ok := b.pkg.Files[spec.Path.Value]; ok {
			if obj := f.Scope.Lookup(ident.Name); nil != obj {
				switch obj.Decl.(*ast.TypeSpec).Type.(type) {
				case *ast.DataType:
					if _, ok := _types[BaseType(ident.Name)]; ok {
						return nil
					}
					obj := b.GetDataType(file, ident.Name)
					if nil != obj {
						ident.Obj = obj
						return nil
					}
					return nil
				}
			}
		}
	}
	return scanner.Error{
		Pos: b.fset.Position(result.TypeExpr.End()),
		Msg: "Type can only be data: " + result.TypeExpr.(*ast.Ident).Name,
	}
}

func (b *Builder) checkServerDuplicateItem(server *ast.ServerType, index int, name string) bool {
	for i := index + 1; i < len(server.Methods); i++ {
		s := server.Methods[i]
		if s.Name.Name == name {
			return true
		}
	}
	return false
}

func (b *Builder) checkServerExtends(file *ast.File, server *ast.ServerType, index int) error {
	for i, item := range server.Extends {
		if _, ok := _keys[BaseType(item.Name)]; ok {
			return scanner.Error{
				Pos: b.fset.Position(server.Name.Pos()),
				Msg: "Invalid name: " + item.Name,
			}
		}
		if b.checkServerDuplicateExtends(server, i, item.Name) {
			return scanner.Error{
				Pos: b.fset.Position(item.NamePos),
				Msg: "Duplicate item: " + item.Name,
			}
		}

		obj := b.getServerExtends(file, index, item.Name)
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

func (b *Builder) getServerExtends(file *ast.File, index int, name string) *ast.Object {
	if obj := file.Scope.Lookup(name); nil != obj {
		switch obj.Decl.(type) {
		case *ast.TypeSpec:
			t := (obj.Decl.(*ast.TypeSpec)).Type
			switch t.(type) {
			case *ast.ServerType:
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
					case *ast.ServerType:
						return obj
					}
				}
			}
		}
	}
	return nil
}

func (b *Builder) checkServerDuplicateExtends(server *ast.ServerType, index int, name string) bool {
	for i := index + 1; i < len(server.Extends); i++ {
		s := server.Extends[i]
		if s.Name == name {
			return true
		}
	}
	return false
}
