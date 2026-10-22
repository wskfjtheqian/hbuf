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
		// 类型符号以内联 type 修饰符导入(verbatimModuleSyntax 下避免 TS1484)
		"    type FieldDesc,",
		"    type HBufDescriptor,",
		"    type HData,",
		// 与旧版后端一致:所有 data 类与 desc 一律 export
		"export class User implements HData {",
		"id: bigint = 0n",
		"name: string | null = null",
		"active: boolean = false",
		"avatar: Uint8Array = new Uint8Array(0)",
		// date 恒为 Date | null(零时间用 null 表示)
		"created: Date | null = null",
		// decimal 字段用 Decimal 类(与旧版后端一致),第三方模块 import
		"import {Decimal} from \"decimal.js\"",
		"balance: Decimal = new Decimal(0)",
		"tags: string[] = []",
		// map 用 Record 对象(与旧版后端一致),键为 JSON 键语义(int64 -> string)
		"scores: Record<string, number | null> = {}",
		// enum 字段用 enum 类(与旧版后端一致),descriptor 仍按 int32 编码
		"status: Status = Status.valueOf(0)",
		"export const userDesc: DataDescriptor = NewDataDescriptor(",
		"[1, {desc: NewInt64Descriptor(false), get: (o: User) => o.id, set: (o: User, v: bigint) => o.id = v}]",
		"[2, {desc: NewStringDescriptor(true), get: (o: User) => o.name, set: (o: User, v: string | null) => o.name = v}]",
		"[4, {desc: NewBytesDescriptor(false), get: (o: User) => o.avatar, set: (o: User, v: Uint8Array) => o.avatar = v}]",
		"[5, {desc: NewTimeDescriptor(false), get: (o: User) => o.created, set: (o: User, v: Date | null) => o.created = v}]",
		// decimal descriptor 以字符串承载:get 取 toString(),set 用 new Decimal 还原
		"[6, {desc: NewDecimalDescriptor(false), get: (o: User) => o.balance.toString(), set: (o: User, v: string) => o.balance = new Decimal(v)}]",
		"[7, {desc: NewListDescriptor(NewStringDescriptor(false), false), get: (o: User) => o.tags, set: (o: User, v: string[]) => o.tags = v}]",
		// map descriptor 层仍操作 Map:get 转 Map(key 恢复 BigInt),set 转回 Record
		"[8, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewInt8Descriptor(true), false), get: (o: User) => (() => { const m = new Map<any, any>(); const r: any = o.scores; for (const k of Object.keys(r)) { m.set(BigInt(k), r[k]) } return m })(), set: (o: User, v: Map<any, any>) => o.scores = (() => { const r: Record<string, any> = {}; for (const [k, e] of v) { r[String(k)] = e } return r })()}]",
		"[9, {desc: NewInt32Descriptor(false), get: (o: User) => o.status.value, set: (o: User, v: number) => o.status = Status.valueOf(v)}]",
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
		"user: User = new User()",
		"userOpt: User | null = null",
		"[1, {desc: NewRefDescriptor(userDesc, false), get: (o: Order) => o.user, set: (o: Order, v: User) => o.user = v}]",
		"[2, {desc: NewRefDescriptor(userDesc, true), get: (o: Order) => o.userOpt, set: (o: Order, v: User | null) => o.userOpt = v}]",
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
		"export const nodeDesc: HBufDescriptor = NewSyncDescriptor(() => NewDataDescriptor(",
		"parent: Node | null = null",
		"children: Node[] = []",
		"[2, {desc: NewRefDescriptor(nodeDesc, true), get: (o: Node) => o.parent, set: (o: Node, v: Node | null) => o.parent = v}]",
		"[3, {desc: NewListDescriptor(NewRefDescriptor(nodeDesc, false), false), get: (o: Node) => o.children, set: (o: Node, v: Node[]) => o.children = v}]",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q, got:\n%s", want, data)
		}
	}
}

// TestGenerateExtends 类 implements 直接父类,祖先字段展平,extends_ Map 按扩展 ID 挂父 desc。
func TestGenerateExtends(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	data := filesByName(resp)["admin/base.data.ts"]
	for _, want := range []string{
		"export class Student implements HData, Base {",
		// 父类也一律 export(仅被同文件 extends 引用的父类同样可被业务代码 import)
		"export class Base implements HData {",
		// 父类字段展平声明(id 不重复在 Student 的 fields Map)
		"id: bigint = 0n",
		"no: string | null = null",
		"[2, {desc: baseDesc, get: (o: Student) => o, set: (o: Student, v: Base) => Object.assign(o, v)}]",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q, got:\n%s", want, data)
		}
	}
	// Student 的 fields Map 只含自身字段 no;extends_ Map 只在有继承时出现
	if strings.Contains(data, "get: (o: Student) => o.id,") {
		t.Error("parent field id must not appear in Student fields Map, got:\n" + data)
	}
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
		"export class PublicInfo implements HData {",
		"export const publicInfoDesc: DataDescriptor = NewDataDescriptor(",
	} {
		if !strings.Contains(pub, want) {
			t.Errorf("public file missing %q, got:\n%s", want, pub)
		}
	}

	admin := files["admin/admin.data.ts"]
	for _, want := range []string{
		// 当前文件在 admin/ 目录,目标在 public/ 目录:相对路径带目录回溯
		`import {PublicInfo, publicInfoDesc} from "../public/public.data"`,
		"info: PublicInfo = new PublicInfo()",
		"[1, {desc: NewRefDescriptor(publicInfoDesc, false), get: (o: Req) => o.info, set: (o: Req, v: PublicInfo) => o.info = v}]",
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
	if !strings.Contains(data, "export class User implements HData {") {
		t.Errorf("missing User, got:\n%s", data)
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

// TestGenerateEnum 枚举类独立生成到 .enum.ts:value/name/type/cssClass 与 UI 注解。
func TestGenerateEnum(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	enum := filesByName(resp)["admin/base.enum.ts"]
	for _, want := range []string{
		"export class Status {",
		"public readonly value: number",
		"public readonly name: string",
		"public readonly type: string",
		"public readonly cssClass: string",
		// 无 ui 注解:type 默认 info,cssClass 为空
		`public static readonly PLAYER = new Status(0, "Player", "info", "")`,
		`public static readonly STOP = new Status(1, "Stop", "info", "")`,
		"public static readonly values: Status[] = [",
		"Status.PLAYER,",
		"toString(): string {",
		"return this.name",
	} {
		if !strings.Contains(enum, want) {
			t.Errorf("enum file missing %q, got:\n%s", want, enum)
		}
	}

	// ui 注解:type 覆盖默认值,class 追加项名
	resp, err = Generate(buildRequest(t, `package ts = "admin";
enum UiStatus {
    [ui:type="success"; class="round"]
    ok = 0;
    fail = 1;
};`))
	if err != nil {
		t.Fatal(err)
	}
	enum = filesByName(resp)["admin/base.enum.ts"]
	for _, want := range []string{
		`public static readonly OK = new UiStatus(0, "Ok", "success", "round ok")`,
		`public static readonly FAIL = new UiStatus(1, "Fail", "info", "")`,
	} {
		if !strings.Contains(enum, want) {
			t.Errorf("ui enum missing %q, got:\n%s", want, enum)
		}
	}
}

// TestGenerateServer 服务生成 .server.ts:接口/客户端/注册函数,void 与 stream 特殊处理。
func TestGenerateServer(t *testing.T) {
	const serverSrc = `package ts = "admin";

data Req {
    string name = 1;
};

data Resp {
    string message = 1;
};

server UserServer {
    Resp Login(Req req) = 1;
    void Notify(void v) = 2;
    stream Download(string name) = 3;
    string Ping(void v) = 4;
};

server Child:UserServer = 2 {
    Resp Login2(Req req) = 1;
};
`
	resp, err := Generate(buildRequest(t, serverSrc))
	if err != nil {
		t.Fatal(err)
	}
	server := filesByName(resp)["admin/base.server.ts"]
	for _, want := range []string{
		// data/server 是独立输出文件:同文件引用的 data 类也必须 import
		`import {Req, Resp} from "./base.data"`,
		// 类型符号以内联 type 修饰符导入(verbatimModuleSyntax 下避免 TS1484)
		"    type BufferType,",
		"    type Option,",
		"    type RequestType,",
		"    type ResponseType,",
		"export interface UserServer {",
		"login(req: Req, _opt?: Option): Promise<Resp>",
		"notify(_opt?: Option): Promise<void>",
		"download(name: string, _opt?: Option): Promise<BufferType>",
		"ping(_opt?: Option): Promise<string>",
		"export class UserServerClient implements UserServer {",
		"protected client: Client",
		`return this.client.invoke<Resp>(1, "user_server", "login", req, Resp.fromMap)`,
		`return this.client.invoke<void>(2, "user_server", "notify", undefined)`,
		`return this.client.invoke<BufferType>(3, "user_server", "download", name)`,
		`return this.client.invoke<string>(4, "user_server", "ping", undefined)`,
		"export function RegisterUserServer(r: Server, server: UserServer) {",
		`r.register(0, "user_server", [`,
		"id: 1,",
		`name: "login",`,
		"return server.login(req as Req, opt)",
		"return server.notify(opt)",
		"from: Req.fromMap,",
		"export function UnRegisterUserServer(r: Server) {",
		`r.unRegister(0, "user_server")`,
		// 继承:接口 extends,客户端合并父类方法
		"export interface Child extends UserServer {",
		"export class ChildClient implements Child {",
		"login2(req: Req, _opt?: Option): Promise<Resp>",
		"export function RegisterChild(r: Server, server: Child) {",
	} {
		if !strings.Contains(server, want) {
			t.Errorf("server file missing %q, got:\n%s", want, server)
		}
	}
	// 同文件 data 被 server 引用时类与 desc 需要 export
	data := filesByName(resp)["admin/base.data.ts"]
	for _, want := range []string{
		"export class Req implements HData {",
		"export const reqDesc: DataDescriptor = NewDataDescriptor(",
		"export class Resp implements HData {",
		"export const respDesc: DataDescriptor = NewDataDescriptor(",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q, got:\n%s", want, data)
		}
	}
}

// TestGenerateServerCrossFile server 引用其他文件的 data 时生成相对 import。
func TestGenerateServerCrossFile(t *testing.T) {
	fset := htoken.NewFileSet()
	fileMain, err := hparser.ParseFile(fset, "/a/admin.hbuf", `package ts = "admin";
import "public.hbuf";
data Resp {
    string message = 1;
};
server UserServer {
    Resp Login(PublicInfo req) = 1;
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
	server := files["admin/admin.server.ts"]
	for _, want := range []string{
		`import {PublicInfo} from "../public/public.data"`,
		"login(req: PublicInfo, _opt?: Option): Promise<Resp>",
		"from: PublicInfo.fromMap,",
	} {
		if !strings.Contains(server, want) {
			t.Errorf("server file missing %q, got:\n%s", want, server)
		}
	}
	pub := files["public/public.data.ts"]
	for _, want := range []string{
		"export class PublicInfo implements HData {",
		"export const publicInfoDesc: DataDescriptor = NewDataDescriptor(",
	} {
		if !strings.Contains(pub, want) {
			t.Errorf("public file missing %q, got:\n%s", want, pub)
		}
	}
}

// TestGenerateJson data 类 fromMap/toMap 覆盖基础类型、数组、映射与引用转换。
func TestGenerateJson(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	data := filesByName(resp)["admin/base.data.ts"]
	for _, want := range []string{
		"static fromMap(map: Record<string, any>): User {",
		"toMap(): Record<string, any> {",
		// int64 非可空:JSON 字符串 -> BigInt
		`ret.id = null == map["id"] ? 0n : BigInt(map["id"] as string)`,
		`"id": this.id.toString(),`,
		// 可空 string
		`ret.name = null == map["name"] ? null : map["name"].toString()`,
		// bytes base64 编解码
		`ret.avatar = null == map["avatar"] ? new Uint8Array(0) : new Uint8Array([...atob(map["avatar"])].map(c => c.charCodeAt(0)))`,
		`"avatar": btoa(String.fromCharCode(...this.avatar)),`,
		// date 毫秒时间戳
		`ret.created = null == map["created"] ? new Date(0) : new Date(Number(map["created"]).valueOf())`,
		`"created": this.created?.getTime(),`,
		// decimal 字段用 Decimal 实例,JSON 层经十进制字符串还原/输出
		`ret.balance = null == map["balance"] ? new Decimal(0) : new Decimal(map["balance"] as string)`,
		`"balance": this.balance.toString(),`,
		// 数组用 convertArray
		`ret.tags = null == map["tags"] ? [] : (!isArray(map["tags"]) ? [] : convertArray(map["tags"], (item) => null == item ? "" : item.toString())!)`,
		`"tags": convertArray(this.tags, (e) => e),`,
		// 映射用 convertRecord:JSON 键为字符串,int64 key 原样、值递归转换
		"ret.scores = null == map[\"scores\"] ? {} : (!isRecord(map[\"scores\"]) ? {} : convertRecord(map[\"scores\"], (key, value) => new RecordEntry(key, null == value ? null : Number(value).valueOf()))!)",
		`"scores": convertRecord(this.scores, (key, value) => new RecordEntry(key, value)),`,
		// enum 字段用 enum 类,JSON 层经 value 数字还原/输出
		"ret.status = null == map[\"status\"] ? Status.valueOf(0) : Status.valueOf(Number(map[\"status\"]).valueOf())",
		`"status": this.status.value,`,
		// 非可空 data 引用递归 fromMap,JSON 缺失时回退空结构
		`ret.user = null == map["user"] ? User.fromMap({}) : User.fromMap(map["user"] as Record<string, any>)`,
		`"user": this.user.toMap(),`,
		// 可空 data 引用
		`ret.userOpt = null == map["user_opt"] ? null : User.fromMap(map["user_opt"] as Record<string, any>)`,
		`"user_opt": this.userOpt?.toMap(),`,
		// 继承字段展平:父类字段与自身字段同层级 JSON 键
		`ret.id = null == map["id"] ? 0n : BigInt(map["id"] as string)`,
		`ret.no = null == map["no"] ? null : map["no"].toString()`,
		`"id": this.id.toString(),`,
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q, got:\n%s", want, data)
		}
	}
}

// TestGenerateAuxMethods 辅助方法 clone/$getChangeField/$fieldMaps/$convertField/fromData/toData。
func TestGenerateAuxMethods(t *testing.T) {
	resp, err := Generate(buildRequest(t, genSrc))
	if err != nil {
		t.Fatal(err)
	}
	data := filesByName(resp)["admin/base.data.ts"]
	for _, want := range []string{
		// 二进制编解码:新版 hbuf_ts 运行时真实序列化
		"public static fromData(data: ArrayBuffer): User {",
		"return NewHBufDecoder()(data, User.fromMap) as User",
		"public toData(): ArrayBuffer {",
		"return NewHBufEncode()(this)",
		// clone:基础类型直接赋值,date/bytes/decimal 重建
		"public clone(): User {",
		"ret.id = this.id",
		"ret.created = null == this.created ? null : new Date(this.created.getTime())",
		"ret.avatar = new Uint8Array([...this.avatar])",
		// clone:Decimal 实例可变,复制构造
		"ret.balance = new Decimal(this.balance)",
		// clone:数组递归拷贝
		"ret.tags = null == this.tags ? [] : convertArray(this.tags, (e) => e)!",
		// clone:映射 convertRecord 逐项拷贝
		"ret.scores = convertRecord(this.scores, (key, value) => new RecordEntry(key, value))!",
		// clone:非可空 data 引用递归 clone,可空 null 透传
		"ret.user = this.user.clone()",
		"ret.userOpt = null == this.userOpt ? null : this.userOpt.clone()",
		// clone:继承展平字段参与拷贝
		"ret.no = this.no",
		// clone:enum 实例为单例,引用直接赋值
		"ret.status = this.status",
		// $getChangeField:enum 按 value、decimal 按字符串、date 按时间戳比较,其余 !=;字段名取原始 IDL 名
		"if (ignore.includes(\"status\") || this.status?.value != val.status?.value) {",
		"if (ignore.includes(\"balance\") || this.balance?.toString() != val.balance?.toString()) {",
		"if (ignore.includes(\"created\") || this.created?.getTime() != val.created?.getTime()) {",
		"ret.fields.push(\"created\")",
		"ret.change.created = this.created",
		"if (ignore.includes(\"id\") || this.id != val.id) {",
		// $fieldMaps 与 $convertField
		"private static $fieldMaps: Record<string, string> = {",
		"\"created\": \"created\",",
		"public static $convertField(fields: string[]): string[] {",
		"list.add(User.$fieldMaps[field])",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("data file missing %q, got:\n%s", want, data)
		}
	}
	// 展平继承:Student 的 $fieldMaps 含父类字段 id
	if !strings.Contains(data, "export class Student implements HData, Base {") {
		t.Skip("Student missing")
	}
	sd := data[strings.Index(data, "class Student"):]
	for _, want := range []string{
		"ret.id = this.id",
		"if (ignore.includes(\"id\") || this.id != val.id) {",
	} {
		if !strings.Contains(sd, want) {
			t.Errorf("Student missing %q, got:\n%s", want, sd)
		}
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
	if !strings.Contains(data, "// CUSTOM User") {
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
