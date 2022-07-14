package compiler

import (
	"fmt"
	"strings"
)

const concatChar byte = '^'

func end(input string) bool {
	return len(input) == 0 || input[0] == ')'
}

func exprSieve(input string, w *strings.Builder) (int, error) {
	n, err := concatSieve(input, w)
	if err != nil {
		return 0, err
	}
	if !end(input[n:]) {
		if input[n] == '|' {
			w.WriteByte('|')
			m, err := exprSieve(input[n+1:], w)
			if err != nil {
				return n + 1, err
			}
			return n + 1 + m, nil
		}
	}
	return n, nil
}

func concatSieve(input string, w *strings.Builder) (int, error) {
	n, err := closedSieve(input, w)
	if err != nil {
		return 0, err
	}
	if !end(input[n:]) {
		var buf strings.Builder
		if m, err := concatSieve(input[n:], &buf); err == nil {
			w.WriteByte(concatChar)
			w.WriteString(buf.String())
			return n + m, nil
		}
	}
	return n, nil
}

func closedSieve(input string, w *strings.Builder) (int, error) {
	n, err := basicSieve(input, w)
	if err != nil {
		return 0, err
	}
	if !end(input[n:]) {
		if c := input[n]; c == '*' || c == '+' {
			w.WriteByte(c)
			return n + 1, nil
		}
	}
	return n, nil
}

func basicSieve(input string, w *strings.Builder) (int, error) {
	// ε is permissible
	if end(input) {
		return 0, nil
	}
	if input[0] == '(' {
		w.WriteByte('(')
		n, err := exprSieve(input[1:], w)
		if err != nil {
			return 1, err
		}
		if input[n+1] != ')' {
			return 0, fmt.Errorf("bracket not closed")
		}
		w.WriteByte(')')
		return n + 2, nil
	}
	return 1, symbol(input[0], w)
}

func symbol(c byte, w *strings.Builder) error {
	if ('0' <= c && c <= '9') || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') {
		w.WriteByte(c)
		return nil
	}
	return fmt.Errorf("%q is not an allowed symbol", c)
}

/*
Sieve validates a regular expression and inserts concatenation breaks.

We make use of the following language-and-translation scheme:
    expr   → concat '|' { print('|') } expr
           | concat

    concat → closed { print('⋅') } concat
           | closed

	closed → basic * { print('*') }
           | basic + { print('+') }
		   | basic

    basic  → ( expr )
           | symbol  { print(symbol) }
           | ε

    symbol → a-Z | A-Z | 0-9
*/
func Sieve(regex string) (string, error) {
	var buf strings.Builder
	if _, err := exprSieve(regex, &buf); err != nil {
		return "", fmt.Errorf("%s: partial result %q", err, buf.String())
	}
	return buf.String(), nil
}
