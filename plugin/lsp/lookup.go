package lsp

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/token"
)

// declKind 声明类别。
type declKind int

const (
	declNone declKind = iota
	declType          // data/enum/server 类型
	declField
	declMethod
	declEnumItem
	declBaseType // 基础类型关键字
)

// decl 查找结果:标识符对应的声明。
type decl struct {
	kind   declKind
	name   string
	spec   *ast.TypeSpec
	field  *ast.Field
	method *ast.FuncType
	item   *ast.EnumItem
	file   *ast.File
	path   string // 声明所在文件路径
	base   string // 基础类型名(kind == declBaseType)
}

// namePosOf 声明的名称位置(用于跳转/悬停范围)。
func (d *decl) namePos() (token.Pos, token.Pos) {
	switch d.kind {
	case declType:
		if d.spec != nil && d.spec.Name != nil {
			return d.spec.Name.Pos(), d.spec.Name.End()
		}
	case declField:
		if d.field != nil && d.field.Name != nil {
			return d.field.Name.Pos(), d.field.Name.End()
		}
	case declMethod:
		if d.method != nil && d.method.Name != nil {
			return d.method.Name.Pos(), d.method.Name.End()
		}
	case declEnumItem:
		if d.item != nil && d.item.Name != nil {
			return d.item.Name.Pos(), d.item.Name.End()
		}
	}
	return token.NoPos, token.NoPos
}

// lookup 在 offset 处查找标识符的声明。
func lookup(pf *parsedFile, doc *Document, offset int) *decl {
	file := pf.pkg.Files[doc.Path]
	if file == nil {
		return nil
	}
	name, _, _ := identifierAt([]byte(doc.Text), offset)
	if name == "" {
		return nil
	}

	// 基础类型关键字
	if _, ok := baseTypes[name]; ok {
		return &decl{kind: declBaseType, name: name, base: name, file: file, path: doc.Path}
	}

	ctx := findContext(pf, file, offset)

	// 字段/方法/枚举项自身的名字
	if ctx.kind == ctxDataField && ctx.field.Name.Name == name {
		return &decl{kind: declField, name: name, field: ctx.field, file: file, path: doc.Path}
	}
	if ctx.kind == ctxServerMethod && ctx.method.Name.Name == name {
		return &decl{kind: declMethod, name: name, method: ctx.method, file: file, path: doc.Path}
	}
	if ctx.kind == ctxEnumItem && ctx.item.Name.Name == name {
		return &decl{kind: declEnumItem, name: name, item: ctx.item, file: file, path: doc.Path}
	}

	// 当前文件作用域(类型)
	if obj := file.Scope.Lookup(name); obj != nil {
		if ts, ok := obj.Decl.(*ast.TypeSpec); ok {
			return &decl{kind: declType, name: name, spec: ts, file: file, path: doc.Path}
		}
	}

	// import 文件作用域(跨文件类型)
	for _, spec := range file.Imports {
		dep := pf.pkg.Files[spec.Path.Value]
		if dep == nil {
			continue
		}
		if obj := dep.Scope.Lookup(name); obj != nil {
			if ts, ok := obj.Decl.(*ast.TypeSpec); ok {
				return &decl{kind: declType, name: name, spec: ts, file: dep, path: spec.Path.Value}
			}
		}
	}

	// data 块内:引用 extends/字段名 → 字段定义
	if ctx.kind == ctxDataField || ctx.kind == ctxData {
		if ctx.data != nil {
			if f := findField(ctx.data, name); f != nil {
				return &decl{kind: declField, name: name, field: f, file: file, path: doc.Path}
			}
		}
	}
	// server 块内:引用方法名
	if ctx.kind == ctxServerMethod || ctx.kind == ctxServer {
		if ctx.server != nil {
			if m := findMethod(ctx.server, name); m != nil {
				return &decl{kind: declMethod, name: name, method: m, file: file, path: doc.Path}
			}
		}
	}
	return nil
}

// findField 在当前 data(含 extends 递归)中查找字段。
func findField(data *ast.DataType, name string) *ast.Field {
	for _, field := range data.Fields.List {
		if field.Name.Name == name {
			return field
		}
	}
	for _, ext := range data.Extends {
		if ext.Name.Obj == nil {
			continue
		}
		if ts, ok := ext.Name.Obj.Decl.(*ast.TypeSpec); ok {
			if parent, ok := ts.Type.(*ast.DataType); ok {
				if f := findField(parent, name); f != nil {
					return f
				}
			}
		}
	}
	return nil
}

// findMethod 在当前 server(含 extends 递归)中查找方法。
func findMethod(server *ast.ServerType, name string) *ast.FuncType {
	for _, method := range server.Methods {
		if method.Name.Name == name {
			return method
		}
	}
	for _, ext := range server.Extends {
		if ext.Name.Obj == nil {
			continue
		}
		if ts, ok := ext.Name.Obj.Decl.(*ast.TypeSpec); ok {
			if parent, ok := ts.Type.(*ast.ServerType); ok {
				if m := findMethod(parent, name); m != nil {
					return m
				}
			}
		}
	}
	return nil
}

// baseTypes hbuf 基础类型与关键字。
var baseTypes = map[string]bool{
	"int8": true, "int16": true, "int32": true, "int64": true,
	"uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"bool": true, "float": true, "double": true, "decimal": true,
	"bytes": true, "string": true, "date": true,
	"void": true, "stream": true,
	"data": true, "enum": true, "server": true,
	"import": true, "package": true,
}

// typeKindOf 类型声明 → LSP SymbolKind。
func typeKindOf(typ ast.Expr) int {
	switch typ.(type) {
	case *ast.EnumType:
		return SymbolEnum
	case *ast.ServerType:
		return SymbolInterface
	case *ast.DataType:
		return SymbolStruct
	}
	return 0
}

// typeWordOf 类型声明 → 语言关键字。
func typeWordOf(typ ast.Expr) string {
	switch typ.(type) {
	case *ast.EnumType:
		return "enum"
	case *ast.ServerType:
		return "server"
	case *ast.DataType:
		return "data"
	}
	return ""
}

// typeNameOf 类型声明名称。
func typeNameOf(typ ast.Expr) string {
	switch t := typ.(type) {
	case *ast.EnumType:
		return t.Name.Name
	case *ast.ServerType:
		return t.Name.Name
	case *ast.DataType:
		return t.Name.Name
	}
	return ""
}

// docTextOf 注释组文本。
func docTextOf(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	return group.Text()
}

// baseTypeDetail 基础类型补全详情。
func baseTypeDetail(name string) string {
	if _, ok := baseTypes[name]; ok {
		return "基础类型"
	}
	return ""
}
