package tsgen

import (
	"hbuf/pkg/ast"
	hparser "hbuf/pkg/parser"
	htoken "hbuf/pkg/token"
	"hbuf/plugin/core/wire"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const genSrc = `package ts = "admin";

enum Status {
    player = 0;
    stop = 1;
};

data User {
    int64 id = 1;
    string? name = 2;
    bool active = 3;
    bytes avatar = 4;
    date created = 5;
    decimal balance = 6;
    string[] tags = 7;
    int8?<int64> scores = 8;
    Status status = 9;
};

data Node {
    string name = 1;
    Node? parent = 2;
    Node[] children = 3;
};

data Order {
    User user = 1;
    User? userOpt = 2;
};

data Base {
    int64 id = 1;
};

data Student:Base = 2 {
    string? no = 3;
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
		Type:     "ts",
		Files:    wire.Convert(pkg, "ts"),
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

// TestGenerateBasic 基础类型映射、默认值与 descriptor 构造。
func TestGenerateBasic(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	files := filesByName(resp)
	data, ok := files["admin/base.data.ts"]
	if !ok {
		t.Fatalf("missing generated file admin/base.data.ts (got %v)", keys(files))
	}
	for _, want := range []string{
		`} from "../src"`,
		"class UserTs implements HData {",
		"id: bigint = 0n",
		"name: string | null = null",
		"active: boolean = false",
		"avatar: Uint8Array = new Uint8Array(0)",
		// date 恒为 Date | null(零时间用 null 表示)
		"created: Date | null = null",
		"balance: string = \"\"",
		"tags: string[] = []",
		"scores: Map<bigint, number | null> = new Map()",
		// enum 字段按 int32 编码
		"status: number = 0",
		"const userDesc: DataDescriptor = NewDataDescriptor(",
		"[1, {desc: NewInt64Descriptor(false), get: (o: UserTs) => o.id, set: (o: UserTs, v: bigint) => o.id = v}]",
		"[2, {desc: NewStringDescriptor(true), get: (o: UserTs) => o.name, set: (o: UserTs, v: string | null) => o.name = v}]",
		"[4, {desc: NewBytesDescriptor(false), get: (o: UserTs) => o.avatar, set: (o: UserTs, v: Uint8Array) => o.avatar = v}]",
		"[5, {desc: NewTimeDescriptor(false), get: (o: UserTs) => o.created, set: (o: UserTs, v: Date | null) => o.created = v}]",
		"[7, {desc: NewListDescriptor(NewStringDescriptor(false), false), get: (o: UserTs) => o.tags, set: (o: UserTs, v: string[]) => o.tags = v}]",
		"[8, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewInt8Descriptor(true), false), get: (o: UserTs) => o.scores, set: (o: UserTs, v: Map<bigint, number | null>) => o.scores = v}]",
		"[9, {desc: NewInt32Descriptor(false), get: (o: UserTs) => o.status, set: (o: UserTs, v: number) => o.status = v}]",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q, got:\n%s", want, data)
		}
	}
	// date 不可空也不应声明为纯 Date
	if strings.Contains(data, "created: Date =") {
		t.Error("date should always be Date | null")
	}
}

// TestGenerateRefs 非可空 data 引用默认零值实例,可空为 null。
func TestGenerateRefs(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	data := filesByName(resp)["admin/base.data.ts"]
	for _, want := range []string{
		"user: UserTs = new UserTs()",
		"userOpt: UserTs | null = null",
		"[1, {desc: NewRefDescriptor(userDesc, false), get: (o: OrderTs) => o.user, set: (o: OrderTs, v: UserTs) => o.user = v}]",
		"[2, {desc: NewRefDescriptor(userDesc, true), get: (o: OrderTs) => o.userOpt, set: (o: OrderTs, v: UserTs | null) => o.userOpt = v}]",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q, got:\n%s", want, data)
		}
	}
}

// TestGenerateRecursive 自引用 data 用 NewSyncDescriptor 延迟构建。
func TestGenerateRecursive(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	data := filesByName(resp)["admin/base.data.ts"]
	for _, want := range []string{
		"const nodeDesc: HBufDescriptor = NewSyncDescriptor(() => NewDataDescriptor(",
		"parent: NodeTs | null = null",
		"children: NodeTs[] = []",
		"[2, {desc: NewRefDescriptor(nodeDesc, true), get: (o: NodeTs) => o.parent, set: (o: NodeTs, v: NodeTs | null) => o.parent = v}]",
		"[3, {desc: NewListDescriptor(NewRefDescriptor(nodeDesc, false), false), get: (o: NodeTs) => o.children, set: (o: NodeTs, v: NodeTs[]) => o.children = v}]",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q, got:\n%s", want, data)
		}
	}
}

// TestGenerateExtends 父类型内嵌字段并挂入 extends_ Map。
func TestGenerateExtends(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	data := filesByName(resp)["admin/base.data.ts"]
	for _, want := range []string{
		"class StudentTs implements HData {",
		"base: BaseTs = new BaseTs()",
		"no: string | null = null",
		"[2, {desc: NewRefDescriptor(baseDesc), get: (o: StudentTs) => o.base, set: (o: StudentTs, v: BaseTs) => o.base = v}]",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q, got:\n%s", want, data)
		}
	}
	// extends_ Map 只在有继承时出现
	if strings.Count(data, "new Map<number, FieldDesc>([") != 6 {
		t.Errorf("expect 6 field maps (5 data + 1 extends), got:\n%s", data)
	}
}

// TestGenerateOrder 被引用 desc 必须先声明(User 在 Order 之前)。
func TestGenerateOrder(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	data := filesByName(resp)["admin/base.data.ts"]
	if strings.Index(data, "const userDesc") > strings.Index(data, "const orderDesc") {
		t.Error("userDesc must be declared before orderDesc")
	}
	if strings.Index(data, "const baseDesc") > strings.Index(data, "const studentDesc") {
		t.Error("baseDesc must be declared before studentDesc")
	}
}

// TestCrossFileImport 跨文件引用:被引用类与 desc 需要 export,引用侧生成相对 import。
func TestCrossFileImport(t *testing.T) {
	fset := htoken.NewFileSet()
	fileMain, err := hparser.ParseFile(fset, "/a/admin.hbuf", `package ts = "admin";
import "public.hbuf";
data Req {
    PublicInfo info = 1;
};`, hparser.AllErrors|hparser.ParseComments)
	if err != nil {
		t.Fatalf("parse main: %v", err)
	}
	filePub, err := hparser.ParseFile(fset, "/a/public.hbuf", `package ts = "public";
data PublicInfo {
    string name = 1;
};`, hparser.AllErrors|hparser.ParseComments)
	if err != nil {
		t.Fatalf("parse public: %v", err)
	}
	pkg := &ast.Package{Files: map[string]*ast.File{
		"/a/admin.hbuf":  fileMain,
		"/a/public.hbuf": filePub,
	}}
	req := &wire.Request{
		Protocol: wire.ProtocolVersion,
		Type:     "ts",
		Files:    wire.Convert(pkg, "ts"),
		Changed:  []string{"/a/admin.hbuf", "/a/public.hbuf"},
	}
	resp, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	files := filesByName(resp)

	pub := files["public/public.data.ts"]
	for _, want := range []string{
		"export class PublicInfoTs implements HData {",
		"export const publicInfoDesc: DataDescriptor = NewDataDescriptor(",
	} {
		if !strings.Contains(pub, want) {
			t.Errorf("public file missing %q, got:\n%s", want, pub)
		}
	}

	admin := files["admin/admin.data.ts"]
	for _, want := range []string{
		// 当前文件在 admin/ 目录,目标在 public/ 目录:相对路径带目录回溯
		`import {PublicInfoTs, publicInfoDesc} from "../public/public.data"`,
		"info: PublicInfoTs = new PublicInfoTs()",
		"[1, {desc: NewRefDescriptor(publicInfoDesc, false), get: (o: ReqTs) => o.info, set: (o: ReqTs, v: PublicInfoTs) => o.info = v}]",
	} {
		if !strings.Contains(admin, want) {
			t.Errorf("admin file missing %q, got:\n%s", want, admin)
		}
	}
}

// TestGenerateSkipUnchanged 未变更文件不生成。
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

// TestGenerateNoTsPackage 无 ts 包注解时与进程内 typescript 后端一致:输出到根目录。
func TestGenerateNoTsPackage(t *testing.T) {
	req := buildRequest(t, genSrc)
	for _, f := range req.Files {
		delete(f.Packages, "ts")
	}
	resp, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	data, ok := filesByName(resp)["base.data.ts"]
	if !ok {
		t.Fatalf("expect root-level base.data.ts without ts package, got %v", keys(filesByName(resp)))
	}
	if !strings.Contains(data, "class UserTs implements HData {") {
		t.Errorf("missing UserTs, got:\n%s", data)
	}
}

// TestGenerateBadType 非 ts 请求报错。
func TestGenerateBadType(t *testing.T) {
	req := buildRequest(t, genSrc)
	req.Type = "go"
	if _, err := Generate(req); err == nil {
		t.Error("expect error for non-ts type")
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
	data := filesByName(resp)["admin/base.data.ts"]
	if !strings.Contains(data, "// CUSTOM UserTs") {
		t.Errorf("expect custom template output, got:\n%s", data)
	}
}

// TestTemplateDirNotExists 模板目录不存在报错。
func TestTemplateDirNotExists(t *testing.T) {
	req := buildRequest(t, genSrc)
	req.Options = []string{"template=/no/such/dir"}
	if _, err := Generate(req); err == nil {
		t.Error("expect error for missing template dir")
	}
}

// TestRuntimeOption 插件参数 runtime 覆盖运行时 import 路径。
func TestRuntimeOption(t *testing.T) {
	req := buildRequest(t, genSrc)
	req.Options = []string{"runtime=@hbuf/runtime"}
	resp, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	data := filesByName(resp)["admin/base.data.ts"]
	if !strings.Contains(data, `} from "@hbuf/runtime"`) {
		t.Errorf("expect runtime option override, got:\n%s", data)
	}
	if strings.Contains(data, `from "../src"`) {
		t.Error("default runtime path should be overridden")
	}
}

func keys(m map[string]string) []string {
	ret := make([]string, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}
