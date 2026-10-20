package lsp

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/token"
)

// fileToken 按路径查找文件 token 对象。
func fileToken(pf *parsedFile, path string) *token.File {
	var tf *token.File
	pf.fset.Iterate(func(f *token.File) bool {
		if f.Name() == path {
			tf = f
			return false
		}
		return true
	})
	return tf
}

// typeString 输出类型文本(如 `int32?`、`string[]`、`string<int>`、`void`)。
func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.VarType:
		s := identText(t.TypeExpr)
		if t.Empty {
			s += "?"
		}
		return s
	case *ast.ArrayType:
		s := typeString(t.VType) + "[]"
		if t.Empty {
			s += "?"
		}
		return s
	case *ast.MapType:
		s := typeString(t.VType) + "<" + typeString(t.Key) + ">"
		if t.Empty {
			s += "?"
		}
		return s
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// identText 提取标识符名。
func identText(expr ast.Expr) string {
	if id, ok := expr.(*ast.Ident); ok && id != nil {
		return id.Name
	}
	return ""
}

// linePrefix 返回 offset 所在行从行首到 offset 的文本。
func linePrefix(src []byte, offset int) string {
	if offset > len(src) {
		offset = len(src)
	}
	start := offset
	for start > 0 && src[start-1] != '\n' {
		start--
	}
	return string(src[start:offset])
}

// lineEnd 返回 offset 所在行的行尾偏移(不含换行)。
func lineEnd(src []byte, offset int) int {
	if offset > len(src) {
		offset = len(src)
	}
	for offset < len(src) && src[offset] != '\n' {
		offset++
	}
	return offset
}

// wordAtPrefix 行前缀中光标正在输入的词及其起始偏移。
// 若紧邻光标的是空白/标点,则没有正在输入的词(返回 "")。
func wordAtPrefix(prefix string) (string, int) {
	i := len(prefix)
	for i > 0 && isIdentChar(prefix[i-1]) {
		i--
	}
	return prefix[i:], i
}
