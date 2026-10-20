// Package templates 内置默认模板,编译进 hbuf-gen-go 二进制。
// 外部自定义模板目录(插件参数 template=<目录>)中同名文件可覆盖默认模板。
package templates

import "embed"

//go:embed data.tmpl
//go:embed enum.tmpl
//go:embed server.tmpl
var FS embed.FS
