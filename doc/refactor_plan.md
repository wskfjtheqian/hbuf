# HBUF 插件化 + 模板技术改造计划

> 目标:将 HBUF 从"单体内嵌各语言生成器"改造为 protoc 风格的可插拔架构,
> 并以 text/template 模板技术替换命令式字符串拼接的代码生成方式。

## 一、任务清单

| # | 任务 | 状态 |
|---|---|---|
| 1 | 创建 pkg/wire 中性描述模型(Request/Response/File/Type 等,实际落 plugin/core/wire) | 已完成 |
| 2 | 实现 AST → wire 转换(convert.go,含引用解析与包名标注) | 已完成 |
| 3 | 创建 pkg/build/plugin.go(插件发现、JSON over stdio 调用、统一落盘) | 已完成 |
| 4 | 改造 build.Build:插件调度 + 变更文件列表 + 进程内兼容回退 | 已完成 |
| 5 | cmd/hbuf 增加 -plugin=hbuf-gen-x=/path 参数 | 已完成 |
| 6 | 创建 pkg/plugin SDK(插件侧 stdin/stdout 主循环、协议校验,实际落 plugin/core/sdk) | 已完成 |
| 7 | 创建 pkg/tpl 模板引擎(内置 embed + 外部覆盖 + 执行缓存,实际落 plugin/core/tpl) | 已完成 |
| 8 | 创建 pkg/tpl 模板函数库(命名转换、类型映射等 FuncMap,实际落 plugin/core/tpl) | 已完成 |
| 9 | 设计 golang 模板上下文(DataContext/EnumContext/ServerContext) | 已完成 |
| 10 | 编写内置模板 templates/data.tmpl(descriptor+struct+Getter/Setter) | 已完成 |
| 11 | 编写 templates/enum.tmpl、server.tmpl 模板 | 已完成 |
| 12 | golang 后端接入模板渲染(template.go,与 Writer 并存) | 已完成 |
| 13 | 创建 cmd/hbuf-gen-go 插件入口(plugin.Run + 模板生成,实际落 plugin/gen-go/cmd) | 已完成 |
| 14 | 创建 cmd/hbuf-gen-dart/java/ts 插件入口 | 待办(仅 go 已建) |
| 15 | 构建验证:go build/go vet + 端到端生成测试 | 已完成 |
| 16 | 更新 AGENTS.md(插件协议、模板机制、目录结构) | 已完成 |
| 17 | 创建 plugin/lsp 目录与 cmd/hbuf-lsp 语言服务器入口 | 已完成 |
| 18 | 实现 LSP JSON-RPC 框架(initialize/textDocument/workspace,纯标准库) | 已完成 |
| 19 | 诊断能力:语法/类型校验/未解析引用实时报错 | 已完成 |
| 20 | 补全能力:类型/枚举/字段/方法/注解 tag 名 | 已完成 |
| 21 | 跳转定义与查找引用(跨文件 import) | 已完成 |
| 22 | 悬停提示与文档符号(Doc 注释、大纲) | 已完成 |
| 23 | 格式化接入 pkg/format(textDocument/formatting) | 已完成 |
| 24 | 编辑器集成:hbuf_idea LSP 接入与 VSCode 扩展说明 | 已完成 |
| 25 | 二进制序列化(全语言)实现 | 待办 |
| 26 | 生成后端单元/端到端测试基线 | 待办 |
| 27 | CI/CD 多平台自动发布(GitHub Actions) | 待办 |
| 28 | 广播 RPC(go/dart)实现 | 待办 |
| 29 | 循环 import 检测与错误信息优化 | 待办 |
| 30 | C/C# 插件后端(首个外部插件验证) | 待办 |
| 31 | 远程 import / 模块仓库 | 待办 |
| 32 | watch 模式(-w 文件变化自动构建) | 待办 |
| 33 | JSON Schema / OpenAPI 导出 | 待办 |
| 34 | examples 示例仓库 | 待办 |
| 35 | LSP 增强(语义高亮/代码操作/inlay hints/workspace 符号) | 已完成 |
| 36 | 数据库生成增强(迁移/事务/join) | 待办 |
| 37 | Golang 端多语言文本支持 | 待办 |
| 38 | parser fuzz 测试 | 待办 |
| 39 | API 文档站生成 | 待办 |

## 二、功能清单

### 1. 插件化架构(protoc 风格)

| # | 功能 | 说明 |
|---|---|---|
| F1 | 中性描述模型 pkg/wire | 对标 protobuf 的 descriptor.proto:Request/Response/File/Type/EnumType/DataType/ServerType/Method/Field/VarType/Tag,带协议版本号,JSON 可序列化 |
| F2 | AST → wire 转换 | 核心侧把 ast.Package 转成 wire.File 列表;引用类型自动标注 enum/data 种类与所属包名(供跨文件引用),字段 ID 解析为 int64 |
| F3 | 插件发现 | 按 hbuf-gen-<type> 命名在 PATH 查找;支持 -plugin 覆盖路径(对标 --plugin=protoc-gen-go=...) |
| F4 | 插件调用协议 | JSON over stdin/stdout:请求写 stdin、响应读 stdout、stderr 透传错误;插件无状态,核心负责落盘 |
| F5 | 增量调度 | 核心保留 FileHash(.build.json)增量检测,计算变更文件列表;每次编译只调一次插件,请求携带全部文件+变更列表 |
| F6 | 兼容回退 | 找不到外部插件时回退到进程内注册的生成器(拆分完成后删除) |
| F7 | CLI 参数扩展 | cmd/hbuf 新增 -plugin 参数(可重复,name=path 形式) |
| F8 | 插件 SDK pkg/plugin | 插件侧 stdio 主循环:读请求、校验协议版本、执行生成、写响应;第三方用任意语言实现协议即可开发插件 |
| F9 | 插件入口 | cmd/hbuf-gen-go/dart/java/ts 各一个 main(后续可迁出独立仓库) |

### 2. 模板技术(text/template,零依赖)

| # | 功能 | 说明 |
|---|---|---|
| F10 | 模板引擎 plugin/core/tpl | 基于标准库 text/template:默认模板 embed 打包进程序、外部自定义模板目录覆盖、模板解析缓存 |
| F11 | 模板函数库 | FuncMap 注册命名转换(hump/underline/allUpper/middleLine/lowerFirst)与语言类型映射(goType/goDefault 等),复用 build 包工具 |
| F12 | 模板上下文 | 结构化数据模型:DataContext/EnumContext/ServerContext/FieldContext(类型、tag、注释、descriptor 表达式预先算好),模板只做渲染 |
| F13 | 内置模板 | templates/data.tmpl(descriptor+struct+Getter/Setter)、enum.tmpl、server.tmpl,替换字符串拼接的 printXxxCode |
| F14 | 渲染接入 | golang 后端 template.go:上下文构建 + 引擎执行,输出与现有 build.Writer 并存,可逐文件灰度切换 |
| F15 | 用户自定义模板 | 通过插件参数 template=<目录> 指定自定义模板目录;同名文件覆盖内置模板,未指定或未命中时回退到默认模板(已编译进程序),不改 Go 代码即可定制生成风格 |

### 3. 验证与文档

| # | 功能 | 说明 |
|---|---|---|
| F16 | 构建验证 | go build ./...、go vet ./...、端到端生成对比(模板输出与旧拼接输出一致) |
| F17 | 文档更新 | AGENTS.md 补充插件协议、模板机制、新目录结构 |

### 4. 语言服务(LSP,方便编辑器集成)

| # | 功能 | 说明 |
|---|---|---|
| F18 | 语言服务器 hbuf-lsp | 独立二进制,JSON-RPC over stdio,纯标准库实现(零第三方依赖),对标 gopls |
| F19 | 诊断 | 语法错误、类型校验错误、未解析引用实时报错(复用 scanner.ErrorList + build 校验),增量解析缓存 |
| F20 | 代码补全 | 基础类型、已声明类型/枚举、data 字段名、server 方法名、注解 tag 名与参数 |
| F21 | 跳转定义与查找引用 | 类型/字段/方法定义跳转,支持跨文件 import 递归 |
| F22 | 悬停提示与文档符号 | 悬停显示 Doc 注释/类型/ID;文档大纲(data/enum/server/字段/方法) |
| F23 | 格式化 | textDocument/formatting 接入 pkg/format 现有格式化器 |
| F24 | 编辑器集成 | IntelliJ(hbuf_idea 通过 LSP 接入)、VSCode LSP 客户端扩展、neovim 等标准 LSP 客户端 |

### 5. 功能补全与生态建设

**高优先级**:

| # | 功能 | 说明 |
|---|---|---|
| F25 | 二进制序列化(全语言) | README 功能表中所有语言二进制序列化未完成,是跨语言高效传输的核心能力,hbuf 对标 protobuf 的最大短板 |
| F26 | 测试基线 | 生成后端(data/server/database 等)几乎无测试;插件化重构前建立单元/端到端测试基线,防止迁移后输出不一致 |
| F27 | CI/CD 自动发布 | build_test.go 已有交叉编译+打包逻辑,接 GitHub Actions 自动构建并发布 hbuf + hbuf-gen-* 二进制 |
| F28 | 广播 RPC(go/dart) | README 未完成项,推送类场景(消息通知)必需 |
| F29 | 循环 import 检测与错误优化 | 确认循环依赖检测;错误信息补充修复建议(也为 LSP 诊断打基础) |

**中优先级**:

| # | 功能 | 说明 |
|---|---|---|
| F30 | C/C# 后端 | 插件化后的首个外部插件,验证第三方可用任意语言开发插件 |
| F31 | 远程 import / 模块仓库 | 目前 import 仅支持本地相对路径;对标 buf registry,支持远程拉取 |
| F32 | watch 模式 | 基于 FileHash 增量,加 -w 监听文件变化自动重编 |
| F33 | JSON Schema / OpenAPI 导出 | data/server 定义导出为 JSON Schema/OpenAPI,对接网关与文档工具 |
| F34 | examples 示例仓库 | 完整示例(含自定义模板、LSP 配置),降低上手门槛 |
| F35 | LSP 增强 | 语义高亮(semantic tokens)、代码操作(code actions)、inlay hints、workspace 符号搜索 |

**长期/可选**:

| # | 功能 | 说明 |
|---|---|---|
| F36 | 数据库生成增强 | 迁移脚本生成、事务封装、join 查询(当前仅单表操作) |
| F37 | Golang 多语言文本 | README 中 Golang 多语言文本标注未完成 |
| F38 | fuzz 测试 | Go 原生 fuzzing 适合 parser,提升对异常 .hbuf 输入的健壮性 |
| F39 | API 文档站 | 基于文档注释生成 API 文档站(可复用 LSP 悬停数据) |

## 三、实施顺序

```
阶段一: 线协议落地      F1→F8 (wire 模型、插件发现/调用、SDK、CLI)          ✔ 已完成
阶段二: 模板引擎        F10→F15 (tpl 引擎、函数库、上下文、模板)              ✔ 已完成
阶段三: 插件入口        F9 (hbuf-gen-* 四个入口,后端切模板渲染)              ◐ 仅 hbuf-gen-go
阶段四: 验证收尾        F16→F17                                               ✔ 已完成
阶段五: 语言服务        F18→F24 (LSP 服务器、诊断/补全/跳转/格式化、编辑器接入) ✔ 已完成
阶段六: 测试与发布      F26→F27 (测试基线、CI/CD,可与阶段一并行,保障重构质量)
阶段七: 核心功能补全    F25、F28、F29、F37 (二进制序列化、广播 RPC、循环检测、Golang 多语言文本)
阶段八: 生态建设        F30→F36、F38→F39 (C/C# 插件、远程 import、watch、导出、示例、LSP 增强、数据库、fuzz、文档站) ◐ LSP 增强(F35)已完成
```

## 四、目标架构

```
hbuf/                          # 核心仓库(不感知任何目标语言)
├── cmd/hbuf/main.go           # 只做: 解析→校验→增量→调度插件
├── pkg/{scanner,token,parser,ast}   # 编译前端(不动)
├── pkg/build/                 # 编排 + 插件发现 + 协议调用 + 落盘
└── plugin/                    # 新插件化架构(可拆分)
    ├── core/wire/             # 中性描述模型(对标 descriptor.proto)
    ├── core/sdk/              # 插件侧 SDK(stdio 循环)
    ├── core/tpl/              # 模板引擎(text/template 封装)
    ├── lsp/cmd/hbuf-lsp/      # 语言服务器(JSON-RPC over stdio)
    └── gen-*/                 # 各语言插件(迁出为独立仓库 hbuf-gen-*)

hbuf-gen-dart/  hbuf-gen-java/  hbuf-gen-ts/   # 以后各自独立仓库
```

## 五、关键设计决策

1. **插件发现**:按 hbuf-gen-<type> 命名在 PATH 查找,支持 -plugin=hbuf-gen-x=/path 覆盖
2. **中性描述模型 pkg/wire**:AST 含 token.Pos、*Object、接口类型,无法直接 JSON 序列化,
   核心先转成稳定的 wire 模型;协议带版本号、可跨语言实现插件
3. **插件无状态**:核心负责落盘和 .build.json 增量,插件只返回文件内容
4. **每次编译只调一次插件**:请求带全部文件(供引用解析)+ 变更文件列表
5. **模板引擎基于标准库 text/template**:零第三方依赖,符合项目纯标准库风格;
   生成逻辑与生成格式分离

## 六、模板机制

生成代码由模板渲染,查找顺序:

1. **自定义模板(优先)**:通过插件参数 `template=<目录>` 指定自定义模板目录
   (核心写入请求 Options,插件解析);目录下同名文件覆盖内置模板
2. **默认模板(兜底)**:未指定自定义模板目录、或目录中无同名文件时,
   使用编译进程序的内置模板

### 默认模板编译进程序

- 内置模板存放于 `plugin/gen-go/templates/*.tmpl`,通过 `//go:embed` 打包进
  `hbuf-gen-go` 二进制,发布时只需分发单个可执行文件
- 模板引擎(plugin/core/tpl)基于 `fs.FS` 接口:内置模板用 `embed.FS`,
  自定义模板用 `os.DirFS`,两层合并后自定义优先

### 自定义模板规则

- 文件名与内置模板同名即可覆盖,如 `data.tmpl`、`enum.tmpl`、`server.tmpl`
- 模板渲染上下文与内置模板一致(DataContext/EnumContext/ServerContext),
  用户可直接拷贝内置模板修改,不必改 Go 代码
- 模板可用 FuncMap 函数:hump、lowerFirst、underline、allUpper、middleLine、
  goType、goDefault 等命名转换与类型映射函数

## 七、语言服务(LSP)

为方便编辑器集成,提供 HBUF 语言服务器 `hbuf-lsp`(对标 gopls):

- **协议**:JSON-RPC over stdio,纯标准库实现(零第三方依赖)
- **复用编译前端**:增量解析 + token.FileSet 位置信息 + build 校验,与编译器同源,
  避免诊断结果与命令行不一致
- **能力**:
  - 诊断:语法错误、类型校验错误、未解析引用实时报错
  - 补全:基础类型、已声明类型/枚举、data 字段名、server 方法名、注解 tag 名
  - 跳转定义与查找引用(支持跨文件 import)
  - 悬停提示(Doc 注释/类型/ID)与文档符号大纲
  - 格式化:接入 pkg/format(textDocument/formatting)
  - 语义高亮(textDocument/semanticTokens/full,delta 编码,scanner+AST 分层)
  - 代码操作(textDocument/codeAction:格式化 quickfix、未使用 import 删除)
  - 内联提示(textDocument/inlayHint:extends 成员摘要、跨文件类型来源)
  - 工作区符号(workspace/symbol:扫描工作区 *.hbuf,大小写不敏感过滤)
- **编辑器集成**:IntelliJ(hbuf_idea 插件 LSP 接入)、VSCode(LSP 客户端扩展)、
  neovim 等标准 LSP 客户端
- **增量同步**:解析缓存按文件内容 hash 失效,支持未保存文档(内存覆盖)
