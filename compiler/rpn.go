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
		var buf strings.Builder
		if m, err := concat(input[n:], &buf); err == nil {
			w.WriteString("⋅")
			w.WriteString(buf.String())
			return n + m, nil
		}
	}
	return n, nil
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
