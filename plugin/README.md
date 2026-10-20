# plugin — HBUF 新插件化架构

本目录存放 HBUF 改造后的可插拔架构代码(详见 `doc/refactor_plan.md`)。
设计对标 protoc 生态:核心负责解析/校验/调度,各语言生成器为独立插件,
通过 JSON over stdio 协议通信,生成逻辑基于 text/template 模板。

## 目录布局

```
plugin/
├── core/                          # 插件化核心库(插件协议 + SDK + 模板引擎)
│   ├── wire/                      # 中性描述模型(对标 descriptor.proto)
│   │                              # Request/Response/File/Type/EnumType/DataType/...
│   │                              # 以及 AST → wire 转换(convert.go)
│   ├── sdk/                       # 插件侧 SDK(stdio 主循环、协议版本校验)
│   │                              # 第三方可用任意语言实现协议,不必使用本 SDK
│   └── tpl/                       # 模板引擎(text/template 封装)
│                                  # 内置模板 embed 打包 + 外部模板覆盖 + 执行缓存
│
├── lsp/                           # HBUF 语言服务(方便编辑器集成)
│   └── cmd/hbuf-lsp/              # 语言服务器入口(JSON-RPC over stdio)
│                                  # 诊断/补全/跳转定义/悬停/符号/格式化
│
└── gen-go/                        # Go 插件(后续迁出为独立仓库 hbuf-gen-go)
    ├── cmd/hbuf-gen-go/           # 插件二进制入口(plugin.Run + 模板生成)
    └── templates/                 # 内置生成模板(*.tmpl,支持外部覆盖)
        data.tmpl / enum.tmpl / server.tmpl / ...

(规划中,尚未创建)
├── gen-dart/cmd/hbuf-gen-dart     # Dart 插件(后续独立仓库)
├── gen-java/cmd/hbuf-gen-java     # Java 插件(后续独立仓库)
└── gen-ts/cmd/hbuf-gen-ts         # TypeScript 插件(后续独立仓库)
```

## 核心约定

1. 插件命名:hbuf-gen-<type>,由核心在 PATH 查找,支持 -plugin 覆盖
2. 协议版本:hbuf-plugin-v1,JSON over stdin/stdout,插件无状态
3. 核心负责落盘与 .build.json 增量;插件只返回文件内容
4. 每次编译只调一次插件:请求携带全部文件(引用解析)+ 变更文件列表
5. 模板引擎零第三方依赖(纯标准库 text/template)

## 模板机制

生成代码由模板渲染,查找顺序:

1. **自定义模板(优先)**:通过插件参数 `template=<目录>` 指定自定义模板目录
   (核心写入请求 Options,插件解析);目录下同名文件覆盖内置模板
2. **默认模板(兜底)**:未指定自定义模板目录、或目录中无同名文件时,
   使用编译进程序的内置模板

### 默认模板编译进程序

- 内置模板存放于 `gen-go/templates/*.tmpl`,通过 `//go:embed` 打包进
  `hbuf-gen-go` 二进制,发布时只需分发单个可执行文件
- 模板引擎(plugin/core/tpl)基于 `fs.FS` 接口:内置模板用 `embed.FS`,
  自定义模板用 `os.DirFS`,两层合并后自定义优先

### 自定义模板规则

- 文件名与内置模板同名即可覆盖,如 `data.tmpl`、`enum.tmpl`、`server.tmpl`
- 模板渲染上下文与内置模板一致(DataContext/EnumContext/ServerContext),
  用户可直接拷贝内置模板修改,不必改 Go 代码
- 模板可用 FuncMap 函数:hump、lowerFirst、underline、allUpper、middleLine、
  goType、goDefault 等命名转换与类型映射函数

## 语言服务(LSP)

为方便编辑器集成,提供 HBUF 语言服务器 `hbuf-lsp`(对标 gopls):

- **协议**:JSON-RPC over stdio,纯标准库实现(零第三方依赖)
- **复用编译前端**:增量解析 + token.FileSet 位置信息 + build 校验,与编译器同源
- **能力**:
  - 诊断:语法错误、类型校验错误、未解析引用实时报错
  - 补全:基础类型、已声明类型/枚举、字段名、方法名、注解 tag 名
  - 跳转定义与查找引用(支持跨文件 import)
  - 悬停提示(Doc 注释/类型/ID)与文档符号大纲
  - 格式化:接入 pkg/format(textDocument/formatting)
- **编辑器集成**:IntelliJ(hbuf_idea 插件 LSP 接入)、VSCode(LSP 客户端扩展)、
  neovim 等标准 LSP 客户端
- **增量同步**:解析缓存按文件内容 hash 失效,支持未保存文档(内存覆盖)

## 与现有代码的关系

- 现阶段与现有 pkg/ 代码并存:核心 `build.Build` 优先调用外部插件,
  找不到时回退到进程内生成器(兼容期,拆分完成后删除)
- gen-* 内的生成逻辑从现有 pkg/golang、pkg/dart、pkg/java、pkg/typescript 迁移而来,
  由消费 ast 改为消费 wire 模型,字符串拼接逐步替换为模板渲染
