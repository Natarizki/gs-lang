package parser

import (
	"fmt"
	"strconv"
        "strings"

	"gs-lang/lexer"
)

const (
	_ int = iota
	LOWEST
	OR_PREC
	AND_PREC
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	POWER_PREC
	PREFIX
	CALL
	INDEX
)

var precedences = map[lexer.TokenType]int{
	lexer.OR:       OR_PREC,
	lexer.AND:      AND_PREC,
	lexer.EQ:       EQUALS,
	lexer.NOT_EQ:   EQUALS,
	lexer.LT:       LESSGREATER,
	lexer.GT:       LESSGREATER,
	lexer.LT_EQ:    LESSGREATER,
	lexer.GT_EQ:    LESSGREATER,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.SLASH:    PRODUCT,
	lexer.ASTERISK: PRODUCT,
	lexer.PERCENT:  PRODUCT,
	lexer.POWER:    POWER_PREC,
	lexer.LPAREN:   CALL,
	lexer.LBRACKET: INDEX,
	lexer.DOT:      INDEX,
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

type Parser struct {
	l *lexer.Lexer

	curToken  lexer.Token
	peekToken lexer.Token

	errors []string

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBoolean)
	p.registerPrefix(lexer.FALSE, p.parseBoolean)
	p.registerPrefix(lexer.NULL, p.parseNull)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LBRACKET, p.parseListLiteral)
	p.registerPrefix(lexer.LBRACE, p.parseMapLiteral)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	for _, tt := range []lexer.TokenType{
		lexer.PLUS, lexer.MINUS, lexer.SLASH, lexer.ASTERISK, lexer.PERCENT, lexer.POWER,
		lexer.EQ, lexer.NOT_EQ, lexer.LT, lexer.GT, lexer.LT_EQ, lexer.GT_EQ,
		lexer.AND, lexer.OR,
	} {
		p.registerInfix(tt, p.parseInfixExpression)
	}
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACKET, p.parseIndexExpression)
	p.registerInfix(lexer.DOT, p.parseDotExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tt lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tt] = fn
}

func (p *Parser) registerInfix(tt lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tt] = fn
}

func (p *Parser) Errors() []string { return p.errors }

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) skipNewlines() {
	for p.curToken.Type == lexer.NEWLINE {
		p.nextToken()
	}
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool  { return p.curToken.Type == t }
func (p *Parser) peekTokenIs(t lexer.TokenType) bool { return p.peekToken.Type == t }

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("line %d: expected token %s, got %s (%q)",
		p.peekToken.Line, t, p.peekToken.Type, p.peekToken.Literal)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("line %d: no parse rule for token %s (%q)",
		p.curToken.Line, t, p.curToken.Literal)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) ParseProgram() *Program {
	return p.parseProgramInternal(false)
}

func (p *Parser) ParseProgramStrict() *Program {
	return p.parseProgramInternal(true)
}

func (p *Parser) parseProgramInternal(strict bool) *Program {
	program := &Program{Statements: []Statement{}}

	p.skipNewlines()
	for !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			if strict && !isAllowedTopLevel(stmt) {
				p.errors = append(p.errors, fmt.Sprintf(
					"line %d: statement not allowed outside a task block { } — only task blocks, import, start, and then are allowed at top level",
					p.curToken.Line))
			}
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
		p.skipNewlines()
	}

	return program
}

func isAllowedTopLevel(stmt Statement) bool {
	switch stmt.(type) {
	case *TaskBlock, *ImportStatement, *StartThenStatement:
		return true
	default:
		return false
	}
}


func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case lexer.FNC:
		return p.parseFuncStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.MATCH:
		return p.parseMatchStatement()
	case lexer.BREAK:
		return &BreakStatement{Token: p.curToken}
	case lexer.CONTINUE:
		return &ContinueStatement{Token: p.curToken}
	case lexer.ERROR:
		return p.parseErrorStatement()
	case lexer.IMPORT:
		return p.parseImportStatement()
	case lexer.USE:
		return p.parseUseStatement()
	case lexer.TRY:
		return p.parseTryStatement()
	case lexer.SPAWN:
		return p.parseSpawnStatement()
	case lexer.WAIT:
		return &WaitStatement{Token: p.curToken}
	case lexer.MAKE:
		return p.parseLetStatement(true)
	case lexer.START, lexer.THEN:
		return p.parseStartThenStatement()
	case lexer.STARTIF:
		return p.parseStartIfStatement()
	case lexer.LBRACE:
		return p.parseTaskOrBlockAsStatement()
	case lexer.IDENT:
		if p.peekTokenIs(lexer.ASSIGN) {
			return p.parseLetStatement(false)
		}
		if p.peekTokenIs(lexer.LBRACE) {
			return p.parseStructStatement()
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement(isMake bool) Statement {
	stmt := &LetStatement{Token: p.curToken}

	if isMake {
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseReturnStatement() Statement {
	stmt := &ReturnStatement{Token: p.curToken}

	p.nextToken()
	if p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.EOF) {
		return stmt
	}

	stmt.ReturnValues = append(stmt.ReturnValues, p.parseExpression(LOWEST))
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		stmt.ReturnValues = append(stmt.ReturnValues, p.parseExpression(LOWEST))
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() Statement {
	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseBlock() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = []Statement{}

	// kasus 1: curToken sendiri sudah '{' (dipanggil dari task block)
	if p.curTokenIs(lexer.LBRACE) {
		p.nextToken()
		p.skipNewlines()
		for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
			stmt := p.parseStatement()
			if stmt != nil {
				block.Statements = append(block.Statements, stmt)
			}
			p.nextToken()
			p.skipNewlines()
		}
		return block
	}

	// kasus 2: peekToken adalah '{' (dipanggil dari if/for/fnc, curToken masih token sebelumnya)
	if !p.expectPeek(lexer.LBRACE) {
		return block
	}
	p.nextToken()
	p.skipNewlines()
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
		p.skipNewlines()
	}
	return block
}

func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.NEWLINE) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: could not parse %q as integer",
			p.curToken.Line, p.curToken.Literal))
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() Expression {
	lit := &FloatLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: could not parse %q as float",
			p.curToken.Line, p.curToken.Literal))
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() Expression {
	raw := p.curToken.Literal

	if !strings.Contains(raw, "{") {
		return &StringLiteral{Token: p.curToken, Value: raw}
	}

	parts, err := parseInterpolationParts(raw)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: %s", p.curToken.Line, err))
		return &StringLiteral{Token: p.curToken, Value: raw}
	}

	return &InterpolatedString{Token: p.curToken, Parts: parts}
}

// parseInterpolationParts memecah string mentah jadi bagian teks statis
// dan ekspresi (di dalam {...}), lalu mem-parse tiap ekspresi itu memakai
// lexer+parser terpisah
func parseInterpolationParts(raw string) ([]InterpolationPart, error) {
	parts := []InterpolationPart{}
	var textBuf strings.Builder
	i := 0

	for i < len(raw) {
		ch := raw[i]
		if ch == '{' {
			if textBuf.Len() > 0 {
				parts = append(parts, InterpolationPart{IsExpr: false, Text: textBuf.String()})
				textBuf.Reset()
			}
			end := strings.IndexByte(raw[i:], '}')
			if end == -1 {
				return nil, fmt.Errorf("unterminated interpolation expression in string")
			}
			exprStr := raw[i+1 : i+end]
			l := lexer.New(exprStr)
			subParser := New(l)
			expr := subParser.parseExpression(LOWEST)
			if len(subParser.Errors()) > 0 {
				return nil, fmt.Errorf("invalid expression in string interpolation: %s", exprStr)
			}
			parts = append(parts, InterpolationPart{IsExpr: true, Expr: expr})
			i += end + 1
		} else {
			textBuf.WriteByte(ch)
			i++
		}
	}

	if textBuf.Len() > 0 {
		parts = append(parts, InterpolationPart{IsExpr: false, Text: textBuf.String()})
	}

	return parts, nil
}

func (p *Parser) parseBoolean() Expression {
	return &BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(lexer.TRUE)}
}

func (p *Parser) parseNull() Expression {
	return &NullLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() Expression {
	expr := &PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	expr := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	prec := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(prec)
	return expr
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	expr := p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	return expr
}

func (p *Parser) parseCallExpression(function Expression) Expression {
	expr := &CallExpression{Token: p.curToken, Function: function}
	expr.Arguments = p.parseExpressionList(lexer.RPAREN)
	return expr
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []Expression {
	list := []Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseIndexExpression(left Expression) Expression {
	expr := &IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	expr.Index = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		p.nextToken()
		expr.End = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}
	return expr
}

func (p *Parser) parseDotExpression(left Expression) Expression {
	expr := &DotExpression{Token: p.curToken, Left: left}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	expr.Field = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return expr
}

func (p *Parser) parseListLiteral() Expression {
	lit := &ListLiteral{Token: p.curToken}
	lit.Elements = p.parseExpressionList(lexer.RBRACKET)
	return lit
}

func (p *Parser) parseMapLiteral() Expression {
	lit := &MapLiteral{Token: p.curToken}
	lit.Pairs = make(map[Expression]Expression)

	p.nextToken()
	for p.curTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	for !p.curTokenIs(lexer.RBRACE) {
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)
		lit.Pairs[key] = value

		p.nextToken()
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
			for p.curTokenIs(lexer.NEWLINE) {
				p.nextToken()
			}
		} else if !p.curTokenIs(lexer.RBRACE) {
			return nil
		}
	}

	return lit
}
