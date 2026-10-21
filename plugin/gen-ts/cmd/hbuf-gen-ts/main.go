// hbuf-gen-ts 插件入口:通过 JSON over stdio 接收核心编译器请求,
// 按 test/ts/hbuf-ts-codegen.md 规范生成 TypeScript 数据类代码。
package main

import (
	"hbuf/plugin/core/sdk"
	tsgen "hbuf/plugin/gen-ts"
)

func main() {
	sdk.Run(tsgen.Generate)
}
