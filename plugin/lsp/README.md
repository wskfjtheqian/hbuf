# hbuf-lsp — HBUF 语言服务器

HBUF 语言服务器(对标 gopls):基于 LSP(JSON-RPC over stdio),纯标准库实现,
零第三方依赖。为编辑器提供诊断、补全、跳转定义/引用、悬停、文档符号与格式化。

## 构建

```bash
go build -o bin/hbuf-lsp ./cmd/hbuf-lsp
# 或随核心一起:./build.sh 产物包含各平台二进制

bin/hbuf-lsp -v    # 打印版本
bin/hbuf-lsp -log  # 调试模式(日志写 stderr,协议流仍只写 stdout)
```

## 支持的能力

| 能力 | 方法 | 说明 |
|---|---|---|
| 诊断 | publishDiagnostics | 语法错误 + 类型校验错误(与命令行 `hbuf` 同源,复用 `build.CheckPackage`) |
| 补全 | textDocument/completion | 关键字、基础类型、已声明类型/枚举(含 import)、字段名、方法名、枚举项、注解 tag 名与键名 |
| 跳转定义 | textDocument/definition | 类型/字段/方法/枚举项,支持跨文件 import |
| 查找引用 | textDocument/references | 全部引用位置,支持 `includeDeclaration` |
| 悬停 | textDocument/hover | Doc 注释、类型信息、ID 等(markdown) |
| 文档符号 | textDocument/documentSymbol | 类型 → 字段/方法/枚举项的树形大纲 |
| 格式化 | textDocument/formatting | 接入 `pkg/format`,全量替换 TextEdit |

协议实现要点:

- **全文同步**(`textDocumentSync: 1`):didChange 的最后一个 change 即全文
- **未保存文档**:打开文档以内存内容为准,依赖文件从磁盘读取
- **位置约定**:0-based 行,列按 UTF-16 code unit(emoji 等 BMP 外字符占 2 列)
- **退出码**:shutdown 后 exit 返回 0,否则返回 1

## 编辑器集成

hbuf-lsp 是标准 LSP 服务器,任何 LSP 客户端均可接入。

### VSCode

自建最小扩展(约 20 行):`package.json` 注册 language + languageServer:

```jsonc
// 1. languages:注册 .hbuf 文件类型
"languages": [
  { "id": "hbuf", "extensions": [".hbuf"] }
]
// 2. languageServer:启动 hbuf-lsp
"languageServer": [
  {
    "id": "hbuf-lsp",
    "label": "HBUF",
    "command": "/path/to/bin/hbuf-lsp",
    "languages": ["hbuf"]
  }
]
```

基于 vscode-languageclient 的完整扩展示例(extension.ts 核心):

```typescript
import * as vscode from 'vscode';
import { LanguageClient, ServerOptions } from 'vscode-languageclient/node';

export function activate(context: vscode.ExtensionContext) {
  const server: ServerOptions = {
    command: '/path/to/bin/hbuf-lsp', // 或 process.env.HBUF_LSP
    args: [],
  };
  const client = new LanguageClient('hbuf-lsp', 'HBUF Language Server', server, {
    documentSelector: [{ scheme: 'file', language: 'hbuf' }],
  });
  context.subscriptions.push(client.start());
}
```

### IntelliJ IDEA(hbuf_idea 插件)

在 `hbuf_idea` 插件(仓库 `github.com/wskfjtheqian/hbuf_idea`)中接入 LSP,
方式二选一:

1. **LSP4IJ**(推荐):IDE 2024.1+ 自带平台,注册 `ServerDescriptor`
   (languageId `hbuf`、扩展名 `hbuf`、启动命令指向 hbuf-lsp),由平台代理
   补全/诊断/跳转等能力,无需自实现各 Inspect/CompletionProvider
2. **自实现桥接**:若插件需要深度定制,可保留现有 PSI 基础设施,
   将 hbuf-lsp 的响应转换为 IntelliJ 的 Inspection/Completion 对象
   (用 `com.intellij.platform.lsp` API 或直接通过 stdin/stdout 启动子进程)

接入后:hbuf_idea 原有的语法高亮继续由 PSI 提供,LSP 提供语义级
诊断/补全/跳转,两者互补。

### Neovim(内置 LSP 客户端)

```lua
vim.api.nvim_create_autocmd('FileType', {
  pattern = 'hbuf',
  callback = function()
    vim.lsp.start({
      name = 'hbuf-lsp',
      cmd = { '/path/to/bin/hbuf-lsp' },
    })
  end,
})
```

若使用 nvim-lspconfig 可写最小配置:

```lua
local configs = require('lspconfig.configs')
configs.hbuf_lsp = {
  default_config = {
    cmd = { '/path/to/bin/hbuf-lsp' },
    filetypes = { 'hbuf' },
    root_dir = function(fname)
      return vim.fs.dirname(vim.fs.find({ 'go.mod', '.git' }, { upward = true })[1] or fname)
    end,
  },
}
require('lspconfig').hbuf_lsp.setup({})
```

### 其他编辑器

任何支持 LSP 的编辑器(Emacs eglot、Sublime LSP、Helix 等)均可:

- 文件类型:`hbuf`,扩展名 `.hbuf`
- 启动命令:`hbuf-lsp`(建议 `-log` 调试时加,日志在 stderr)
- 初始化参数无特殊要求,服务器返回 capabilities 后即可使用

## 测试

```bash
go test ./plugin/lsp/ -v
```

- `lsp_test.go`:能力级单元测试(位置转换往返、补全、定义/引用、悬停、符号、格式化、诊断)
- `serve_test.go`:协议级集成测试(io.Pipe 完整会话:initialize → didOpen → 诊断通知 → completion → shutdown → exit,含退出码与未知方法错误)

## 实现结构

```
plugin/lsp/
├── protocol.go     LSP 类型与常量(Position/Range/Diagnostic/CompletionItem/...)
├── jsonrpc.go      JSON-RPC 2.0 帧(Content-Length 头,Wire 读写)
├── server.go       Serve 主循环 + dispatch 方法分发 + 通知发送
├── view.go         文档视图(Open/Change/Close)+ import 依赖图递归解析
├── position.go     token.Pos ↔ LSP Position(UTF-16 列、EOF 特判)
├── context.go      光标 → AST 上下文(findContext + walkIdents 遍历)
├── diagnostics.go  诊断(语法错误 + build.CheckPackage 同源校验)
├── lookup.go       标识符 → 声明(基础类型/字段/方法/枚举项/类型)
├── completion.go   补全(关键字/类型/字段/方法/枚举项/tag)
├── definition.go   跳转定义与查找引用
├── hover.go        悬停(markdown)
├── symbols.go      文档符号(树形大纲)
└── format.go       格式化(独立解析,接入 pkg/format)
```
