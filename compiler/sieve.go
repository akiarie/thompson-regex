package compiler

import (
	"fmt"
	"strings"
)

func end(input []rune) bool {
	return len(input) == 0 || input[0] == ')'
}

func exprSieve(input []rune, w *strings.Builder) (int, error) {
	n, err := concatSieve(input, w)
	if err != nil {
		return 0, err
	}
	m, err := unionSieve(input[n:], w)
	if err != nil {
		return n, err
	}
	return n + m, nil
}

func unionSieve(input []rune, w *strings.Builder) (int, error) {
	// ε is permissible
	if end(input) {
		return 0, nil
	}
	if input[0] != '|' {
		return 0, fmt.Errorf("nonempty union must start with '|'")
	}
	w.WriteRune('|')
	n, err := concatSieve(input[1:], w)
	if err != nil {
		return 1, err
	}
	m, err := unionSieve(input[1+n:], w)
	if err != nil {
		return 1 + n, err
	}
	return 1 + n + m, nil
}

func concatSieve(input []rune, w *strings.Builder) (int, error) {
	n, err := closedSieve(input, w)
	if err != nil {
		return 0, err
	}
	m, err := restSieve(input[n:], w)
	if err != nil {
		return n, err
	}
	return n + m, nil
}

func restSieve(input []rune, w *strings.Builder) (int, error) {
	// ε is permissible
	if end(input) {
		return 0, nil
	}
	var buf strings.Builder
	n, err := closedSieve(input, &buf)
	if err != nil {
		return 0, nil // allows backtracking
	}
	w.WriteRune('⋅')
	w.WriteString(buf.String())
	m, err := restSieve(input[n:], w)
	if err != nil {
		return n, err
	}
	return n + m, nil
}

func closedSieve(input []rune, w *strings.Builder) (int, error) {
	n, err := basicSieve(input, w)
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

func basicSieve(input []rune, w *strings.Builder) (int, error) {
	// ε is permissible
	if end(input) {
		return 0, nil
	}
	if input[0] == '(' {
		w.WriteRune('(')
		n, err := exprSieve(input[1:], w)
		if err != nil {
			return 1, err
		}
		if input[n+1] != ')' {
			return 0, fmt.Errorf("bracket not closed")
		}
		w.WriteRune(')')
		return n + 2, nil
	}
	return 1, symbol(input[0], w)
}

func symbol(c rune, w *strings.Builder) error {
	if ('0' <= c && c <= '9') || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') {
		w.WriteRune(c)
		return nil
	}
	return fmt.Errorf("%q is not an allowed symbol", c)
}

/*
Sieve validates a regular expression and inserts concatenation breaks.

We make use of the following language-and-translation scheme:
    expr   → concat union
    union  → '|' { print('|') } concat union
           | ε

	concat → closed rest
    rest   → { print('⋅') } closed rest
		   | ε

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
	if _, err := exprSieve([]rune(regex), &buf); err != nil {
		return "", fmt.Errorf("%s: partial result %q", err, buf.String())
	}
	return buf.String(), nil
}
