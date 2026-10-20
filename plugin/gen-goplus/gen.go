package goplus

import (
	"fmt"
	"go/format"
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"hbuf/plugin/gen-goplus/templates"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Generate hbuf-gen-goplus 插件入口:按请求生成 Go 的 database/verify/export/mq 代码。
// 默认模板编译进二进制,插件参数 template=<目录> 可覆盖同名模板。
func Generate(req *wire.Request) (*wire.Response, error) {
	if req.Type != "go" {
		return nil, fmt.Errorf("unsupported type: %s", req.Type)
	}
	index := NewTypeIndex(req.Files)

	engine, err := newEngine(req)
	if err != nil {
		return nil, err
	}

	resp := &wire.Response{}
	for _, file := range req.Files {
		if file == nil {
			continue
		}
		if !req.IsChanged(file.Path) {
			continue
		}
		pack := file.Packages["go"]
		if pack == "" {
			continue
		}
		name := baseName(file.Path)
		ctx := NewCtx(req, file, index, pack)
		fc, err := NewFileContext(ctx, file)
		if err != nil {
			return nil, err
		}
		// 各输出文件构建期登记的 import 集合(四类独立,避免互相污染)
		baseByTemplate := map[string]Imports{
			"database.tmpl": fc.databaseImports,
			"verify.tmpl":   fc.verifyImports,
			"export.tmpl":   fc.exportImports,
			"mq.tmpl":       fc.mqImports,
		}
		// 与进程内后端一致:输出到 <包名点分转目录>/<name>.<kind>.go
		packDir := filepath.Join(strings.Split(pack, ".")...)

		for _, tmplName := range []string{"database.tmpl", "verify.tmpl", "export.tmpl", "mq.tmpl"} {
			body, err := engine.Execute(tmplName, fc)
			if err != nil {
				return nil, fmt.Errorf("%s %s: %w", file.Path, tmplName, err)
			}
			if strings.TrimSpace(body) == "" {
				continue
			}
			suffix := strings.TrimSuffix(tmplName, ".tmpl")
			content := header(pack, baseByTemplate[tmplName]) + body
			// 语法校验兼格式化预检:输出保持原样(与进程内后端一致),
			// 仅用 format.Source 验证生成代码可被解析
			if _, err := format.Source([]byte(content)); err != nil {
				return nil, fmt.Errorf("%s %s format: %w", file.Path, tmplName, err)
			}
			// 与进程内后端一致:去掉末尾多余的换行
			content = strings.TrimSuffix(content, "\n")
			resp.Files = append(resp.Files, &wire.GeneratedFile{
				Name:    filepath.Join(packDir, name+"."+suffix+".go"),
				Content: content,
			})
		}
	}
	sort.Slice(resp.Files, func(i, j int) bool { return resp.Files[i].Name < resp.Files[j].Name })
	return resp, nil
}

// newEngine 构建模板引擎:默认模板 embed + 外部自定义模板覆盖。
func newEngine(req *wire.Request) (*tpl.Engine, error) {
	var customFS fs.FS
	if dir, ok := req.GetOption("template"); ok && dir != "" {
		info, err := os.Stat(dir)
		if err != nil {
			return nil, fmt.Errorf("template dir: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("template dir is not a directory: %s", dir)
		}
		customFS = os.DirFS(dir)
	}
	return tpl.New(templates.FS, customFS, tpl.Funcs()), nil
}
