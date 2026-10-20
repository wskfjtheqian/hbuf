package gogen

import (
	"fmt"
	"go/format"
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"hbuf/plugin/gen-go/templates"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// headerImports 各输出文件的固定 import 路径。
var headerImports = map[string][]string{
	"data.tmpl": {
		"reflect",
		"unsafe",
		"github.com/wskfjtheqian/hbuf_golang/pkg/hbuf",
	},
	"enum.tmpl": {
		"context",
		"strconv",
		"strings",
		"github.com/wskfjtheqian/hbuf_golang/pkg/herror",
	},
	"server.tmpl": {
		"context",
		"github.com/wskfjtheqian/hbuf_golang/pkg/hrpc",
		"github.com/wskfjtheqian/hbuf_golang/pkg/hservice",
	},
}

// Generate hbuf-gen-go 插件入口:按请求生成 Go 代码文件。
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
		fc := NewFileContext(ctx, file)
		// 各输出文件构建期登记的 import 集合(data/enum/server 独立,避免互相污染)
		baseByTemplate := map[string]map[string]bool{
			"data.tmpl":   fc.dataImports,
			"enum.tmpl":   fc.enumImports,
			"server.tmpl": fc.serverImports,
		}
		// 与进程内后端一致:输出到 <包名点分转目录>/<name>.<kind>.go
		packDir := filepath.Join(strings.Split(pack, ".")...)

		for _, tmplName := range []string{"data.tmpl", "enum.tmpl", "server.tmpl"} {
			// 每个输出文件从对应类别基线 import 开始
			base := baseByTemplate[tmplName]
			ctx.Imports = make(map[string]bool, len(base))
			for k := range base {
				ctx.Imports[k] = true
			}
			body, err := engine.Execute(tmplName, fc)
			if err != nil {
				return nil, fmt.Errorf("%s %s: %w", file.Path, tmplName, err)
			}
			if strings.TrimSpace(body) == "" {
				continue
			}
			// 按文件内容动态登记条件性 import,避免未使用 import
			if tmplName == "server.tmpl" {
				for _, s := range fc.Servers {
					for _, m := range s.Methods {
						if m.StreamIn || m.StreamOut {
							ctx.Import("io")
						}
						if m.Bind == nil {
							ctx.Import("github.com/wskfjtheqian/hbuf_golang/pkg/herror")
						}
						if !m.StreamIn {
							ctx.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf")
						}
					}
				}
			}
			suffix := strings.TrimSuffix(tmplName, ".tmpl")
			content := header(ctx, tmplName) + body
			// 自动格式化,保证输出对齐与标准风格
			formatted, err := format.Source([]byte(content))
			if err != nil {
				return nil, fmt.Errorf("%s %s format: %w", file.Path, tmplName, err)
			}
			resp.Files = append(resp.Files, &wire.GeneratedFile{
				Name:    filepath.Join(packDir, name+"."+suffix+".go"),
				Content: string(formatted),
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

// header 生成文件头:package 声明 + import 块。
func header(ctx *Ctx, tmplName string) string {
	for _, path := range headerImports[tmplName] {
		ctx.Import(path)
	}
	var sb strings.Builder
	sb.WriteString("package " + ctx.Pack + "\n\n")
	if len(ctx.Imports) > 0 {
		keys := make([]string, 0, len(ctx.Imports))
		for key := range ctx.Imports {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		sb.WriteString("import (\n")
		for _, key := range keys {
			sb.WriteString("\t\"" + key + "\"\n")
		}
		sb.WriteString(")\n\n")
	}
	return sb.String()
}
