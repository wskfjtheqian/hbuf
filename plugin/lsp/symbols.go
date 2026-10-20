package lsp

import (
	"hbuf/pkg/ast"
)

// documentSymbols 文档符号大纲:类型 → 字段/方法/枚举项。
func (s *Server) documentSymbols(doc *Document) []DocumentSymbol {
	pf := s.view.parse(doc)
	file := pf.pkg.Files[doc.Path]
	if file == nil {
		return nil
	}
	src := pf.srcs[doc.Path]
	tf := fileToken(pf, doc.Path)
	if src == nil || tf == nil {
		return nil
	}

	syms := make([]DocumentSymbol, 0, len(file.Specs))
	for _, spec := range file.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		sym := DocumentSymbol{
			Name:           ts.Name.Name,
			Kind:           typeKindOf(ts.Type),
			Detail:         typeWordOf(ts.Type),
			Range:          toRange(src, tf, ts.Type.Pos(), ts.Type.End()),
			SelectionRange: toRange(src, tf, ts.Name.Pos(), ts.Name.End()),
		}
		switch typ := ts.Type.(type) {
		case *ast.DataType:
			for _, field := range typ.Fields.List {
				sym.Children = append(sym.Children, DocumentSymbol{
					Name:           field.Name.Name,
					Detail:         typeString(field.Type) + " = " + field.Id.Value,
					Kind:           SymbolField,
					Range:          toRange(src, tf, field.Pos(), field.End()),
					SelectionRange: toRange(src, tf, field.Name.Pos(), field.Name.End()),
				})
			}
		case *ast.ServerType:
			for _, method := range typ.Methods {
				sym.Children = append(sym.Children, DocumentSymbol{
					Name:           method.Name.Name,
					Detail:         typeString(method.Result) + "(" + typeString(method.Param) + " " + method.ParamName.Name + ")",
					Kind:           SymbolMethod,
					Range:          toRange(src, tf, method.Pos(), method.End()),
					SelectionRange: toRange(src, tf, method.Name.Pos(), method.Name.End()),
				})
			}
		case *ast.EnumType:
			for _, item := range typ.Items {
				sym.Children = append(sym.Children, DocumentSymbol{
					Name:           item.Name.Name,
					Detail:         "= " + item.Id.Value,
					Kind:           SymbolEnumMember,
					Range:          toRange(src, tf, item.Pos(), item.End()),
					SelectionRange: toRange(src, tf, item.Name.Pos(), item.Name.End()),
				})
			}
		}
		syms = append(syms, sym)
	}
	return syms
}
