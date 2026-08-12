package parser

import "gs-lang/lexer"

func (p *Parser) parseFuncStatement() Statement {
	fncToken := p.curToken

	// deteksi method: fnc (o orang) sapa { ... }
	if p.peekTokenIs(lexer.LPAREN) {
		return p.parseMethodStatement(fncToken)
	}

	stmt := &FuncStatement{Token: fncToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken()
		stmt.Parameters = p.parseFunctionParameters()
	}

	stmt.Body = p.parseBlock()
	return stmt
}

// parseMethodStatement: fnc (o orang) sapa(params) { ... }
func (p *Parser) parseMethodStatement(fncToken lexer.Token) Statement {
	stmt := &MethodStatement{Token: fncToken}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.ReceiverName = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.ReceiverType = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken()
		stmt.Parameters = p.parseFunctionParameters()
	}

	stmt.Body = p.parseBlock()
	return stmt
}

func (p *Parser) parseFunctionParameters() []*Identifier {
	identifiers := []*Identifier{}

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()
	identifiers = append(identifiers, &Identifier{Token: p.curToken, Value: p.curToken.Literal})

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		identifiers = append(identifiers, &Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseIfStatement() Statement {
	stmt := &IfStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	stmt.Consequence = p.parseBlock()

	for p.peekTokenIs(lexer.BUT) {
		p.nextToken() // ke BUT
		clause := &ButIfClause{Token: p.curToken}
		if !p.expectPeek(lexer.IF) {
			return stmt
		}
		p.nextToken()
		clause.Condition = p.parseExpression(LOWEST)
		clause.Consequence = p.parseBlock()
		stmt.ButIfs = append(stmt.ButIfs, clause)
	}

	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken() // ke ELSE
		stmt.Alternative = p.parseBlock()
	}

	return stmt
}

func (p *Parser) parseMatchStatement() Statement {
	stmt := &MatchStatement{Token: p.curToken}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.LBRACE) {
		return stmt
	}
	p.nextToken()
	p.skipNewlines()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		c := &MatchCase{}
		c.Value = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.RPAREN) {
			return stmt
		}
		c.Body = p.parseBlock()
		stmt.Cases = append(stmt.Cases, c)
		p.nextToken()
		p.skipNewlines()
	}

	return stmt
}

func (p *Parser) parseForStatement() Statement {
	stmt := &ForStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Variable = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.IN) {
		return nil
	}

	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	stmt.Body = p.parseBlock()
	return stmt
}

func (p *Parser) parseErrorStatement() Statement {
	stmt := &ErrorStatement{Token: p.curToken}
	p.nextToken()
	stmt.Message = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseTryStatement() Statement {
	stmt := &TryStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Target = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Call = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseImportStatement() Statement {
	stmt := &ImportStatement{Token: p.curToken}

	if !p.expectPeek(lexer.STRING) {
		return nil
	}
	stmt.Path = p.curToken.Literal

	return stmt
}

func (p *Parser) parseUseStatement() Statement {
	stmt := &UseStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return stmt
}

func (p *Parser) parseStructStatement() Statement {
	nameToken := p.curToken
	name := &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}
	p.nextToken()
	p.skipNewlines()

	// deteksi: field pertama diikuti '=' berarti instance literal, bukan definisi
	isInstance := p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.ASSIGN)

	if isInstance {
		inst := &StructInstanceLiteral{Token: nameToken, Name: name, Fields: make(map[string]Expression)}
		for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
			if p.curTokenIs(lexer.IDENT) {
				fieldName := p.curToken.Literal
				if !p.expectPeek(lexer.ASSIGN) {
					return nil
				}
				p.nextToken()
				inst.Fields[fieldName] = p.parseExpression(LOWEST)
			}
			p.nextToken()
			p.skipNewlines()
		}
		return inst
	}

	stmt := &StructStatement{Token: nameToken, Name: name}
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.IDENT) {
			stmt.Fields = append(stmt.Fields, &Identifier{Token: p.curToken, Value: p.curToken.Literal})
		}
		p.nextToken()
		p.skipNewlines()
	}
	return stmt
}

func (p *Parser) parseSpawnStatement() Statement {
	stmt := &SpawnStatement{Token: p.curToken}
	p.nextToken()
	stmt.Call = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseStartThenStatement() Statement {
	stmt := &StartThenStatement{Token: p.curToken, IsStart: p.curTokenIs(lexer.START)}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	if !p.expectPeek(lexer.DOUBLE_COLON) {
		return nil
	}
	if !p.expectPeek(lexer.INT) {
		return nil
	}
	stmt.TaskIDs = append(stmt.TaskIDs, p.curToken.Literal)

	for p.peekTokenIs(lexer.PIPE) {
		p.nextToken()
		if !p.expectPeek(lexer.IDENT) {
			return stmt
		}
		if !p.expectPeek(lexer.DOUBLE_COLON) {
			return stmt
		}
		if !p.expectPeek(lexer.INT) {
			return stmt
		}
		stmt.TaskIDs = append(stmt.TaskIDs, p.curToken.Literal)
	}

	return stmt
}

var taskCounter = 0

func (p *Parser) parseTaskOrBlockAsStatement() Statement {
	taskCounter++
	tb := &TaskBlock{Token: p.curToken, ID: formatTaskID(taskCounter)}
	tb.Body = p.parseBlock()
	return tb
}

func formatTaskID(n int) string {
	if n < 10 {
		return "00" + itoa(n)
	} else if n < 100 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// parseStartIfStatement: startif kondisi \n start = id::00X \n elstart \n start = id::00Y
func (p *Parser) parseStartIfStatement() Statement {
	stmt := &StartIfStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	// setelah kondisi, harapkan block berisi "start = id::00X" (mode brace atau indentasi)
	consBlock := p.parseBlock()
	stmt.Consequence = extractTaskIDsFromBlock(consBlock)

	if p.peekTokenIs(lexer.ELSTART) {
		p.nextToken() // ke ELSTART
		altBlock := p.parseBlock()
		stmt.Alternative = extractTaskIDsFromBlock(altBlock)
	}

	return stmt
}

// extractTaskIDsFromBlock mengambil semua task ID dari statement "start = id::00X"
// di dalam sebuah block (dipakai untuk isi startif/elstart)
func extractTaskIDsFromBlock(block *BlockStatement) []string {
	ids := []string{}
	for _, s := range block.Statements {
		if sts, ok := s.(*StartThenStatement); ok {
			ids = append(ids, sts.TaskIDs...)
		}
	}
	return ids
}
