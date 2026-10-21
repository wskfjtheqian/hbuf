package tsgen

import (
	"fmt"
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"hbuf/plugin/gen-ts/templates"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// defaultRuntime 运行时包默认 import 路径(可用插件参数 runtime 覆盖)
const defaultRuntime = "../src"

// Generate hbuf-gen-ts 插件入口:按请求生成 TypeScript 数据类/枚举/服务文件。
// 数据类含 descriptor 与 JSON 编解码(fromMap/toMap),可直接对接 hbuf_ts 运行时;
// 输出文件名与进程内 typescript 后端一致:<包名点分转目录>/<name>.<data|enum|server>.ts。
func Generate(req *wire.Request) (*wire.Response, error) {
	if req.Type != "ts" {
		return nil, fmt.Errorf("unsupported type: %s", req.Type)
	}
	index := NewTypeIndex(req.Files)

	engine, err := newEngine(req)
	if err != nil {
		return nil, err
	}

	runtime := defaultRuntime
	if v, ok := req.GetOption("runtime"); ok && v != "" {
		runtime = v
	}

	exported := exportedDescs(req.Files, index)

	resp := &wire.Response{}
	for _, file := range req.Files {
		if file == nil {
			continue
		}
		if !req.IsChanged(file.Path) {
			continue
		}
		pack := file.Packages["ts"]
		// 与进程内 typescript 后端一致:无 ts 包注解时输出到根目录,
		// 有包注解时输出到 <包名点分转目录>/<name>.<data|enum|server>.ts
		packDir := ""
		if pack != "" {
			packDir = filepath.Join(strings.Split(pack, ".")...)
		}
		name := baseName(file.Path)

		// data 文件
		dctx := NewCtx(req, file, index, pack, runtime)
		body, err := engine.Execute("data.tmpl", NewFileContext(dctx, file, exported))
		if err != nil {
			return nil, fmt.Errorf("%s data.tmpl: %w", file.Path, err)
		}
		if strings.TrimSpace(body) != "" {
			resp.Files = append(resp.Files, &wire.GeneratedFile{
				Name:    filepath.Join(packDir, name+".data.ts"),
				Content: dctx.header() + body,
			})
		}

		// enum 文件
		ectx := NewCtx(req, file, index, pack, runtime)
		body, err = engine.Execute("enum.tmpl", NewEnumFileContext(ectx, file))
		if err != nil {
			return nil, fmt.Errorf("%s enum.tmpl: %w", file.Path, err)
		}
		if strings.TrimSpace(body) != "" {
			resp.Files = append(resp.Files, &wire.GeneratedFile{
				Name:    filepath.Join(packDir, name+".enum.ts"),
				Content: ectx.header() + body,
			})
		}

		// server 文件
		sctx := NewCtx(req, file, index, pack, runtime)
		body, err = engine.Execute("server.tmpl", NewServerFileContext(sctx, file))
		if err != nil {
			return nil, fmt.Errorf("%s server.tmpl: %w", file.Path, err)
		}
		if strings.TrimSpace(body) != "" {
			resp.Files = append(resp.Files, &wire.GeneratedFile{
				Name:    filepath.Join(packDir, name+".server.ts"),
				Content: sctx.header() + body,
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

// header 生成文件头:运行时 import 块 + 跨文件 import 块。
func (c *Ctx) header() string {
	var sb strings.Builder
	symbols := make([]string, 0, len(c.Imports))
	for s := range c.Imports {
		symbols = append(symbols, s)
	}
	sort.Strings(symbols)
	if len(symbols) > 0 {
		sb.WriteString("import {\n")
		for _, s := range symbols {
			sb.WriteString("    " + s + ",\n")
		}
		sb.WriteString("} from \"" + c.Runtime + "\"\n")
	}
	paths := make([]string, 0, len(c.CrossRef))
	for p := range c.CrossRef {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		syms := make([]string, 0, len(c.CrossRef[p]))
		for s := range c.CrossRef[p] {
			syms = append(syms, s)
		}
		sort.Strings(syms)
		sb.WriteString("import {" + strings.Join(syms, ", ") + "} from \"" + p + "\"\n")
	}
	if sb.Len() == 0 {
		return ""
	}
	return sb.String() + "\n"
}
