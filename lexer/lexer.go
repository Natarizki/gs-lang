package lexer

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipSpacesAndComments()

	line, col := l.line, l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: EQ, Literal: "==", Line: line, Column: col}
		} else {
			tok = Token{Type: ASSIGN, Literal: "=", Line: line, Column: col}
		}
	case '+':
		tok = Token{Type: PLUS, Literal: "+", Line: line, Column: col}
	case '-':
		tok = Token{Type: MINUS, Literal: "-", Line: line, Column: col}
	case '*':
		if l.peekChar() == '*' {
			l.readChar()
			tok = Token{Type: POWER, Literal: "**", Line: line, Column: col}
		} else {
			tok = Token{Type: ASTERISK, Literal: "*", Line: line, Column: col}
		}
	case '/':
		tok = Token{Type: SLASH, Literal: "/", Line: line, Column: col}
	case '%':
		tok = Token{Type: PERCENT, Literal: "%", Line: line, Column: col}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: NOT_EQ, Literal: "!=", Line: line, Column: col}
		} else {
			tok = Token{Type: NOT, Literal: "!", Line: line, Column: col}
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: LT_EQ, Literal: "<=", Line: line, Column: col}
		} else {
			tok = Token{Type: LT, Literal: "<", Line: line, Column: col}
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: GT_EQ, Literal: ">=", Line: line, Column: col}
		} else {
			tok = Token{Type: GT, Literal: ">", Line: line, Column: col}
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = Token{Type: AND, Literal: "&&", Line: line, Column: col}
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: line, Column: col}
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = Token{Type: OR, Literal: "||", Line: line, Column: col}
		} else {
			tok = Token{Type: PIPE, Literal: "|", Line: line, Column: col}
		}
	case ',':
		tok = Token{Type: COMMA, Literal: ",", Line: line, Column: col}
	case ':':
		if l.peekChar() == ':' {
			l.readChar()
			tok = Token{Type: DOUBLE_COLON, Literal: "::", Line: line, Column: col}
		} else {
			tok = Token{Type: COLON, Literal: ":", Line: line, Column: col}
		}
	case '.':
		tok = Token{Type: DOT, Literal: ".", Line: line, Column: col}
	case '(':
		tok = Token{Type: LPAREN, Literal: "(", Line: line, Column: col}
	case ')':
		tok = Token{Type: RPAREN, Literal: ")", Line: line, Column: col}
	case '{':
		tok = Token{Type: LBRACE, Literal: "{", Line: line, Column: col}
	case '}':
		tok = Token{Type: RBRACE, Literal: "}", Line: line, Column: col}
	case '[':
		tok = Token{Type: LBRACKET, Literal: "[", Line: line, Column: col}
	case ']':
		tok = Token{Type: RBRACKET, Literal: "]", Line: line, Column: col}
	case '"':
		tok = Token{Type: STRING, Literal: l.readString(), Line: line, Column: col}
	case '\n':
		tok = Token{Type: NEWLINE, Literal: "\\n", Line: line, Column: col}
		l.line++
		l.column = 0
	case 0:
		tok = Token{Type: EOF, Literal: "", Line: line, Column: col}
	default:
		if isLetter(l.ch) {
			lit := l.readIdentifier()
			tok = Token{Type: LookupIdent(lit), Literal: lit, Line: line, Column: col}
			return tok
		} else if isDigit(l.ch) {
			litType, lit := l.readNumber()
			tok = Token{Type: litType, Literal: lit, Line: line, Column: col}
			return tok
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: line, Column: col}
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipSpacesAndComments() {
	for {
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
			l.readChar()
		}

		if l.ch == '/' && l.peekChar() == '/' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			continue
		}

		if l.ch == '#' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			continue
		}

		if l.ch == '/' && l.peekChar() == '*' {
			l.readChar()
			l.readChar()
			for !(l.ch == '*' && l.peekChar() == '/') && l.ch != 0 {
				if l.ch == '\n' {
					l.line++
					l.column = 0
				}
				l.readChar()
			}
			l.readChar()
			l.readChar()
			continue
		}

		break
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() (TokenType, string) {
	start := l.position
	litType := INT
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		litType = FLOAT
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return litType, l.input[start:l.position]
}

func (l *Lexer) readString() string {
	start := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
		if l.ch == '\\' {
			l.readChar()
		}
	}
	return l.input[start:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
