package compiler

import (
	"fmt"
)

type parser struct {
	input, output string
	pos           int
	exits         int
}

const exitchar = ')'

func (p *parser) end() bool {
	if p.pos >= len(p.input) {
		return true
	}
	if p.exits > 0 && p.input[p.pos] == exitchar {
		return true
	}
	return false
}

func (p *parser) write(s string) {
	p.output += s
}

func (p *parser) expr() error {
	if err := p.concat(); err != nil {
		c := p.input[p.pos]
		if c == '*' || c == '+' {
			p.write(string(c))
			p.pos++
			return p.expr()
		}
		return err
	}
	if p.end() {
		p.write("⋅")
		return nil
	}
	// concat only stops if at end of expr or closure thus, if neither of these
	// has occurred, there must be some error
	return fmt.Errorf("premature end of expr on %q, exits: %d", p.input[p.pos], p.exits)
}

func (p *parser) concat() error {
	if err := p.union(); err != nil {
		return err
	}
	if p.end() {
		return nil
	}
	return p.concat()
}

func (p *parser) union() error {
	if err := p.basic(); err != nil {
		return err
	}
	if p.pos < len(p.input) && p.input[p.pos] == '|' {
		p.pos++
		if err := p.union(); err != nil {
			return err
		}
		p.write("|")
	}
	return nil
}

func (p *parser) basic() error {
	// ε is permissible
	if p.end() {
		return nil
	}
	if p.input[p.pos] == '(' {
		p.pos++
		p.exits++
		if err := p.expr(); err != nil {
			return err
		}
		p.exits--
		if p.input[p.pos] != ')' {
			return fmt.Errorf("bracket not closed")
		}
		p.pos++
		return nil
	}
	return p.symbol()
}

func (p *parser) symbol() error {
	c := p.input[p.pos]
	if ('0' <= c && c <= '9') || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') {
		p.write(string(c))
		p.pos++
		return nil
	}
	return fmt.Errorf("%q is not an allowed symbol", c)
}

/*
RPNConvert validates a regular expression and (if valid) converts it to a
reverse Polish regular expression.

We make use of the following language-and-translation scheme:
    expr   → concat * expr    { print('*.') }
           | concat + expr    { print('+.') }
           | concat           { print('.') }

    concat → union concat
           | union

    union  → basic '|' union  { print('|') }
           | basic

    basic  → ( expr )
           | symbol           { print(symbol) }
           | ε

    symbol → a-Z | A-Z | 0-9
*/
func RPNConvert(regex string) (string, error) {
	p := &parser{input: regex}
	if err := p.expr(); err != nil {
		return "", fmt.Errorf("%s: partial result %q", err, p.output)
	}
	return p.output, nil
}
