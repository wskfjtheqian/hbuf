package lsp

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
	"sort"
	"strconv"
)

// 语义高亮:按 LSP 规范的标准 token 类型/修饰符顺序(索引即协议值)。
var semanticTokenTypes = []string{
	"namespace", "type", "class", "enum", "interface", "struct",
	"typeParameter", "parameter", "variable", "property", "enumMember",
	"event", "function", "method", "macro", "keyword", "modifier",
	"comment", "string", "number", "regexp", "operator", "decorator",
}

var semanticTokenModifiers = []string{
	"declaration", "definition", "readonly", "static", "deprecated",
	"abstract", "async", "modification", "documentation", "defaultLibrary",
}

// token 类型索引(见 semanticTokenTypes)。
const (
	semTokType       = 1  // 基础类型/类型引用
	semTokEnum       = 3  // enum 类型名
	semTokInterface  = 4  // server 类型名
	semTokStruct     = 5  // data 类型名
	semTokParameter  = 7  // 方法参数名
	semTokProperty   = 9  // 字段名
	semTokEnumMember = 10 // 枚举项名
	semTokMethod     = 13 // 方法名
	semTokMacro      = 14 // 注解键名
	semTokKeyword    = 15 // 语言关键字
	semTokComment    = 17 // 注释
	semTokString     = 18 // 字符串(package 值/import 路径等)
	semTokNumber     = 19 // 数字 ID
	semTokDecorator  = 22 // 注解 tag 名
)

// token 修饰符位(见 semanticTokenModifiers)。
const semModDeclaration = 1 << 0

// semToken 一条待编码的语义 token(已折算为 LSP 位置)。
type semToken struct {
	line   int // 0-based 行
	char   int // 0-based UTF-16 列
	length int // UTF-16 长度
	typ    int
	mod    uint32
}

// semanticTokens 计算整个文档的语义 token(delta 编码,见 LSP SemanticTokens)。
// 关键字/注释/字符串/数字来自 scanner(语法错误时也能尽量收集),语义来自 AST。
func (s *Server) semanticTokens(doc *Document) *SemanticTokens {
	pf := s.view.parse(doc)
	src := []byte(doc.Text)
	tf := fileToken(pf, doc.Path)
	if src == nil || tf == nil {
		return &SemanticTokens{Data: []uint32{}}
	}
	file := pf.pkg.Files[doc.Path]
	if file == nil {
		file = &ast.File{Scope: ast.NewScope(nil)}
	}

	var toks []semToken
	seen := map[string]bool{} // "line:char:len" 去重
	add := func(t semToken) {
		if t.length <= 0 {
			return
		}
		key := strconv.Itoa(t.line) + ":" + strconv.Itoa(t.char) + ":" + strconv.Itoa(t.length)
		if seen[key] {
			return
		}
		seen[key] = true
		toks = append(toks, t)
	}
	// 区间 → semToken(跨行 token 如块注释按行拆分)
	addRange := func(start, end token.Pos, typ int, mod uint32) {
		for off := tf.Offset(start); off < tf.Offset(end); {
			seg := off
			for seg < tf.Offset(end) && src[seg] != '\n' {
				seg++
			}
			p := toPosition(src, tf, token.Pos(int(tf.Base())+off))
			add(semToken{line: p.Line, char: p.Character,
				length: utf16Len(src[off:seg]), typ: typ, mod: mod})
			off = seg + 1
		}
	}

	// 1. scanner:关键字/注释/字符串/数字
	fset := token.NewFileSet()
	stf := fset.AddFile(doc.Path, fset.Base(), len(src))
	var scan scanner.Scanner
	scan.Init(stf, src, func(token.Position, string) {}, scanner.ScanComments)
	for {
		pos, tok, lit := scan.Scan()
		if tok == token.EOF {
			break
		}
		switch {
		case tok == token.COMMENT:
			addRange(pos, pos+token.Pos(len(lit)), semTokComment, 0)
		case tok == token.STRING:
			addRange(pos, pos+token.Pos(len(lit)), semTokString, 0)
		case tok == token.INT:
			addRange(pos, pos+token.Pos(len(lit)), semTokNumber, 0)
		case tok.IsKeyword():
			addRange(pos, pos+token.Pos(len(tok.String())), semTokKeyword, 0)
		}
	}

	// 2. AST:标识符语义(类型/字段/方法/枚举项/注解 tag 与键)
	emitTypeRef := func(id *ast.Ident) {
		if id == nil || id.Name == "_" {
			return
		}
		typ := semTokType
		if obj := id.Obj; obj != nil {
			if ts, ok := obj.Decl.(*ast.TypeSpec); ok {
				typ = typeSemKind(ts.Type)
			}
		} else if ts, _ := findTypeSpec(pf, file, id.Name); ts != nil {
			// 跨文件 import 类型(parser 单文件解析时未链接 Obj)
			typ = typeSemKind(ts.Type)
		}
		addRange(id.Pos(), id.End(), typ, 0)
	}
	for _, spec := range file.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		switch t := ts.Type.(type) {
		case *ast.DataType:
			for _, tag := range t.Tags {
				emitTagSem(tag, addRange)
			}
			if ts.Name != nil {
				addRange(ts.Name.Pos(), ts.Name.End(), semTokStruct, semModDeclaration)
			}
			for _, ext := range t.Extends {
				emitTypeRef(ext.Name)
			}
			if t.Fields != nil {
				for _, field := range t.Fields.List {
					for _, tag := range field.Tags {
						emitTagSem(tag, addRange)
					}
					if field.Name != nil {
						addRange(field.Name.Pos(), field.Name.End(), semTokProperty, semModDeclaration)
					}
					walkExpr(field.Type, func(id *ast.Ident) { emitTypeRef(id) })
				}
			}
		case *ast.ServerType:
			for _, tag := range t.Tags {
				emitTagSem(tag, addRange)
			}
			if ts.Name != nil {
				addRange(ts.Name.Pos(), ts.Name.End(), semTokInterface, semModDeclaration)
			}
			for _, ext := range t.Extends {
				emitTypeRef(ext.Name)
			}
			for _, method := range t.Methods {
				for _, tag := range method.Tags {
					emitTagSem(tag, addRange)
				}
				if method.Name != nil {
					addRange(method.Name.Pos(), method.Name.End(), semTokMethod, semModDeclaration)
				}
				if method.ParamName != nil {
					addRange(method.ParamName.Pos(), method.ParamName.End(), semTokParameter, semModDeclaration)
				}
				if method.Result != nil {
					walkExpr(method.Result, func(id *ast.Ident) { emitTypeRef(id) })
				}
				if method.Param != nil {
					walkExpr(method.Param, func(id *ast.Ident) { emitTypeRef(id) })
				}
			}
		case *ast.EnumType:
			for _, tag := range t.Tags {
				emitTagSem(tag, addRange)
			}
			if ts.Name != nil {
				addRange(ts.Name.Pos(), ts.Name.End(), semTokEnum, semModDeclaration)
			}
			for _, item := range t.Items {
				for _, tag := range item.Tags {
					emitTagSem(tag, addRange)
				}
				if item.Name != nil {
					addRange(item.Name.Pos(), item.Name.End(), semTokEnumMember, semModDeclaration)
				}
			}
		}
	}

	// 3. 排序 + delta 编码(同起点取长 token,保证稳定)
	sort.Slice(toks, func(i, j int) bool {
		if toks[i].line != toks[j].line {
			return toks[i].line < toks[j].line
		}
		if toks[i].char != toks[j].char {
			return toks[i].char < toks[j].char
		}
		return toks[i].length > toks[j].length
	})
	data := make([]uint32, 0, len(toks)*5)
	prevLine, prevChar := 0, 0
	for _, t := range toks {
		deltaLine := uint32(t.line - prevLine)
		deltaChar := uint32(t.char)
		if deltaLine == 0 {
			deltaChar = uint32(t.char - prevChar)
		}
		data = append(data, deltaLine, deltaChar, uint32(t.length), uint32(t.typ), t.mod)
		prevLine, prevChar = t.line, t.char
	}
	return &SemanticTokens{Data: data}
}

// typeSemKind 类型声明 → semantic token 类型。
func typeSemKind(typ ast.Expr) int {
	switch typ.(type) {
	case *ast.EnumType:
		return semTokEnum
	case *ast.ServerType:
		return semTokInterface
	case *ast.DataType:
		return semTokStruct
	}
	return semTokType
}

// emitTagSem 注解 tag:tag 名为 decorator,键名为 macro。
func emitTagSem(tag *ast.Tag, addRange func(token.Pos, token.Pos, int, uint32)) {
	if tag == nil {
		return
	}
	if tag.Name != nil {
		addRange(tag.Name.Pos(), tag.Name.End(), semTokDecorator, 0)
	}
	for _, kv := range tag.KV {
		if kv.Name != nil {
			addRange(kv.Name.Pos(), kv.Name.End(), semTokMacro, 0)
		}
	}
}
