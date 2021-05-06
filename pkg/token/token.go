package token

import (
	"strconv"
	"unicode"
	"unicode/utf8"
)

type Token int

const (
	ILLEGAL Token = iota
	EOF
	COMMENT

	literal_beg
	IDENT  // main
	INT    // 12345
	STRING // "abc"
	literal_end

	operator_beg
	ASSIGN    // =
	LSS       // <
	LPAREN    // (
	LBRACK    // [
	LBRACE    // {
	COMMA     // ,
	GTR       // >
	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :
	operator_end

	keyword_beg
	IMPORT
	PACKAGE
	DATA
	SERVER
	ENUM
	keyword_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:  "IDENT",
	INT:    "INT",
	STRING: "STRING",

	LSS:    "<",
	GTR:    ">",
	ASSIGN: "=",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",

	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",
	IMPORT:    "import",
	PACKAGE:   "package",

	DATA:   "data",
	SERVER: "server",
	ENUM:   "enum",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 7
)

func (op Token) Precedence() int {
	switch op {
	case LSS, GTR:
	}
	return LowestPrec
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

func Lookup(ident string) Token {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	return IDENT
}

func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

func (tok Token) IsOperator() bool { return operator_beg < tok && tok < operator_end }

func (tok Token) IsKeyword() bool { return keyword_beg < tok && tok < keyword_end }

func IsExported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}

func IsKeyword(name string) bool {
	// TODO: opt: use a perfect hash function instead of a global map.
	_, ok := keywords[name]
	return ok
}

func IsIdentifier(name string) bool {
	for i, c := range name {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return name != "" && !IsKeyword(name)
}
