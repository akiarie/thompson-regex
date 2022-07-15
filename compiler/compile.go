package compiler

import (
	"fmt"
	"log"

	"thompson-regex/assembler"
)

// A RuneMatcher matches a single character.
type RuneMatcher rune

func (c RuneMatcher) Generate(codegens *assembler.CodeGenerators) (string, error) {
	return codegens.Rune(rune(c))
}

// An BinOpMatcher matches based on the provided matchers and binary operation.
type BinOpMatcher struct {
	a, b assembler.MatcherGenerator
	op   rune
}

func (m *BinOpMatcher) Generate(codegens *assembler.CodeGenerators) (string, error) {
	amc, err := m.a.Generate(codegens)
	if err != nil {
		return "", err
	}
	bmc, err := m.b.Generate(codegens)
	if err != nil {
		return "", err
	}
	return map[rune]func(string, string) (string, error){
		'|': codegens.Or,
		'⋅': codegens.Concat,
	}[m.op](amc, bmc)
}

// A ClosureMatcher matches based on the provided matcher and unary operation.
type ClosureMatcher struct {
	a  assembler.MatcherGenerator
	op rune
}

func (m *ClosureMatcher) Generate(codegens *assembler.CodeGenerators) (string, error) {
	amc, err := m.a.Generate(codegens)
	if err != nil {
		return "", err
	}
	return codegens.Closure(
		amc,
		map[rune]int{'*': 0, '+': 1}[m.op],
	)
}

// Compile returns a string containing the Go source for a command-line program
// that takes a string as its input and outputs the matches of the given regex.
// The regex must be in reverse Polish notation.
func Compile(regex string) (assembler.MatcherGenerator, error) {
	stack := []assembler.MatcherGenerator{}
	pop := func() assembler.MatcherGenerator {
		mc := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return mc
	}
	push := func(m assembler.MatcherGenerator) {
		stack = append(stack, m)
	}
	for _, c := range regex {
		switch c {
		case '+', '*':
			push(&ClosureMatcher{pop(), c})
			continue
		case '|', '⋅':
			if len(stack) < 2 {
				return nil, fmt.Errorf("cannot use %q with less than 2 elements", c)
			}
			// reversal for the sequential matcher case
			a, b := pop(), pop()
			push(&BinOpMatcher{b, a, c})
			continue
		default:
			push(RuneMatcher(c))
		}
	}
	if len(stack) != 1 {
		log.Fatalln("stack length not 1")
	}
	return stack[0], nil
}
