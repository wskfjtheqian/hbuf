我先检索项目记忆并分析代码库结构,然后生成一份 AI 快速理解项目的文档。已获取项目记忆。现在读取关键文件深入了解项目结构。继续深入了解核心包结构和文档。build 目录显示为空,但 main.go 引用了它。我再确认一下并查看语言生成包。奇怪,`pkg/build` 似乎没有文件。检查一下 git 状态和 .gitignore。build 目录文件确实存在。我继续查看其余核心文件,了解整体架构。最后确认 build 目录其余文件和 Java/format 包的功能。我已完成对项目的全面分析。项目是 **HBUF**——一个用 Go 编写的跨语言代码生成器(类似 protobuf/thrift),核心流程是:`.hbuf` IDL 文件 → 词法/语法分析 → AST → 校验 → 按语言后端生成 Go/Dart/Java/TypeScript 代码。现在生成 `AGENTS.md`(AI 通用约定文件)放在项目根目录。

```markdown
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
- 目标语言代码通过字符串拼接生成(不依赖目标语言的 AST 库)

## 构建与运行

```bash
# 编译(交叉编译脚本见 build.sh / build.bat)
go build -ldflags "-X main.version=<version>" -o bin/hbuf.darwin ./cmd/main.go
# 运行测试(含跨平台交叉编译测试 build_test.go)
go test ./...

# CLI 用法
hbuf -i <输入目录或通配符,如 ./test/*.hbuf> \
     -o <输出目录> \
     -t <构建类型: go|dart|java|ts> \
     -p <目标包名,如 admin> \
     [-l <生成项类型列表,如 "DS">] \
     [-a 全量构建,跳过增量检测] \
     [-v 显示版本]
```

## 架构与目录结构

```
cmd/main.go            CLI 入口:注册 4 种构建类型,解析 flag,调用 build.Build
pkg/
├── scanner/           词法分析(仿 go/scanner,含 ErrorList)
├── token/             Token 定义、FileSet、Pos 位置管理
├── parser/            语法分析(递归下降,支持 Trace 调试模式、import 递归解析)
├── ast/               AST 节点定义、Scope/Object 符号表、Package
├── build/             ★ 构建编排核心
│   ├── build.go       Build 主流程、BuildType 注册、BaseType 常量、
│   │                  标签/字段/方法遍历工具、命名转换工具
│   ├── writer.go      Writer:代码输出缓冲 + import 管理 + 多语言文本
│   ├── hash.go        FileHash:基于 MD5 + .build.json 的增量构建检测
│   └── ...
├── golang/            Go 后端(data/enum/server/database/verify/encoder/export/mq)
├── dart/              Dart 后端(data/enum/server/ui/verify)
├── java/              Java 后端(data/enum/server)
├── typescript/        TS 后端(data/enum/server/lang/ui.tsx/verify)
└── format/            .hbuf 格式化器
doc/                   hbuf.md(IDL 语法)、hbuf_db.md(数据库注解)、serialize.md
```

### 数据流(核心处理流程)

```
.hbuf 源文件 → scanner 词法 → parser 语法(AST) → build.checkFiles 类型校验
    → FileHash 增量检测(MD5 + .build.json,import 变化级联重编)
    → 按 -t 分发到语言后端 Build(file, fset, param)
    → 各后端遍历 AST,用 build.Writer 拼代码 → 写出 {name}.data.go 等文件
```

### 语言后端统一接口(新增语言必看)

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
package go = "admin"        // 按语言指定包名(dart 不支持)

enum Status { player = 0; stop = 1 }

data Base = 0 { int64 id = 0; string name = 1 }      // data 支持多继承
data Student:Base, Class = 2 { int32? no = 0 }        // ? 表示可空

server StudentServer:BaseServer = 0 {                  // server 支持继承
    GetClassResp GetClass(GetClassReq req) = 0         // 函数式方法
    SendMessage(MessageReq req) = 1                    // 广播式方法(无返回值)
}

data User = 3 {
    @marshal @in string name = 0                       // 注解: marshal/in/out
    @verify regex="..." string phone = 1               // 验证注解
    @db { key="true" order="id $ DESC" } int64 id = 2  // 数据库注解(见 hbuf_db.md)
}
```

- 基础类型:`int8..int64`、`uint8..uint64`、`bool`、`float`、`double`、`decimal`、`string`、`date`、`bytes`
- 复合类型:`数组[]`、`Map`(key 仅限基础类型和 enum)
- 每个字段/类型必须带唯一整数 ID(序列化用)
- 数据库注解(db):index、key、force、insert/inserts/update/set/del/get/list/map/count、table、rm、where、offset、limit、order、group、converter

## 代码约定

- 注释语言:中文;标识符沿用原语言
- `build` 包是共享工具层,**公共逻辑放这里,勿在语言后端重复实现**
- 命名转换统一用 build 包工具:`StringToHumpName`(驼峰)、`StringToUnderlineName`(下划线)、`StringToAllUpper`、`StringToMiddleLine`(中横线)、`StringToFirstLower`
- 类型判断统一用 `build.IsNumber/IsBool/IsArray/IsMap/IsEnum/IsNil/GetBaseType`
- 字段/方法遍历用 `build.EnumField/EnumMethod`(自动处理多继承去重)
- 新增 AST 节点需同步各语言后端的 `printTypeSpec` switch
- 输出文件在 `test/out/`(已被 .gitignore 忽略)

## 常见任务指引

1. **给某语言新增 SQL 生成**:参考 `pkg/golang/database.go` 的注解驱动模式
2. **新增语言后端(如 C#)**:复制 `pkg/java`,改类型映射和 writerFile 模板,在 `cmd/main.go` 注册
3. **扩展 IDL 语法**:改 `pkg/token/token.go` → `pkg/scanner` → `pkg/parser` → `pkg/ast` → 各后端
4. **修改序列化方案**:见 `pkg/golang/data.go`(JSON/二进制)与 `doc/serialize.md`
5. **验证 CLI 行为**:`build_test.go` 内含跨平台交叉编译与发布打包流程
```

已生成 `AGENTS.md`(项目根目录),内容涵盖:

1. **项目概述** — HBUF 是什么、各语言生成能力矩阵、关联运行时库
2. **技术栈** — Go 1.27、零第三方依赖、仿 Go 编译器架构
3. **构建与运行** — 编译命令、测试命令、CLI 参数说明
4. **架构与目录结构** — 11 个包的职责,重点标注 `pkg/build` 编排核心
5. **数据流** — 从 .hbuf 源文件到多语言代码的完整处理链路
6. **语言后端扩展指南** — 统一 `Build` 接口、各后端一致的内部模式
7. **.hbuf IDL 语法要点** — 精简语法示例 + 基础类型 + 注解体系
8. **代码约定** — 共享工具层原则、命名转换、类型判断工具
9. **常见任务指引** — 5 类高频改动场景的切入点

你可以直接复制上述内容保存为 `AGENTS.md` 放入项目根目录(现代 AI 编程工具会自动读取该文件)。如需调整详略程度或补充某部分(如序列化协议细节),告诉我即可。