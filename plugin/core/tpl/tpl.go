// Package tpl 提供基于 text/template 的模板引擎。
//
// 模板查找顺序(自定义优先):
//  1. 外部自定义模板目录(插件参数 template=<目录>),同名文件覆盖默认模板;
//  2. 默认模板(通过 //go:embed 编译进插件二进制,兜底)。
//
// 引擎把所有 *.tmpl 解析到同一个命名空间,模板之间可以用 {{template "xxx.tmpl" .}}
// 相互调用,解析结果按名字缓存。
package tpl

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"sync"
	"text/template"
)

// Engine 模板引擎
type Engine struct {
	fs   fs.FS // 合并后的文件系统(自定义覆盖默认)
	fm   template.FuncMap
	once sync.Once
	tmpl *template.Template
	err  error
}

// New 创建引擎。
//   - defaultFS: 内置默认模板文件系统(通常来自 //go:embed);
//   - customFS:  外部自定义模板目录,可为 nil;
//   - fm:        模板函数库,可为 nil。
func New(defaultFS fs.FS, customFS fs.FS, fm template.FuncMap) *Engine {
	fsys := fs.FS(defaultFS)
	if customFS != nil {
		fsys = &overlayFS{base: defaultFS, overlay: customFS}
	}
	return &Engine{fs: fsys, fm: fm}
}

// Execute 执行指定模板并返回渲染结果。
func (e *Engine) Execute(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := e.ExecuteTo(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ExecuteTo 执行指定模板并把结果写入 w。
func (e *Engine) ExecuteTo(w io.Writer, name string, data interface{}) error {
	e.once.Do(e.parse)
	if e.err != nil {
		return e.err
	}
	if t := e.tmpl.Lookup(name); t != nil {
		return t.Execute(w, data)
	}
	return fmt.Errorf("template not found: %s", name)
}

// parse 惰性解析所有 *.tmpl,同名文件由自定义模板覆盖(overlayFS 已处理)。
func (e *Engine) parse() {
	names, err := fs.Glob(e.fs, "*.tmpl")
	if err != nil {
		e.err = err
		return
	}
	root := template.New("root").Funcs(e.fm)
	for _, name := range names {
		data, err := fs.ReadFile(e.fs, name)
		if err != nil {
			e.err = err
			return
		}
		if _, err := root.New(name).Parse(string(data)); err != nil {
			e.err = fmt.Errorf("parse template %s: %w", name, err)
			return
		}
	}
	e.tmpl = root
}

// overlayFS 合并两个文件系统:overlay 优先,缺失时回落到 base。
// 同时实现 Glob/ReadDir 以支持 fs.Glob 枚举合并后的文件列表(同名文件 overlay 优先)。
type overlayFS struct {
	base    fs.FS
	overlay fs.FS
}

func (o *overlayFS) Open(name string) (fs.File, error) {
	if f, err := o.overlay.Open(name); err == nil {
		return f, nil
	}
	return o.base.Open(name)
}

func (o *overlayFS) Glob(pattern string) ([]string, error) {
	return globMerge(o.overlay, o.base, pattern)
}

func (o *overlayFS) ReadDir(name string) ([]fs.DirEntry, error) {
	overlayEntries, err1 := fs.ReadDir(o.overlay, name)
	baseEntries, err2 := fs.ReadDir(o.base, name)
	if err1 != nil && err2 != nil {
		return nil, err2
	}
	seen := make(map[string]bool, len(overlayEntries)+len(baseEntries))
	out := make([]fs.DirEntry, 0, len(overlayEntries)+len(baseEntries))
	for _, e := range overlayEntries {
		seen[e.Name()] = true
		out = append(out, e)
	}
	for _, e := range baseEntries {
		if !seen[e.Name()] {
			out = append(out, e)
		}
	}
	return out, nil
}

// globMerge 合并两个文件系统的 Glob 结果,overlay 优先且去重。
func globMerge(overlay, base fs.FS, pattern string) ([]string, error) {
	overlayMatches, err := fs.Glob(overlay, pattern)
	if err != nil {
		return nil, err
	}
	baseMatches, err := fs.Glob(base, pattern)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(overlayMatches)+len(baseMatches))
	out := make([]string, 0, len(overlayMatches)+len(baseMatches))
	for _, m := range overlayMatches {
		seen[m] = true
		out = append(out, m)
	}
	for _, m := range baseMatches {
		if !seen[m] {
			out = append(out, m)
		}
	}
	return out, nil
}

// Names 返回引擎可见的模板名列表(用于测试与诊断)。
func (e *Engine) Names() ([]string, error) {
	e.once.Do(e.parse)
	if e.err != nil {
		return nil, e.err
	}
	var names []string
	for _, t := range e.tmpl.Templates() {
		if t.Name() != "root" {
			names = append(names, t.Name())
		}
	}
	return names, nil
}
