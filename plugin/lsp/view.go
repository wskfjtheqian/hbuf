package lsp

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/parser"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Document 一个打开的文档。
type Document struct {
	URI        string
	LanguageID string
	Version    int
	Text       string
	Path       string // 磁盘路径(file URI 转换)
}

// parsedFile 一次完整解析的结果:根文件及其 import 依赖图。
type parsedFile struct {
	fset *token.FileSet
	pkg  *ast.Package
	srcs map[string][]byte            // path → 文件内容
	errs map[string]scanner.ErrorList // path → 语法错误
}

// View 文档视图:打开文档集合 + import 依赖解析。
type View struct {
	mu     sync.Mutex
	docs   map[string]*Document // uri → doc
	byPath map[string]string    // 磁盘路径 → uri
	roots  []string             // 工作区根目录(initialize 的 workspaceFolders)
}

func NewView() *View {
	return &View{
		docs:   map[string]*Document{},
		byPath: map[string]string{},
	}
}

// Open 打开/更新文档并返回该文档对象。
func (v *View) Open(uri, languageID string, version int, text string) *Document {
	v.mu.Lock()
	defer v.mu.Unlock()
	doc := &Document{
		URI:        uri,
		LanguageID: languageID,
		Version:    version,
		Text:       text,
		Path:       uriToPath(uri),
	}
	v.docs[uri] = doc
	v.byPath[doc.Path] = uri
	return doc
}

// Change 更新文档内容。
func (v *View) Change(uri string, version int, text string) (*Document, bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	doc, ok := v.docs[uri]
	if !ok {
		return nil, false
	}
	doc.Version = version
	doc.Text = text
	return doc, true
}

// Close 关闭文档。
func (v *View) Close(uri string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	doc, ok := v.docs[uri]
	if !ok {
		return
	}
	delete(v.docs, uri)
	if v.byPath[doc.Path] == uri {
		delete(v.byPath, doc.Path)
	}
}

// Doc 按 uri 取文档。
func (v *View) Doc(uri string) *Document {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.docs[uri]
}

// Docs 返回所有打开文档(快照)。
func (v *View) Docs() []*Document {
	v.mu.Lock()
	defer v.mu.Unlock()
	docs := make([]*Document, 0, len(v.docs))
	for _, d := range v.docs {
		docs = append(docs, d)
	}
	return docs
}

// SetWorkspaceRoots 设置工作区根目录(来自 initialize 的 workspaceFolders/rootUri)。
func (v *View) SetWorkspaceRoots(roots []string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.roots = append([]string(nil), roots...)
}

// Roots 返回工作区根目录快照。
func (v *View) Roots() []string {
	v.mu.Lock()
	defer v.mu.Unlock()
	return append([]string(nil), v.roots...)
}

// PathURI 按磁盘路径查打开文档的 uri(未打开返回 false)。
func (v *View) PathURI(path string) (string, bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	uri, ok := v.byPath[path]
	return uri, ok
}

// parse 解析文档及其 import 依赖图。
// 打开文档内容以内存为准(支持未保存文档),依赖文件从磁盘读取。
func (v *View) parse(doc *Document) *parsedFile {
	v.mu.Lock()
	docs := make(map[string]*Document, len(v.docs))
	for k, d := range v.docs {
		docs[k] = d
	}
	byPath := make(map[string]string, len(v.byPath))
	for k, d := range v.byPath {
		byPath[k] = d
	}
	v.mu.Unlock()

	pf := &parsedFile{
		fset: token.NewFileSet(),
		pkg:  ast.NewPackage(),
		srcs: map[string][]byte{},
		errs: map[string]scanner.ErrorList{},
	}
	v.loadFile(pf, docs, byPath, doc.Path, map[string]bool{})
	return pf
}

// loadFile 递归加载文件及其 import(仿 build 的 parser.ParseDir,内存覆盖优先)。
func (v *View) loadFile(pf *parsedFile, docs map[string]*Document, byPath map[string]string, path string, visited map[string]bool) {
	path = filepath.Clean(path)
	if visited[path] {
		return
	}
	visited[path] = true

	var src []byte
	if uri, ok := byPath[path]; ok {
		src = []byte(docs[uri].Text) // 未保存文档:内存覆盖
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return // 依赖文件不存在:由诊断报告(import 未解析)
		}
		src = data
	}
	pf.srcs[path] = src

	f, err := parser.ParseFile(pf.fset, path, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		if el, ok := err.(scanner.ErrorList); ok {
			pf.errs[path] = append(pf.errs[path], el...)
		}
	}
	if f == nil {
		f = &ast.File{Scope: ast.NewScope(nil)}
	}
	f.Path = path
	pf.pkg.Files[path] = f

	for _, spec := range f.Imports {
		imp, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		if !filepath.IsAbs(imp) {
			imp = filepath.Join(filepath.Dir(path), imp)
		}
		imp = filepath.Clean(imp)
		// 与 ParseDir 行为一致:import 路径归一为绝对路径,供校验跨文件解析
		spec.Path.Value = imp
		v.loadFile(pf, docs, byPath, imp, visited)
	}
}

// uriToPath 将 file:// URI 转为磁盘路径。
func uriToPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return strings.TrimPrefix(uri, "file://")
	}
	return filepath.FromSlash(u.Path)
}

// pathToURI 将磁盘路径转为 file:// URI。
func pathToURI(path string) string {
	return "file://" + filepath.ToSlash(path)
}
