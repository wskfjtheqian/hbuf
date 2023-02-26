package parser

import (
	"fmt"
	"hbuf/pkg/ast"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
	"strconv"
	"strings"
	"unicode"
)

type parser struct {
	file    *token.File
	errors  scanner.ErrorList
	scanner scanner.Scanner

	// Tracing/debugging
	mode   Mode // parsing mode
	trace  bool // == (mode & Trace != 0)
	indent int  // indentation used for tracing output

	// Comments
	comments    []*ast.CommentGroup
	leadComment *ast.CommentGroup // last lead comment
	lineComment *ast.CommentGroup // last line comment

	// Next token
	pos token.Pos   // token position
	tok token.Token // one token look-ahead
	lit string      // token literal

	// Error recovery
	// (used to limit the number of calls to parser.advance
	// w/o making scanning progress - avoids potential endless
	// loops across multiple parser functions during error recovery)
	syncPos token.Pos // last synchronization position
	syncCnt int       // number of parser.advance calls without progress

	// Non-syntactic parser control
	exprLev int  // < 0: in control clause, >= 0: in expression
	inRhs   bool // if set, the parser is parsing a rhs expression

	// Ordinary identifier scopes
	pkgScope   *ast.Scope        // pkgScope.Outer == nil
	topScope   *ast.Scope        // top-most scope; may be pkgScope
	unresolved []*ast.Ident      // unresolved identifiers
	imports    []*ast.ImportSpec // list of imports

	// Label scopes
	// (maintained by open/close LabelScope)
	labelScope  *ast.Scope     // label scope for current function
	targetStack [][]*ast.Ident // stack of unresolved labels
}

func (p *parser) init(fset *token.FileSet, filename string, src []byte, mode Mode) {
	p.file = fset.AddFile(filename, -1, len(src))
	var m scanner.Mode
	if mode&ParseComments != 0 {
		m = scanner.ScanComments
	}
	eh := func(pos token.Position, msg string) {
		p.errors.Add(pos, msg)
	}
	p.scanner.Init(p.file, src, eh, m)

	p.mode = mode
	p.trace = mode&Trace != 0 // for convenience (p.trace is used frequently)

	p.next()
}

func (p *parser) openScope() {
	p.topScope = ast.NewScope(p.topScope)
}

func (p *parser) closeScope() {
	p.topScope = p.topScope.Outer
}

func (p *parser) openLabelScope() {
	p.labelScope = ast.NewScope(p.labelScope)
	p.targetStack = append(p.targetStack, nil)
}

func (p *parser) closeLabelScope() {
	// resolve labels
	n := len(p.targetStack) - 1
	scope := p.labelScope
	for _, ident := range p.targetStack[n] {
		ident.Obj = scope.Lookup(ident.Name)
		if ident.Obj == nil && p.mode&DeclarationErrors != 0 {
			p.error(ident.Pos(), fmt.Sprintf("label %s undefined", ident.Name))
		}
	}
	// pop label scope
	p.targetStack = p.targetStack[0:n]
	p.labelScope = p.labelScope.Outer
}

func (p *parser) declare(decl, data interface{}, scope *ast.Scope, kind ast.ObjKind, idents ...*ast.Ident) {
	for _, ident := range idents {
		assert(ident.Obj == nil, "identifier already declared or resolved")
		obj := ast.NewObj(kind, ident.Name)
		// remember the corresponding declaration for redeclaration
		// errors and global variable resolution/typechecking phase
		obj.Decl = decl
		obj.Data = data
		ident.Obj = obj
		if ident.Name != "_" {
			if alt := scope.Insert(obj); alt != nil && p.mode&DeclarationErrors != 0 {
				prevDecl := ""
				if pos := alt.Pos(); pos.IsValid() {
					prevDecl = fmt.Sprintf("\n\tprevious declaration at %s", p.file.Position(pos))
				}
				p.error(ident.Pos(), fmt.Sprintf("%s redeclared in this block%s", ident.Name, prevDecl))
			}
		}
	}
}

var unresolved = new(ast.Object)

func (p *parser) tryResolve(x ast.Expr, collectUnresolved bool) {
	// nothing to do if x is not an identifier or the blank identifier
	ident, _ := x.(*ast.Ident)
	if ident == nil {
		return
	}
	assert(ident.Obj == nil, "identifier already declared or resolved")
	if ident.Name == "_" {
		return
	}
	// try to resolve the identifier
	for s := p.topScope; s != nil; s = s.Outer {
		if obj := s.Lookup(ident.Name); obj != nil {
			ident.Obj = obj
			return
		}
	}
	// all local scopes are known, so any unresolved identifier
	// must be found either in the file scope, package scope
	// (perhaps in another file), or universe scope --- collect
	// them so that they can be resolved later
	if collectUnresolved {
		ident.Obj = unresolved
		p.unresolved = append(p.unresolved, ident)
	}
}

func (p *parser) resolve(x ast.Expr) {
	p.tryResolve(x, true)
}

// ----------------------------------------------------------------------------
// Parsing support

func (p *parser) printTrace(a ...interface{}) {
	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = len(dots)
	pos := p.file.Position(p.pos)
	fmt.Printf("%5d:%3d: ", pos.Line, pos.Column)
	i := 2 * p.indent
	for i > n {
		fmt.Print(dots)
		i -= n
	}
	// i <= n
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *parser, msg string) *parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

// Usage pattern: defer un(trace(p, "..."))
func un(p *parser) {
	p.indent--
	p.printTrace(")")
}

// Advance to the next token.
func (p *parser) next0() {
	// Because of one-token look-ahead, print the previous token
	// when tracing as it provides a more readable output. The
	// very first token (!p.pos.IsValid()) is not initialized
	// (it is token.ILLEGAL), so don't print it .
	if p.trace && p.pos.IsValid() {
		s := p.tok.String()
		switch {
		case p.tok.IsLiteral():
			p.printTrace(s, p.lit)
		case p.tok.IsOperator(), p.tok.IsKeyword():
			p.printTrace("\"" + s + "\"")
		default:
			p.printTrace(s)
		}
	}

	p.pos, p.tok, p.lit = p.scanner.Scan()
}

// Consume a comment and return it and the line on which it ends.
func (p *parser) consumeComment() (comment *ast.Comment, endline int) {
	// /*-style comments may end on a different line than where they start.
	// Scan the comment for '\n' chars and adjust endline accordingly.
	endline = p.file.Line(p.pos)
	if p.lit[1] == '*' {
		// don't use range here - no need to decode Unicode code points
		for i := 0; i < len(p.lit); i++ {
			if p.lit[i] == '\n' {
				endline++
			}
		}
	}

	comment = &ast.Comment{Slash: p.pos, Text: p.lit}
	p.next0()

	return
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. A non-comment token or n
// empty lines terminate a comment group.
//
func (p *parser) consumeCommentGroup(n int) (comments *ast.CommentGroup, endline int) {
	var list []*ast.Comment
	endline = p.file.Line(p.pos)
	for p.tok == token.COMMENT && p.file.Line(p.pos) <= endline+n {
		var comment *ast.Comment
		comment, endline = p.consumeComment()
		list = append(list, comment)
	}

	// add comment group to the comments list
	comments = &ast.CommentGroup{List: list}
	p.comments = append(p.comments, comments)

	return
}

// Advance to the next non-comment token. In the process, collect
// any comment groups encountered, and remember the last lead and
// line comments.
//
// A lead comment is a comment group that starts and ends in a
// line without any other tokens and that is followed by a non-comment
// token on the line immediately after the comment group.
//
// A line comment is a comment group that follows a non-comment
// token on the same line, and that has no tokens after it on the line
// where it ends.
//
// Lead and line comments may be considered documentation that is
// stored in the AST.
//
func (p *parser) next() {
	p.leadComment = nil
	p.lineComment = nil
	prev := p.pos
	p.next0()

	if p.tok == token.COMMENT {
		var comment *ast.CommentGroup
		var endline int

		if p.file.Line(p.pos) == p.file.Line(prev) {
			// The comment is on same line as the previous token; it
			// cannot be a lead comment but may be a line comment.
			comment, endline = p.consumeCommentGroup(0)
			if p.file.Line(p.pos) != endline || p.tok == token.EOF {
				// The next token is on a different line, thus
				// the last comment group is a line comment.
				p.lineComment = comment
			}
		}

		// consume successor comments, if any
		endline = -1
		for p.tok == token.COMMENT {
			comment, endline = p.consumeCommentGroup(1)
		}

		if endline+1 == p.file.Line(p.pos) {
			// The next token is following on the line immediately after the
			// comment group, thus the last comment group is a lead comment.
			p.leadComment = comment
		}
	}
}

// A bailout panic is raised to indicate early termination.
type bailout struct{}

func (p *parser) error(pos token.Pos, msg string) {
	epos := p.file.Position(pos)

	// If AllErrors is not set, discard errors reported on the same line
	// as the last recorded error and stop parsing if there are more than
	// 10 errors.
	if p.mode&AllErrors == 0 {
		n := len(p.errors)
		if n > 0 && p.errors[n-1].Pos.Line == epos.Line {
			return // discard - likely a spurious error
		}
		if n > 10 {
			panic(bailout{})
		}
	}

	p.errors.Add(epos, msg)
}

func (p *parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		switch {
		case p.tok == token.SEMICOLON && p.lit == "\n":
			msg += ", found newline"
		case p.tok.IsLiteral():
			// print 123 rather than 'INT', etc.
			msg += ", found " + p.lit
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}
	p.error(pos, msg)
}

func (p *parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next() // make progress
	return pos
}

// expect2 is like expect, but it returns an invalid position
// if the expected token is not found.
func (p *parser) expect2(tok token.Token) (pos token.Pos) {
	if p.tok == tok {
		pos = p.pos
	} else {
		p.errorExpected(p.pos, "'"+tok.String()+"'")
	}
	p.next() // make progress
	return
}

// expectClosing is like expect but provides a better error message
// for the common case of a missing comma before a newline.
//
func (p *parser) expectClosing(tok token.Token, context string) token.Pos {
	if p.tok != tok && p.tok == token.SEMICOLON && p.lit == "\n" {
		p.error(p.pos, "missing ',' before newline in "+context)
		p.next()
	}
	return p.expect(tok)
}

func (p *parser) expectSemi() {
	// semicolon is optional before a closing ')' or '}'
	if p.tok != token.RPAREN && p.tok != token.RBRACE {
		switch p.tok {
		case token.COMMA:
			// permit a ',' instead of a ';' but complain
			p.errorExpected(p.pos, "';'")
			fallthrough
		case token.SEMICOLON:
			p.next()
		default:
			p.errorExpected(p.pos, "';'")
			//p.advance(stmtStart)
		}
	}
}

func (p *parser) atComma(context string, follow token.Token) bool {
	if p.tok == token.COMMA {
		return true
	}
	if p.tok != follow {
		msg := "missing ','"
		if p.tok == token.SEMICOLON && p.lit == "\n" {
			msg += " before newline"
		}
		p.error(p.pos, msg+" in "+context)
		return true // "insert" comma and continue
	}
	return false
}

func assert(cond bool, msg string) {
	if !cond {
		panic("hbuf/pkg/parser internal error: " + msg)
	}
}

// advance consumes tokens until the current token p.tok
// is in the 'to' set, or token.EOF. For error recovery.
func (p *parser) advance(to map[token.Token]bool) {
	for ; p.tok != token.EOF; p.next() {
		if to[p.tok] {
			// Return only if parser made some progress since last
			// sync or if it has not reached 10 advance calls without
			// progress. Otherwise consume at least one token to
			// avoid an endless parser loop (it is possible that
			// both parseOperand and parseStmt call advance and
			// correctly do not advance, thus the need for the
			// invocation limit p.syncCnt).
			if p.pos == p.syncPos && p.syncCnt < 10 {
				p.syncCnt++
				return
			}
			if p.pos > p.syncPos {
				p.syncPos = p.pos
				p.syncCnt = 0
				return
			}
			// Reaching here indicates a parser bug, likely an
			// incorrect token list in this function, but it only
			// leads to skipping of possibly correct code if a
			// previous error is present, and thus is preferred
			// over a non-terminating parse.
		}
	}
}

var declStart = map[token.Token]bool{
	token.DATA:   true,
	token.SERVER: true,
}

var exprEnd = map[token.Token]bool{
	token.COMMA:     true,
	token.COLON:     true,
	token.SEMICOLON: true,
	token.RPAREN:    true,
	token.RBRACK:    true,
	token.RBRACE:    true,
	token.LSS:       true,
}

// ----------------------------------------------------------------------------
// Identifiers
func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	if p.tok == token.IDENT {
		name = p.lit
		p.next()
	} else {
		p.expect(token.IDENT) // use expect() error handling
	}
	return &ast.Ident{NamePos: pos, Name: name}
}

// ----------------------------------------------------------------------------
// Types
func (p *parser) parseType() ast.Type {
	if p.trace {
		defer un(trace(p, "Type"))
	}

	typ := p.tryType()

	if typ == nil {
		pos := p.pos
		p.errorExpected(pos, "type")
		p.advance(exprEnd)
		return &ast.VarType{
			TypeExpr: &ast.BadExpr{
				From: pos,
				To:   p.pos,
			},
		}
	}

	return typ
}

// If the result is an identifier, it is not resolved.
func (p *parser) parseTypeName() *ast.Ident {
	if p.trace {
		defer un(trace(p, "TypeName"))
	}

	ident := p.parseIdent()
	return ident
}

func (p *parser) parseArrayType(elt *ast.VarType) *ast.ArrayType {
	if p.trace {
		defer un(trace(p, "ArrayType"))
	}

	var lBrack = p.expect(token.LBRACK)
	var rBrack = p.expect(token.RBRACK)
	return &ast.ArrayType{Lbrack: lBrack, Rbrack: rBrack, VType: elt}
}

func (p *parser) parseId() *ast.BasicLit {
	var id *ast.BasicLit
	if p.tok != token.ASSIGN {
		p.errorExpected(p.pos, "not find id key")
		return nil
	}
	p.next()
	if p.tok != token.INT {
		p.errorExpected(p.pos, "not find int")
	}
	id = &ast.BasicLit{ValuePos: p.pos, Kind: p.tok, Value: p.lit}
	p.next()
	return id
}

func (p *parser) parseFieldDecl(scope *ast.Scope) *ast.Field {
	if p.trace {
		defer un(trace(p, "FieldDecl"))
	}
	doc := p.leadComment
	tags := p.parseTags()
	typ := p.parseVarType()
	if p.tok != token.IDENT {
		p.errorExpected(p.pos, "not find Type")
	}
	name := p.parseIdent()
	var id = p.parseId()

	// Tag
	var tag *ast.BasicLit
	if p.tok == token.STRING {
		tag = &ast.BasicLit{ValuePos: p.pos, Kind: p.tok, Value: p.lit}
		p.next()
	}

	p.expectSemi() // call before accessing p.linecomment

	field := &ast.Field{
		Tags:    tags,
		Doc:     doc,
		Name:    name,
		Type:    typ,
		Id:      id,
		Tag:     tag,
		Comment: p.lineComment,
	}
	p.declare(field, nil, scope, ast.Var, name)
	p.resolve(typ)

	return field
}

///解析继承的
func (p *parser) parseExtend() []*ast.Ident {
	var extends []*ast.Ident
	if p.tok == token.COLON {
		for p.tok == token.COLON || p.tok == token.COMMA {
			p.next()
			if p.tok != token.IDENT {
				p.error(p.pos, "Data not found Extend Type")
			}
			extends = append(extends, p.parseIdent())
		}
	}
	return extends
}

func (p *parser) parseVarType() ast.Type {
	typ := p.tryIdentOrType()
	if typ == nil {
		pos := p.pos
		p.errorExpected(pos, "type")
		p.next() // make progress
		typ = &ast.VarType{
			TypeExpr: &ast.BadExpr{From: pos, To: p.pos},
		}
	}
	return typ
}

func (p *parser) parseMethodSpec() *ast.FuncType {
	if p.trace {
		defer un(trace(p, "MethodSpec"))
	}

	doc := p.leadComment
	tags := p.parseTags()
	result := p.parseVarType()
	name := p.parseIdent()
	p.expect(token.LPAREN)
	param := p.parseVarType()
	paramName := p.parseIdent()
	p.expect(token.RPAREN)
	id := p.parseId()

	var typ = &ast.FuncType{
		Result:    result.(*ast.VarType),
		Name:      name,
		Param:     param.(*ast.VarType),
		ParamName: paramName,
		Doc:       doc,
		Comment:   doc,
		Id:        id,
		Tags:      tags,
	}
	return typ
}

func (p *parser) parseMapType(value *ast.VarType) *ast.MapType {
	if p.trace {
		defer un(trace(p, "MapType"))
	}

	lss := p.expect(token.LSS)
	key := p.parseType()
	gtr := p.expect(token.GTR)
	return &ast.MapType{
		LSS:   lss,
		GTR:   gtr,
		Key:   key,
		VType: value,
	}
}

func (p *parser) tryIdentOrType() ast.Type {
	if p.tok != token.IDENT {
		defer un(trace(p, "TypeName"))
	}
	v := &ast.VarType{
		TypeExpr: p.parseTypeName(),
	}
	if p.tok == token.Question {
		v.Empty = true
		p.next()
	}
	if p.tok == token.LBRACK {
		a := p.parseArrayType(v)
		if p.tok == token.Question {
			a.Empty = true
			p.next()
		}
		return a
	} else if p.tok == token.LSS {
		m := p.parseMapType(v)
		if p.tok == token.Question {
			m.Empty = true
			p.next()
		}
		return m
	} else {
		return v
	}
}

func (p *parser) tryType() ast.Type {
	typ := p.tryIdentOrType()
	if typ != nil {
		p.resolve(typ)
	}
	return typ
}

// ----------------------------------------------------------------------------
// Statements

// Parsing modes for parseSimpleStmt.
const (
	basic = iota
	labelOk
	rangeOk
)

func (p *parser) parseTypeList() (list []ast.Expr) {
	if p.trace {
		defer un(trace(p, "TypeList"))
	}

	list = append(list, p.parseType())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseType())
	}

	return
}

func isValidImport(lit string) bool {
	const illegalChars = `!"#$%&'()*,:;<=>?[\]^{|}` + "`\uFFFD"
	s, _ := strconv.Unquote(lit) // hbuf/pkg/scanner returns a legal string literal
	for _, r := range s {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) || strings.ContainsRune(illegalChars, r) {
			return false
		}
	}
	return s != ""
}

func (p *parser) parseImportSpec(doc *ast.CommentGroup, _ token.Token, _ int) ast.Spec {
	if p.trace {
		defer un(trace(p, "ImportSpec"))
	}
	pos := p.pos
	p.expect(token.IMPORT)
	var path string
	if p.tok == token.STRING {
		path = p.lit
		if !isValidImport(path) {
			p.error(pos, "invalid import path: "+path)
		}
		p.next()
	} else {
		p.expect(token.STRING) // use expect() error handling
	}
	p.expectSemi() // call before accessing p.linecomment

	// collect imports
	spec := &ast.ImportSpec{
		Doc:     doc,
		Path:    &ast.BasicLit{ValuePos: pos, Kind: token.STRING, Value: path},
		Comment: p.lineComment,
	}
	p.imports = append(p.imports, spec)

	return spec
}

func (p *parser) parseDataSpec(doc *ast.CommentGroup, tags []*ast.Tag) ast.Spec {
	if p.trace {
		defer un(trace(p, "DataType"))
	}

	pos := p.pos
	p.expect(token.DATA)
	name := p.parseIdent()
	extends := p.parseExtend()
	id := p.parseId()

	lbrace := p.expect(token.LBRACE)
	scope := ast.NewScope(nil) // struct scope
	var list []*ast.Field
	for p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseFieldDecl(scope))
	}
	rbrace := p.expect(token.RBRACE)

	spec := &ast.TypeSpec{Doc: doc, Name: name}
	p.declare(spec, nil, p.topScope, ast.Data, name)
	if p.tok == token.ASSIGN {
		spec.Assign = p.pos
		p.next()
	}
	spec.Type = &ast.DataType{
		Tags:    tags,
		Data:    pos,
		Name:    name,
		Extends: extends,
		Id:      id,
		Fields: &ast.FieldList{
			Opening: lbrace,
			List:    list,
			Closing: rbrace,
		},
		Doc: doc,
	}
	p.expectSemi()
	spec.Comment = p.lineComment
	return spec
}

func (p *parser) parseEnumSpec(doc *ast.CommentGroup, tags []*ast.Tag) ast.Spec {
	if p.trace {
		defer un(trace(p, "EnumType"))
	}

	p.expect(token.ENUM)
	name := p.parseIdent()

	lbrace := p.expect(token.LBRACE)
	scope := ast.NewScope(nil) // struct scope
	var list []*ast.EnumItem
	for p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseEnumItem(scope))
	}
	rbrace := p.expect(token.RBRACE)

	spec := &ast.TypeSpec{Doc: doc, Name: name}
	p.declare(spec, nil, p.topScope, ast.Enum, name)
	if p.tok == token.ASSIGN {
		spec.Assign = p.pos
		p.next()
	}
	spec.Type = &ast.EnumType{
		Name:    name,
		Opening: lbrace,
		Closing: rbrace,
		Items:   list,
		Tags:    tags,
		Doc:     doc,
	}
	p.expectSemi()
	spec.Comment = p.lineComment
	return spec
}

func (p *parser) parseEnumItem(scope *ast.Scope) *ast.EnumItem {
	if p.trace {
		defer un(trace(p, "FieldDecl"))
	}
	doc := p.leadComment
	tags := p.parseTags()
	if p.tok != token.IDENT {
		p.errorExpected(p.pos, "not find Type")
	}
	name := p.parseIdent()

	var id *ast.BasicLit
	p.next()
	if p.tok != token.INT {
		p.errorExpected(p.pos, "not find int")
	}
	id = &ast.BasicLit{ValuePos: p.pos, Kind: p.tok, Value: p.lit}
	p.next()

	p.expectSemi() // call before accessing p.linecomment
	return &ast.EnumItem{
		Doc:     doc,
		Name:    name,
		Id:      id,
		Comment: p.lineComment,
		Tags:    tags,
	}
}

func (p *parser) parseServerSpec(doc *ast.CommentGroup, tags []*ast.Tag) ast.Spec {
	if p.trace {
		defer un(trace(p, "ServerType"))
	}

	pos := p.pos
	p.expect(token.SERVER)
	name := p.parseIdent()
	extends := p.parseExtend()
	id := p.parseId()

	lbrace := p.expect(token.LBRACE)
	var list []*ast.FuncType
	for p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseMethodSpec())
		p.next()
	}
	rbrace := p.expect(token.RBRACE)

	spec := &ast.TypeSpec{Doc: doc, Name: name}
	p.declare(spec, nil, p.topScope, ast.Server, name)
	if p.tok == token.ASSIGN {
		spec.Assign = p.pos
		p.next()
	}
	spec.Type = &ast.ServerType{
		Tags:    tags,
		Server:  pos,
		Name:    name,
		Extends: extends,
		Opening: lbrace,
		Methods: list,
		Closing: rbrace,
		Id:      id,
		Doc:     doc,
	}
	p.expectSemi()
	spec.Comment = p.lineComment
	return spec
}

func (p *parser) parseDecl(sync map[token.Token]bool) ast.Spec {
	if p.trace {
		defer un(trace(p, "Declaration"))
	}
	doc := p.leadComment
	tags := p.parseTags()
	switch p.tok {
	case token.DATA:
		return p.parseDataSpec(doc, tags)
	case token.SERVER:
		return p.parseServerSpec(doc, tags)
	case token.ENUM:
		return p.parseEnumSpec(doc, tags)

	default:
		pos := p.pos
		p.errorExpected(pos, "declaration")
		p.advance(sync)
		return &ast.BadSpec{From: pos, To: p.pos}
	}
}

func (p *parser) parseFile() *ast.File {
	if p.trace {
		defer un(trace(p, "File"))
	}
	if p.errors.Len() != 0 {
		return nil
	}

	packages := make(map[string]*ast.PackageSpec, 0)
	for token.PACKAGE == p.tok {
		pack := p.parsePackageSpec()
		if nil != pack && nil != pack.Name {
			packages[pack.Name.Name] = pack
		}
	}

	p.openScope()
	p.pkgScope = p.topScope
	var specs []ast.Spec
	if p.mode&PackageClauseOnly == 0 {
		// import decls
		for p.tok == token.IMPORT {
			specs = append(specs, p.parseImportSpec(nil, token.IMPORT, 0))
		}

		if p.mode&ImportsOnly == 0 {
			// rest of package body
			for p.tok != token.EOF {
				specs = append(specs, p.parseDecl(declStart))
			}
		}
	}
	p.closeScope()
	assert(p.topScope == nil, "unbalanced scopes")
	assert(p.labelScope == nil, "unbalanced label scopes")

	// resolve global identifiers within the same file
	i := 0
	for _, ident := range p.unresolved {
		// i <= index for current ident
		assert(ident.Obj == unresolved, "object already resolved")
		ident.Obj = p.pkgScope.Lookup(ident.Name) // also removes unresolved sentinel
		if ident.Obj == nil {
			p.unresolved[i] = ident
			i++
		}
	}

	return &ast.File{
		//Doc:        doc,
		Packages:   packages,
		Specs:      specs,
		Scope:      p.pkgScope,
		Imports:    p.imports,
		Unresolved: p.unresolved[0:i],
		Comments:   p.comments,
	}
}

//解析包名
func (p *parser) parsePackageSpec() *ast.PackageSpec {
	pos := p.expect(token.PACKAGE)
	if p.trace {
		return nil
	}

	var ident *ast.Ident
	if token.IDENT == p.tok {
		ident = p.parseIdent()
	}

	if token.ASSIGN != p.tok {
		p.error(pos, "invalid import path:= ")
	}
	p.next()
	var path string
	if p.tok == token.STRING {
		path = p.lit
		p.next()
	} else {
		p.expect(token.STRING)
	}
	p.expectSemi()

	return &ast.PackageSpec{
		Doc:     p.lineComment,
		Name:    ident,
		Value:   &ast.BasicLit{ValuePos: pos, Kind: token.STRING, Value: path},
		Comment: p.lineComment,
	}
}

func (p *parser) parseTags() []*ast.Tag {
	var ret []*ast.Tag
	for token.LBRACK == p.tok {
		tag := p.parseTag()
		if nil != tag {
			ret = append(ret, tag)
		}
	}
	return ret
}

func (p *parser) parseTag() *ast.Tag {
	pos := p.pos
	p.next()
	name := p.parseIdent()
	if token.COLON != p.tok {
		p.error(p.pos, "Not find tag name ")
		return nil
	}
	p.next()

	var kvs []*ast.KeyValue
	for token.EOF != p.tok && token.RBRACK != p.tok {
		kvs = append(kvs, p.parseKeyValue())
		if token.SEMICOLON == p.tok {
			p.next()
		}
	}
	p.expect(token.RBRACK)
	p.expectSemi()
	return &ast.Tag{
		Name:    name,
		KV:      kvs,
		Opening: pos,
	}
}

func (p *parser) parseKeyValue() *ast.KeyValue {
	name := p.parseIdent()
	if token.ASSIGN != p.tok {
		p.error(p.pos, "syntax error ")
		return nil
	}
	p.next()

	values := make([]*ast.BasicLit, 0)
	for {
		pos := p.pos
		var value string
		if p.tok == token.STRING {
			value = p.lit
			p.next()
		} else {
			p.expect(token.STRING)
		}
		values = append(values, &ast.BasicLit{ValuePos: pos, Kind: token.STRING, Value: value})
		if token.COMMA != p.tok {
			break
		}
		p.next()
	}
	return &ast.KeyValue{
		Name:   name,
		Values: values,
	}
}
