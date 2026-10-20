// Package sdk 提供插件侧的通用骨架:JSON over stdio 协议循环与协议校验。
// 插件进程从 stdin 读取编译请求,调用 handler 生成代码,把结果写到 stdout。
// 核心编译器(hbuf)负责启动插件进程、传递请求与落盘生成文件。
package sdk

import (
	"encoding/json"
	"fmt"
	"hbuf/plugin/core/wire"
	"io"
	"os"
)

// Handler 插件生成逻辑:接收编译请求,返回生成文件列表。
type Handler func(req *wire.Request) (*wire.Response, error)

// Run 运行插件主循环:从 stdin 读取请求,结果写入 stdout。
// 任何失败都会输出错误到 stderr 并以非零状态退出,便于核心编译器诊断。
func Run(handler Handler) {
	if err := RunIO(os.Stdin, os.Stdout, handler); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// RunIO 可注入读写流的 Run 实现,便于单元测试。
func RunIO(in io.Reader, out io.Writer, handler Handler) error {
	req, err := ReadRequest(in)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	if req.Protocol != wire.ProtocolVersion {
		return fmt.Errorf("unsupported protocol version: %q, expect %q", req.Protocol, wire.ProtocolVersion)
	}
	resp, err := handler(req)
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	if resp == nil {
		resp = &wire.Response{}
	}
	return json.NewEncoder(out).Encode(resp)
}

// ReadRequest 从流中读取并解析一次编译请求。
func ReadRequest(r io.Reader) (*wire.Request, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	req := &wire.Request{}
	if err := json.Unmarshal(data, req); err != nil {
		return nil, err
	}
	return req, nil
}
