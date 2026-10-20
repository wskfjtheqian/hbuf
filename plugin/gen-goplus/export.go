package goplus

import (
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"sort"
	"strconv"
	"strings"
)

// ExportContext export 输出模板的 data 类型上下文。
type ExportContext struct {
	Name    string
	Keys    []*ExportKeyContext
	All     []*ExportFieldContext
	HasDate bool
}

// ExportKeyContext 单个 filter 分组。
type ExportKeyContext struct {
	Key    string
	Fields []*ExportFieldContext
}

// ExportFieldContext 单个导出字段:Name 为驼峰名,
// Line 为 ExportData 中的赋值行(含缩进与空值保护,预渲染)。
type ExportFieldContext struct {
	Name string
	Line string
}

// buildExport 为带 export 注解的 data 类型构建 export 模板上下文。
// 与进程内 printExportCode 一致:无 export 注解或字段为空时返回 nil。
func (c *Ctx) buildExport(imp Imports, data *wire.DataType) *ExportContext {
	if getExport(data.Tags) == nil {
		return nil
	}

	maps := make(map[string][]*wire.Field)
	var lists []*wire.Field
	_ = c.enumField(data, func(field *wire.Field, data *wire.DataType) error {
		export := getExport(field.Tags)
		if export == nil {
			return nil
		}
		lists = append(lists, field)
		for _, key := range export.Filter {
			maps[key] = append(maps[key], field)
		}
		return nil
	})

	if len(maps) == 0 && len(lists) == 0 {
		return nil
	}

	keys := make([]string, 0, len(maps))
	for key := range maps {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ec := &ExportContext{Name: tpl.HumpName(data.Name)}
	for _, key := range keys {
		kc := &ExportKeyContext{Key: key}
		for i, item := range maps[key] {
			kc.Fields = append(kc.Fields, c.buildExportField(imp, i, item))
		}
		ec.Keys = append(ec.Keys, kc)
	}
	for i, item := range lists {
		ec.All = append(ec.All, c.buildExportField(imp, i, item))
		if isDate(item.Type) {
			ec.HasDate = true
		}
	}
	if ec.HasDate {
		imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/htime", "")
	}
	return ec
}

// buildExportField 渲染单个字段的数据行,与进程内 printExportDataItemCode 一致。
func (c *Ctx) buildExportField(imp Imports, i int, item *wire.Field) *ExportFieldContext {
	name := tpl.HumpName(item.Name)
	isNilType := isNil(item.Type)
	var sb strings.Builder
	tab := "\t\t"
	if isNilType {
		sb.WriteString("\t\tif g." + name + " != nil {\n")
		tab = "\t\t\t"
	}

	if isArray(item.Type) || isMap(item.Type) {
		imp.Import("encoding/json", "")
		sb.WriteString(tab + "bytes, err := json.Marshal(g.Get" + name + "())\n")
		sb.WriteString(tab + "if err!= nil {\n")
		sb.WriteString(tab + "\treturn nil, err\n")
		sb.WriteString(tab + "}\n")
		sb.WriteString(tab + "list[" + strconv.Itoa(i) + "] = string(bytes)\n")
	} else if isEnum(item.Type) {
		sb.WriteString(tab + "list[" + strconv.Itoa(i) + "] = g.Get" + name + "().ToName()\n")
	} else if isDate(item.Type) {
		sb.WriteString(tab + "list[" + strconv.Itoa(i) + "] = g.Get" + name + "().In(loc)\n")
	} else {
		sb.WriteString(tab + "list[" + strconv.Itoa(i) + "] = g.Get" + name + "()\n")
	}
	if isNilType {
		sb.WriteString("\t\t}\n")
	}
	return &ExportFieldContext{Name: name, Line: strings.TrimSuffix(sb.String(), "\n")}
}
