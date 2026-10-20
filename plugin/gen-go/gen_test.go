package gogen

import (
	"go/parser"
	"go/token"
	"hbuf/pkg/ast"
	hparser "hbuf/pkg/parser"
	htoken "hbuf/pkg/token"
	"hbuf/plugin/core/wire"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const genSrc = `package go = "admin";
package java = "com.hbuf.admin";

data Base {
    int64 id = 1;

    string name = 2;
};

enum Status {
    player = 0;

    stop = 1;
};

data Student:Base = 2 {
    int32? no = 3;

    Status status = 4;

    string[] names = 5;

    string phone = 6;
};

data Message {
    string message = 1;
};

server BaseServer {
    Student GetStudent(Student req) = 1;

    void SendMessage(Message message) = 2;
};
`

// buildRequest 解析内联源码并构造编译请求。
func buildRequest(t *testing.T, src string) *wire.Request {
	t.Helper()
	fset := htoken.NewFileSet()
	file, err := hparser.ParseFile(fset, "/a/base.hbuf", src, hparser.AllErrors|hparser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	pkg := &ast.Package{Files: map[string]*ast.File{"/a/base.hbuf": file}}
	return &wire.Request{
		Protocol: wire.ProtocolVersion,
		Type:     "go",
		Pack:     "github.com/test/hbuf",
		Files:    wire.Convert(pkg, "go"),
		Changed:  []string{"/a/base.hbuf"},
	}
}

func filesByName(resp *wire.Response) map[string]string {
	ret := map[string]string{}
	for _, f := range resp.Files {
		ret[f.Name] = f.Content
	}
	return ret
}

func TestGenerateBasic(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	files := filesByName(resp)
	for _, name := range []string{"admin/base.data.go", "admin/base.enum.go", "admin/base.server.go"} {
		if _, ok := files[name]; !ok {
			t.Errorf("missing generated file: %s (got %v)", name, keys(files))
		}
	}

	data := files["admin/base.data.go"]
	for _, want := range []string{
		"package admin",
		`import (`,
		`"reflect"`,
		`"github.com/wskfjtheqian/hbuf_golang/pkg/hbuf"`,
		"var baseDescriptor = hbuf.NewSyncDescriptor",
		"hbuf.NewDataDescriptor(0, false, reflect.TypeOf(&base)",
		"unsafe.Offsetof(base.Id)",
		"type Base struct {",
		`json:"id,omitempty" hbuf:"1"`,
		"func (g *Base) GetId() int64 {",
		"func (g *Base) SetId(val int64) {",
		"type Student struct {",
		"Base",
		"`hbuf:\"2\"`",
		"GetNo() int32 {",
		"if nil == g.No {",
		"return int32(0)",
		"return *g.No",
		"SetNo(val int32) {",
		"g.No = &val",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q", want)
		}
	}

	enum := files["admin/base.enum.go"]
	for _, want := range []string{
		"type Status int32",
		"const StatusPlayer Status = 0",
		"func (t Status) MarshalJSON()",
		"var statusMap = map[Status]string{",
		"func (e Status) ToName() string {",
		"var statusValues = map[string]Status{",
		"func StatusValues() map[string]Status {",
	} {
		if !strings.Contains(enum, want) {
			t.Errorf("enum file missing %q", want)
		}
	}

	server := files["admin/base.server.go"]
	for _, want := range []string{
		"type BaseServer interface {",
		"Init(ctx context.Context)",
		"GetStudent(ctx context.Context, req *Student) (*Student, error)",
		"SendMessage(ctx context.Context, message *Message) error",
		"type DefaultBaseServer struct {",
		"func (s *DefaultBaseServer) Init(ctx context.Context) {",
		"herror.NewError(\"not find server base_server\")",
		"type BaseServerClient struct {",
		"func NewBaseServerClient(client *hrpc.Client) BaseServer {",
		"r.client.Invoke(ctx, 0, \"base_server\", \"get_student\"",
		"hrpc.NewResultResponse[*Student]()",
		"func RegisterBaseServer(r hrpc.ServerRegister, server BaseServer) {",
		"r.Register(0, \"base_server\", server,",
		"&hrpc.Method{",
		"Name: \"get_student\",",
		"decoder(&Student{})",
		"var NotFoundBaseServer = &DefaultBaseServer{}",
		"func GetBaseServer(ctx context.Context) BaseServer {",
		"hservice.GetClient(ctx, \"base_server\")",
		"func BaseServerName() string {",
	} {
		if !strings.Contains(server, want) {
			t.Errorf("server file missing %q", want)
		}
	}
}

// TestGeneratedSyntax 用 go/parser 校验生成文件语法合法。
func TestGeneratedSyntax(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range resp.Files {
		fset := token.NewFileSet()
		if _, err := parser.ParseFile(fset, f.Name, f.Content, parser.AllErrors); err != nil {
			t.Errorf("syntax error in %s: %v\n%s", f.Name, err, f.Content)
		}
	}
}

func TestGenerateSkipUnchanged(t *testing.T) {
	req := buildRequest(t, genSrc)
	req.Changed = []string{"/a/other.hbuf"}
	resp, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Files) != 0 {
		t.Errorf("expect no files for unchanged path, got %v", keys(filesByName(resp)))
	}
}

func TestGenerateNoGoPackage(t *testing.T) {
	req := buildRequest(t, genSrc)
	for _, f := range req.Files {
		delete(f.Packages, "go")
	}
	resp, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Files) != 0 {
		t.Errorf("expect no files without go package, got %v", keys(filesByName(resp)))
	}
}

func TestGenerateBadType(t *testing.T) {
	req := buildRequest(t, genSrc)
	req.Type = "dart"
	if _, err := Generate(req); err == nil {
		t.Error("expect error for non-go type")
	}
}

// TestCustomTemplateOverride 自定义模板覆盖默认模板。
func TestCustomTemplateOverride(t *testing.T) {
	dir := t.TempDir()
	custom := `{{range .Datas}}// CUSTOM {{.Name}}
{{end}}`
	if err := os.WriteFile(filepath.Join(dir, "data.tmpl"), []byte(custom), 0644); err != nil {
		t.Fatal(err)
	}
	req := buildRequest(t, genSrc)
	req.Options = []string{"template=" + dir}
	resp, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	data := filesByName(resp)["admin/base.data.go"]
	if !strings.Contains(data, "// CUSTOM Base") {
		t.Errorf("expect custom template output, got:\n%s", data)
	}
	// 未覆盖的 enum 模板仍来自默认
	if !strings.Contains(filesByName(resp)["admin/base.enum.go"], "type Status int32") {
		t.Error("enum should still use default template")
	}
}

func TestTemplateDirNotExists(t *testing.T) {
	req := buildRequest(t, genSrc)
	req.Options = []string{"template=/no/such/dir"}
	if _, err := Generate(req); err == nil {
		t.Error("expect error for missing template dir")
	}
}

func keys(m map[string]string) []string {
	ret := make([]string, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}
