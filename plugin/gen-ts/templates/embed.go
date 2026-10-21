// Package templates 内置 gen-ts 默认模板(通过 //go:embed 编译进插件二进制)。
// 外部自定义模板目录(插件参数 template=<目录>)同名文件覆盖默认模板。
package templates

import "embed"

//go:embed data.tmpl enum.tmpl server.tmpl
var FS embed.FS
