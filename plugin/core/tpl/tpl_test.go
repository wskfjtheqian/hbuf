package tpl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func defaultFS() fstest.MapFS {
	return fstest.MapFS{
		"data.tmpl":    {Data: []byte(`type {{.Name}} struct{}`)},
		"header.tmpl":  {Data: []byte(`// header`)},
		"compose.tmpl": {Data: []byte(`{{template "header.tmpl" .}} {{.Name}}`)},
	}
}

func TestExecuteDefault(t *testing.T) {
	e := New(defaultFS(), nil, Funcs())
	out, err := e.Execute("data.tmpl", map[string]string{"Name": "User"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "type User struct{}" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestCustomOverridesDefault(t *testing.T) {
	custom := fstest.MapFS{
		"data.tmpl": {Data: []byte(`custom {{.Name}}`)},
	}
	e := New(defaultFS(), custom, Funcs())
	out, err := e.Execute("data.tmpl", map[string]string{"Name": "User"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "custom User" {
		t.Errorf("expect custom override, got %q", out)
	}
	// 未覆盖的模板仍来自默认
	out, err = e.Execute("header.tmpl", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "// header" {
		t.Errorf("expect default header, got %q", out)
	}
}

func TestTemplateCompose(t *testing.T) {
	e := New(defaultFS(), nil, Funcs())
	out, err := e.Execute("compose.tmpl", map[string]string{"Name": "User"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "// header User" {
		t.Errorf("unexpected compose output: %q", out)
	}
}

func TestMissingTemplate(t *testing.T) {
	e := New(defaultFS(), nil, Funcs())
	if _, err := e.Execute("missing.tmpl", nil); err == nil {
		t.Error("expect error for missing template")
	}
}

func TestParseError(t *testing.T) {
	e := New(fstest.MapFS{"bad.tmpl": {Data: []byte(`{{.Name`)}}, nil, Funcs())
	if _, err := e.Execute("bad.tmpl", nil); err == nil {
		t.Error("expect parse error")
	}
}

func TestFuncs(t *testing.T) {
	cases := []struct {
		fn   func(string) string
		in   string
		want string
	}{
		{HumpName, "user_name", "UserName"},
		{LowerFirstName, "user_name", "userName"},
		{UnderlineName, "UserName", "user_name"},
		{AllUpperName, "UserName", "USER_NAME"},
		{MiddleLine, "UserName", "user-name"},
		{LowerFirst, "UserName", "userName"},
	}
	for _, c := range cases {
		if got := c.fn(c.in); got != c.want {
			t.Errorf("%q -> %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEngineWithDirFS(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data.tmpl"), []byte(`dir {{.Name}}`), 0644); err != nil {
		t.Fatal(err)
	}
	e := New(defaultFS(), os.DirFS(dir), Funcs())
	out, err := e.Execute("data.tmpl", map[string]string{"Name": "User"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "dir User") {
		t.Errorf("expect dir override, got %q", out)
	}
	names, err := e.Names()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range names {
		if n == "data.tmpl" {
			found = true
		}
	}
	if !found {
		t.Errorf("Names should include data.tmpl, got %v", names)
	}
}
