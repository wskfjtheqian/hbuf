// Package wire 定义 HBUF 插件协议的中性描述模型(对标 protobuf 的 descriptor.proto)。
// 核心编译器把 AST 转换为 wire 模型后,通过 JSON over stdio 传给插件,
// 插件基于该模型生成目标语言代码。模型是协议的稳定部分,字段只增不改。
package wire

// ProtocolVersion 插件协议版本,核心与插件必须一致
const ProtocolVersion = "hbuf-plugin-v1"

// Request 核心编译器发给插件的一次编译请求(对标 CodeGeneratorRequest)
type Request struct {
	Protocol  string   `json:"protocol"`             // 协议版本,插件必须校验
	Type      string   `json:"type"`                 // 生成类型: go / dart / java / ts
	OutDir    string   `json:"out_dir"`              // 输出根目录(仅参考,核心负责落盘)
	Pack      string   `json:"pack,omitempty"`       // -p 目标包路径
	BuildList string   `json:"build_list,omitempty"` // -l 生成项类型列表
	Options   []string `json:"options,omitempty"`    // 插件参数,如 template=<目录>
	Files     []*File  `json:"files"`                // 全部文件(含类型定义,供跨文件引用)
	Changed   []string `json:"changed"`              // 本次需要重新生成的文件路径
}

// IsChanged 判断文件是否需要(重新)生成
func (r *Request) IsChanged(path string) bool {
	for _, item := range r.Changed {
		if item == path {
			return true
		}
	}
	return false
}

// GetOption 获取插件参数值,如 GetOption("template")
func (r *Request) GetOption(key string) (string, bool) {
	prefix := key + "="
	for _, item := range r.Options {
		if len(item) > len(prefix) && item[:len(prefix)] == prefix {
			return item[len(prefix):], true
		}
	}
	return "", false
}

// Response 插件返回的生成结果(对标 CodeGeneratorResponse)
type Response struct {
	Files []*GeneratedFile `json:"files"`
}

// GeneratedFile 一个待写出的生成文件
type GeneratedFile struct {
	Name    string `json:"name"`    // 相对输出目录的文件名,如 base.data.go
	Content string `json:"content"` // 完整文件内容
}

// File 单个 .hbuf 文件的中性描述
type File struct {
	Path     string            `json:"path"`               // 源文件路径
	Packages map[string]string `json:"packages,omitempty"` // 各语言包名: go/java/ts
	Imports  []string          `json:"imports,omitempty"`  // import 的文件路径
	Types    []*Type           `json:"types,omitempty"`    // 本文件声明的类型
}

// Type 类型声明(Kind 判别字段,避免接口序列化)
type Type struct {
	Kind   string      `json:"kind"` // enum / data / server
	Enum   *EnumType   `json:"enum,omitempty"`
	Data   *DataType   `json:"data,omitempty"`
	Server *ServerType `json:"server,omitempty"`
}

// TypeBase 类型声明公共部分
type TypeBase struct {
	Name    string     `json:"name"`              // 类型名
	Tags    []*Tag     `json:"tags,omitempty"`    // 类型注解
	Extends []*Extends `json:"extends,omitempty"` // 继承列表
	Doc     string     `json:"doc,omitempty"`     // 文档注释
}

// EnumType 枚举声明
type EnumType struct {
	TypeBase
	Items []*EnumItem `json:"items,omitempty"`
}

// EnumItem 枚举项
type EnumItem struct {
	Name string `json:"name"`
	Id   int64  `json:"id"`
	Tags []*Tag `json:"tags,omitempty"`
	Doc  string `json:"doc,omitempty"`
}

// DataType 数据声明
type DataType struct {
	TypeBase
	Fields []*Field `json:"fields,omitempty"`
}

// ServerType 服务声明
type ServerType struct {
	TypeBase
	Methods []*Method `json:"methods,omitempty"`
}

// Method 服务方法
type Method struct {
	Name      string   `json:"name"`
	Id        int64    `json:"id"`
	Result    *VarType `json:"result,omitempty"` // nil 表示广播(无返回值)
	Param     *VarType `json:"param,omitempty"`  // 入参
	ParamName string   `json:"param_name,omitempty"`
	Tags      []*Tag   `json:"tags,omitempty"`
	Doc       string   `json:"doc,omitempty"`
}

// Field 数据字段
type Field struct {
	Name string   `json:"name"`
	Id   int64    `json:"id"`
	Type *VarType `json:"type"`
	Tags []*Tag   `json:"tags,omitempty"`
	Doc  string   `json:"doc,omitempty"`
}

// VarType 统一扁平化的类型描述(替代 ast 的 VarType/ArrayType/MapType/Ident 分支)
type VarType struct {
	Kind    string   `json:"kind"`               // base / ref / array / map
	Name    string   `json:"name,omitempty"`     // kind=base: 基础类型名; kind=ref: 引用类型名
	Empty   bool     `json:"empty"`              // 是否可空
	RefKind string   `json:"ref_kind,omitempty"` // kind=ref: enum / data
	RefPack string   `json:"ref_pack,omitempty"` // kind=ref: 引用类型所在文件的目标语言包名
	Element *VarType `json:"element,omitempty"`  // kind=array: 元素类型
	Key     string   `json:"key,omitempty"`      // kind=map: key 类型名(基础类型或 enum 名)
	Value   *VarType `json:"value,omitempty"`    // kind=map: value 类型
}

// Tag 注解
type Tag struct {
	Name string      `json:"name"`
	KV   []*KeyValue `json:"kv,omitempty"`
}

// KeyValue 注解参数
type KeyValue struct {
	Name   string   `json:"name"`
	Values []string `json:"values,omitempty"`
}

// Extends 继承声明
type Extends struct {
	Name string `json:"name"`
	Id   int64  `json:"id"`
}
