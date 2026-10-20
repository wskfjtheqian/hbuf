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
	Path    string   // 显式插件二进制路径(优先于 PATH 查找)
	Options []string // 插件参数,如 template=<目录>
}

// FindPlugin 查找类型对应的插件二进制。
// 查找顺序:显式路径(可选,带 hbuf-gen-<typ>= 前缀则剥离)→ PATH 中的 hbuf-gen-<typ>。
// 未找到时返回空路径(调用方回退到进程内后端)。
func FindPlugin(typ string, cfg *PluginConfig) (string, error) {
	if cfg != nil && cfg.Path != "" {
		path := cfg.Path
		if name, rest, ok := strings.Cut(path, "="); ok {
			if name != "hbuf-gen-"+typ {
				return "", fmt.Errorf("plugin name mismatch: %s, expect hbuf-gen-%s", name, typ)
			}
			path = rest
		}
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("plugin binary: %w", err)
		}
		return path, nil
	}
	path, err := exec.LookPath("hbuf-gen-" + typ)
	if err != nil {
		return "", nil
	}
	return path, nil
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
