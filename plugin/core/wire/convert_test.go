package wire

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
	"testing"
)

// parseSrc 解析内联 .hbuf 源码
func parseSrc(t *testing.T, path string, src string) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return file
}

const convertSrc = `package go = "admin";
package java = "com.hbuf.admin";

data Base {
    int64 id = 0;

    string name = 1;
};

enum Status {
    player = 0;

    stop = 1;
};

data Student:Base = 2 {
    int32? no = 0;

    Status status = 1;

    string[] names = 2;

    [verify: regex="^1[0-9]{10}$"];
    string phone = 4;
};

server BaseServer {
    Student GetStudent(int64 id) = 0;
};
`

func TestConvertFile(t *testing.T) {
	file := parseSrc(t, "/a/base.hbuf", convertSrc)
	pkg := &ast.Package{Files: map[string]*ast.File{"/a/base.hbuf": file}}

	files := Convert(pkg, "go")
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1", len(files))
	}
	wf := files[0]
	if wf.Path != "/a/base.hbuf" {
		t.Errorf("path = %q", wf.Path)
	}
	if wf.Packages["go"] != "admin" || wf.Packages["java"] != "com.hbuf.admin" {
		t.Errorf("packages = %v", wf.Packages)
	}
	if len(wf.Types) != 4 {
		t.Fatalf("types = %d, want 4", len(wf.Types))
	}

	// data Base
	base := wf.Types[0]
	if base.Kind != "data" || base.Data.Name != "Base" {
		t.Fatalf("types[0] = %+v", base)
	}
	if len(base.Data.Fields) != 2 {
		t.Fatalf("Base fields = %d", len(base.Data.Fields))
	}
	id := base.Data.Fields[0]
	if id.Name != "id" || id.Id != 0 || id.Type.Kind != "base" || id.Type.Name != "int64" || id.Type.Empty {
		t.Errorf("Base.id = %+v", id)
	}
	name := base.Data.Fields[1]
	if name.Name != "name" || name.Id != 1 || name.Type.Name != "string" {
		t.Errorf("Base.name = %+v", name)
	}

	// enum Status
	status := wf.Types[1]
	if status.Kind != "enum" || status.Enum.Name != "Status" || len(status.Enum.Items) != 2 {
		t.Fatalf("types[1] = %+v", status)
	}
	if status.Enum.Items[0].Name != "player" || status.Enum.Items[0].Id != 0 {
		t.Errorf("item0 = %+v", status.Enum.Items[0])
	}
	if status.Enum.Items[1].Name != "stop" || status.Enum.Items[1].Id != 1 {
		t.Errorf("item1 = %+v", status.Enum.Items[1])
	}

	// data Student:Base = 2
	student := wf.Types[2]
	if student.Kind != "data" || student.Data.Name != "Student" {
		t.Fatalf("types[2] = %+v", student)
	}
	if len(student.Data.Extends) != 1 || student.Data.Extends[0].Name != "Base" || student.Data.Extends[0].Id != 2 {
		t.Errorf("Student extends = %+v", student.Data.Extends)
	}
	if len(student.Data.Fields) != 4 {
		t.Fatalf("Student fields = %d", len(student.Data.Fields))
	}

	no := student.Data.Fields[0]
	if no.Name != "no" || !no.Type.Empty || no.Type.Kind != "base" || no.Type.Name != "int32" {
		t.Errorf("Student.no = %+v", no)
	}
	statusField := student.Data.Fields[1]
	if statusField.Type.Kind != "ref" || statusField.Type.Name != "Status" || statusField.Type.RefKind != "enum" || statusField.Type.RefPack != "admin" {
		t.Errorf("Student.status = %+v", statusField.Type)
	}
	names := student.Data.Fields[2]
	if names.Type.Kind != "array" || names.Type.Element == nil || names.Type.Element.Name != "string" {
		t.Errorf("Student.names = %+v", names.Type)
	}
	phone := student.Data.Fields[3]
	if len(phone.Tags) != 1 || phone.Tags[0].Name != "verify" {
		t.Fatalf("Student.phone tags = %+v", phone.Tags)
	}
	if len(phone.Tags[0].KV) != 1 || phone.Tags[0].KV[0].Name != "regex" || len(phone.Tags[0].KV[0].Values) != 1 || phone.Tags[0].KV[0].Values[0] != "^1[0-9]{10}$" {
		t.Errorf("phone regex = %+v", phone.Tags[0].KV)
	}

	// server BaseServer
	server := wf.Types[3]
	if server.Kind != "server" || server.Server.Name != "BaseServer" || len(server.Server.Methods) != 1 {
		t.Fatalf("types[3] = %+v", server)
	}
	method := server.Server.Methods[0]
	if method.Name != "GetStudent" || method.Id != 0 || method.ParamName != "id" {
		t.Errorf("method = %+v", method)
	}
	if method.Result == nil || method.Result.Kind != "ref" || method.Result.Name != "Student" {
		t.Errorf("method result = %+v", method.Result)
	}
	if method.Param == nil || method.Param.Kind != "base" || method.Param.Name != "int64" {
		t.Errorf("method param = %+v", method.Param)
	}
}

// TestConvertCrossFile 跨文件引用:RefKind 与 RefPack 需从包内解析
func TestConvertCrossFile(t *testing.T) {
	srcEnum := `package go = "public";

enum State {
    on = 0;
};
`
	srcMain := `package go = "admin";

import "public.hbuf";

data Device {
    State state = 0;
};
`
	enumFile := parseSrc(t, "/a/public.hbuf", srcEnum)
	mainFile := parseSrc(t, "/a/device.hbuf", srcMain)
	pkg := &ast.Package{Files: map[string]*ast.File{
		"/a/public.hbuf": enumFile,
		"/a/device.hbuf": mainFile,
	}}

	files := Convert(pkg, "go")
	if len(files) != 2 {
		t.Fatalf("files = %d, want 2", len(files))
	}
	// 文件按路径排序,device.hbuf 在前
	device := files[0]
	if device.Path != "/a/device.hbuf" || len(device.Imports) != 1 || device.Imports[0] != "public.hbuf" {
		t.Fatalf("device file = %+v", device)
	}
	if len(device.Types) != 1 || device.Types[0].Data == nil {
		t.Fatalf("device types = %+v", device.Types)
	}
	field := device.Types[0].Data.Fields[0]
	if field.Type.Kind != "ref" || field.Type.Name != "State" || field.Type.RefKind != "enum" {
		t.Errorf("cross file ref = %+v", field.Type)
	}
	if field.Type.RefPack != "public" {
		t.Errorf("cross file ref pack = %q, want public", field.Type.RefPack)
	}
}
