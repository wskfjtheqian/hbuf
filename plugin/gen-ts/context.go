package tsgen

import (
	"fmt"
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"strconv"
	"strings"
)

// FileContext 一个 .hbuf 文件的模板根上下文。
type FileContext struct {
	*Ctx
	Package string         // ts 包名
	Datas   []*DataContext // data 类型列表(按 desc 依赖拓扑排序)
}

// NewFileContext 构建文件级模板上下文。
// 同文件内 data 按 desc 引用关系拓扑排序;环上节点用 NewSyncDescriptor 延迟构建。
func NewFileContext(ctx *Ctx, file *wire.File, exported map[string]bool) *FileContext {
	fc := &FileContext{Ctx: ctx, Package: ctx.Pack}
	for _, od := range orderDatas(ctx, file) {
		fc.Datas = append(fc.Datas, NewDataContext(ctx, od.Data, od.Sync, exported))
	}
	return fc
}

// orderedData 拓扑排序结果。
type orderedData struct {
	Data *wire.DataType
	Sync bool
}

// orderDatas 同文件内 data 按 desc 依赖拓扑排序(被引用的 desc 必须先声明)。
// 存在环(自引用/互引用)时,取声明序最早的环上节点标记 Sync,
// 其 NewDataDescriptor 会被 NewSyncDescriptor 包裹以延迟求值,从而打破循环。
func orderDatas(ctx *Ctx, file *wire.File) []*orderedData {
	var datas []*wire.DataType
	for _, t := range file.Types {
		if t != nil && t.Kind == "data" && t.Data != nil {
			datas = append(datas, t.Data)
		}
	}
	// dependents[A] = {B...}:B 的声明依赖 A(即 A 必须在 B 之前声明)
	dependents := map[string]map[string]bool{}
	degree := map[string]int{}
	for _, d := range datas {
		deps := map[string]bool{}
		for _, e := range d.Extends {
			if e == nil {
				continue
			}
			if t, owner := ctx.index.Find(e.Name); t != nil && t.Kind == "data" && owner != nil && owner.Path == file.Path {
				deps[e.Name] = true
			}
		}
		for _, f := range d.Fields {
			if f == nil {
				continue
			}
			// 自引用也计入依赖:自环同样由环检测标记 Sync(NewSyncDescriptor 延迟构建)
			collectRefs(f.Type, func(name string) {
				if t, owner := ctx.index.Find(name); t != nil && t.Kind == "data" && owner != nil && owner.Path == file.Path {
					deps[name] = true
				}
			})
		}
		for dep := range deps {
			if dependents[dep] == nil {
				dependents[dep] = map[string]bool{}
			}
			dependents[dep][d.Name] = true
			degree[d.Name]++
		}
	}
	remaining := map[string]*wire.DataType{}
	for _, d := range datas {
		remaining[d.Name] = d
	}
	var out []*orderedData
	var queue []string
	for _, d := range datas {
		if degree[d.Name] == 0 {
			queue = append(queue, d.Name)
		}
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if remaining[n] == nil {
			continue
		}
		out = append(out, &orderedData{Data: remaining[n]})
		delete(remaining, n)
		for dname := range dependents[n] {
			if remaining[dname] == nil {
				continue
			}
			degree[dname]--
			if degree[dname] == 0 {
				queue = append(queue, dname)
			}
		}
		if len(queue) == 0 && len(remaining) > 0 {
			// 剩余节点构成环:取声明序最早的节点标记 Sync 并移除
			for _, d := range datas {
				if remaining[d.Name] == nil {
					continue
				}
				out = append(out, &orderedData{Data: d, Sync: true})
				delete(remaining, d.Name)
				for dname := range dependents[d.Name] {
					if remaining[dname] == nil {
						continue
					}
					degree[dname]--
					if degree[dname] == 0 {
						queue = append(queue, dname)
					}
				}
				break
			}
		}
	}
	// 防御:理论不可达
	for _, d := range datas {
		if remaining[d.Name] != nil {
			out = append(out, &orderedData{Data: d, Sync: true})
			delete(remaining, d.Name)
		}
	}
	return out
}

// collectRefs 收集类型树中的 data 引用名(嵌套数组/映射递归)。
func collectRefs(vt *wire.VarType, fn func(name string)) {
	if vt == nil {
		return
	}
	switch vt.Kind {
	case "ref":
		if vt.RefKind == "data" {
			fn(vt.Name)
		}
	case "array":
		collectRefs(vt.Element, fn)
	case "map":
		collectRefs(vt.Value, fn)
	}
}

// exportedDescs 统计被其他文件引用的 data 名(其类与 desc 需要 export)。
func exportedDescs(files []*wire.File, index *TypeIndex) map[string]bool {
	exported := map[string]bool{}
	for _, f := range files {
		if f == nil {
			continue
		}
		for _, t := range f.Types {
			if t == nil || t.Kind != "data" || t.Data == nil {
				continue
			}
			for _, e := range t.Data.Extends {
				if e == nil {
					continue
				}
				if rt, owner := index.Find(e.Name); owner != nil && rt != nil && rt.Kind == "data" && owner.Path != f.Path {
					exported[e.Name] = true
				}
			}
			for _, field := range t.Data.Fields {
				if field == nil {
					continue
				}
				collectRefs(field.Type, func(name string) {
					if rt, owner := index.Find(name); owner != nil && rt != nil && rt.Kind == "data" && owner.Path != f.Path {
						exported[name] = true
					}
				})
			}
		}
	}
	return exported
}

// DataContext data 类型的模板上下文。
type DataContext struct {
	*Ctx
	Name     string           // 类名:XxxTs
	DescName string           // descriptor 单例名:xxxDesc
	DescOpen string           // descriptor 声明开头(含类型注解与构造调用)
	Export   string           // 被其他文件引用时为 "export "
	Sync     bool             // 递归:NewSyncDescriptor 包裹
	Extends  []*ExtendContext // 父类型内嵌字段
	Fields   []*FieldContext
	Doc      string
}

// NewDataContext 构建 data 类型的模板上下文。
func NewDataContext(ctx *Ctx, data *wire.DataType, sync bool, exported map[string]bool) *DataContext {
	dc := &DataContext{
		Ctx:      ctx,
		Name:     tpl.HumpName(data.Name) + "Ts",
		DescName: lowerCamel(data.Name) + "Desc",
		Sync:     sync,
		Doc:      data.Doc,
	}
	if exported[data.Name] {
		dc.Export = "export "
	}
	if sync {
		ctx.Import("NewSyncDescriptor")
		dc.DescOpen = fmt.Sprintf("%sconst %s: HBufDescriptor = NewSyncDescriptor(() => NewDataDescriptor(", dc.Export, dc.DescName)
	} else {
		dc.DescOpen = fmt.Sprintf("%sconst %s: DataDescriptor = NewDataDescriptor(", dc.Export, dc.DescName)
	}
	for _, e := range data.Extends {
		if e == nil {
			continue
		}
		ref := &wire.VarType{Kind: "ref", Name: e.Name, RefKind: "data"}
		ext := &ExtendContext{
			Ctx:     ctx,
			Name:    lowerCamel(e.Name),
			Type:    ctx.refType(ref),
			Default: ctx.tsDefault(ref),
			Id:      strconv.FormatInt(e.Id, 10),
			Desc:    "NewRefDescriptor(" + ctx.refDesc(ref) + ")",
		}
		ctx.Import("NewRefDescriptor")
		dc.Extends = append(dc.Extends, ext)
	}
	for _, f := range data.Fields {
		if f == nil || f.Type == nil {
			continue
		}
		fc := &FieldContext{
			Ctx:     ctx,
			Name:    lowerCamel(f.Name),
			Id:      strconv.FormatInt(f.Id, 10),
			Type:    ctx.tsType(f.Type),
			Default: ctx.tsDefault(f.Type),
			Desc:    ctx.tsDescriptor(f.Type),
			Comment: strings.Trim(strings.ReplaceAll(f.Doc, "\n", " "), " "),
		}
		dc.Fields = append(dc.Fields, fc)
	}
	return dc
}

// FieldContext 数据字段的模板上下文。
type FieldContext struct {
	*Ctx
	Name    string // 小驼峰属性名
	Id      string // id 字符串
	Type    string // TS 类型声明
	Default string // 默认值表达式
	Desc    string // descriptor 构造表达式
	Comment string // 文档注释(已压缩单行)
}

// ExtendContext 父类型内嵌字段的模板上下文。
type ExtendContext struct {
	*Ctx
	Name    string // 内嵌属性名(父类型名小驼峰)
	Type    string // 父类型类名
	Default string // 默认值表达式
	Id      string // 扩展 ID
	Desc    string // descriptor 构造表达式
}
