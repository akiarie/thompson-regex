package assembler

import (
	"strings"
	"text/template"
)

func Go(root MatcherGenerator) (string, error) {
	tmpl, err := template.New("program").Parse(`package main

import (
	"fmt"
	"log"
	"os"
)

// A matcher is a function representing a particular expression, returning true
// if the given rune slice matches the expression with the length of the match,
// or false otherwise with an undefined integer.
type matcher func([]rune) (bool, int)

// char returns a matcher for the given rune.
func char(c rune) matcher {
	return func(input []rune) (bool, int) {
		if len(input) > 0 {
			return input[0] == c, 1
		}
		return false, 0
	}
}

// or returns a single matcher for strings matching any of the
// given matchers
func or(matchers ...matcher) matcher {
	return func(input []rune) (bool, int) {
		for _, m := range matchers {
			if ok, n := m(input); ok {
				return true, n
			}		
		}
		return false, 0
	}
}

// concat returns a single matcher for strings matching the
// concatenation of the given matchers
func concat(matchers ...matcher) matcher {
	return func(input []rune) (bool, int) {
		p := 0
		for _, m := range matchers {
			if ok, n := m(input[p:]); ok {
				p += n
				continue
			}
			return false, p
		}
		return true, p
	}
}

// kmatcher returns a single matcher for strings matching exactly k occurrences
// of the given matcher
func kmatcher(m matcher, k int) matcher {
	matchers := make([]matcher, k)
	for i := 0; i < k; i++ {
		matchers[i] = m
	}
	return concat(matchers...)
}

// closure returns a single matcher for strings matching the closure of
// the given matcher
func closure(m matcher, min int) matcher {
	return func(input []rune) (bool, int) {
		// match min occurrences first
		if ok, n := kmatcher(m, min)(input); ok {
			// then match 1 or more additional 
			if ok, subn := closure(m, 1)(input[n:]); ok {
				n += subn
			}
			return true, n
		}
		return false, 0
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("must supply input string")
	}
	input := []rune(os.Args[1])

	match := {{ . }}

	matches := []string{}
	for i := 0; i < len(input); {
		if ok, n := match(input[i:]); ok {
			matches = append(matches, string(input[i:i+n]))
			i += n
			continue
		}
		i++
	}

	fmt.Printf("matches: %q\n", matches)
}
`)
	if err != nil {
		return "", err
	}

	gens, err := codeGens(
		"char('{{ . }}')",
		`or(
	{{ .MatcherFuncA }},
	{{ .MatcherFuncB }},
)`,
		`concat(
	{{ .MatcherFuncA }},
	{{ .MatcherFuncB }},
)`,
		`closure(
	{{ .MatcherFuncA }},
	{{ .Min }},
)`,
	)
	if err != nil {
		return "", err
	}

	data, err := root.Generate(gens)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
