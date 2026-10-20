package lsp

import (
	"fmt"
	"strings"
)

// hover 悬停提示:Doc 注释、类型、ID。
func (s *Server) hover(doc *Document, pos Position) *Hover {
	pf := s.view.parse(doc)
	src := []byte(doc.Text)
	offset := toOffset(src, pos)
	d := lookup(pf, doc, offset)
	if d == nil {
		return nil
	}

	var b strings.Builder
	switch d.kind {
	case declType:
		fmt.Fprintf(&b, "```hbuf\n%s %s\n```\n", typeWordOf(d.spec.Type), d.name)
		if docText := docTextOf(d.spec.Doc); docText != "" {
			b.WriteString("\n" + docText)
		}
	case declField:
		fmt.Fprintf(&b, "```hbuf\n字段 %s %s = %s\n```\n", d.name, typeString(d.field.Type), d.field.Id.Value)
		if docText := docTextOf(d.field.Doc); docText != "" {
			b.WriteString("\n" + docText)
		}
	case declMethod:
		fmt.Fprintf(&b, "```hbuf\n方法 %s %s(%s %s) = %s\n```\n",
			typeString(d.method.Result), d.name, typeString(d.method.Param), d.method.ParamName.Name, d.method.Id.Value)
		if docText := docTextOf(d.method.Doc); docText != "" {
			b.WriteString("\n" + docText)
		}
	case declEnumItem:
		fmt.Fprintf(&b, "```hbuf\n枚举项 %s = %s\n```\n", d.name, d.item.Id.Value)
		if docText := docTextOf(d.item.Doc); docText != "" {
			b.WriteString("\n" + docText)
		}
	case declBaseType:
		fmt.Fprintf(&b, "```hbuf\n基础类型 %s\n```\n", d.base)
	}

	// 标识符范围
	name, start, end := identifierAt(src, offset)
	if name == "" {
		return &Hover{Contents: MarkupContent{Kind: "markdown", Value: b.String()}}
	}
	tf := fileToken(pf, doc.Path)
	r := toRange(src, tf, tf.Pos(start), tf.Pos(end))
	return &Hover{Contents: MarkupContent{Kind: "markdown", Value: b.String()}, Range: &r}
}
