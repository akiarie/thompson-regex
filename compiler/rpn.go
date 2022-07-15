package compiler

import (
	"fmt"
	"strings"
)

func expr(input []rune, w *strings.Builder) (int, error) {
	n, err := concat(input, w)
	if err != nil {
		return 0, err
	}
	m, err := union(input[n:], w)
	if err != nil {
		return n, err
	}
	return n + m, nil
}

func union(input []rune, w *strings.Builder) (int, error) {
	// ε is permissible
	if end(input) {
		return 0, nil
	}
	if input[0] != '|' {
		return 0, fmt.Errorf("nonempty union must start with '|'")
	}
	n, err := concat(input[1:], w)
	if err != nil {
		return 1, err
	}
	w.WriteRune('|')
	m, err := union(input[1+n:], w)
	if err != nil {
		return 1 + n, err
	}
	return 1 + n + m, nil
}

func concat(input []rune, w *strings.Builder) (int, error) {
	n, err := closed(input, w)
	if err != nil {
		return 0, err
	}
	m, err := rest(input[n:], w)
	if err != nil {
		return n, err
	}
	return n + m, nil
}

func rest(input []rune, w *strings.Builder) (int, error) {
	// ε is permissible
	if end(input) {
		return 0, nil
	}
	if input[0] != '⋅' {
		return 0, nil // allow backtracking
	}
	n, err := closed(input[1:], w)
	if err != nil {
		return 1, err
	}
	w.WriteRune('⋅')
	m, err := rest(input[1+n:], w)
	if err != nil {
		return 1 + n, err
	}
	return 1 + n + m, nil
}

func closed(input []rune, w *strings.Builder) (int, error) {
	n, err := basic(input, w)
	if err != nil {
		return 0, err
	}
	if !end(input[n:]) {
		if c := input[n]; c == '*' || c == '+' {
			w.WriteRune(c)
			return n + 1, nil
		}
	}
	return n, nil
}

// A ntparser is a parser for a nonterminal.
type ntparser func(input []rune, w *strings.Builder) (int, error)

func basic(input []rune, w *strings.Builder) (int, error) {
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
    expr   → concat union
    union  → '|' concat { print('|') } union
           | ε

    concat → closed rest
    rest   → '⋅' closed { print('⋅') } rest
           | ε

    closed → basic *  { print('*') }
           | basic +  { print('+') }
           | basic

    basic  → ( expr )
           | symbol   { print(symbol) }
           | ε

    symbol → a-Z | A-Z | 0-9
*/
func RPNConvert(regex string) (string, error) {
	var buf strings.Builder
	if _, err := expr([]rune(regex), &buf); err != nil {
		return "", fmt.Errorf("%s: partial result %q", err, buf.String())
	}
	return buf.String(), nil
}
