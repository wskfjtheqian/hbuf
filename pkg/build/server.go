package build

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/scanner"
)

func (b *Builder) checkServer(file *ast.File, server *ast.ServerType, index int) error {
	name := server.Name.Name
	if _, ok := _keys[name]; ok {
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

	err := b.checkServerItem(file, server)
	if err != nil {
		return err
	}

	return nil
}

func (b *Builder) checkServerItem(file *ast.File, server *ast.ServerType) error {
	for index, item := range server.Methods {
		err := b.checkServerItemType(file, item.Result)
		if err != nil {
			return err
		}

		if _, ok := _keys[item.Name.Name]; ok {
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

		if _, ok := _keys[item.ParamName.Name]; ok {
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
			if _, ok := _types[ident.Name]; ok {
				return nil
			}
			obj := b.getDataType(file, ident.Name)
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
					if _, ok := _types[ident.Name]; ok {
						return nil
					}
					obj := b.getDataType(file, ident.Name)
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
		Msg: "Type can only be data",
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
