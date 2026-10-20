package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestLSPProcessStdio 端到端冒烟:以真实子进程(OS 级 stdin/stdout 管道)运行 Serve,
// 验证完整会话。通过环境变量 HBUF_LSP_HELPER=1 让测试二进制以语言服务器模式
// 运行(helper process 模式,与 gopls 测试一致),父进程负责构造帧并校验响应。
func TestLSPProcessStdio(t *testing.T) {
	if os.Getenv("HBUF_LSP_HELPER") == "1" {
		// 子进程:以语言服务器模式运行。此处绝不能调用 t 方法
		// (输出会写入 stdout 污染协议流),退出码即 Serve 返回值。
		os.Exit(Serve(os.Stdin, os.Stdout, Options{Name: "hbuf-lsp", Version: "test"}))
	}

	cmd := exec.Command(os.Args[0], "-test.run", "^TestLSPProcessStdio$")
	cmd.Env = append(os.Environ(), "HBUF_LSP_HELPER=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	const uri = "file:///tmp/student.hbuf"
	const src = "package go = \"admin\"\n" +
		"\n" +
		"data Base { int64 id = 1; }\n" +
		"\n" +
		"enum Status { player = 0; stop = 1; }\n" +
		"\n" +
		"data Student:Base = 2 {\n" +
		"    int32? no = 3;\n" +
		"    Status status = 4;\n" +
		"}\n"

	// 发送完整会话:initialize → didOpen → completion → shutdown → exit
	w := bufio.NewWriter(stdin)
	send := func(msg interface{}) {
		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.WriteString("Content-Length: " + strconv.Itoa(len(data)) + "\r\n\r\n"); err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	req := func(id int, method string, params interface{}) {
		send(map[string]interface{}{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	}
	notify := func(method string, params interface{}) {
		send(map[string]interface{}{"jsonrpc": "2.0", "method": method, "params": params})
	}

	req(1, "initialize", map[string]interface{}{"capabilities": map[string]interface{}{}})
	notify("initialized", map[string]interface{}{})
	notify("textDocument/didOpen", map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri, "languageId": "hbuf", "version": 1, "text": src,
		},
	})
	req(2, "textDocument/completion", map[string]interface{}{
		"textDocument": map[string]interface{}{"uri": uri},
		"position":     map[string]interface{}{"line": 5, "character": 0},
	})
	req(3, "textDocument/semanticTokens/full", map[string]interface{}{
		"textDocument": map[string]interface{}{"uri": uri},
	})
	req(4, "shutdown", nil)
	notify("exit", nil)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	stdin.Close()

	// 读取 5 条输出:initialize 响应、publishDiagnostics 通知、completion 响应、
	// semanticTokens 响应、shutdown 响应
	wire := NewWire(stdout, io.Discard)
	var got []*Message
	for i := 0; i < 5; i++ {
		msg, err := wire.Read()
		if err != nil {
			t.Fatalf("read message %d: %v (stderr: %s)", i, err, stderr)
		}
		got = append(got, msg)
	}

	// 1. initialize 响应
	if got[0].ID == nil || string(*got[0].ID) != "1" || !bytes.Contains(got[0].Result, []byte(`"capabilities"`)) {
		t.Errorf("initialize response mismatch: %+v", got[0])
	}
	// 2. publishDiagnostics 通知(文档无错误)
	if got[1].Method != "textDocument/publishDiagnostics" ||
		!bytes.Contains(got[1].Params, []byte(`"diagnostics":[]`)) {
		t.Errorf("publishDiagnostics mismatch: %+v", got[1])
	}
	// 3. completion 响应(顶层空行:关键字 + 已声明类型)
	if got[2].ID == nil || string(*got[2].ID) != "2" ||
		!bytes.Contains(got[2].Result, []byte(`"items":[`)) ||
		!bytes.Contains(got[2].Result, []byte(`"label":"Student"`)) {
		t.Errorf("completion response mismatch: %+v", got[2])
	}
	// 4. semanticTokens 响应(非空且 5 元组对齐)
	if got[3].ID == nil || string(*got[3].ID) != "3" ||
		!bytes.Contains(got[3].Result, []byte(`"data":[`)) {
		t.Errorf("semanticTokens response mismatch: %+v", got[3])
	}
	// 5. shutdown 响应
	if got[4].ID == nil || string(*got[4].ID) != "4" || !bytes.Equal(bytes.TrimSpace(got[4].Result), []byte("null")) {
		t.Errorf("shutdown response mismatch: %+v", got[4])
	}

	// 退出码必须为 0(shutdown 后 exit)
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("process exit error: %v (stderr: %s)", err, strings.TrimSpace(stderr.String()))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for process exit")
	}
}
