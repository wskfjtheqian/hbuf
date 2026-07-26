package build

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/scanner"
	"strconv"
	"strings"
)

type Marshal struct {
	In  []string
	Out []string
}

func GetFieldTag(tag []*ast.Tag) map[string][]string {
	val, ok := GetTag(tag, "tag")
	if !ok {
		return nil
	}

	ret := make(map[string][]string)
	if nil != val.KV {
		for _, item := range val.KV {
			name := strings.ToLower(item.Name.Name)
			if name == "hbuf" {
				continue
			}
			for _, value := range item.Values {
				v := strings.Trim(value.Value[1:len(value.Value)-1], "")
				if len(v) > 0 {
					ret[name] = append(ret[name], v)
				}
			}
		}
	}
	return ret
}

func GetMarshal(tag []*ast.Tag) *Marshal {
	val, ok := GetTag(tag, "marshal")
	if !ok {
		return nil
	}
	marshal := Marshal{}
	if nil != val.KV {
		for _, item := range val.KV {
			if "out" == item.Name.Name {
				for _, value := range item.Values {
					marshal.Out = append(marshal.Out, value.Value[1:len(value.Value)-1])
				}
			} else if "in" == item.Name.Name {
				for _, value := range item.Values {
					marshal.In = append(marshal.In, value.Value[1:len(value.Value)-1])
				}
			}
		}
	}
	return &marshal
}

func (b *Builder) checkData(file *ast.File, data *ast.DataType, index int) error {
	err := b.checkTags(data.Tags)
	if err != nil {
		return err
	}
	name := data.Name.Name
	if _, ok := _keys[BaseType(name)]; ok {
		return scanner.Error{
			Pos: b.fset.Position(data.Name.Pos()),
			Msg: "Invalid Name: " + name,
		}
	}

	if b.checkDuplicateType(file, index, name) {
		return scanner.Error{
			Pos: b.fset.Position(data.Name.Pos()),
			Msg: "Duplicate type: " + name,
		}
	}

	err = b.checkDataExtends(file, data, index)
	if err != nil {
		return err
	}

	err = b.checkDataItem(file, data)
	if err != nil {
		return err
	}

	data.Name.Obj.Data = file
	return nil
}

func (b *Builder) checkDataExtends(file *ast.File, data *ast.DataType, index int) error {
	for i, item := range data.Extends {
		if _, ok := _keys[BaseType(item.Name.Name)]; ok {
			return scanner.Error{
				Pos: b.fset.Position(data.Name.Pos()),
				Msg: "Invalid Name: " + item.Name.Name,
			}
		}
		if b.checkDataDuplicateExtends(data, i, item.Name.Name) {
			return scanner.Error{
				Pos: b.fset.Position(item.Name.NamePos),
				Msg: "Duplicate item: " + item.Name.Name,
			}
		}
		if b.checkDataDuplicateValue(data, i, item.Id.Value) {
			return scanner.Error{
				Pos: b.fset.Position(item.Id.Pos()),
				Msg: "Duplicate id: " + item.Id.Value,
			}
		}
		id, err := strconv.ParseInt(item.Id.Value, 10, 64)
		if err != nil {
			return scanner.Error{
				Pos: b.fset.Position(item.Id.Pos()),
				Msg: "Invalid id: " + item.Id.Value,
			}
		}
		if id < 1 || id > 0xFFFF {
			return scanner.Error{
				Pos: b.fset.Position(item.Id.Pos()),
				Msg: "Invalid id: " + item.Id.Value,
			}
		}

		obj := b.getDataExtends(file, index, item.Name.Name)
		if nil == obj {
			return scanner.Error{
				Pos: b.fset.Position(item.Name.NamePos),
				Msg: "Not find: " + item.Name.Name,
			}
		}
		item.Name.Obj = obj
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

func (b *Builder) GetDataType(file *ast.File, name string) *ast.Object {
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
		if _, ok := _types[BaseType(ident.Name)]; ok {
			return nil
		}
		obj := b.GetDataType(file, ident.Name)
		if nil != obj {
			ident.Obj = obj
			return nil
		}
		return scanner.Error{
			Pos: b.fset.Position(ident.Pos()),
			Msg: "Invalid Name: " + ident.Name,
		}
	}

	return scanner.Error{
		Pos: b.fset.Position(typ.Pos()),
		Msg: "Type Error",
	}
}

func (b *Builder) checkDataItem(file *ast.File, data *ast.DataType) error {
	for index, item := range data.Fields.List {
		err := b.checkTags(item.Tags)
		if err != nil {
			return err
		}
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
		if _, ok := _keys[BaseType(item.Name.Name)]; ok {
			return scanner.Error{
				Pos: b.fset.Position(item.Name.Pos()),
				Msg: "Invalid Name: " + item.Name.Name,
			}
		}
		if b.checkDataDuplicateItem(data, index, item.Name.Name) {
			return scanner.Error{
				Pos: b.fset.Position(item.Name.Pos()),
				Msg: "Duplicate item: " + item.Name.Name,
			}
		}
		id, err := strconv.ParseInt(item.Id.Value, 10, 64)
		if err != nil {
			return scanner.Error{
				Pos: b.fset.Position(item.Id.Pos()),
				Msg: "Invalid id: " + item.Id.Value,
			}
		}

		if id < 1 || id > 0xFFFF {
			return scanner.Error{
				Pos: b.fset.Position(item.Id.Pos()),
				Msg: "Invalid id: " + item.Id.Value,
			}
		}
	}
	return nil
}

func (b *Builder) checkDataDuplicateExtends(data *ast.DataType, index int, name string) bool {
	for i := index + 1; i < len(data.Extends); i++ {
		s := data.Extends[i]
		if s.Name.Name == name {
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
