package parser

import (
	"errors"
	"fmt"
	"strings"
	"ti/base"
	"ti/lexer"
)

func (p *Parser) getToken() {
	if p.ungetFlg {
		p.ungetFlg = false
		p.Lexer.IsSpace = p.Lexer.IsSpacePrev

		return
	}

	if p.Lexer.Advance() {
		p.token = p.Lexer.Token()

		switch p.token {
		case '\n':
			p.Row++

		default:
			p.ErrorRow = p.Row
		}

		return
	}

	p.token = base.EOS
}

func (p *Parser) Unget() {
	p.ungetFlg = true
}

func (p *Parser) ReadWithCheck(target string) (*base.T, bool, error) {
	t, err := p.Read()
	if err != nil {
		return nil, false, err
	}

	if t.IsTargetIdentifier(target) {
		return t, true, nil
	}

	return t, false, nil
}

func (p *Parser) SkipNewline() error {
	for {
		t, err := p.Read()
		if err != nil {
			return err
		}

		if t == nil {
			return nil
		}

		if t.IsNewLineIdentifier() {
			continue
		}

		p.Unget()

		return nil
	}
}

func (p *Parser) ReadAhead() (*base.T, error) {
	t, err := p.Read()
	if err != nil {
		return nil, err
	}

	p.Unget()

	return t, nil
}

func (p *Parser) Read() (*base.T, error) {
	p.LastT = p.CurrentT

	p.getToken()
	var t *base.T

	switch p.token {
	case base.INT:
		if v, ok := p.Lexer.Value().(int64); ok {
			t = base.MakeInt(v)
		} else {
			t = base.MakeInt(0)
		}

	case base.FLOAT:
		if v, ok := p.Lexer.Value().(float64); ok {
			t = base.MakeFloat(v)
		} else {
			t = base.MakeFloat(0)
		}

	case base.STRING:
		stringValue, _ := p.Lexer.Value().(string)
		t = base.MakeString(stringValue)

		if p.BeforeString != stringValue {
			// Count newlines in string and increment p.Row accordingly
			newlineCount := strings.Count(stringValue, "\n")
			p.Row += newlineCount
		}

		p.BeforeString = stringValue

	case base.NIL:
		t = base.MakeNil()

	case
		';',
		'^',
		'+', '-', '/', '*',
		'>', '<',
		'(', ')',
		',', '\n',
		'{', '}',
		'[', ']',
		'!',
		'|',
		'=',
		'.':

		t = base.MakeIdentifier(string(p.token))

	case base.UNKNOWN:
		id, ok := p.Lexer.Value().(lexer.Identifier)
		if !ok {
			break
		}

		t = base.MakeIdentifier(id.GetName())

		if t.IsBoolIdentifier() {
			t = base.MakeBool()
			break
		}

		if t.IsClassIdentifier() {
			t = base.MakeClass(id.GetName())
			break
		}

		if t.IsConstIdentifier() {
			t = base.MakeConst(id.GetName())
			break
		}

		if t.IsSymbolIdentifier() {
			t = base.MakeSymbol(id.GetName())
			break
		}

	case base.EOS:
		return nil, nil

	default:
		return nil, errors.New("read error")
	}

	t.IsBeforeSpace = p.Lexer.IsSpace
	p.Lexer.IsSpacePrev = p.Lexer.IsSpace

	p.Lexer.IsSpace = false

	if p.Debug {
		fmt.Println(t)
	}

	p.CurrentT = *t

	return t, nil
}

func (p *Parser) SkipToTargetToken(target string) error {
	for {
		nextT, err := p.Read()
		if err != nil {
			return err
		}

		if nextT.IsTargetIdentifier(target) {
			break
		}
	}

	return nil
}

func (p *Parser) Skip() {
	p.getToken()
}

func (p *Parser) ReadTwice() (*base.T, error) {
	_, err := p.Read()
	if err != nil {
		return nil, err
	}

	return p.Read()
}
