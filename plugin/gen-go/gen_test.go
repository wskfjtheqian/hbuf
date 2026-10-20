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

    Status[]? statuses = 7;
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
		// 可空数组:slice 自身可表示空,字段与 Getter 均不加指针
		"func (g *Student) GetStatuses() []Status {",
		"return g.Statuses",
		"func (g *Student) SetStatuses(val []Status) {",
		"g.Statuses = val",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q", want)
		}
	}
	if strings.Contains(data, "*[]Status") {
		t.Error("nullable array should not be rendered as pointer slice")
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

// TestCrossFileImportPath 跨文件类型引用时,import 路径应为 "<构建包前缀>/<目标包名>"。
func TestCrossFileImportPath(t *testing.T) {
	fset := htoken.NewFileSet()
	fileMain, err := hparser.ParseFile(fset, "/a/sms.public.hbuf", `package go = "sms";
import "public.hbuf";
data Req {
    SmsProvider provider = 1;
};`, hparser.AllErrors|hparser.ParseComments)
	if err != nil {
		t.Fatalf("parse main: %v", err)
	}
	filePub, err := hparser.ParseFile(fset, "/a/public.hbuf", `package go = "public";
enum SmsProvider {
    aliyun = 0;
};`, hparser.AllErrors|hparser.ParseComments)
	if err != nil {
		t.Fatalf("parse public: %v", err)
	}
	pkg := &ast.Package{Files: map[string]*ast.File{
		"/a/sms.public.hbuf": fileMain,
		"/a/public.hbuf":     filePub,
	}}
	req := &wire.Request{
		Protocol: wire.ProtocolVersion,
		Type:     "go",
		Pack:     "github.com/test/hbuf",
		Files:    wire.Convert(pkg, "go"),
		Changed:  []string{"/a/sms.public.hbuf"},
	}
	resp, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	data := filesByName(resp)["sms/sms.public.data.go"]
	if !strings.Contains(data, `"github.com/test/hbuf/public"`) {
		t.Errorf("expect cross-file import %q, got:\n%s", "github.com/test/hbuf/public", data)
	}

	// 构建包前缀为空时,import 退化为末段包名
	req.Pack = ""
	resp, err = Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	data = filesByName(resp)["sms/sms.public.data.go"]
	if !strings.Contains(data, "\n\t\"public\"\n") {
		t.Errorf("expect bare import %q when pack empty, got:\n%s", "public", data)
	}
}

// TestBindForward bind 转发调用应取 req 中嵌入的目标参数类型字段,
// 且跨文件引用的 import 必须出现在 server 文件中。
func TestBindForward(t *testing.T) {
	fset := htoken.NewFileSet()
	fileBase, err := hparser.ParseFile(fset, "/a/base.hbuf", `package go = "base";
data Student {
    string name = 1;
};
data StudentResp {
    string name = 1;
};
server BaseServer {
    StudentResp GetStudent(Student req) = 1;
};`, hparser.AllErrors|hparser.ParseComments)
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	fileManage, err := hparser.ParseFile(fset, "/a/manage.hbuf", `package go = "manage";
import "base.hbuf";
data MgStudent:Student = 2 {
};
data MgStudentResp:StudentResp = 2 {
};
server ManageServer {
    [bind:value="BaseServer.GetStudent"]
    MgStudentResp GetStudent(MgStudent req) = 1;
};`, hparser.AllErrors|hparser.ParseComments)
	if err != nil {
		t.Fatalf("parse manage: %v", err)
	}
	pkg := &ast.Package{Files: map[string]*ast.File{
		"/a/base.hbuf":   fileBase,
		"/a/manage.hbuf": fileManage,
	}}
	req := &wire.Request{
		Protocol: wire.ProtocolVersion,
		Type:     "go",
		Pack:     "github.com/test/hbuf",
		Files:    wire.Convert(pkg, "go"),
		Changed:  []string{"/a/base.hbuf", "/a/manage.hbuf"},
	}
	resp, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	server := filesByName(resp)["manage/manage.server.go"]
	for _, want := range []string{
		// bind 转发:取 req 中嵌入的目标参数类型字段
		`base.GetBaseServer(ctx).GetStudent(ctx, &req.Student)`,
		// 响应包装:当前结果内嵌目标结果
		`return &MgStudentResp{StudentResp: *resp}, nil`,
		// 跨文件引用的 import 出现在 server 文件
		`"github.com/test/hbuf/base"`,
	} {
		if !strings.Contains(server, want) {
			t.Errorf("manage server file missing %q, got:\n%s", want, server)
		}
	}
	if strings.Contains(server, "req.MgStudent") {
		t.Error("bind forward should use target param type field, not current param type")
	}
}

func keys(m map[string]string) []string {
	ret := make([]string, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}
