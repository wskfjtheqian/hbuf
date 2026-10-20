// hbuf-lsp 是 HBUF 语言服务器(Language Server Protocol,JSON-RPC over stdio)。
// 对标 gopls:诊断、补全、跳转定义/引用、悬停、文档符号、格式化。
package main

import (
	"flag"
	"fmt"
	"hbuf/plugin/lsp"
	"log"
	"os"
)

var (
	version = flag.Bool("v", false, "print version and exit")
	debug   = flag.Bool("log", false, "write logs to stderr (never stdout)")
)

const lspVersion = "0.1.0"

func main() {
	flag.Parse()
	if *version {
		fmt.Println("hbuf-lsp " + lspVersion)
		return
	}

	var logf func(string, ...interface{})
	if *debug {
		logger := log.New(os.Stderr, "[hbuf-lsp] ", log.LstdFlags)
		logf = func(format string, args ...interface{}) {
			logger.Printf(format, args...)
		}
	}

	os.Exit(lsp.Serve(os.Stdin, os.Stdout, lsp.Options{
		Name:    "hbuf-lsp",
		Version: lspVersion,
		Logf:    logf,
	}))
}
