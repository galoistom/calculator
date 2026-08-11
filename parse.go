package main

import "fmt"

func (l *Lexer) getNumber() Token {
	res := Token{
		Type: NUMBER,
	}
	var ch byte
	for l.position < len(l.input) {
		ch = l.input[l.position]
		if isDigit(ch) || ch == '.' || ch == '/' {
			res.Value += string(ch)
			l.position++
		} else {
			break
		}
	}
	return res
}

func (l *Lexer) getOp() Token {
	res := Token{
		Type: IDENT,
	}
	var ch byte
	for l.position < len(l.input) {
		ch = l.input[l.position]
		if isSymbol(ch) {
			res.Value += string(ch)
			l.position++
		} else {
			break
		}
	}
	return res
}
func isQuote(ch byte) bool {
	return ch == '\''
}
func isOperator(ch byte) bool {
	switch ch {
	case '+', '-', '*', '/':
		return true
	}
	return false
}

func isSymbol(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || isOperator(ch) || isQuote(ch)
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch == '_') || (ch == '!') ||
		(ch == '=') || (ch == '<') ||
		(ch == '>') || (ch == '?')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func (l *Lexer) skipWhitespace() {
	for l.position < len(l.input) {
		if l.input[l.position] == ' ' ||
			l.input[l.position] == '\n' ||
			l.input[l.position] == '\t' {
			l.position++
		} else {
			break
		}
	}
}

func (l *Lexer) getString() Token {
	res := Token{Type: STRING}
	l.position++
	for l.position < len(l.input) {
		ch := l.input[l.position]
		switch ch {
		case '"':
			l.position++
			return res
		case '\\':
			l.position++
			if l.position >= len(l.input) {
				panic("unterminated string")
			}
			res.Value += string(l.input[l.position])
			l.position++
		default:
			res.Value += string(ch)
			l.position++
		}
	}
	panic("unterminated string")
}

func (l *Lexer) nextToken() Token {
	l.skipWhitespace()
	if l.position >= len(l.input) {
		return Token{Type: EOF}
	}
	ch := l.input[l.position]
	switch ch {
	case '(':
		l.position++
		return Token{Type: LPAREN, Value: "("}
	case ')':
		l.position++
		return Token{Type: RPAREN, Value: ")"}
	case '\'':
		l.position++
		return Token{Type: QUOTE, Value: "'"}
	case '"':
		return l.getString()
	case '`':
		l.position++
		return Token{Type: QUASIQUOTE, Value: "`"}
	case ',':
		l.position++
		if l.position < len(l.input) && l.input[l.position] == '@' {
			l.position++
			return Token{Type: SPLICE, Value: ",@"}
		}
		return Token{Type: UNQUOTE, Value: ","}
	}
	if isOperator(ch) {
		l.position++
		return Token{Type: IDENT, Value: string(ch)}
	} else if isDigit(ch) {
		return l.getNumber()
	} else if isLetter(ch) {
		return l.getOp()
	}
	panic("invalid symbol")
}

func (l *Lexer) GetToken() []Token {
	l.position = 0
	res := []Token{l.nextToken()}
	for res[len(res)-1].Type != EOF {
		res = append(res, l.nextToken())
	}
	return res
}

func (p *Parser) Parse() (Expr, error) {
	if p.position >= len(p.tokens) {
		return nil, fmt.Errorf("no tokens")
	}
	switch x := p.tokens[p.position]; x.Type {
	case NUMBER:
		p.position++
		res, err := stringNumber{Value: x.Value}.getValue()
		if err != nil {
			return nil, err
		}
		return res, nil
	case IDENT:
		p.position++
		return Symbol{content: x.Value}, nil
	case QUOTE:
		p.position++
		content, err := p.Parse()
		if err != nil {
			return nil, err
		}
		return List{args: []Expr{Symbol{content: "quote"}, content}}, nil
	case QUASIQUOTE:
		p.position++
		content, err := p.Parse()
		if err != nil {
			return nil, err
		}
		return List{args: []Expr{Symbol{content: "quasiquote"}, content}}, nil
	case UNQUOTE:
		p.position++
		content, err := p.Parse()
		if err != nil {
			return nil, err
		}
		return List{args: []Expr{Symbol{content: "unquote"}, content}}, nil
	case SPLICE:
		p.position++
		content, err := p.Parse()
		if err != nil {
			return nil, err
		}
		return List{args: []Expr{Symbol{content: "unquote-splicing"}, content}}, nil
	case STRING:
		p.position++
		return String{content: x.Value}, nil
	case LPAREN:
		p.position++
		res := List{args: []Expr{}}
		for {
			y := p.tokens[p.position]
			if y.Type == RPAREN {
				break
			}
			if y.Type == EOF {
				return nil, fmt.Errorf("invalid sentences, ( is not closed")
			}
			t, err := p.Parse()
			if err != nil {
				return nil, err
			}
			res.args = append(res.args, t)
		}
		p.position++
		return res, nil
	case EOF:
		p.position++
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid sentences")
	}
}

func (p *Parser) ParseSequence() ([]Expr, error) {
	p.position = 0
	l := len(p.tokens)
	res := []Expr{}
	for p.position < l {
		t, err := p.Parse()
		if err != nil {
			return nil, err
		}
		if t != nil {
			res = append(res, t)
		}
	}
	return res, nil
}
