package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"hbuf/plugin/core/wire"
	"strings"
	"testing"
)

func TestRunIOOk(t *testing.T) {
	req := &wire.Request{
		Protocol: wire.ProtocolVersion,
		Type:     "go",
		Changed:  []string{"/a/base.hbuf"},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	handler := func(req *wire.Request) (*wire.Response, error) {
		if req.Type != "go" {
			t.Errorf("handler got type %q", req.Type)
		}
		return &wire.Response{Files: []*wire.GeneratedFile{
			{Name: "base.data.go", Content: "package admin\n"},
		}}, nil
	}
	if err := RunIO(bytes.NewReader(data), &out, handler); err != nil {
		t.Fatalf("RunIO: %v", err)
	}

	resp := &wire.Response{}
	if err := json.Unmarshal(out.Bytes(), resp); err != nil {
		t.Fatalf("response parse: %v", err)
	}
	if len(resp.Files) != 1 || resp.Files[0].Name != "base.data.go" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestRunIOBadProtocol(t *testing.T) {
	req := &wire.Request{Protocol: "unknown"}
	data, _ := json.Marshal(req)
	err := RunIO(bytes.NewReader(data), &bytes.Buffer{}, func(req *wire.Request) (*wire.Response, error) {
		return nil, nil
	})
	if err == nil || !strings.Contains(err.Error(), "protocol") {
		t.Errorf("expect protocol error, got %v", err)
	}
}

func TestRunIOHandlerError(t *testing.T) {
	req := &wire.Request{Protocol: wire.ProtocolVersion}
	data, _ := json.Marshal(req)
	err := RunIO(bytes.NewReader(data), &bytes.Buffer{}, func(req *wire.Request) (*wire.Response, error) {
		return nil, errors.New("boom")
	})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("expect handler error, got %v", err)
	}
}

func TestReadRequestBadJSON(t *testing.T) {
	if _, err := ReadRequest(bytes.NewReader([]byte("{"))); err == nil {
		t.Error("expect parse error for bad json")
	}
}
