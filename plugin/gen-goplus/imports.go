package goplus

import (
	"sort"
	"strings"
)

// Imports 输出文件的 import 集合:路径 -> 别名(空串表示无别名)。
// 构建期各动作上下文按需登记,渲染后统一生成 header。
type Imports map[string]string

// Import 登记 import 路径,alias 为空表示无别名。
func (m Imports) Import(path, alias string) {
	if path == "" {
		return
	}
	if _, ok := m[path]; !ok {
		m[path] = alias
	}
}

// AddImports 合并其他 import 集合。
func (m Imports) AddImports(other Imports) {
	for path, alias := range other {
		m.Import(path, alias)
	}
}

// header 生成文件头:package 声明 + import 块。
// 与进程内后端一致:无别名 import 在前,有别名在后,中间空行分隔。
func header(pack string, imps Imports) string {
	var sb strings.Builder
	sb.WriteString("package " + pack + "\n\n")
	if len(imps) > 0 {
		keys := make([]string, 0, len(imps))
		for path := range imps {
			keys = append(keys, path)
		}
		sort.Strings(keys)
		sb.WriteString("import (\n")
		for _, path := range keys {
			if imps[path] == "" {
				sb.WriteString("\t\"" + path + "\"\n")
			}
		}
		hasAlias := false
		for _, path := range keys {
			if imps[path] != "" {
				hasAlias = true
				break
			}
		}
		if hasAlias {
			sb.WriteString("\n")
			for _, path := range keys {
				if imps[path] != "" {
					sb.WriteString("\t" + imps[path] + " \"" + path + "\"\n")
				}
			}
		}
		sb.WriteString(")\n\n")
	}
	return sb.String()
}
