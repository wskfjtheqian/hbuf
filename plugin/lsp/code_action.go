package lsp

import (
	"fmt"
	"hbuf/pkg/ast"
	"hbuf/pkg/token"
	"path/filepath"
	"strconv"
)

// codeAction 代码操作:
//  1. source.fixAll.hbuf —— 全文格式化(无语法错误时可用)
//  2. quickfix —— 删除未使用的 import
func (s *Server) codeAction(doc *Document, params CodeActionParams) []CodeAction {
	var actions []CodeAction

	// 1. 格式化文档(语法错误时 formatting 失败,自然不提供)
	if edits, err := s.formatting(doc); err == nil {
		actions = append(actions, CodeAction{
			Title: "格式化文档",
			Kind:  "source.fixAll.hbuf",
			Edit: &WorkspaceEdit{
				Changes: map[string][]TextEdit{doc.URI: edits},
			},
		})
	}

	// 2. 未使用的 import 删除
	pf := s.view.parse(doc)
	src := []byte(doc.Text)
	tf := fileToken(pf, doc.Path)
	file := pf.pkg.Files[doc.Path]
	if src != nil && tf != nil && file != nil {
		for _, spec := range file.Imports {
			if importUsed(pf, file, spec) {
				continue
			}
			path, err := unquoteImport(spec)
			if err != nil {
				continue
			}
			// title 展示文件名(import 路径已归一为绝对路径)
			name := path
			if filepath.IsAbs(path) {
				name = filepath.Base(path)
			}
			actions = append(actions, CodeAction{
				Title: fmt.Sprintf("删除未使用的 import %q", name),
				Kind:  "quickfix",
				Edit: &WorkspaceEdit{
					Changes: map[string][]TextEdit{doc.URI: {{
						Range:   importLineRange(src, tf, spec),
						NewText: "",
					}}},
				},
			})
		}
	}
	return actions
}

// importUsed 判断 import 是否被当前文件使用(其声明的类型是否被引用)。
// import 依赖文件解析失败时视为已使用,避免误删。
func importUsed(pf *parsedFile, file *ast.File, spec *ast.ImportSpec) bool {
	dep := pf.pkg.Files[spec.Path.Value]
	if dep == nil || dep.Scope == nil {
		return true
	}
	for _, obj := range dep.Scope.Objects {
		ts, ok := obj.Decl.(*ast.TypeSpec)
		if !ok {
			continue
		}
		// 当前文件有同名类型声明时以本地声明优先,该 import 不算被使用
		if local := file.Scope.Lookup(obj.Name); local != nil && local != obj {
			continue
		}
		if fileRefersType(file, obj.Name, ts) {
			return true
		}
	}
	return false
}

// fileRefersType 当前文件是否引用(非声明)指定类型名。
func fileRefersType(file *ast.File, name string, ts *ast.TypeSpec) bool {
	used := false
	walkIdents(file, func(id *ast.Ident) {
		if used || id.Name != name {
			return
		}
		// 排除声明位置本身(ts.Name)
		if ts.Name != nil && id.Pos() == ts.Name.Pos() {
			return
		}
		used = true
	})
	return used
}

// unquoteImport 还原 import 的字面路径(如 "base.hbuf" → base.hbuf)。
// 注意:view.loadFile 已将 import 路径归一为无引号的绝对路径。
func unquoteImport(spec *ast.ImportSpec) (string, error) {
	if s, err := strconv.Unquote(spec.Path.Value); err == nil {
		return s, nil
	}
	return spec.Path.Value, nil
}

// importLineRange 计算 import 声明整行(含换行)的范围,用于删除。
func importLineRange(src []byte, tf *token.File, spec *ast.ImportSpec) Range {
	line := tf.Position(spec.Pos()).Line // 1-based
	lineStart := int(tf.LineStart(line)) - tf.Base()
	lineEnd := lineStart
	for lineEnd < len(src) && src[lineEnd] != '\n' {
		lineEnd++
	}
	if lineEnd < len(src) {
		lineEnd++ // 连同换行一起删除
	}
	return toRange(src, tf,
		token.Pos(tf.Base()+lineStart), token.Pos(tf.Base()+lineEnd))
}
