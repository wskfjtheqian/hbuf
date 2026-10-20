package lsp

import (
	"hbuf/pkg/token"
	"unicode/utf8"
)

// toPosition 将 token.Pos 转为 LSP Position(0-based,UTF-16 列)。
// src 为该文件完整内容。
func toPosition(src []byte, tf *token.File, pos token.Pos) Position {
	if nil == tf || !pos.IsValid() {
		return Position{}
	}
	offset := tf.Offset(pos)
	if offset >= len(src) {
		// EOF:token 层 lines 表不含尾随空行,统一映射到文档末尾(0-based 末行行首)
		line := 0
		for _, b := range src {
			if b == '\n' {
				line++
			}
		}
		return Position{Line: line, Character: 0}
	}
	tp := tf.Position(pos)
	line := tp.Line - 1
	lineStart := int(tf.LineStart(tp.Line)) - tf.Base()
	if lineStart < 0 || lineStart > len(src) {
		lineStart = 0
	}
	if tp.Offset < lineStart {
		tp.Offset = lineStart
	}
	return Position{Line: line, Character: utf16Len(src[lineStart:tp.Offset])}
}

// toRange 将起止 Pos 转为 LSP Range。
func toRange(src []byte, tf *token.File, start, end token.Pos) Range {
	return Range{
		Start: toPosition(src, tf, start),
		End:   toPosition(src, tf, end),
	}
}

// toOffset 将 LSP Position(0-based,UTF-16 列)转为字节偏移。
func toOffset(src []byte, pos Position) int {
	// 定位行起始
	line, lineStart := 0, 0
	for i := 0; i < len(src) && line < pos.Line; i++ {
		if src[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}
	// 行内按 UTF-16 列前进
	offset := lineStart
	col := 0
	for offset < len(src) && src[offset] != '\n' && col < pos.Character {
		r, size := utf8.DecodeRune(src[offset:])
		if r == utf8.RuneError && size <= 1 {
			col++
			offset++
		} else if r > 0xFFFF {
			col += 2
			offset += size
		} else {
			col++
			offset += size
		}
	}
	return offset
}

// utf16Len 计算字节序列的 UTF-16 code unit 长度。
func utf16Len(b []byte) int {
	n := 0
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r == utf8.RuneError && size <= 1 {
			n++
			b = b[1:]
		} else if r > 0xFFFF {
			n += 2
			b = b[size:]
		} else {
			n++
			b = b[size:]
		}
	}
	return n
}

// identifierAt 提取 offset 处的标识符及其字节区间。
func identifierAt(src []byte, offset int) (string, int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(src) {
		offset = len(src)
	}
	start, end := offset, offset
	for start > 0 && isIdentChar(src[start-1]) {
		start--
	}
	for end < len(src) && isIdentChar(src[end]) {
		end++
	}
	return string(src[start:end]), start, end
}

func isIdentChar(c byte) bool {
	return c == '_' ||
		(c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}
