package lsp

import (
	"fmt"
	"hbuf/pkg/ast"
	"hbuf/pkg/token"
)

// definition 跳转定义(支持跨文件 import)。
// 未找到时返回 nil(协议中编码为 null)。
func (s *Server) definition(doc *Document, pos Position) []Location {
	pf := s.view.parse(doc)
	offset := toOffset([]byte(doc.Text), pos)
	d := lookup(pf, doc, offset)
	if d == nil || d.kind == declBaseType {
		return nil
	}
	start, end := d.namePos()
	if !start.IsValid() {
		return nil
	}
	return []Location{locationOfPos(pf, d.path, start, end)}
}

// locationOfPos 由 Pos 区间构造 Location。
func locationOfPos(pf *parsedFile, path string, start, end token.Pos) Location {
	src := pf.srcs[path]
	tf := fileToken(pf, path)
	return Location{
		URI:   pathToURI(path),
		Range: toRange(src, tf, start, end),
	}
}

// references 查找引用(支持跨文件 import)。
func (s *Server) references(doc *Document, pos Position, includeDecl bool) []Location {
	pf := s.view.parse(doc)
	offset := toOffset([]byte(doc.Text), pos)
	d := lookup(pf, doc, offset)
	if d == nil {
		return nil
	}
	declStart, declEnd := d.namePos()

	var locs []Location
	for path, file := range pf.pkg.Files {
		src := pf.srcs[path]
		tf := fileToken(pf, path)
		if src == nil || tf == nil {
			continue
		}
		seen := map[string]bool{} // 同一 Ident(如 TypeSpec.Name 与内嵌 Name)只记一次
		walkIdents(file, func(id *ast.Ident) {
			if id.Name != d.name {
				return
			}
			if !sameDecl(id, d) {
				return
			}
			// 声明位置按需排除
			if !includeDecl && path == d.path && id.Pos() == declStart && id.End() == declEnd {
				return
			}
			key := fmt.Sprintf("%d:%d", id.Pos(), id.End())
			if seen[key] {
				return
			}
			seen[key] = true
			locs = append(locs, Location{
				URI:   pathToURI(path),
				Range: toRange(src, tf, id.Pos(), id.End()),
			})
		})
	}
	return locs
}

// sameDecl 判断标识符是否引用同一声明。
func sameDecl(id *ast.Ident, d *decl) bool {
	switch d.kind {
	case declType:
		return id.Obj != nil && id.Obj.Decl == d.spec
	case declField:
		return id.Obj != nil && id.Obj.Decl == d.field
	case declMethod, declEnumItem:
		// 方法与枚举项无符号表条目,按名称匹配(块内语义)
		return id.Obj == nil
	}
	return false
}
