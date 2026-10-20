package lsp

import (
	"bytes"
	"encoding/json"
	"hbuf/pkg/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

const lspDepSrc = `package go = "admin";

data Base {
    int64 id = 1;

    string name = 2;
};

data Message {
    string message = 1;
};
`

const lspSrc = `package go = "admin";

import "base.hbuf";

//学生信息
data Student : Base = 2 {
    int32? no = 3;

    Status status = 4;
};

enum Status {
    player = 0;

    stop = 1;
};

server BaseServer {
    Student GetStudent(Student req) = 1;

    void SendMessage(Message message) = 2;
};
`

// newTestServer 构造带文档的服务器(输出丢弃)。
func newTestServer(t *testing.T, src string) (*Server, *Document, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "base.hbuf"), []byte(lspDepSrc), 0644); err != nil {
		t.Fatal(err)
	}
	s := &Server{
		wire: NewWire(bytes.NewReader(nil), io.Discard),
		view: NewView(),
	}
	// 根文件用独立文件名,避免 src 的 import 指向自身
	path := filepath.Join(dir, "student.hbuf")
	doc := s.view.Open(pathToURI(path), "hbuf", 1, src)
	return s, doc, dir
}

// posAt 定位子串起始位置的 LSP Position。
func posAt(t *testing.T, src string, substr string) Position {
	t.Helper()
	i := strings.Index(src, substr)
	if i < 0 {
		t.Fatalf("substr %q not found", substr)
	}
	line := 0
	lineStart := 0
	for j := 0; j < i; j++ {
		if src[j] == '\n' {
			line++
			lineStart = j + 1
		}
	}
	return Position{Line: line, Character: utf16Len([]byte(src[lineStart:i]))}
}

func TestToOffsetRoundTrip(t *testing.T) {
	src := "// 学生信息\n数据 data X {\n    string 名字 = 1;\n};\n"
	fset := token.NewFileSet()
	file := fset.AddFile("/a/b.hbuf", -1, len(src))
	file.SetLinesForContent([]byte(src))

	// 每个字符边界的 Pos 转 Position 再转回 offset 应一致
	// (LSP 客户端给出的位置始终在字符边界上,字符中间字节无需往返)
	for off := 0; off < len(src); off++ {
		if !utf8.RuneStart(src[off]) {
			continue
		}
		pos := toPosition([]byte(src), file, file.Pos(off))
		back := toOffset([]byte(src), pos)
		if back != off {
			t.Fatalf("offset %d round-trip to %d (pos %+v)", off, back, pos)
		}
	}
	// UTF-16 列:中文字符按 1 个 code unit,BMP 外字符按 2 个
	src2 := "a😀b\n"
	fset2 := token.NewFileSet()
	file2 := fset2.AddFile("/x.hbuf", -1, len(src2))
	file2.SetLinesForContent([]byte(src2))
	p := toPosition([]byte(src2), file2, file2.Pos(len(src2)))
	if p.Line != 1 || p.Character != 0 { // EOF 映射到末行行首
		t.Fatalf("emoji line pos: %+v", p)
	}
	// 同一行内 emoji 占 2 列
	pos3 := toPosition([]byte(src2), file2, file2.Pos(5)) // 'b' 前
	if pos3.Character != 3 {                              // 'a'=1 + emoji=2
		t.Fatalf("emoji col: %+v", pos3)
	}
}

func TestWireRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	w := NewWire(bytes.NewReader(nil), &buf)
	id := json.RawMessage(jsonRaw(`7`))
	msg := &Message{JSONRPC: "2.0", ID: &id, Method: "textDocument/hover", Params: jsonRaw(`{"a":1}`)}
	if err := w.Write(msg); err != nil {
		t.Fatal(err)
	}
	r := NewWire(&buf, io.Discard)
	got, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if got.Method != msg.Method || got.ID == nil || string(*got.ID) != "7" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	// 多消息连续读
	if err := w.Write(msg); err != nil {
		t.Fatal(err)
	}
	if err := w.Write(msg); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Read(); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Read(); err != nil {
		t.Fatal(err)
	}
}

func jsonRaw(s string) (raw []byte) { return []byte(s) }

func TestCompletion(t *testing.T) {
	s, doc, _ := newTestServer(t, lspSrc)

	// 顶层:关键字
	res := s.completion(doc, posAt(t, lspSrc, "//学生信息"))
	labels := map[string]bool{}
	for _, item := range res.Items {
		labels[item.Label] = true
	}
	for _, want := range []string{"data", "enum", "server", "package", "import", "int64", "Student", "Base", "Message"} {
		if !labels[want] {
			t.Errorf("top-level completion missing %q (got %d items)", want, len(res.Items))
		}
	}

	// data 块内字段行:已有字段名
	res = s.completion(doc, posAt(t, lspSrc, "no = 3;"))
	labels = map[string]bool{}
	for _, item := range res.Items {
		labels[item.Label] = true
	}
	if !labels["no"] || !labels["status"] {
		t.Errorf("field completion missing members: %v", labels)
	}

	// server 块内方法行:已有方法名
	res = s.completion(doc, posAt(t, lspSrc, "GetStudent(Student req) = 1;"))
	labels = map[string]bool{}
	for _, item := range res.Items {
		labels[item.Label] = true
	}
	if !labels["GetStudent"] || !labels["SendMessage"] {
		t.Errorf("method completion missing members: %v", labels)
	}

	// tag 上下文(光标在 "[v" 之后)
	tagSrc := "data X {\n    [v\n    string name = 1;\n};\n"
	doc2 := s.view.Open(pathToURI(filepath.Join(t.TempDir(), "tag.hbuf")), "hbuf", 1, tagSrc)
	res = s.completion(doc2, posAt(t, tagSrc, "v\n"))
	found := false
	for _, item := range res.Items {
		if item.Label == "verify" {
			found = true
		}
	}
	if !found {
		t.Errorf("tag completion missing verify: %+v", res.Items)
	}

	// package 语言名(光标在 "g" 之后)
	pkgSrc := "package g\n"
	doc3 := s.view.Open(pathToURI(filepath.Join(t.TempDir(), "pkg.hbuf")), "hbuf", 1, pkgSrc)
	res = s.completion(doc3, posAt(t, pkgSrc, "\n"))
	for _, item := range res.Items {
		if item.Label == "go" {
			return
		}
	}
	t.Errorf("package completion missing go: %+v", res.Items)
}

func TestDefinition(t *testing.T) {
	s, doc, _ := newTestServer(t, lspSrc)

	// 类型定义:字段类型 Status → enum Status 声明
	locs := s.definition(doc, posAt(t, lspSrc, "Status status = 4;"))
	if len(locs) != 1 {
		t.Fatalf("expect 1 location, got %d", len(locs))
	}
	if locs[0].URI != doc.URI {
		t.Errorf("unexpected uri: %s", locs[0].URI)
	}
	// 位置应指向 enum Status 的名字
	if locs[0].Range.Start.Line != 11 { // "enum Status {" 所在行
		t.Errorf("expect definition at line 11, got %+v", locs[0].Range)
	}

	// 跨文件:extends Base → 依赖文件 data Base
	locs = s.definition(doc, posAt(t, lspSrc, "Base = 2 {"))
	if len(locs) != 1 {
		t.Fatalf("cross-file: expect 1 location, got %d", len(locs))
	}
	if !strings.HasSuffix(locs[0].URI, "/base.hbuf") {
		t.Errorf("expect definition in base.hbuf, got %s", locs[0].URI)
	}

	// 字段自身名 → 字段定义
	locs = s.definition(doc, posAt(t, lspSrc, "no = 3;"))
	if len(locs) != 1 {
		t.Fatalf("field: expect 1 location, got %d", len(locs))
	}
	if locs[0].Range.Start.Line != 6 {
		t.Errorf("expect field definition at line 6, got %+v", locs[0].Range)
	}

	// 无定义处:标识符不存在 → nil
	locs = s.definition(doc, posAt(t, lspSrc, "//学生信息"))
	if locs != nil {
		t.Errorf("expect nil for comment position, got %+v", locs)
	}
}

func TestReferences(t *testing.T) {
	s, doc, _ := newTestServer(t, lspSrc)

	// Status 类型:定义 1 处 + 字段引用 1 处
	locs := s.references(doc, posAt(t, lspSrc, "Status status = 4;"), false)
	if len(locs) != 1 {
		t.Fatalf("expect 1 reference (exclude declaration), got %d: %+v", len(locs), locs)
	}
	if locs[0].Range.Start.Line != 8 {
		t.Errorf("expect reference at line 8, got %+v", locs[0].Range)
	}

	// 包含声明:2 处
	locs = s.references(doc, posAt(t, lspSrc, "Status status = 4;"), true)
	if len(locs) != 2 {
		t.Fatalf("expect 2 (include declaration), got %d: %+v", len(locs), locs)
	}
}

func TestHover(t *testing.T) {
	s, doc, _ := newTestServer(t, lspSrc)

	// 类型悬停
	h := s.hover(doc, posAt(t, lspSrc, "Status status = 4;"))
	if h == nil {
		t.Fatal("expect hover for type")
	}
	if !strings.Contains(h.Contents.Value, "enum Status") {
		t.Errorf("hover missing type info: %s", h.Contents.Value)
	}

	// 字段悬停(含 Doc)
	h = s.hover(doc, posAt(t, lspSrc, "no = 3;"))
	if h == nil {
		t.Fatal("expect hover for field")
	}
	if !strings.Contains(h.Contents.Value, "no") {
		t.Errorf("hover missing field info: %s", h.Contents.Value)
	}

	// 方法悬停
	h = s.hover(doc, posAt(t, lspSrc, "GetStudent(Student req) = 1;"))
	if h == nil {
		t.Fatal("expect hover for method")
	}
	if !strings.Contains(h.Contents.Value, "GetStudent") {
		t.Errorf("hover missing method info: %s", h.Contents.Value)
	}

	// 基础类型悬停
	h = s.hover(doc, posAt(t, lspSrc, "int32? no = 3;"))
	if h == nil || !strings.Contains(h.Contents.Value, "基础类型") {
		t.Errorf("hover missing base type info: %+v", h)
	}
}

func TestDocumentSymbols(t *testing.T) {
	s, doc, _ := newTestServer(t, lspSrc)
	syms := s.documentSymbols(doc)
	if len(syms) != 3 {
		t.Fatalf("expect 3 symbols, got %d", len(syms))
	}
	byName := map[string]DocumentSymbol{}
	for _, sym := range syms {
		byName[sym.Name] = sym
	}
	if byName["Student"].Kind != SymbolStruct {
		t.Errorf("Student kind: %d", byName["Student"].Kind)
	}
	if len(byName["Student"].Children) != 2 {
		t.Errorf("Student children: %d", len(byName["Student"].Children))
	}
	if byName["Status"].Kind != SymbolEnum || len(byName["Status"].Children) != 2 {
		t.Errorf("Status symbol wrong: %+v", byName["Status"])
	}
	if byName["BaseServer"].Kind != SymbolInterface || len(byName["BaseServer"].Children) != 2 {
		t.Errorf("BaseServer symbol wrong: %+v", byName["BaseServer"])
	}
}

func TestFormatting(t *testing.T) {
	messy := "package go=\"admin\";\nimport \"base.hbuf\";\n\ndata Student : Base=2{\nint32? no=3;\n};\n"
	s, doc, _ := newTestServer(t, messy)
	edits, err := s.formatting(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(edits) != 1 {
		t.Fatalf("expect 1 edit, got %d", len(edits))
	}
	out := edits[0].NewText
	for _, want := range []string{"package go = \"admin\";", "import \"base.hbuf\";", "data Student : Base = 2 {", "    int32? no = 3;"} {
		if !strings.Contains(out, want) {
			t.Errorf("formatted output missing %q:\n%s", want, out)
		}
	}
}

func TestDiagnostics(t *testing.T) {
	// 语法错误
	s, doc, _ := newTestServer(t, "data X {\n    int64 id = 1;\n")
	diags, _ := s.diagnosticsOf(doc)
	if len(diags) == 0 {
		t.Fatal("expect syntax diagnostics")
	}
	found := false
	for _, d := range diags {
		if d.Code == "syntax" {
			found = true
		}
	}
	if !found {
		t.Errorf("expect syntax code, got %+v", diags)
	}

	// 校验错误(id = 0 非法)
	s2, doc2, _ := newTestServer(t, "package go = \"admin\";\n\ndata X {\n    int64 id = 0;\n};\n")
	diags2, _ := s2.diagnosticsOf(doc2)
	found = false
	for _, d := range diags2 {
		if strings.Contains(d.Message, "Invalid id") {
			found = true
		}
	}
	if !found {
		t.Errorf("expect invalid id diagnostic, got %+v", diags2)
	}

	// 无错误文档:无诊断
	s3, doc3, _ := newTestServer(t, lspSrc)
	diags3, _ := s3.diagnosticsOf(doc3)
	if len(diags3) != 0 {
		t.Errorf("expect no diagnostics, got %+v", diags3)
	}
}
