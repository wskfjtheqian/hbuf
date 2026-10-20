package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// JSON-RPC 2.0 标准错误码。
const (
	ErrParseError     = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternalError  = -32603
)

// RPCError JSON-RPC 错误对象。
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Message 一条 JSON-RPC 消息(请求/通知/响应共用)。
// 请求:Method != "" 且 ID != nil;通知:Method != "" 且 ID == nil。
type Message struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *RPCError        `json:"error,omitempty"`
}

// IsRequest 判断是否为请求(需要响应)。
func (m *Message) IsRequest() bool {
	return m.Method != "" && m.ID != nil
}

// IsNotification 判断是否为通知(无需响应)。
func (m *Message) IsNotification() bool {
	return m.Method != "" && m.ID == nil
}

// IsResponse 判断是否为响应(服务器发起的请求的回复,当前实现未发起请求)。
func (m *Message) IsResponse() bool {
	return m.Method == "" && m.ID != nil
}

// newResponse 构造成功响应。
func newResponse(id *json.RawMessage, result interface{}) *Message {
	data, _ := json.Marshal(result)
	raw := json.RawMessage(data)
	return &Message{JSONRPC: "2.0", ID: id, Result: raw}
}

// newErrorResponse 构造错误响应。
func newErrorResponse(id *json.RawMessage, code int, msg string) *Message {
	return &Message{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg}}
}

// Wire LSP 消息帧通道(Content-Length 头 + JSON body)。
type Wire struct {
	r *bufio.Reader
	w io.Writer
}

func NewWire(r io.Reader, w io.Writer) *Wire {
	return &Wire{r: bufio.NewReader(r), w: w}
}

// Read 读取一条消息。解析失败返回 ParseError 语义错误。
func (w *Wire) Read() (*Message, error) {
	var length int = -1
	for {
		line, err := w.r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if i := strings.Index(line, ":"); i > 0 {
			key := strings.TrimSpace(line[:i])
			value := strings.TrimSpace(line[i+1:])
			if strings.EqualFold(key, "Content-Length") {
				n, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("invalid Content-Length: %q", value)
				}
				length = n
			}
		}
	}
	if length < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(w.r, body); err != nil {
		return nil, err
	}
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	if msg.JSONRPC != "" && msg.JSONRPC != "2.0" {
		return nil, fmt.Errorf("unsupported jsonrpc version %q", msg.JSONRPC)
	}
	return &msg, nil
}

// Write 写入一条消息。
func (w *Wire) Write(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w.w, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
		return err
	}
	_, err = w.w.Write(data)
	return err
}
