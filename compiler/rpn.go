package compiler

import (
	"fmt"
)

type parser struct {
	input, output string
	pos           int
}

func (p *parser) write(s string) {
	p.output += s
}

func (p *parser) expr(nested bool) error {
	if err := p.union(); err != nil {
		return err
	}
	for p.pos < len(p.input) {
		if err := p.union(); err != nil {
			c := p.input[p.pos]
			if c == '*' || c == '+' {
				p.write(string(c))
				p.pos++
				continue
			} else if nested && c == ')' {
				return nil
			}
			return err
		}
		// Thompson adds a concatenation operator at this point, but I have
		// omitted it because I'm not sure what its purpose is:
		// p.write("⋅")
	}
	return nil
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
	if p.input[p.pos] == '(' {
		p.pos++
		if err := p.expr(true); err != nil {
			return err
		}
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

We make use of the following language and translation scheme:
	expr   → expr *  		  { print('*') }
           | expr +           { print('+') }
           | seq

	seq    → union seq
           | union

	union  → basic '|' union  { print('|') }
           | basic

	basic  → ( expr )
           | symbol           { print(symbol) }

	symbol → a-Z | A-Z | 0-9
*/
func RPNConvert(regex string) (string, error) {
	p := &parser{input: regex}
	if err := p.expr(false); err != nil {
		return "", fmt.Errorf("%s: partial result %q", err, p.output)
	}
	return p.output, nil
}
