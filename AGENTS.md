# AGENTS.md — HBUF 项目 AI 开发指南

## 项目概述

HBUF 是一个**跨语言代码生成工具**(类似 protobuf/thrift,但功能更全),用 **Go** 编写。
它以自定义 IDL 语言(`.hbuf` 文件)为输入,生成多种目标语言的数据结构、序列化、RPC 接口、
SQL 数据库操作、表单验证、UI 组件和多语言文本代码。

**生成能力一览**(详见 `README.md`):
| 能力 | golang | dart | java | typescript |
|---|---|---|---|---|
| 数据结构 + JSON 序列化 | ✅ | ✅ | ✅ | ✅ |
| RPC 接口(函数) | ✅ | ✅ | ✅ | ✅ |
| SQL 操作(Insert/Update/Del/List…) | ✅ | - | - | - |
| 正则表单验证 | ✅ | ✅ | - | ✅ |
| Flutter 表格/表单 UI | - | ✅ | - | ✅(tsx) |
| 多语言文本 | - | ✅ | - | ✅ |

**各语言运行时库**(生成代码依赖它们,均为独立仓库):
- golang: `github.com/wskfjtheqian/hbuf_golang`
- dart: `github.com/wskfjtheqian/hbuf_dart`
- java: `github.com/wskfjtheqian/hbuf_java`
- flutter: `github.com/wskfjtheqian/hbuf_flutter`
- typescript: `github.com/wskfjtheqian/hbuf_ts`
- IDE 插件源码: `github.com/wskfjtheqian/hbuf_idea`

## 技术栈

- **Go 1.27**(`go.mod`: module `hbuf`),纯标准库实现,**无第三方依赖**
- 词法/语法分析器架构仿 Go 编译器(`go/scanner`、`go/parser`、`go/ast` 风格)
- 目标语言代码生成分两条路径并存:
  - 进程内后端(`pkg/golang` 等):字符串拼接(兼容期)
  - 外部插件(见下文插件架构):JSON over stdio + text/template 模板

## 构建与运行

```bash
# 编译(交叉编译脚本见 build.sh / build.bat)
go build -ldflags "-X main.version=<version>" -o bin/hbuf ./cmd
# 运行测试(根包 build_test.go 含跨平台交叉编译,较慢,可用 -short 跳过)
go test ./...
go test -short ./...
go test ./plugin/... ./pkg/build/

# CLI 用法
hbuf -i <输入目录或通配符,如 ./test/*.hbuf> \
     -o <输出目录> \
     -t <构建类型: go|dart|java|ts> \
     -p <目标包名,如 admin> \
     [-l <生成项类型列表,如 "DS">] \
     [-a 全量构建,跳过增量检测] \
     [-plugin hbuf-gen-go=/path/to/binary] \
     [-template /path/to/custom/templates] \
     [-v 显示版本]
```

## 架构与目录结构

```
cmd/main.go            CLI 入口:注册 4 种构建类型,解析 flag,调用 build.BuildWithPlugin
cmd/hbuf-lsp/         语言服务器入口(-v/-log flags,日志写 stderr)
pkg/
├── scanner/           词法分析(仿 go/scanner,含 ErrorList)
├── token/             Token 定义、FileSet、Pos 位置管理
├── parser/            语法分析(递归下降,支持 Trace 调试模式、import 递归解析)
├── ast/               AST 节点定义、Scope/Object 符号表、Package
├── build/             ★ 构建编排核心
│   ├── build.go       Build/BuildWithPlugin 主流程、BuildType 注册、BaseType 常量、
│   │                  标签/字段/方法遍历工具、命名转换工具
│   ├── plugin.go      插件协议客户端:FindPlugin(发现)、CallPlugin(JSON over stdio)、
│   │                  WriteGeneratedFiles(统一落盘 + 目录逃逸防护)
│   ├── writer.go      Writer:代码输出缓冲 + import 管理 + 多语言文本
│   ├── hash.go        FileHash:基于 MD5 + .build.json 的增量构建检测
│   └── ...
├── golang/            Go 后端(data/enum/server/database/verify/encoder/export/mq)
├── dart/              Dart 后端(data/enum/server/ui/verify)
├── java/              Java 后端(data/enum/server)
├── typescript/        TS 后端(data/enum/server/lang/ui.tsx/verify)
└── format/            .hbuf 格式化器(AST → 源码,round-trip 幂等)
plugin/                ★ 新插件化架构(详见 plugin/README.md)
├── core/wire/         中性描述模型(Request/Response/File/Type 等)+ AST→wire 转换
├── core/sdk/          插件侧 SDK(stdio 主循环、协议版本校验)
├── core/tpl/          模板引擎(内置模板 embed + 外部覆盖 + 执行缓存)
├── lsp/               ★ HBUF 语言服务(JSON-RPC over stdio,纯标准库,对标 gopls)
│   ├── protocol.go / jsonrpc.go   LSP 类型与 JSON-RPC 2.0 帧
│   ├── server.go       Serve 主循环 + 方法分发(initialize/didOpen/...)
│   ├── view.go         文档视图 + import 依赖图解析(内存覆盖优先,支持未保存文档)
│   ├── diagnostics.go  诊断:语法错误 + build.CheckPackage 同源校验
│   ├── completion.go / definition.go / hover.go / symbols.go / format.go
│   ├── semantic_tokens.go  语义高亮(scanner 基础 + AST 语义,delta 编码)
│   ├── code_action.go  代码操作(source.fixAll 格式化、未使用 import 删除)
│   ├── inlay_hint.go   内联提示(extends 成员摘要、跨文件类型来源)
│   └── workspace_symbol.go 工作区符号(workspaceFolders 扫描 + 符号搜索)
└── gen-go/            Go 插件(已完成):funcs.go(类型映射)、context.go(模板上下文)、
                       templates/*.tmpl、cmd/hbuf-gen-go(入口)
doc/                   hbuf.md(IDL 语法)、hbuf_db.md(数据库注解)、serialize.md、
                       refactor_plan.md(插件化改造计划)
```

### 数据流(核心处理流程)

```
.hbuf 源文件 → scanner 词法 → parser 语法(AST) → build.checkFiles 类型校验
    → FileHash 增量检测(MD5 + .build.json,import 变化级联重编)
    → 收集变更文件列表 changed
    → 按 -t 查找插件(FindPlugin:显式路径 → PATH 查找 hbuf-gen-<type>)
        ├─ 找到:AST→wire 转换 → JSON over stdio 调用插件一次
        │       (请求携带全部文件 + changed,插件无状态)
        │       → 响应为文件列表,核心统一落盘(WriteGeneratedFiles)+ FileHash.Save
        └─ 未找到:进程内回退,遍历 changed 调用后端 Build(file, fset, param)
```

### 插件协议(新增语言生成器必看)

- 协议名 `hbuf-plugin-v1`:JSON over stdio,每次编译调用插件一次,插件无状态
- 核心请求 `wire.Request{Protocol,Type,OutDir,Pack,BuildList,Options,Files,Changed}`,
  插件响应 `wire.Response{Files []*wire.GeneratedFile{Name,Content}}`
- 插件命名 `hbuf-gen-<type>`,核心在 PATH 查找;`-plugin hbuf-gen-go=/path` 显式指定
- 核心负责落盘与增量,插件只负责生成内容
- 插件侧可用 `plugin/core/sdk` 写主循环;第三方可用任意语言实现同一 JSON 协议
- Go 插件生成基于模板:内置模板 `//go:embed` 打包,`-template <目录>` 下同名文件覆盖
- 模板输出后经 `go/format` 自动格式化,文件名约定 `<包名点分转目录>/<name>.data.go` 等

### 语言服务(LSP)要点

- 协议输出仅写 stdout,日志(-log)写 stderr;错误绝不可调 `scanner.ErrorList.Error()`
  (该方法在多个错误时会 fmt.Println 污染协议流,必须遍历 ErrorList 项逐个转换)
- 位置转换:`token.Position`(1-based 行/字节列)→ LSP `Position`(0-based 行/UTF-16 列),
  见 `toPosition`/`toOffset`;EOF 位置 token 层 lines 表不含尾随空行,需特判映射到末行行首
- 诊断与命令行同源:`build.CheckPackage(fset, pkg)` 复用 `checkFiles` 校验规则;
  存在语法错误时跳过类型校验,避免残缺 AST 误报
- 文档解析:`view.parse` 递归加载 import 图,打开文档内存内容优先;
  import 路径归一为绝对路径并回写 spec.Path.Value(与 ParseDir 一致,供跨文件校验);
  格式化(textDocument/formatting)必须独立解析,避免绝对路径写入输出
- 同一 Ident 可能被遍历两次(TypeSpec.Name 与 DataType/EnumType 内嵌 Name),
  references 收集需按 (Pos,End) 去重
- 语义高亮:关键字/注释/字符串/数字由 scanner 提供(语法错误文档也能尽量着色),
  AST 只补语义;跨文件 import 类型引用的 Ident.Obj 为 nil(parser 单文件解析),
  需按名字查 import 文件声明类别;跨行 token(块注释)需按行拆分;delta 编码同 line 0 时列相对
- codeAction:view.loadFile 已把 spec.Path.Value 改写为无引号绝对路径,
  还原 import 字面量时需对 strconv.Unquote 失败容错(直接返回原值)
- workspace/symbol:扫描根目录来自 initialize 的 workspaceFolders/rootUri,
  未提供时回退为打开文档目录;跳过隐藏目录,上限 1000 文件、结果 100 条

### 进程内语言后端统一接口(兼容期)

每个后端包实现同名函数并通过注册接入:
```go
func Build(file *ast.File, fset *token.FileSet, param *build.Param) error
```
在 `cmd/main.go` 中 `build.AddBuildType("go", golang.Build)` 注册。
各后端内部结构高度一致(可复制任一后端改造):
1. 定义 `_types map[build.BaseType]string` 基础类型映射表(如 dart 的 int→`int`、Int64→`Int64`)
2. 定义 XxxWriter 结构,持有多个 `build.Writer`(每个输出文件一个)
3. `Builder.Node()` 遍历 `file.Specs`,对 `ast.EnumType / ast.DataType / ast.ServerType` 分别生成代码
4. 用 `build.GetTag()` / `build.GetKeyValue()` 读取注解标签控制生成行为
5. 输出文件名约定:go 为 `{name}.data.go`/`.enum.go`/`.server.go`/`.database.go`/`.verify.go`/`.encoder.go`/`.export.go`/`.mq.go`;dart 为 `.data.dart`/`.enum.dart`/`.server.dart`/`.ui.dart`/`.verify.dart`;ts 为 `.data.ts`/`.enum.ts`/`.server.ts`/`.lang.ts`/`.ui.tsx`/`.verify.ts`;java 为 `XxxData.java`/`XxxEnum.java`/`XxxServer.java`

## .hbuf IDL 语言要点(详见 doc/hbuf.md、doc/hbuf_db.md)

```hbuf
package go = "admin"        // 按语言指定包名

enum Status { player = 0; stop = 1; }

data Base { int64 id = 1; string name = 2; }   // data 支持多继承
data Student:Base = 2 { int32? no = 3; }        // ? 表示可空,extends 带唯一 ID

server BaseServer {
    Student GetStudent(Student req) = 1;        // 方法参数只能是 data 类型
    void SendMessage(Message message) = 2;      // void 前缀:广播式方法
}

data User = 3 {
    [verify: regex="..."];                      // 注解独立分号结尾
    string phone = 4;
}
```

- 基础类型:`int8..int64`、`uint8..uint64`、`bool`、`float`、`double`、`decimal`、`string`、`date`、`bytes`;方法返回值另有 `void`(广播)、`stream`(流式)
- 复合类型:`数组[]`(后置)、`Map<Key>`(key 仅限基础类型和 enum)
- 每个字段/类型/方法必须带唯一整数 ID(字段与方法 ID 必须 >= 1,enum 项从 0 开始)
- 注解 tag 语法:`[tagName: k="v" k2="v2"];`(键值对空格分隔,值逗号分隔)
- 数据库注解(db):index、key、force、insert/inserts/update/set/del/get/list/map/count、table、rm、where、offset、limit、order、group、converter

## 代码约定

- 注释语言:中文;标识符沿用原语言
- `build` 包是共享工具层,**公共逻辑放这里,勿在语言后端重复实现**
- 命名转换统一用 build 包工具:`StringToHumpName`(驼峰)、`StringToUnderlineName`(下划线)、`StringToAllUpper`、`StringToMiddleLine`(中横线)、`StringToFirstLower`
- 类型判断统一用 `build.IsNumber/IsBool/IsArray/IsMap/IsEnum/IsNil/GetBaseType`
- 字段/方法遍历用 `build.EnumField/EnumMethod`(自动处理多继承去重)
- 新增 AST 节点需同步各语言后端的 `printTypeSpec` switch 与 `plugin/core/wire/convert.go`
- 输出文件在 `test/out/`(已被 .gitignore 忽略)

## 常见任务指引

1. **给某语言新增 SQL 生成**:参考 `pkg/golang/database.go` 的注解驱动模式
2. **新增语言生成器(如 C#)**:实现 `hbuf-gen-<type>` 插件——可复制 `plugin/gen-go`,用任意语言实现 hbuf-plugin-v1 协议,经 wire 模型 + 模板生成(优先于进程内后端)
3. **扩展 IDL 语法**:改 `pkg/token/token.go` → `pkg/scanner` → `pkg/parser` → `pkg/ast` → 各后端 + `plugin/core/wire/convert.go`
4. **修改序列化方案**:见 `pkg/golang/data.go`(JSON/二进制)与 `doc/serialize.md`
5. **验证 CLI 行为**:`build_test.go` 内含跨平台交叉编译与发布打包流程(仅全量测试运行,`-short` 跳过);插件链路测试见 `pkg/build/plugin_test.go`、`plugin/gen-go/gen_test.go`
6. **扩展语言服务能力**:LSP 各能力分散在 `plugin/lsp/*.go`(diagnostics/completion/definition/hover/symbols/format),测试见 `lsp_test.go`(能力级)与 `serve_test.go`(协议级集成);编辑器接入说明见 `plugin/lsp/README.md`
