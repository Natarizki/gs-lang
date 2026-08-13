package parser

import "gs-lang/lexer"

// Node adalah interface dasar untuk semua node AST
type Node interface {
	TokenLiteral() string
}

// Statement adalah node yang tidak menghasilkan value
type Statement interface {
	Node
	statementNode()
}

// Expression adalah node yang menghasilkan value
type Expression interface {
	Node
	expressionNode()
}

// ============ ROOT ============

// Program adalah root dari setiap file GS
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// ============ STATEMENTS ============

// LetStatement: nama = nilai
type LetStatement struct {
	Token lexer.Token // token IDENT
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

// ReturnStatement: return a, b
type ReturnStatement struct {
	Token       lexer.Token
	ReturnValues []Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// ExpressionStatement: statement yang cuma berisi satu expression
// (misal pemanggilan fungsi berdiri sendiri: sapa("Budi"))
type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// BlockStatement: kumpulan statement dalam satu blok ({} atau indentasi)
type BlockStatement struct {
	Token      lexer.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// IfStatement: if / but if / else
type IfStatement struct {
	Token       lexer.Token
	Condition   Expression
	Consequence *BlockStatement
	ButIfs      []*ButIfClause // "but if" chain
	Alternative *BlockStatement // else
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }

// ButIfClause: satu klausa "but if kondisi { ... }"
type ButIfClause struct {
	Token       lexer.Token
	Condition   Expression
	Consequence *BlockStatement
}

// MatchCase: satu case dalam match, misal "18) print(...)"
type MatchCase struct {
	Value Expression
	Body  *BlockStatement
}

// MatchStatement: match nama { 1) ... 2) ... }
type MatchStatement struct {
	Token lexer.Token
	Value Expression
	Cases []*MatchCase
}

func (ms *MatchStatement) statementNode()       {}
func (ms *MatchStatement) TokenLiteral() string { return ms.Token.Literal }

// ForStatement: for i in range(5) { ... }
type ForStatement struct {
	Token    lexer.Token
	Variable *Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }

// BreakStatement: break
type BreakStatement struct {
	Token lexer.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }

// ContinueStatement: continue
type ContinueStatement struct {
	Token lexer.Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }

// ErrorStatement: error "pesan"
type ErrorStatement struct {
	Token   lexer.Token
	Message Expression
}

func (es *ErrorStatement) statementNode()       {}
func (es *ErrorStatement) TokenLiteral() string { return es.Token.Literal }

// FuncStatement: fnc nama(params) { ... } ATAU fnc nama(params) return-block
type FuncStatement struct {
	Token      lexer.Token
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fs *FuncStatement) statementNode()       {}
func (fs *FuncStatement) TokenLiteral() string { return fs.Token.Literal }

// StructStatement: nama_struct { field1 \n field2 }
type StructStatement struct {
	Token  lexer.Token
	Name   *Identifier
	Fields []*Identifier
}

func (ss *StructStatement) statementNode()       {}
func (ss *StructStatement) TokenLiteral() string { return ss.Token.Literal }

// ImportStatement: import "dir/file.gs" ATAU import "GS/paket"
type ImportStatement struct {
	Token lexer.Token
	Path  string
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }

// UseStatement: use nama_struct_instance
type UseStatement struct {
	Token lexer.Token
	Name  *Identifier
}

func (us *UseStatement) statementNode()       {}
func (us *UseStatement) TokenLiteral() string { return us.Token.Literal }

// TryStatement: try x = bagi(10, 0)
type TryStatement struct {
	Token       lexer.Token
	Target      *Identifier
	Call        Expression
}

func (ts *TryStatement) statementNode()       {}
func (ts *TryStatement) TokenLiteral() string { return ts.Token.Literal }

// TaskBlock: { ... } top-level task (dipakai untuk concurrency)
type TaskBlock struct {
	Token lexer.Token
	ID    string // "001", "002", dst -- diisi otomatis
	Body  *BlockStatement
}

func (tb *TaskBlock) statementNode()       {}
func (tb *TaskBlock) TokenLiteral() string { return tb.Token.Literal }

// StartThenStatement: start = id::001  /  then = id::002 | id::003
type StartThenStatement struct {
	Token   lexer.Token
	IsStart bool // true = "start", false = "then"
	TaskIDs []string
}

func (sts *StartThenStatement) statementNode()       {}
func (sts *StartThenStatement) TokenLiteral() string { return sts.Token.Literal }

// SpawnStatement: spawn tugas1()
type SpawnStatement struct {
	Token lexer.Token
	Call  Expression
}

func (ss *SpawnStatement) statementNode()       {}
func (ss *SpawnStatement) TokenLiteral() string { return ss.Token.Literal }

// WaitStatement: wait
type WaitStatement struct {
	Token lexer.Token
}

func (ws *WaitStatement) statementNode()       {}
func (ws *WaitStatement) TokenLiteral() string { return ws.Token.Literal }

// ============ EXPRESSIONS ============

// Identifier: nama, umur, foo
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// IntegerLiteral: 123
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// FloatLiteral: 12.3
type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }

// StringLiteral: "halo {nama}"
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

// BooleanLiteral: true / false
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }

// NullLiteral: null / _
type NullLiteral struct {
	Token lexer.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }

// ListLiteral: [1, 2, 3]
type ListLiteral struct {
	Token    lexer.Token
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Literal }

// MapLiteral: {"kunci": "nilai"}
type MapLiteral struct {
	Token lexer.Token
	Pairs map[Expression]Expression
}

func (ml *MapLiteral) expressionNode()      {}
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }

// PrefixExpression: -5, !aktif
type PrefixExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

// InfixExpression: a + b, umur >= 18
type InfixExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

// CallExpression: bagi(10, 2)
type CallExpression struct {
	Token     lexer.Token
	Function  Expression // Identifier atau FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

// IndexExpression: angka[0], angka[1:3]
type IndexExpression struct {
	Token lexer.Token
	Left  Expression
	Index Expression
	// dipakai untuk slicing angka[1:3], End != nil artinya ini slice
	End Expression
}

func (ix *IndexExpression) expressionNode()      {}
func (ix *IndexExpression) TokenLiteral() string { return ix.Token.Literal }

// DotExpression: orang.nama (akses field struct)
type DotExpression struct {
	Token  lexer.Token
	Left   Expression
	Field  *Identifier
}

func (de *DotExpression) expressionNode()      {}
func (de *DotExpression) TokenLiteral() string { return de.Token.Literal }

// FunctionLiteral: function sebagai value (buat first-class function)
type FunctionLiteral struct {
	Token      lexer.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }

// StartIfStatement: startif kondisi { start = id::00X } elstart { start = id::00Y }
// Dipakai HANYA di dalam task block (di luar task block pakai StartThenStatement biasa)
type StartIfStatement struct {
	Token       lexer.Token
	Condition   Expression
	Consequence []string // task ID yang dipanggil kalau kondisi true
	Alternative []string // task ID yang dipanggil kalau kondisi false (elstart)
}

func (sis *StartIfStatement) statementNode()       {}
func (sis *StartIfStatement) TokenLiteral() string { return sis.Token.Literal }

// StructInstanceLiteral: budi { nama = "Budi" \n umur = 20 }
type StructInstanceLiteral struct {
	Token  lexer.Token
	Name   *Identifier
	Fields map[string]Expression
}

func (sil *StructInstanceLiteral) statementNode()       {}
func (sil *StructInstanceLiteral) TokenLiteral() string { return sil.Token.Literal }

// MethodStatement: fnc (o orang) sapa { ... }
type MethodStatement struct {
	Token         lexer.Token
	ReceiverName  *Identifier // "o"
	ReceiverType  *Identifier // "orang"
	Name          *Identifier // "sapa"
	Parameters    []*Identifier
	Body          *BlockStatement
}

func (ms *MethodStatement) statementNode()       {}
func (ms *MethodStatement) TokenLiteral() string { return ms.Token.Literal }

// InterpolatedString: "halo {nama}, umur {umur}"
// Parts berisi campuran string statis dan Expression (untuk bagian {..})
type InterpolatedString struct {
	Token lexer.Token
	Parts []InterpolationPart
}

type InterpolationPart struct {
	IsExpr bool
	Text   string     // dipakai kalau IsExpr == false
	Expr   Expression // dipakai kalau IsExpr == true
}

func (is *InterpolatedString) expressionNode()      {}
func (is *InterpolatedString) TokenLiteral() string { return is.Token.Literal }

// GetLine mengembalikan nomor baris dari sebuah node, dipakai untuk
// line-tracking bytecode generator (lihat bytecode.nodeLine)
func (es *ExpressionStatement) GetLine() int { return es.Token.Line }
func (ls *LetStatement) GetLine() int        { return ls.Token.Line }
func (rs *ReturnStatement) GetLine() int     { return rs.Token.Line }
func (is *IfStatement) GetLine() int         { return is.Token.Line }
func (fs *ForStatement) GetLine() int        { return fs.Token.Line }
func (ms *MatchStatement) GetLine() int      { return ms.Token.Line }
func (es2 *ErrorStatement) GetLine() int     { return es2.Token.Line }
func (ts *TryStatement) GetLine() int        { return ts.Token.Line }
func (fs2 *FuncStatement) GetLine() int      { return fs2.Token.Line }
func (ss *StructStatement) GetLine() int     { return ss.Token.Line }
func (ms2 *MethodStatement) GetLine() int    { return ms2.Token.Line }
func (ce *CallExpression) GetLine() int      { return ce.Token.Line }
func (ie *InfixExpression) GetLine() int     { return ie.Token.Line }
func (pe *PrefixExpression) GetLine() int    { return pe.Token.Line }
func (id *Identifier) GetLine() int          { return id.Token.Line }
func (ix *IndexExpression) GetLine() int     { return ix.Token.Line }
func (de *DotExpression) GetLine() int       { return de.Token.Line }
