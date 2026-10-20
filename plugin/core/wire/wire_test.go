package wire

import (
	"encoding/json"
	"testing"
)

func TestRequestIsChanged(t *testing.T) {
	req := &Request{Changed: []string{"/a/b.hbuf", "/a/c.hbuf"}}
	if !req.IsChanged("/a/b.hbuf") {
		t.Error("IsChanged should be true for listed path")
	}
	if req.IsChanged("/a/d.hbuf") {
		t.Error("IsChanged should be false for unlisted path")
	}
	if (&Request{}).IsChanged("/a/b.hbuf") {
		t.Error("IsChanged should be false for empty request")
	}
}

func TestRequestGetOption(t *testing.T) {
	req := &Request{Options: []string{"template=/tmp/tpl", "verbose=true"}}
	if val, ok := req.GetOption("template"); !ok || val != "/tmp/tpl" {
		t.Errorf("GetOption template = %q, %v", val, ok)
	}
	if val, ok := req.GetOption("verbose"); !ok || val != "true" {
		t.Errorf("GetOption verbose = %q, %v", val, ok)
	}
	if _, ok := req.GetOption("missing"); ok {
		t.Error("GetOption missing should be false")
	}
}

// TestResponseJSON 验证 Response 可被 JSON 往返且字段名符合协议
func TestResponseJSON(t *testing.T) {
	resp := &Response{Files: []*GeneratedFile{
		{Name: "base.data.go", Content: "package admin\n"},
	}}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var got Response
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Files) != 1 || got.Files[0].Name != "base.data.go" || got.Files[0].Content != "package admin\n" {
		t.Errorf("round trip mismatch: %+v", got.Files)
	}
}

// TestRequestJSON 验证 Request 可被 JSON 往返
func TestRequestJSON(t *testing.T) {
	req := &Request{
		Protocol: ProtocolVersion,
		Type:     "go",
		OutDir:   "/out",
		Pack:     "admin",
		Files: []*File{{
			Path:     "/a/base.hbuf",
			Packages: map[string]string{"go": "admin"},
			Types: []*Type{{
				Kind: "data",
				Data: &DataType{
					TypeBase: TypeBase{Name: "Base"},
					Fields: []*Field{{
						Name: "id",
						Id:   0,
						Type: &VarType{Kind: "base", Name: "int64"},
					}},
				},
			}},
		}},
		Changed: []string{"/a/base.hbuf"},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got Request
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Protocol != ProtocolVersion || got.Type != "go" {
		t.Errorf("round trip mismatch: %+v", got)
	}
	if len(got.Files) != 1 || len(got.Files[0].Types) != 1 {
		t.Fatalf("round trip files mismatch: %+v", got.Files)
	}
	typ := got.Files[0].Types[0]
	if typ.Kind != "data" || typ.Data.Name != "Base" || len(typ.Data.Fields) != 1 {
		t.Errorf("round trip type mismatch: %+v", typ)
	}
	if typ.Data.Fields[0].Type.Kind != "base" || typ.Data.Fields[0].Type.Name != "int64" {
		t.Errorf("round trip field mismatch: %+v", typ.Data.Fields[0])
	}
}
