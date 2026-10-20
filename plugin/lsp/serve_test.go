package lsp

import (
	"encoding/json"
	"io"
	"path/filepath"
	"testing"
)

// testClient 协议级测试客户端:与 Serve 主循环通过 io.Pipe 交互。
type testClient struct {
	t    *testing.T
	wire *Wire
	done chan int // Serve 退出码
}

func newTestClient(t *testing.T) *testClient {
	t.Helper()
	serverR, clientW := io.Pipe() // 客户端写 → 服务器读
	clientR, serverW := io.Pipe() // 服务器写 → 客户端读
	done := make(chan int, 1)
	go func() {
		done <- Serve(serverR, serverW, Options{Name: "hbuf-lsp", Version: "test"})
	}()
	return &testClient{
		t:    t,
		wire: NewWire(clientR, clientW),
		done: done,
	}
}

// send 发送请求。
func (c *testClient) send(id, method string, params interface{}) {
	c.t.Helper()
	data, err := json.Marshal(params)
	if err != nil {
		c.t.Fatal(err)
	}
	rawID := json.RawMessage(json.RawMessage(id))
	msg := &Message{JSONRPC: "2.0", ID: &rawID, Method: method, Params: data}
	if err := c.wire.Write(msg); err != nil {
		c.t.Fatal(err)
	}
}

// notify 发送通知。
func (c *testClient) notify(method string, params interface{}) {
	c.t.Helper()
	data, err := json.Marshal(params)
	if err != nil {
		c.t.Fatal(err)
	}
	msg := &Message{JSONRPC: "2.0", Method: method, Params: data}
	if err := c.wire.Write(msg); err != nil {
		c.t.Fatal(err)
	}
}

// recv 读取下一条消息。
func (c *testClient) recv() *Message {
	c.t.Helper()
	msg, err := c.wire.Read()
	if err != nil {
		c.t.Fatal(err)
	}
	return msg
}

// TestServeSession 完整协议会话:initialize → didOpen(诊断) → completion → shutdown → exit。
func TestServeSession(t *testing.T) {
	c := newTestClient(t)

	// 1. initialize 请求
	c.send("1", "initialize", map[string]interface{}{})
	resp := c.recv()
	if !resp.IsResponse() || string(*resp.ID) != "1" {
		t.Fatalf("expect initialize response, got %+v", resp)
	}
	var initRes InitializeResult
	if err := json.Unmarshal(resp.Result, &initRes); err != nil {
		t.Fatal(err)
	}
	if !initRes.Capabilities.DefinitionProvider || !initRes.Capabilities.HoverProvider {
		t.Errorf("missing capabilities: %+v", initRes.Capabilities)
	}
	if initRes.ServerInfo == nil || initRes.ServerInfo.Name != "hbuf-lsp" {
		t.Errorf("server info: %+v", initRes.ServerInfo)
	}
	c.notify("initialized", struct{}{})

	// 2. didOpen → publishDiagnostics 通知
	dir := t.TempDir()
	uri := pathToURI(filepath.Join(dir, "main.hbuf"))
	src := "package go = \"admin\";\n\ndata X {\n    int64 id = 1;\n};\n"
	c.notify("textDocument/didOpen", DidOpenParams{
		TextDocument: TextDocumentItem{URI: uri, LanguageID: "hbuf", Version: 1, Text: src},
	})
	notif := c.recv()
	if notif.Method != "textDocument/publishDiagnostics" {
		t.Fatalf("expect publishDiagnostics, got %+v", notif)
	}
	var diagParams PublishDiagnosticsParams
	if err := json.Unmarshal(notif.Params, &diagParams); err != nil {
		t.Fatal(err)
	}
	if diagParams.URI != uri || len(diagParams.Diagnostics) != 0 {
		t.Errorf("diagnostics: uri=%s n=%d", diagParams.URI, len(diagParams.Diagnostics))
	}

	// 3. completion 请求(顶层位置:关键字 + 类型)
	c.send("2", "textDocument/completion", CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 0, Character: 0},
	})
	resp = c.recv()
	if !resp.IsResponse() || string(*resp.ID) != "2" {
		t.Fatalf("expect completion response, got %+v", resp)
	}
	var list CompletionList
	if err := json.Unmarshal(resp.Result, &list); err != nil {
		t.Fatal(err)
	}
	labels := map[string]bool{}
	for _, item := range list.Items {
		labels[item.Label] = true
	}
	for _, want := range []string{"data", "enum", "X"} {
		if !labels[want] {
			t.Errorf("completion missing %q: %v", want, labels)
		}
	}

	// 4. didChange 后诊断更新(语法错误)
	c.notify("textDocument/didChange", DidChangeParams{
		TextDocument: VersionedTextDocumentIdentifier{URI: uri, Version: 2},
		ContentChanges: []TextDocumentContentChangeEvent{
			{Text: "data X {\n"},
		},
	})
	notif = c.recv()
	if notif.Method != "textDocument/publishDiagnostics" {
		t.Fatalf("expect publishDiagnostics after change, got %+v", notif)
	}
	if err := json.Unmarshal(notif.Params, &diagParams); err != nil {
		t.Fatal(err)
	}
	if len(diagParams.Diagnostics) == 0 {
		t.Error("expect syntax diagnostics after change")
	}

	// 5. shutdown + exit(正常退出码 0)
	c.send("3", "shutdown", nil)
	resp = c.recv()
	if !resp.IsResponse() || string(*resp.ID) != "3" {
		t.Fatalf("expect shutdown response, got %+v", resp)
	}
	c.notify("exit", nil)
	if code := <-c.done; code != 0 {
		t.Fatalf("expect exit code 0, got %d", code)
	}
}

// TestServeExitWithoutShutdown 未 shutdown 直接 exit 应返回 1。
func TestServeExitWithoutShutdown(t *testing.T) {
	c := newTestClient(t)
	c.notify("exit", nil)
	if code := <-c.done; code != 1 {
		t.Fatalf("expect exit code 1 without shutdown, got %d", code)
	}
}

// TestServeUnknownMethod 未知方法返回 MethodNotFound。
func TestServeUnknownMethod(t *testing.T) {
	c := newTestClient(t)
	c.send("1", "textDocument/unknown", nil)
	resp := c.recv()
	if resp.Error == nil || resp.Error.Code != ErrMethodNotFound {
		t.Fatalf("expect MethodNotFound, got %+v", resp)
	}
	c.notify("exit", nil)
	<-c.done
}
