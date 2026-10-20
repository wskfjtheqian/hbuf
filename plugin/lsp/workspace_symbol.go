package lsp

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// workspaceSymbols 工作区符号搜索:扫描 workspace 根目录下的 .hbuf 文件,
// 收集类型/字段/方法/枚举项符号,按 query(大小写不敏感子串)过滤。
func (s *Server) workspaceSymbols(query string) []WorkspaceSymbol {
	var syms []WorkspaceSymbol
	for _, path := range s.workspaceFiles() {
		src := s.workspaceFileContent(path)
		if src == nil {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, src, parser.AllErrors)
		if err != nil || f == nil {
			continue
		}
		collectWorkspaceSymbols(f, path, src, fset, query, &syms)
		if len(syms) >= 500 {
			break
		}
	}
	sort.Slice(syms, func(i, j int) bool {
		if syms[i].Name != syms[j].Name {
			return syms[i].Name < syms[j].Name
		}
		return syms[i].Location.URI < syms[j].Location.URI
	})
	if len(syms) > 100 {
		syms = syms[:100]
	}
	return syms
}

// workspaceFiles 收集工作区候选文件(根目录递归 *.hbuf + 打开文档)。
func (s *Server) workspaceFiles() []string {
	roots := s.view.Roots()
	if len(roots) == 0 {
		// 无 workspace 配置:回退为打开文档所在目录
		for _, d := range s.view.Docs() {
			if d.Path != "" {
				roots = append(roots, filepath.Dir(d.Path))
			}
		}
	}
	seen := map[string]bool{}
	var files []string
	for _, root := range roots {
		if root == "" {
			continue
		}
		filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if p != root && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir // 跳过隐藏目录(.git 等)
				}
				return nil
			}
			if strings.HasSuffix(p, ".hbuf") && !seen[p] {
				seen[p] = true
				files = append(files, p)
			}
			return nil
		})
		if len(files) >= 1000 {
			break
		}
	}
	// 未保存的打开文档也纳入(可能不在磁盘上)
	for _, d := range s.view.Docs() {
		if d.Path != "" && !seen[d.Path] {
			seen[d.Path] = true
			files = append(files, d.Path)
		}
	}
	return files
}

// workspaceFileContent 工作区文件内容:打开文档以内存为准,否则读磁盘。
func (s *Server) workspaceFileContent(path string) []byte {
	if uri, ok := s.view.PathURI(path); ok {
		if doc := s.view.Doc(uri); doc != nil {
			return []byte(doc.Text)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}

// tokenFileOf 取 FileSet 中的第一个(也是唯一一个)文件对象。
func tokenFileOf(fset *token.FileSet) *token.File {
	var tf *token.File
	fset.Iterate(func(f *token.File) bool {
		tf = f
		return false
	})
	return tf
}

// collectWorkspaceSymbols 收集单个文件的符号并过滤 query。
func collectWorkspaceSymbols(f *ast.File, path string, src []byte, fset *token.FileSet, query string, syms *[]WorkspaceSymbol) {
	q := strings.ToLower(query)
	matched := func(name string) bool {
		return q == "" || strings.Contains(strings.ToLower(name), q)
	}
	file := tokenFileOf(fset)
	add := func(name, container string, kind int, pos, end token.Pos) {
		if !matched(name) || file == nil {
			return
		}
		*syms = append(*syms, WorkspaceSymbol{
			Name:          name,
			Kind:          kind,
			ContainerName: container,
			Location:      Location{URI: pathToURI(path), Range: toRange(src, file, pos, end)},
		})
	}
	fileName := filepath.Base(path)
	for _, spec := range f.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok || ts.Name == nil {
			continue
		}
		switch t := ts.Type.(type) {
		case *ast.DataType:
			add(ts.Name.Name, fileName, SymbolStruct, ts.Name.Pos(), ts.Name.End())
			if t.Fields != nil {
				for _, field := range t.Fields.List {
					if field.Name != nil {
						add(field.Name.Name, ts.Name.Name, SymbolField, field.Name.Pos(), field.Name.End())
					}
				}
			}
		case *ast.ServerType:
			add(ts.Name.Name, fileName, SymbolInterface, ts.Name.Pos(), ts.Name.End())
			for _, method := range t.Methods {
				if method.Name != nil {
					add(method.Name.Name, ts.Name.Name, SymbolMethod, method.Name.Pos(), method.Name.End())
				}
			}
		case *ast.EnumType:
			add(ts.Name.Name, fileName, SymbolEnum, ts.Name.Pos(), ts.Name.End())
			for _, item := range t.Items {
				if item.Name != nil {
					add(item.Name.Name, ts.Name.Name, SymbolEnumMember, item.Name.Pos(), item.Name.End())
				}
			}
		}
	}
}
