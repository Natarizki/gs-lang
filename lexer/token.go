package lexer

// TokenType mendefinisikan jenis-jenis token dalam bahasa GS
type TokenType string

const (
	// Khusus
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"
	NEWLINE TokenType = "NEWLINE"
	INDENT  TokenType = "INDENT"
	DEDENT  TokenType = "DEDENT"

	// Literal
	IDENT  TokenType = "IDENT"  // nama, umur, foo
	INT    TokenType = "INT"    // 123
	FLOAT  TokenType = "FLOAT"  // 12.3
	STRING TokenType = "STRING" // "halo"

	// Keyword
	FNC      TokenType = "FNC"
	RETURN   TokenType = "RETURN"
	IF       TokenType = "IF"
	BUT      TokenType = "BUT"   // "but if"
	ELSE     TokenType = "ELSE"
	MATCH    TokenType = "MATCH"
	FOR      TokenType = "FOR"
	IN       TokenType = "IN"
	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"
	ERROR    TokenType = "ERROR"
	TRY      TokenType = "TRY"
	IMPORT   TokenType = "IMPORT"
	USE      TokenType = "USE"
	MAKE     TokenType = "MAKE"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	NULL     TokenType = "NULL" // null atau _
	SPAWN    TokenType = "SPAWN"
	WAIT     TokenType = "WAIT"
	START    TokenType = "START"
	THEN     TokenType = "THEN"
	STARTIF  TokenType = "STARTIF"
	ELSTART  TokenType = "ELSTART"

	// Operator
	ASSIGN   TokenType = "="
	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	ASTERISK TokenType = "*"
	SLASH    TokenType = "/"
	PERCENT  TokenType = "%"
	POWER    TokenType = "**"

	EQ     TokenType = "=="
	NOT_EQ TokenType = "!="
	LT     TokenType = "<"
	GT     TokenType = ">"
	LT_EQ  TokenType = "<="
	GT_EQ  TokenType = ">="

	AND TokenType = "&&"
	OR  TokenType = "||"
	NOT TokenType = "!"

	PIPE TokenType = "|" // dipakai di "then = id::002 | id::003"

	// Delimiter
	COMMA     TokenType = ","
	COLON     TokenType = ":"
	DOUBLE_COLON TokenType = "::" // dipakai di id::001
	DOT       TokenType = "."
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	LBRACE    TokenType = "{"
	RBRACE    TokenType = "}"
	LBRACKET  TokenType = "["
	RBRACKET  TokenType = "]"
)

// keywords memetakan kata kunci bahasa GS ke TokenType-nya
var keywords = map[string]TokenType{
	"fnc":      FNC,
	"return":   RETURN,
	"if":       IF,
	"but":      BUT,
	"else":     ELSE,
	"match":    MATCH,
	"for":      FOR,
	"in":       IN,
	"break":    BREAK,
	"continue": CONTINUE,
	"error":    ERROR,
	"try":      TRY,
	"import":   IMPORT,
	"use":      USE,
	"make":     MAKE,
	"true":     TRUE,
	"false":    FALSE,
	"null":     NULL,
	"spawn":    SPAWN,
	"wait":     WAIT,
	"start":    START,
	"then":     THEN,
	"startif":  STARTIF,
	"elstart":  ELSTART,
        "_":        NULL,
}
// Token merepresentasikan satu token hasil lexing
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// LookupIdent mengecek apakah sebuah identifier adalah keyword,
// jika bukan, dianggap IDENT biasa.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
