package ast

import (
	"hbuf/pkg/token"
	"strings"
)

type Node interface {
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

type Expr interface {
	Node
	exprNode()
}

type Comment struct {
	Slash token.Pos // position of "/" starting the comment
	Text  string    // comment text (excluding '\n' for //-style comments)
}

func (c *Comment) Pos() token.Pos { return c.Slash }
func (c *Comment) End() token.Pos { return token.Pos(int(c.Slash) + len(c.Text)) }

type CommentGroup struct {
	List []*Comment // len(List) > 0
}

func (g *CommentGroup) Pos() token.Pos { return g.List[0].Pos() }
func (g *CommentGroup) End() token.Pos { return g.List[len(g.List)-1].End() }

func isWhitespace(ch byte) bool { return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' }

func stripTrailingWhitespace(s string) string {
	i := len(s)
	for i > 0 && isWhitespace(s[i-1]) {
		i--
	}
	return s[0:i]
}

func (g *CommentGroup) Text() string {
	if g == nil {
		return ""
	}
	comments := make([]string, len(g.List))
	for i, c := range g.List {
		comments[i] = c.Text
	}

	lines := make([]string, 0, 10) // most comments are less than 10 lines
	for _, c := range comments {
		// Remove comment markers.
		// The parser has given us exactly the comment text.
		switch c[1] {
		case '/':
			//-style comment (no newline at the end)
			c = c[2:]
			if len(c) == 0 {
				// empty line
				break
			}
			if c[0] == ' ' {
				// strip first space - required for Example tests
				c = c[1:]
				break
			}
			if isDirective(c) {
				// Ignore //go:noinline, //line, and so on.
				continue
			}
		case '*':
			/*-style comment */
			c = c[2 : len(c)-2]
		}

		// Split on newlines.
		cl := strings.Split(c, "\n")

		// Walk lines, stripping trailing white space and adding to list.
		for _, l := range cl {
			lines = append(lines, stripTrailingWhitespace(l))
		}
	}

	// Remove leading blank lines; convert runs of
	// interior blank lines to a single blank line.
	n := 0
	for _, line := range lines {
		if line != "" || n > 0 && lines[n-1] != "" {
			lines[n] = line
			n++
		}
	}
	lines = lines[0:n]

	// Add final "" entry to get trailing newline from Join.
	if n > 0 && lines[n-1] != "" {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// isDirective reports whether c is a comment directive.
func isDirective(c string) bool {
	// "//line " is a line directive.
	// (The // has been removed.)
	if strings.HasPrefix(c, "line ") {
		return true
	}

	// "//[a-z0-9]+:[a-z0-9]"
	// (The // has been removed.)
	colon := strings.Index(c, ":")
	if colon <= 0 || colon+1 >= len(c) {
		return false
	}
	for i := 0; i <= colon+1; i++ {
		if i == colon {
			continue
		}
		b := c[i]
		if !('a' <= b && b <= 'z' || '0' <= b && b <= '9') {
			return false
		}
	}
	return true
}

type Field struct {
	Tags    []*Tag
	Doc     *CommentGroup // associated documentation; or nil
	Name    *Ident        // field/method/parameter names; or nil
	Type    Type          // field/method/parameter type
	Id      *BasicLit     // field tag; or nil
	Tag     *BasicLit     // field tag; or nil
	Comment *CommentGroup // line comments; or nil
}

func (f *Field) Pos() token.Pos {
	//if len(f.Type) > 0 {
	//	return f.Names[0].Pos()
	//}
	return f.Type.Pos()
}

func (f *Field) End() token.Pos {
	if f.Tag != nil {
		return f.Tag.End()
	}
	return f.Type.End()
}

type FieldList struct {
	Opening token.Pos // position of opening parenthesis/brace, if any
	List    []*Field  // field list; or nil
	Closing token.Pos // position of closing parenthesis/brace, if any
}

func (f *FieldList) Pos() token.Pos {
	if f.Opening.IsValid() {
		return f.Opening
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if len(f.List) > 0 {
		return f.List[0].Pos()
	}
	return token.NoPos
}

func (f *FieldList) End() token.Pos {
	if f.Closing.IsValid() {
		return f.Closing + 1
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if n := len(f.List); n > 0 {
		return f.List[n-1].End()
	}
	return token.NoPos
}

func (f *FieldList) NumFields() int {
	//n := 0
	//if f != nil {
	//	for _, g := range f.List {
	//		m := len(g.Names)
	//		if m == 0 {
	//			m = 1
	//		}
	//		n += m
	//	}
	//}
	return 0
}

type (
	BadExpr struct {
		From, To token.Pos // position range of bad expression
	}

	Ident struct {
		NamePos token.Pos // identifier position
		Name    string    // identifier name
		Obj     *Object   // denoted object; or nil
	}

	BasicLit struct {
		ValuePos token.Pos   // literal position
		Kind     token.Token // token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
		Value    string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	}

	// A FuncLit node represents a function literal.
	FuncLit struct {
		Type *FuncType // function type
	}

	// A CompositeLit node represents a composite literal.
	CompositeLit struct {
		Type       Expr      // literal type; or nil
		Lbrace     token.Pos // position of "{"
		Elts       []Expr    // list of composite elements; or nil
		Rbrace     token.Pos // position of "}"
		Incomplete bool      // true if (source) expressions are missing in the Elts list
	}
)

type ChanDir int

const (
	SEND ChanDir = 1 << iota
	RECV
)

type Type interface {
	Expr
	Type() Expr
	IsEmpty() bool
	typeNode()
}

type (
	KeyValue struct {
		Name  *Ident    // local package name (including "."); or nil
		Value *BasicLit // import path
	}

	Tag struct {
		Name    *Ident // local package name (including "."); or nil
		KV      []*KeyValue
		Opening token.Pos
		Closing token.Pos
	}

	// VarType 变量类型
	VarType struct {
		TypeExpr Expr //类型名称
		Empty    bool //是否可为空
	}

	ArrayType struct {
		VType  *VarType
		Empty  bool      //是否可为空
		Lbrack token.Pos // position of "["
		Rbrack token.Pos // position of "]"
	}

	MapType struct {
		VType *VarType
		Empty bool //是否可为空
		Key   Expr
		LSS   token.Pos // position of "<"
		GTR   token.Pos // position of ">"
	}

	DataType struct {
		Tags       []*Tag
		Data       token.Pos  // position of "data" keyword
		Fields     *FieldList // list of field declarations
		Incomplete bool       // true if (source) fields are missing in the Fields list
		Name       *Ident
		Extends    []*Ident
		Id         *BasicLit // field Id; or nil
	}

	ServerType struct {
		Tags    []*Tag
		Server  token.Pos // position of "server" keyword
		Name    *Ident
		Extends []*Ident
		Opening token.Pos // position of opening parenthesis/brace, if any
		Methods []*FuncType
		Closing token.Pos     // position of closing parenthesis/brace, if any
		Doc     *CommentGroup // associated documentation; or nil
		Comment *CommentGroup // line comments; or nil
		Id      *BasicLit     // field Id; or nil
	}

	FuncType struct {
		Tags      []*Tag
		Result    *VarType
		Name      *Ident
		Param     *VarType // (incoming) parameters; non-nil
		ParamName *Ident
		Doc       *CommentGroup // associated documentation; or nil
		Comment   *CommentGroup // line comments; or nil
		Id        *BasicLit     // field Id; or nil
	}

	EnumType struct {
		Enum    token.Pos
		Name    *Ident
		Items   []*EnumItem
		Opening token.Pos
		Closing token.Pos
		Doc     *CommentGroup // associated documentation; or nil
		Comment *CommentGroup // line comments; or nil
	}

	EnumItem struct {
		Tags    []*Tag
		Name    *Ident        // field/method/parameter names; or nil
		Id      *BasicLit     // field tag; or nil
		Doc     *CommentGroup // associated documentation; or nil
		Comment *CommentGroup // line comments; or nil
	}
)

func (x *BadExpr) Pos() token.Pos  { return x.From }
func (x *Ident) Pos() token.Pos    { return x.NamePos }
func (x *BasicLit) Pos() token.Pos { return x.ValuePos }
func (x *FuncLit) Pos() token.Pos  { return x.Type.Pos() }
func (x *CompositeLit) Pos() token.Pos {
	if x.Type != nil {
		return x.Type.Pos()
	}
	return x.Lbrace
}
func (x *ArrayType) Pos() token.Pos { return x.VType.Pos() }
func (x *DataType) Pos() token.Pos  { return x.Data }
func (x *FuncType) Pos() token.Pos {
	return x.Result.Pos()
}
func (x *ServerType) Pos() token.Pos { return x.Server }
func (x *MapType) Pos() token.Pos    { return x.VType.Pos() }
func (x *EnumType) Pos() token.Pos   { return x.Enum }
func (x *EnumItem) Pos() token.Pos   { return x.Name.Pos() }
func (x *VarType) Pos() token.Pos    { return x.TypeExpr.Pos() }
func (x *Tag) Pos() token.Pos        { return x.Opening }
func (x *KeyValue) Pos() token.Pos   { return x.Name.Pos() }

func (x *BadExpr) End() token.Pos      { return x.To }
func (x *Ident) End() token.Pos        { return token.Pos(int(x.NamePos) + len(x.Name)) }
func (x *BasicLit) End() token.Pos     { return token.Pos(int(x.ValuePos) + len(x.Value)) }
func (x *FuncLit) End() token.Pos      { return x.Type.End() }
func (x *CompositeLit) End() token.Pos { return x.Rbrace + 1 }
func (x *ArrayType) End() token.Pos    { return x.Rbrack }
func (x *DataType) End() token.Pos     { return x.Fields.End() }
func (x *FuncType) End() token.Pos     { return x.Param.End() }
func (x *ServerType) End() token.Pos   { return x.Closing }
func (x *MapType) End() token.Pos      { return x.GTR }
func (x *EnumType) End() token.Pos     { return x.Items[len(x.Items)-1].End() }
func (x *EnumItem) End() token.Pos     { return x.Comment.End() }
func (x *VarType) End() token.Pos      { return x.TypeExpr.End() }
func (x *Tag) End() token.Pos          { return x.Closing }
func (x *KeyValue) End() token.Pos     { return x.Value.End() }

func (*BadExpr) exprNode()      {}
func (*Ident) exprNode()        {}
func (*BasicLit) exprNode()     {}
func (*FuncLit) exprNode()      {}
func (*CompositeLit) exprNode() {}
func (*ArrayType) exprNode()    {}
func (*DataType) exprNode()     {}
func (*FuncType) exprNode()     {}
func (*ServerType) exprNode()   {}
func (*MapType) exprNode()      {}
func (*EnumType) exprNode()     {}
func (*EnumItem) exprNode()     {}
func (*VarType) exprNode()      {}

func (x *VarType) Type() Expr   { return x.TypeExpr }
func (x *ArrayType) Type() Expr { return x.VType }
func (x *MapType) Type() Expr   { return x.VType }

func (x *VarType) IsEmpty() bool   { return x.Empty }
func (x *ArrayType) IsEmpty() bool { return x.Empty }
func (x *MapType) IsEmpty() bool   { return x.Empty }

func (*VarType) typeNode()   {}
func (*ArrayType) typeNode() {}
func (*MapType) typeNode()   {}

func (id *Ident) IsExported() bool { return token.IsExported(id.Name) }
func (id *Ident) String() string {
	if id != nil {
		return id.Name
	}
	return "<nil>"
}

type (
	Spec interface {
		Node
		specNode()
	}

	BadSpec struct {
		From token.Pos
		To   token.Pos
	}

	ImportSpec struct {
		Doc     *CommentGroup // associated documentation; or nil
		Path    *BasicLit     // import path
		Comment *CommentGroup // line comments; or nil
		EndPos  token.Pos     // end of spec (overrides Value.Pos if nonzero)
	}

	PackageSpec struct {
		Doc     *CommentGroup // associated documentation; or nil
		Name    *Ident        // local package name (including "."); or nil
		Value   *BasicLit     // import path
		Comment *CommentGroup // line comments; or nil
		EndPos  token.Pos     // end of spec (overrides Value.Pos if nonzero)
	}

	TypeSpec struct {
		Doc     *CommentGroup // associated documentation; or nil
		Name    *Ident        // type name
		Assign  token.Pos     // position of '=', if any
		Type    Expr          // *Ident, *ParenExpr, *SelectorExpr, *StarExpr, or any of the *XxxTypes
		Comment *CommentGroup // line comments; or nil
	}
)

func (s *ImportSpec) Pos() token.Pos {
	return s.Path.Pos()
}
func (s *TypeSpec) Pos() token.Pos { return s.Name.Pos() }

func (s *ImportSpec) End() token.Pos {
	if s.EndPos != 0 {
		return s.EndPos
	}
	return s.Path.End()
}

func (s *TypeSpec) End() token.Pos { return s.Type.End() }

func (s *BadSpec) Pos() token.Pos {
	return s.From
}

func (s *BadSpec) End() token.Pos {
	return s.To
}

func (*ImportSpec) specNode() {}

func (*TypeSpec) specNode() {}

func (*BadSpec) specNode() {}

type File struct {
	Doc        *CommentGroup // associated documentation; or nil
	Packages   map[string]*PackageSpec
	Specs      []Spec
	Scope      *Scope          // package scope (this file only)
	Imports    []*ImportSpec   // imports in this file
	Unresolved []*Ident        // unresolved identifiers in this file
	Comments   []*CommentGroup // list of all comments in the source file
	Path       string
}

func (f *File) Pos() token.Pos {
	for _, item := range f.Packages {
		return item.EndPos
	}
	return token.NoPos
}
func (f *File) End() token.Pos {
	if n := len(f.Specs); n > 0 {
		return f.Specs[n-1].End()
	}

	var temp *PackageSpec
	for _, item := range f.Packages {
		temp = item
	}
	if nil != temp {
		return temp.EndPos
	}
	return token.NoPos
}

type Package struct {
	Scope   *Scope             // package scope across all files
	Imports map[string]*Object // map of package id -> package object
	Files   map[string]*File   // Go source files by filename
}

func NewPackage() *Package {
	return &Package{
		Files:   map[string]*File{},
		Imports: map[string]*Object{},
	}
}
func (p *Package) Pos() token.Pos { return token.NoPos }
func (p *Package) End() token.Pos { return token.NoPos }
