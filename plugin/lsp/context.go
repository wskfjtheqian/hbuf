package lsp

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/token"
)

// contextKind 光标所在 AST 上下文类别。
type contextKind int

const (
	ctxTop contextKind = iota
	ctxData
	ctxDataField
	ctxDataExtends
	ctxServer
	ctxServerMethod
	ctxServerExtends
	ctxEnum
	ctxEnumItem
)

// context 光标位置在 AST 中的上下文。
type context struct {
	kind   contextKind
	spec   *ast.TypeSpec // 所在类型声明
	data   *ast.DataType
	server *ast.ServerType
	enum   *ast.EnumType
	field  *ast.Field
	method *ast.FuncType
	item   *ast.EnumItem
}

// offsetOf 将 Pos 转为文件内字节偏移。
func offsetOf(fset *token.FileSet, pos token.Pos) int {
	f := fset.File(pos)
	if f == nil {
		return -1
	}
	return f.Offset(pos)
}

func inRange(fset *token.FileSet, start, end token.Pos, offset int) bool {
	s, e := offsetOf(fset, start), offsetOf(fset, end)
	return s <= offset && offset <= e
}

// findContext 定位 offset 所在的 AST 上下文。
func findContext(pf *parsedFile, file *ast.File, offset int) *context {
	ctx := &context{kind: ctxTop}
	for _, spec := range file.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		switch typ := ts.Type.(type) {
		case *ast.DataType:
			if !inRange(pf.fset, typ.Pos(), typ.End(), offset) {
				continue
			}
			ctx.kind, ctx.spec, ctx.data = ctxData, ts, typ
			for _, field := range typ.Fields.List {
				if inRange(pf.fset, field.Pos(), field.End(), offset) {
					ctx.kind, ctx.field = ctxDataField, field
					return ctx
				}
			}
			for _, ext := range typ.Extends {
				if inRange(pf.fset, ext.Pos(), ext.End(), offset) {
					ctx.kind = ctxDataExtends
					return ctx
				}
			}
			return ctx
		case *ast.ServerType:
			if !inRange(pf.fset, typ.Pos(), typ.End(), offset) {
				continue
			}
			ctx.kind, ctx.spec, ctx.server = ctxServer, ts, typ
			for _, method := range typ.Methods {
				if inRange(pf.fset, method.Pos(), method.End(), offset) {
					ctx.kind, ctx.method = ctxServerMethod, method
					return ctx
				}
			}
			for _, ext := range typ.Extends {
				if inRange(pf.fset, ext.Pos(), ext.End(), offset) {
					ctx.kind = ctxServerExtends
					return ctx
				}
			}
			return ctx
		case *ast.EnumType:
			if !inRange(pf.fset, typ.Pos(), typ.End(), offset) {
				continue
			}
			ctx.kind, ctx.spec, ctx.enum = ctxEnum, ts, typ
			for _, item := range typ.Items {
				if inRange(pf.fset, item.Pos(), item.End(), offset) {
					ctx.kind, ctx.item = ctxEnumItem, item
					return ctx
				}
			}
			return ctx
		}
	}
	return ctx
}

// walkIdents 遍历文件中所有标识符节点。
func walkIdents(file *ast.File, fn func(ident *ast.Ident)) {
	for _, spec := range file.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		if ts.Name != nil {
			fn(ts.Name)
		}
		walkExpr(ts.Type, fn)
	}
}

// walkExpr 递归遍历表达式中的标识符。
func walkExpr(expr ast.Expr, fn func(ident *ast.Ident)) {
	if expr == nil {
		return
	}
	switch t := expr.(type) {
	case *ast.Ident:
		fn(t)
	case *ast.VarType:
		walkExpr(t.TypeExpr, fn)
	case *ast.ArrayType:
		if t.VType != nil {
			walkExpr(t.VType, fn)
		}
	case *ast.MapType:
		if t.VType != nil {
			walkExpr(t.VType, fn)
		}
		walkExpr(t.Key, fn)
	case *ast.DataType:
		for _, tag := range t.Tags {
			walkTag(tag, fn)
		}
		if t.Name != nil {
			fn(t.Name)
		}
		for _, ext := range t.Extends {
			if ext.Name != nil {
				fn(ext.Name)
			}
		}
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				walkField(field, fn)
			}
		}
	case *ast.ServerType:
		for _, tag := range t.Tags {
			walkTag(tag, fn)
		}
		if t.Name != nil {
			fn(t.Name)
		}
		for _, ext := range t.Extends {
			if ext.Name != nil {
				fn(ext.Name)
			}
		}
		for _, method := range t.Methods {
			walkMethod(method, fn)
		}
	case *ast.EnumType:
		for _, tag := range t.Tags {
			walkTag(tag, fn)
		}
		if t.Name != nil {
			fn(t.Name)
		}
		for _, item := range t.Items {
			for _, tag := range item.Tags {
				walkTag(tag, fn)
			}
			if item.Name != nil {
				fn(item.Name)
			}
		}
	}
}

func walkField(field *ast.Field, fn func(ident *ast.Ident)) {
	if field == nil {
		return
	}
	for _, tag := range field.Tags {
		walkTag(tag, fn)
	}
	if field.Name != nil {
		fn(field.Name)
	}
	walkExpr(field.Type, fn)
}

func walkMethod(method *ast.FuncType, fn func(ident *ast.Ident)) {
	if method == nil {
		return
	}
	for _, tag := range method.Tags {
		walkTag(tag, fn)
	}
	if method.Result != nil {
		walkExpr(method.Result, fn)
	}
	if method.Name != nil {
		fn(method.Name)
	}
	if method.Param != nil {
		walkExpr(method.Param, fn)
	}
	if method.ParamName != nil {
		fn(method.ParamName)
	}
}

func walkTag(tag *ast.Tag, fn func(ident *ast.Ident)) {
	if tag == nil {
		return
	}
	if tag.Name != nil {
		fn(tag.Name)
	}
	for _, kv := range tag.KV {
		if kv.Name != nil {
			fn(kv.Name)
		}
	}
}

// walkIdentsIn 遍历单个文件的标识符并按位置过滤。
// 返回位于 [start, end) 内的标识符。
func identsIn(pf *parsedFile, file *ast.File, offset int) []*ast.Ident {
	var ret []*ast.Ident
	walkIdents(file, func(ident *ast.Ident) {
		if offsetOf(pf.fset, ident.Pos()) <= offset && offset <= offsetOf(pf.fset, ident.End()) {
			ret = append(ret, ident)
		}
	})
	return ret
}
