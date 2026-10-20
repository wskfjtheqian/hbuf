package build

import (
	"encoding/json"
	"hbuf/plugin/core/wire"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestFindPluginExplicitPath 显式插件路径。
func TestFindPluginExplicitPath(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "hbuf-gen-go")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	got, err := FindPlugin("go", &PluginConfig{Path: bin})
	if err != nil {
		t.Fatal(err)
	}
	if got != bin {
		t.Errorf("got %s, want %s", got, bin)
	}
}

// TestFindPluginNamePrefix 带 hbuf-gen-x= 前缀的显式路径。
func TestFindPluginNamePrefix(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "gen")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	got, err := FindPlugin("go", &PluginConfig{Path: "hbuf-gen-go=" + bin})
	if err != nil {
		t.Fatal(err)
	}
	if got != bin {
		t.Errorf("got %s, want %s", got, bin)
	}
	// 非 hbuf-gen- 前缀应报错
	if _, err := FindPlugin("go", &PluginConfig{Path: "foo-bar=" + bin}); err == nil {
		t.Error("expect invalid name error")
	}
	// 路径不存在应报错
	if _, err := FindPlugin("go", &PluginConfig{Path: "/no/such/plugin"}); err == nil {
		t.Error("expect missing binary error")
	}
}

// TestFindPluginsMultiple 逗号分隔多个插件一次全部返回。
func TestFindPluginsMultiple(t *testing.T) {
	dir := t.TempDir()
	bins := make([]string, 3)
	for i, name := range []string{"hbuf-gen-go", "hbuf-gen-goplus", "hbuf-gen-dart"} {
		bins[i] = filepath.Join(dir, name)
		if err := os.WriteFile(bins[i], []byte("#!/bin/sh\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	cfg := &PluginConfig{Path: "hbuf-gen-go=" + bins[0] + ",hbuf-gen-goplus=" + bins[1] + "," + bins[2]}
	got, err := FindPlugins("go", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != bins[0] || got[1] != bins[1] || got[2] != bins[2] {
		t.Errorf("got %v, want %v", got, bins)
	}
	// 多个插件时单插件接口应报错,避免静默丢失
	if _, err := FindPlugin("go", cfg); err == nil {
		t.Error("expect multiple plugins error")
	}
}

// TestFindPluginPathLookup 通过 PATH 查找 hbuf-gen-<typ>。
func TestFindPluginPathLookup(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "hbuf-gen-go")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	got, err := FindPlugin("go", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != bin {
		t.Errorf("got %s, want %s", got, bin)
	}
	// 未找到时返回空路径
	if got, err := FindPlugin("dart", nil); err != nil || got != "" {
		t.Errorf("expect empty for missing plugin, got %q, %v", got, err)
	}
}

// TestHelperProcess 通过子进程自调用模拟插件:回显固定响应。
func TestHelperProcess(t *testing.T) {
	if os.Getenv("HBUF_PLUGIN_HELPER") != "1" {
		if os.Getenv("HBUF_PLUGIN_HELPER") != "" {
			os.Exit(3)
		}
		t.Skip("helper process")
	}
	req := &wire.Request{}
	if err := json.NewDecoder(os.Stdin).Decode(req); err != nil {
		os.Exit(2)
	}
	resp := &wire.Response{Files: []*wire.GeneratedFile{
		{Name: "pkg/out.go", Content: "package pkg\n"},
	}}
	data, _ := json.Marshal(resp)
	_, _ = os.Stdout.Write(data)
	os.Exit(0)
}

// helperCmd 构造调用 helper 进程的命令。
func helperCmd() *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(), "HBUF_PLUGIN_HELPER=1")
	return cmd
}

// TestCallPlugin 完整协议调用:启动 helper 进程,传递请求,解析响应。
func TestCallPlugin(t *testing.T) {
	req := &wire.Request{Protocol: wire.ProtocolVersion, Type: "go"}
	resp, err := callPluginCmd(helperCmd(), "helper", req)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Files) != 1 || resp.Files[0].Name != "pkg/out.go" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

// TestCallPluginExitError 插件进程失败时携带 stderr 信息。
func TestCallPluginExitError(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(), "HBUF_PLUGIN_HELPER=2")
	if _, err := callPluginCmd(cmd, "helper", &wire.Request{}); err == nil {
		t.Error("expect plugin exit error")
	}
}

// TestWriteGeneratedFiles 落盘与目录逃逸防护。
func TestWriteGeneratedFiles(t *testing.T) {
	out := t.TempDir()
	resp := &wire.Response{Files: []*wire.GeneratedFile{
		{Name: "admin/base.data.go", Content: "package admin\n"},
		{Name: "nested/deep/file.go", Content: "package nested\n"},
	}}
	if err := WriteGeneratedFiles(out, resp); err != nil {
		t.Fatal(err)
	}
	for _, f := range resp.Files {
		data, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(f.Name)))
		if err != nil {
			t.Fatalf("read %s: %v", f.Name, err)
		}
		if string(data) != f.Content {
			t.Errorf("content mismatch for %s", f.Name)
		}
	}

	for _, name := range []string{"../escape.go", "sub/../../escape.go"} {
		if err := WriteGeneratedFiles(out, &wire.Response{Files: []*wire.GeneratedFile{{Name: name, Content: "x"}}}); err == nil {
			t.Errorf("expect escape rejection for %s", name)
		}
	}
}
