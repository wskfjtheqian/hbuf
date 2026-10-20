// Package lsp 实现 HBUF 语言服务器(JSON-RPC over stdio,纯标准库)。
// 对标 gopls:诊断、补全、跳转定义/引用、悬停、文档符号、格式化。
package lsp

// 基本位置类型(LSP 3.16+ 通用子集)。

// Position 0-based 行/列,列按 UTF-16 code unit 计。
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// DiagnosticSeverity 诊断级别。
const (
	SeverityError       = 1
	SeverityWarning     = 2
	SeverityInformation = 3
	SeverityHint        = 4
)

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity,omitempty"`
	Code     string `json:"code,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

// CompletionItemKind 补全项类别(常用子集)。
const (
	CompletionText       = 1
	CompletionMethod     = 2
	CompletionFunction   = 3
	CompletionField      = 5
	CompletionClass      = 7
	CompletionInterface  = 8
	CompletionValue      = 12
	CompletionEnum       = 13
	CompletionKeyword    = 14
	CompletionEnumMember = 20
	CompletionConstant   = 21
	CompletionStruct     = 22
)

type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// SymbolKind 文档符号类别(常用子集)。
const (
	SymbolMethod     = 6
	SymbolField      = 8
	SymbolEnum       = 10
	SymbolInterface  = 11
	SymbolEnumMember = 22
	SymbolStruct     = 23
)

type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// Hover 悬停提示(markdown)。
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"` // "markdown"
	Value string `json:"value"`
}

// TextDocumentSyncKind 全文同步。
const TextDocumentSyncFull = 1

// ServerCapabilities 服务器能力声明。
type ServerCapabilities struct {
	TextDocumentSync           int                    `json:"textDocumentSync"`
	HoverProvider              bool                   `json:"hoverProvider"`
	DefinitionProvider         bool                   `json:"definitionProvider"`
	ReferencesProvider         bool                   `json:"referencesProvider"`
	DocumentSymbolProvider     bool                   `json:"documentSymbolProvider"`
	DocumentFormattingProvider bool                   `json:"documentFormattingProvider"`
	CompletionProvider         *CompletionOptions     `json:"completionProvider,omitempty"`
	SemanticTokensProvider     *SemanticTokensOptions `json:"semanticTokensProvider,omitempty"`
	CodeActionProvider         bool                   `json:"codeActionProvider,omitempty"`
	InlayHintProvider          bool                   `json:"inlayHintProvider,omitempty"`
	WorkspaceSymbolProvider    bool                   `json:"workspaceSymbolProvider,omitempty"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// 文档同步参数。

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type DidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

type DidCloseParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DidSaveParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// 各类请求参数。

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

// InitializeParams 初始化参数(提取 workspaceFolders/rootUri 用于工作区符号)。
type InitializeParams struct {
	Capabilities     interface{}       `json:"capabilities"`
	WorkspaceFolders []WorkspaceFolder `json:"workspaceFolders,omitempty"`
	RootURI          string            `json:"rootUri,omitempty"`
}

type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// 语义高亮(SemanticTokens,LSP 3.16)。
// TokenTypes/TokenModifiers 索引即 LSP 规范顺序,见 semantic_tokens.go。
type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type SemanticTokens struct {
	ResultID string   `json:"resultId,omitempty"`
	Data     []uint32 `json:"data"`
}

type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

type SemanticTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Full   bool                 `json:"full"`
}

// 代码操作(CodeAction,LSP 3.16)。
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type CodeAction struct {
	Title string         `json:"title"`
	Kind  string         `json:"kind,omitempty"`
	Edit  *WorkspaceEdit `json:"edit,omitempty"`
}

// WorkspaceEdit 编辑(按 uri 分组的文本变更)。
type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes,omitempty"`
}

// InlayHint(LSP 3.17)。
const (
	InlayHintType      = 1 // 类型提示
	InlayHintParameter = 2 // 参数名提示
)

type InlayHintParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

type InlayHint struct {
	Position     Position `json:"position"`
	Label        string   `json:"label"`
	Kind         int      `json:"kind,omitempty"`
	PaddingLeft  bool     `json:"paddingLeft,omitempty"`
	PaddingRight bool     `json:"paddingRight,omitempty"`
	Tooltip      string   `json:"tooltip,omitempty"`
}

// 工作区符号(workspace/symbol)。
type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

type WorkspaceSymbol struct {
	Name          string   `json:"name"`
	Kind          int      `json:"kind"`
	Location      Location `json:"location"`
	ContainerName string   `json:"containerName,omitempty"`
}
