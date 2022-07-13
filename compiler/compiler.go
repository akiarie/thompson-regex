package compiler

import "fmt"

// Sieve validates regular expressions for errors.
func Sieve(regex string) (string, error) {
	/*
		We make use of the following language and translation scheme:
			expr   → expr *
			       | expr +
			       | seq

			seq    → union | union seq

			union  → basic | basic '|' basic

			basic  → ( expr )
			       | symbol

			symbol → a-Z | A-Z | 0-9
	*/
	return "", fmt.Errorf("NOT IMPLEMENTED")
}

// RPNConvert converts a regular expression to a reverse Polish regular
// expression.
func RPNConvert(regex string) (string, error) {
	return "", fmt.Errorf("NOT IMPLEMENTED")
}

// Compile returns a string containing the Go source for a command-line program
// that takes a string as its input and outputs the matches of the given regex.
func Compile(regex string) (string, error) {
	return "", fmt.Errorf("NOT IMPLEMENTED")
}
