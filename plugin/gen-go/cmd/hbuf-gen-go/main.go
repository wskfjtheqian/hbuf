// hbuf-gen-go 是 HBUF 的 Go 代码生成插件。
//
// 协议:JSON over stdio(hbuf-plugin-v1)。从 stdin 读取编译请求,
// 基于模板渲染 Go 代码,结果写到 stdout,由核心编译器统一落盘。
//
// 模板查找顺序:外部自定义模板目录(template=<目录> 插件参数)→ 内置默认模板。
package main

import (
	"hbuf/plugin/core/sdk"
	gogen "hbuf/plugin/gen-go"
)

func main() {
	sdk.Run(gogen.Generate)
}
