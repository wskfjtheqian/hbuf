package lsp

import (
	"encoding/json"
	"fmt"
	"io"
)

// Options 服务器配置。
type Options struct {
	Name    string // 服务器名(initialize 响应)
	Version string
	Logf    func(format string, args ...interface{}) // 日志(建议写 stderr,不得写 stdout)
}

// Server LSP 服务器。
type Server struct {
	wire        *Wire
	view        *View
	initialized bool
	shutdown    bool
	opts        Options
}

// Serve 运行 LSP 主循环,返回进程退出码(0 正常)。
// reader/writer 即 stdio 通道;所有协议输出仅写 writer。
func Serve(reader io.Reader, writer io.Writer, opts Options) int {
	s := &Server{
		wire: NewWire(reader, writer),
		view: NewView(),
		opts: opts,
	}
	for {
		msg, err := s.wire.Read()
		if err != nil {
			if err == io.EOF {
				return s.exitCode()
			}
			s.logf("read message: %v", err)
			return 1
		}
		if msg.IsRequest() {
			result, rpcErr := s.dispatch(msg.Method, msg.Params)
			var resp *Message
			if rpcErr != nil {
				resp = newErrorResponse(msg.ID, rpcErr.Code, rpcErr.Message)
			} else {
				resp = newResponse(msg.ID, result)
			}
			if err := s.wire.Write(resp); err != nil {
				s.logf("write response: %v", err)
				return 1
			}
			continue
		}
		if msg.IsNotification() {
			if msg.Method == "exit" {
				return s.exitCode()
			}
			s.dispatch(msg.Method, msg.Params)
			continue
		}
		s.logf("unexpected message: %+v", msg)
	}
}

func (s *Server) exitCode() int {
	if s.shutdown {
		return 0
	}
	return 1
}

func (s *Server) logf(format string, args ...interface{}) {
	if s.opts.Logf != nil {
		s.opts.Logf(format, args...)
	}
}

// notify 向客户端发送通知。
func (s *Server) notify(method string, params interface{}) {
	data, err := json.Marshal(params)
	if err != nil {
		s.logf("marshal notify %s: %v", method, err)
		return
	}
	msg := &Message{JSONRPC: "2.0", Method: method, Params: data}
	if err := s.wire.Write(msg); err != nil {
		s.logf("write notify %s: %v", method, err)
	}
}

// dispatch 方法分发。返回 (result, rpcError)。
func (s *Server) dispatch(method string, params json.RawMessage) (interface{}, *RPCError) {
	switch method {
	case "initialize":
		if err := json.Unmarshal(params, &struct{}{}); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		s.initialized = true
		name, version := s.opts.Name, s.opts.Version
		if name == "" {
			name = "hbuf-lsp"
		}
		return InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync:           TextDocumentSyncFull,
				HoverProvider:              true,
				DefinitionProvider:         true,
				ReferencesProvider:         true,
				DocumentSymbolProvider:     true,
				DocumentFormattingProvider: true,
				CompletionProvider: &CompletionOptions{
					TriggerCharacters: []string{".", ":"},
				},
			},
			ServerInfo: &ServerInfo{Name: name, Version: version},
		}, nil

	case "initialized":
		return nil, nil

	case "shutdown":
		s.shutdown = true
		return nil, nil

	case "textDocument/didOpen":
		var p DidOpenParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		td := p.TextDocument
		s.view.Open(td.URI, td.LanguageID, td.Version, td.Text)
		s.publishAllDiagnostics()
		return nil, nil

	case "textDocument/didChange":
		var p DidChangeParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		text := ""
		if n := len(p.ContentChanges); n > 0 {
			text = p.ContentChanges[n-1].Text // 全文同步:最后一个变更即全文
		}
		s.view.Change(p.TextDocument.URI, p.TextDocument.Version, text)
		s.publishAllDiagnostics()
		return nil, nil

	case "textDocument/didClose":
		var p DidCloseParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		s.view.Close(p.TextDocument.URI)
		s.sendDiagnostics(p.TextDocument.URI, nil)
		return nil, nil

	case "textDocument/didSave":
		var p DidSaveParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		s.publishAllDiagnostics()
		return nil, nil

	case "textDocument/completion":
		var p CompletionParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		doc := s.view.Doc(p.TextDocument.URI)
		if doc == nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: "document not open: " + p.TextDocument.URI}
		}
		return s.completion(doc, p.Position), nil

	case "textDocument/definition":
		var p DefinitionParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		doc := s.view.Doc(p.TextDocument.URI)
		if doc == nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: "document not open: " + p.TextDocument.URI}
		}
		return s.definition(doc, p.Position), nil

	case "textDocument/references":
		var p ReferenceParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		doc := s.view.Doc(p.TextDocument.URI)
		if doc == nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: "document not open: " + p.TextDocument.URI}
		}
		return s.references(doc, p.Position, p.Context.IncludeDeclaration), nil

	case "textDocument/hover":
		var p HoverParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		doc := s.view.Doc(p.TextDocument.URI)
		if doc == nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: "document not open: " + p.TextDocument.URI}
		}
		return s.hover(doc, p.Position), nil

	case "textDocument/documentSymbol":
		var p DocumentSymbolParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		doc := s.view.Doc(p.TextDocument.URI)
		if doc == nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: "document not open: " + p.TextDocument.URI}
		}
		return s.documentSymbols(doc), nil

	case "textDocument/formatting":
		var p DocumentFormattingParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: err.Error()}
		}
		doc := s.view.Doc(p.TextDocument.URI)
		if doc == nil {
			return nil, &RPCError{Code: ErrInvalidParams, Message: "document not open: " + p.TextDocument.URI}
		}
		edits, err := s.formatting(doc)
		if err != nil {
			return nil, &RPCError{Code: ErrInternalError, Message: err.Error()}
		}
		return edits, nil

	default:
		return nil, &RPCError{Code: ErrMethodNotFound, Message: fmt.Sprintf("method not found: %s", method)}
	}
}
