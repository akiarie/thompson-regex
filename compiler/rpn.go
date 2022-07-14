package compiler

import (
	"fmt"
	"strings"
)

func expr(input string, w *strings.Builder) (int, error) {
	n, err := concat(input, w)
	if err != nil {
		return 0, err
	}
	if !end(input[n:]) {
		if input[n] == '|' {
			m, err := expr(input[n+1:], w)
			if err != nil {
				return n + 1, err
			}
			w.WriteByte('|')
			return n + 1 + m, nil
		}
	}
	return n, nil
}

func concat(input string, w *strings.Builder) (int, error) {
	n, err := closed(input, w)
	if err != nil {
		return 0, err
	}
	if !end(input[n:]) {
		if input[n] == concatChar {
			var buf strings.Builder
			m, err := concat(input[n+1:], &buf)
			if err != nil {
				return n + 1, err
			}
			w.WriteString(buf.String())
			w.WriteByte(concatChar)
			return n + 1 + m, nil
		}
	}
	return n, nil
}

func closed(input string, w *strings.Builder) (int, error) {
	n, err := basic(input, w)
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

func basic(input string, w *strings.Builder) (int, error) {
	// ε is permissible
	if end(input) {
		return 0, nil
	}
	if input[0] == '(' {
		n, err := expr(input[1:], w)
		if err != nil {
			return 1, err
		}
		if input[n+1] != ')' {
			return 0, fmt.Errorf("bracket not closed")
		}
		return n + 2, nil
	}
	return 1, symbol(input[0], w)
}

/*
RPNConvert validates a regular expression and (if valid) converts it to a
reverse Polish regular expression.

We make use of the following language-and-translation scheme:
    expr   → concat '|' expr   { print('|') }
           | concat

    concat → closed '⋅' concat { print('⋅') }
           | closed

	closed → basic *           { print('*') }
           | basic +           { print('+') }
		   | basic

    basic  → ( expr )
           | symbol            { print(symbol) }
           | ε

    symbol → a-Z | A-Z | 0-9
*/
func RPNConvert(regex string) (string, error) {
	var buf strings.Builder
	if _, err := expr(regex, &buf); err != nil {
		return "", fmt.Errorf("%s: partial result %q", err, buf.String())
	}
	return buf.String(), nil
}
