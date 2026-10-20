package lsp

import (
	"fmt"
	"hbuf/pkg/ast"
	"hbuf/pkg/token"
	"path/filepath"
	"strings"
)

// inlayHints 内联提示:
//  1. extends 成员摘要 —— 继承类型名后提示父类型成员数量
//  2. 跨文件类型来源 —— import 类型引用后提示声明文件名
//
// 仅返回 rng 范围内的提示。
func (s *Server) inlayHints(doc *Document, rng Range) []InlayHint {
	pf := s.view.parse(doc)
	src := []byte(doc.Text)
	tf := fileToken(pf, doc.Path)
	file := pf.pkg.Files[doc.Path]
	if src == nil || tf == nil || file == nil {
		return nil
	}

	var hints []InlayHint
	seen := map[string]bool{} // "line:char" 去重
	addHint := func(pos token.Pos, label string, kind int, tooltip string) {
		p := toPosition(src, tf, pos)
		if p.Line < rng.Start.Line || p.Line > rng.End.Line {
			return
		}
		key := fmt.Sprintf("%d:%d", p.Line, p.Character)
		if seen[key] {
			return
		}
		seen[key] = true
		hints = append(hints, InlayHint{
			Position:    p,
			Label:       label,
			Kind:        kind,
			PaddingLeft: true,
			Tooltip:     tooltip,
		})
	}

	for _, spec := range file.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		switch t := ts.Type.(type) {
		case *ast.DataType:
			for _, ext := range t.Extends {
				addExtendsHint(pf, file, ext, addHint)
				// extends 类型来源(跨文件)
				extTS, extFile := findTypeSpec(pf, file, ext.Name.Name)
				if extTS != nil && extFile != nil && extFile.Path != doc.Path {
					addHint(ext.Name.End(), filepath.Base(extFile.Path), InlayHintType,
						"定义于 "+extFile.Path)
				}
			}
			if t.Fields != nil {
				for _, field := range t.Fields.List {
					walkExpr(field.Type, func(id *ast.Ident) {
						addTypeOriginHint(pf, file, doc, id, addHint)
					})
				}
			}
		case *ast.ServerType:
			for _, ext := range t.Extends {
				addExtendsHint(pf, file, ext, addHint)
				extTS, extFile := findTypeSpec(pf, file, ext.Name.Name)
				if extTS != nil && extFile != nil && extFile.Path != doc.Path {
					addHint(ext.Name.End(), filepath.Base(extFile.Path), InlayHintType,
						"定义于 "+extFile.Path)
				}
			}
			for _, method := range t.Methods {
				if method.Result != nil {
					walkExpr(method.Result, func(id *ast.Ident) {
						addTypeOriginHint(pf, file, doc, id, addHint)
					})
				}
				if method.Param != nil {
					walkExpr(method.Param, func(id *ast.Ident) {
						addTypeOriginHint(pf, file, doc, id, addHint)
					})
				}
			}
		}
	}
	return hints
}

// addExtendsHint 继承类型名后提示父类型成员数量。
func addExtendsHint(pf *parsedFile, file *ast.File, ext *ast.Extends, addHint func(token.Pos, string, int, string)) {
	if ext == nil || ext.Name == nil {
		return
	}
	ts, _ := findTypeSpec(pf, file, ext.Name.Name)
	if ts == nil {
		return
	}
	switch t := ts.Type.(type) {
	case *ast.DataType:
		n := 0
		if t.Fields != nil {
			n = len(t.Fields.List)
		}
		names := make([]string, 0, n)
		if t.Fields != nil {
			for _, f := range t.Fields.List {
				if f.Name != nil {
					names = append(names, f.Name.Name)
				}
			}
		}
		addHint(ext.Name.End(), fmt.Sprintf("%d 个字段", n), InlayHintType,
			ext.Name.Name+" 字段: "+strings.Join(names, ", "))
	case *ast.ServerType:
		n := len(t.Methods)
		names := make([]string, 0, n)
		for _, m := range t.Methods {
			if m.Name != nil {
				names = append(names, m.Name.Name)
			}
		}
		addHint(ext.Name.End(), fmt.Sprintf("%d 个方法", n), InlayHintType,
			ext.Name.Name+" 方法: "+strings.Join(names, ", "))
	}
}

// addTypeOriginHint 类型引用来自 import 文件时,提示声明文件名。
func addTypeOriginHint(pf *parsedFile, file *ast.File, doc *Document, id *ast.Ident, addHint func(token.Pos, string, int, string)) {
	if id == nil {
		return
	}
	ts, declFile := findTypeSpec(pf, file, id.Name)
	if ts == nil || declFile == nil || declFile.Path == doc.Path {
		return
	}
	addHint(id.End(), filepath.Base(declFile.Path), InlayHintType, "定义于 "+declFile.Path)
}

// findTypeSpec 在当前文件及 import 依赖中查找类型声明。
// 返回 (声明, 声明所在文件);找不到返回 (nil, nil)。
func findTypeSpec(pf *parsedFile, file *ast.File, name string) (*ast.TypeSpec, *ast.File) {
	if file.Scope != nil {
		if obj := file.Scope.Lookup(name); obj != nil {
			if ts, ok := obj.Decl.(*ast.TypeSpec); ok {
				return ts, file
			}
		}
	}
	for _, spec := range file.Imports {
		dep := pf.pkg.Files[spec.Path.Value]
		if dep == nil || dep.Scope == nil {
			continue
		}
		if obj := dep.Scope.Lookup(name); obj != nil {
			if ts, ok := obj.Decl.(*ast.TypeSpec); ok {
				return ts, dep
			}
		}
	}
	return nil, nil
}
