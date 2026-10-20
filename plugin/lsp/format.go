package lsp

import (
	"bytes"
	"fmt"
	"hbuf/pkg/format"
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
	"strings"
)

// formatting 全文格式化:接入 pkg/format,返回全量替换 TextEdit。
// 独立解析(不复用 view 的解析缓存),避免 import 路径被归一为绝对路径后写入结果。
func (s *Server) formatting(doc *Document) ([]TextEdit, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, doc.Path, []byte(doc.Text), parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("cannot format: document has syntax errors")
	}
	var buf bytes.Buffer
	if err := format.Node(&buf, file); err != nil {
		return nil, err
	}
	// 全量替换:从文档开头到末尾
	lineCount := strings.Count(doc.Text, "\n")
	return []TextEdit{{
		Range: Range{
			Start: Position{},
			End:   Position{Line: lineCount, Character: 0},
		},
		NewText: buf.String(),
	}}, nil
}
