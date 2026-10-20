package lsp

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
)

// diagnosticsOf 计算某文档的完整诊断:语法错误 + 类型校验错误。
// 依赖文件中的错误按文件归属返回(供各打开文档分发)。
func (s *Server) diagnosticsOf(doc *Document) ([]Diagnostic, map[string][]Diagnostic) {
	pf := s.view.parse(doc)

	byPath := map[string][]Diagnostic{}

	// 语法错误(按文件归属)
	for path, errs := range pf.errs {
		for _, e := range errs {
			byPath[path] = append(byPath[path], diagnosticOf(pf, e.Pos, "syntax", e.Msg))
		}
	}

	// 类型校验错误:与命令行同源(build.CheckPackage)。
	// 存在语法错误时跳过校验,避免在残缺 AST 上产生误报。
	if len(pf.errs[doc.Path]) == 0 {
		if err := build.CheckPackage(pf.fset, pf.pkg); err != nil {
			collectCheckError(pf, byPath, err)
		}
	}

	return byPath[doc.Path], byPath
}

// collectCheckError 将校验错误按文件归属收集。
func collectCheckError(pf *parsedFile, byPath map[string][]Diagnostic, err error) {
	switch e := err.(type) {
	case scanner.ErrorList:
		for _, item := range e {
			byPath[item.Pos.Filename] = append(byPath[item.Pos.Filename],
				diagnosticOf(pf, item.Pos, "check", item.Msg))
		}
	case scanner.Error:
		byPath[e.Pos.Filename] = append(byPath[e.Pos.Filename],
			diagnosticOf(pf, e.Pos, "check", e.Msg))
	}
}

// diagnosticOf 将错误位置与消息转为诊断。
func diagnosticOf(pf *parsedFile, pos token.Position, code, msg string) Diagnostic {
	diag := Diagnostic{
		Severity: SeverityError,
		Code:     code,
		Source:   "hbuf",
		Message:  msg,
	}
	if pos.IsValid() {
		src := pf.srcs[pos.Filename]
		var tf *token.File
		pf.fset.Iterate(func(f *token.File) bool {
			if f.Name() == pos.Filename {
				tf = f
				return false
			}
			return true
		})
		if tf != nil && src != nil {
			start := tf.Pos(pos.Offset)
			end := start
			// 尝试延伸到下一个字符,给出可见的下划线范围
			if pos.Offset < len(src) {
				end = tf.Pos(pos.Offset + 1)
			}
			diag.Range = toRange(src, tf, start, end)
		}
	}
	return diag
}

// publishAllDiagnostics 对所有打开文档重新诊断并发布。
func (s *Server) publishAllDiagnostics() {
	for _, doc := range s.view.Docs() {
		diags, byPath := s.diagnosticsOf(doc)
		// 将依赖文件诊断发布到对应打开文档(跳过当前文档,避免重复通知)
		for path, list := range byPath {
			if uri, ok := s.viewPathURI(path); ok && uri != doc.URI {
				s.sendDiagnostics(uri, list)
			}
		}
		s.sendDiagnostics(doc.URI, diags)
	}
}

// viewPathURI 路径 → 已打开文档 uri。
func (s *Server) viewPathURI(path string) (string, bool) {
	if doc := s.view.Doc(pathToURI(path)); doc != nil {
		return doc.URI, true
	}
	return "", false
}

// sendDiagnostics 发布诊断通知。
func (s *Server) sendDiagnostics(uri string, diags []Diagnostic) {
	if diags == nil {
		diags = []Diagnostic{}
	}
	s.notify("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diags,
	})
}

// fileOf 取路径对应的已解析 AST 文件。
func fileOf(pf *parsedFile, path string) *ast.File {
	return pf.pkg.Files[path]
}
