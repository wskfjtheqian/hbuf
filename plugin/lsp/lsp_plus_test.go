package lsp

import (
	"os"
	"path/filepath"
	"testing"
)

// decodeSemTokens delta 解码为绝对位置的 token 列表。
func decodeSemTokens(data []uint32) []semToken {
	var toks []semToken
	line, char := 0, 0
	for i := 0; i+4 < len(data); i += 5 {
		line += int(data[i])
		if data[i] == 0 {
			char += int(data[i+1])
		} else {
			char = int(data[i+1])
		}
		toks = append(toks, semToken{
			line:   line,
			char:   char,
			length: int(data[i+2]),
			typ:    int(data[i+3]),
			mod:    data[i+4],
		})
	}
	return toks
}

// findSemToken 按 (line,char) 查找语义 token。
func findSemToken(t *testing.T, toks []semToken, pos Position) semToken {
	t.Helper()
	for _, tok := range toks {
		if tok.line == pos.Line && tok.char == pos.Character {
			return tok
		}
	}
	t.Fatalf("semantic token not found at %+v (have %d tokens)", pos, len(toks))
	return semToken{}
}

func TestSemanticTokens(t *testing.T) {
	s, doc, _ := newTestServer(t, lspSrc)
	res := s.semanticTokens(doc)
	if len(res.Data) == 0 || len(res.Data)%5 != 0 {
		t.Fatalf("empty or misaligned data: %d", len(res.Data))
	}
	toks := decodeSemTokens(res.Data)

	// 排序检查:逐 token 位置递增
	for i := 1; i < len(toks); i++ {
		a, b := toks[i-1], toks[i]
		if a.line > b.line || (a.line == b.line && a.char >= b.char) {
			t.Fatalf("tokens not ordered: %+v then %+v", a, b)
		}
	}

	check := func(substr string, typ int) {
		t.Helper()
		tok := findSemToken(t, toks, posAt(t, lspSrc, substr))
		if tok.typ != typ {
			t.Errorf("token %q at %+v: type=%d want %d", substr, posAt(t, lspSrc, substr), tok.typ, typ)
		}
	}
	// 声明
	check("Student", semTokStruct)       // data 类型名
	check("Status {", semTokEnum)        // enum 类型名("Status status" 先出现,取 enum 声明处)
	check("BaseServer", semTokInterface) // server 类型名
	check("no = 3", semTokProperty)      // 字段名
	check("GetStudent", semTokMethod)    // 方法名
	check("player", semTokEnumMember)    // 枚举项名
	check("req", semTokParameter)        // 方法参数名
	// 关键字/注释/字符串/数字
	check("data Student", semTokKeyword)
	check("//学生信息", semTokComment)
	check(`"base.hbuf"`, semTokString)
	check("3;", semTokNumber) // 数字 ID
	// 类型引用:基础类型/同文件类型/跨文件类型
	check("int32?", semTokType)
	check("Status status", semTokEnum) // 引用 enum 类型名
	check("Message message", semTokStruct)

	// 注解 tag 名与键名
	tagSrc := "data X {\n    [verify: regex=\"a\"];\n    string name = 1;\n};\n"
	doc2 := s.view.Open(pathToURI(filepath.Join(t.TempDir(), "tag.hbuf")), "hbuf", 1, tagSrc)
	res2 := s.semanticTokens(doc2)
	toks2 := decodeSemTokens(res2.Data)
	check2 := func(substr string, typ int) {
		t.Helper()
		tok := findSemToken(t, toks2, posAt(t, tagSrc, substr))
		if tok.typ != typ {
			t.Errorf("tag token %q: type=%d want %d", substr, tok.typ, typ)
		}
	}
	check2("verify", semTokDecorator)
	check2("regex", semTokMacro)
}

func TestCodeAction(t *testing.T) {
	// 未使用 import → 提供删除 action
	unusedSrc := "package go = \"admin\";\n\nimport \"base.hbuf\";\n\ndata X {\n    int64 id = 1;\n};\n"
	s, doc, _ := newTestServer(t, unusedSrc)
	actions := s.codeAction(doc, CodeActionParams{})
	foundDel := false
	for _, a := range actions {
		if a.Kind == "quickfix" && a.Edit != nil {
			foundDel = true
			edits := a.Edit.Changes[doc.URI]
			if len(edits) != 1 || edits[0].NewText != "" {
				t.Errorf("import delete edit mismatch: %+v", edits)
			}
			if edits[0].Range.Start.Line != 2 || edits[0].Range.End.Line != 3 {
				t.Errorf("import delete range mismatch: %+v", edits[0].Range)
			}
		}
	}
	if !foundDel {
		t.Errorf("missing unused-import quickfix: %+v", actions)
	}

	// 已使用的 import → 不提供删除 action
	s2, doc2, _ := newTestServer(t, lspSrc)
	for _, a := range s2.codeAction(doc2, CodeActionParams{}) {
		if a.Kind == "quickfix" {
			t.Errorf("unexpected quickfix for used import: %+v", a)
		}
	}

	// 无语法错误 → 提供格式化 action
	foundFormat := false
	for _, a := range actions {
		if a.Kind == "source.fixAll.hbuf" {
			foundFormat = true
		}
	}
	if !foundFormat {
		t.Error("missing source.fixAll.hbuf action")
	}

	// 语法错误 → 不提供格式化 action
	badSrc := "data X {\n"
	doc3 := s.view.Open(pathToURI(filepath.Join(t.TempDir(), "bad.hbuf")), "hbuf", 1, badSrc)
	for _, a := range s.codeAction(doc3, CodeActionParams{}) {
		if a.Kind == "source.fixAll.hbuf" {
			t.Errorf("unexpected format action for syntax-error document: %+v", a)
		}
	}
}

func TestInlayHint(t *testing.T) {
	s, doc, _ := newTestServer(t, lspSrc)
	full := Range{Start: Position{}, End: Position{Line: 100, Character: 0}}
	hints := s.inlayHints(doc, full)

	var labels []string
	for _, h := range hints {
		labels = append(labels, h.Label)
	}
	has := func(want string) bool {
		for _, l := range labels {
			if l == want {
				return true
			}
		}
		return false
	}
	if !has("2 个字段") { // Base(id/name) 继承摘要
		t.Errorf("missing extends summary hint: %v", labels)
	}
	if !has("base.hbuf") { // Message/Base 来自 import 文件
		t.Errorf("missing import origin hint: %v", labels)
	}
	// 范围过滤:只含 line 5(extends 行)
	line5 := Range{Start: Position{Line: 5, Character: 0}, End: Position{Line: 5, Character: 100}}
	for _, h := range s.inlayHints(doc, line5) {
		if h.Position.Line != 5 {
			t.Errorf("hint out of range: %+v", h)
		}
	}
}

func TestWorkspaceSymbol(t *testing.T) {
	dir := t.TempDir()
	aSrc := "package go = \"admin\";\n\ndata Alpha {\n    int64 x = 1;\n};\n\nenum Color {\n    red = 0;\n};\n"
	bSrc := "package go = \"admin\";\n\nserver BetaServer {\n    Alpha GetAlpha(Alpha req) = 1;\n};\n"
	if err := os.WriteFile(filepath.Join(dir, "a.hbuf"), []byte(aSrc), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.hbuf"), []byte(bSrc), 0644); err != nil {
		t.Fatal(err)
	}
	s := &Server{view: NewView()}
	s.view.SetWorkspaceRoots([]string{dir})

	names := func(query string) []string {
		var ret []string
		for _, sym := range s.workspaceSymbols(query) {
			ret = append(ret, sym.Name)
		}
		return ret
	}
	has := func(list []string, want string) bool {
		for _, n := range list {
			if n == want {
				return true
			}
		}
		return false
	}

	// 类型查询
	got := names("alpha")
	if !has(got, "Alpha") || !has(got, "GetAlpha") {
		t.Errorf("query alpha: %v", got)
	}
	// 字段查询
	got = names("x")
	if !has(got, "x") {
		t.Errorf("query x: %v", got)
	}
	// 枚举项查询(含 container)
	found := false
	for _, sym := range s.workspaceSymbols("red") {
		if sym.Name == "red" && sym.ContainerName == "Color" {
			found = true
		}
	}
	if !found {
		t.Error("query red: missing enum member with container")
	}
	// 空查询返回全部(含另一文件类型)
	got = names("")
	if !has(got, "BetaServer") || len(got) == 0 {
		t.Errorf("empty query: %v", got)
	}
	// 不匹配
	if got = names("zzz"); len(got) != 0 {
		t.Errorf("query zzz: %v", got)
	}
}
