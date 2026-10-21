package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hbuf/plugin/core/wire"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PluginConfig 插件调度配置。
type PluginConfig struct {
	Path    string   // 显式插件二进制路径,逗号分隔多个(优先于 PATH 查找)
	Options []string // 插件参数,如 template=<目录>
}

// FindPlugins 查找类型对应的插件二进制列表。
// 显式路径(逗号分隔多个,条目可选带 hbuf-gen-<name>= 前缀则剥离)
// 允许同时挂载多个插件(如 hbuf-gen-go 与 hbuf-gen-goplus 一次运行全部生成);
// 条目中 k=v 形式且不以 hbuf-gen- 开头者视为插件参数(追加到 cfg.Options),
// 如 "hbuf-gen-ts=bin/hbuf-gen-ts,runtime=hbuf_ts"。
// 未显式指定时通过 PATH 查找 hbuf-gen-<typ>。未找到时返回空列表(调用方回退到进程内后端)。
func FindPlugins(typ string, cfg *PluginConfig) ([]string, error) {
	if cfg != nil && cfg.Path != "" {
		var bins []string
		for _, item := range strings.Split(cfg.Path, ",") {
			item = strings.TrimSpace(item)
			if "" == item {
				continue
			}
			// 插件参数:k=v 形式、不以 hbuf-gen- 前缀且值不含路径分隔符,
			// 如 "hbuf-gen-ts=bin/hbuf-gen-ts,runtime=hbuf_ts"。
			// 值含路径分隔符的 k=v 视为插件名不合法(落入下方报错),而非参数。
			if eq := strings.IndexByte(item, '='); eq >= 0 &&
				!strings.HasPrefix(item, "hbuf-gen-") &&
				!strings.ContainsAny(item[eq+1:], `/\`) {
				cfg.Options = append(cfg.Options, item)
				continue
			}
			path := item
			if name, rest, ok := strings.Cut(item, "="); ok {
				if !strings.HasPrefix(name, "hbuf-gen-") {
					return nil, fmt.Errorf("plugin name invalid: %s, expect hbuf-gen-<name>", name)
				}
				path = rest
			}
			if _, err := os.Stat(path); err != nil {
				return nil, fmt.Errorf("plugin binary: %w", err)
			}
			bins = append(bins, path)
		}
		return bins, nil
	}
	path, err := exec.LookPath("hbuf-gen-" + typ)
	if err != nil {
		return nil, nil
	}
	return []string{path}, nil
}

// FindPlugin 查找类型对应的单个插件二进制(多插件场景请使用 FindPlugins)。
// 显式指定多个插件时返回错误,避免静默丢失。
func FindPlugin(typ string, cfg *PluginConfig) (string, error) {
	bins, err := FindPlugins(typ, cfg)
	if err != nil {
		return "", err
	}
	if len(bins) > 1 {
		return "", fmt.Errorf("multiple plugins specified, use FindPlugins")
	}
	if len(bins) == 0 {
		return "", nil
	}
	return bins[0], nil
}

// CallPlugin 启动插件进程,通过 JSON over stdio 传递编译请求并解析响应。
func CallPlugin(bin string, req *wire.Request) (*wire.Response, error) {
	return callPluginCmd(exec.Command(bin), bin, req)
}

// callPluginCmd 执行插件进程(命令可注入,便于测试)。
func callPluginCmd(cmd *exec.Cmd, bin string, req *wire.Request) (*wire.Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	cmd.Stdin = bytes.NewReader(data)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("plugin %s: %w: %s", bin, err, strings.TrimSpace(stderr.String()))
	}
	resp := &wire.Response{}
	if err := json.Unmarshal(stdout.Bytes(), resp); err != nil {
		return nil, fmt.Errorf("plugin %s response: %w", bin, err)
	}
	return resp, nil
}

// WriteGeneratedFiles 把插件返回的生成文件统一落盘到输出目录。
// 文件名是相对路径,拒绝目录逃逸。
func WriteGeneratedFiles(out string, resp *wire.Response) error {
	for _, f := range resp.Files {
		if f == nil {
			continue
		}
		rel := filepath.Clean(filepath.FromSlash(f.Name))
		if rel == "." || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("invalid generated file path: %s", f.Name)
		}
		path := filepath.Join(out, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(f.Content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}
