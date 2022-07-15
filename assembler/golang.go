package assembler

import (
	"strings"
	"text/template"
)

// A MatcherGenerator represents the matcher for a given expression.
type MatcherGenerator interface {
	// Generate returns the name of a function to match the expression of the
	// MatcherGenerator.
	Generate(tmpl *CodeGenerators) (string, error)
}

// CodeGenerators houses functions which can be called by the MatcherGenerators
// in order to represent matching in whatever format has been configured in the
// assembler.
type CodeGenerators struct {
	Rune       func(rune) (string, error)
	Or, Concat func(string, string) (string, error)
	Closure    func(string, int) (string, error)
}

// codeGens returns CodeGenerators which can be provided to each
// MatcherGenerator enabling it to represent matching in the format configured
// in the assembler.
func codeGens(char, or, concat, closure string) (*CodeGenerators, error) {
	tmplRune, err := template.New("c").Parse(char)
	if err != nil {
		return nil, err
	}

	tmplOr, err := template.New("cnode").Parse(or)
	if err != nil {
		return nil, err
	}

	tmplConcat, err := template.New("concat").Parse(concat)
	if err != nil {
		return nil, err
	}

	tmplClosure, err := template.New("closure").Parse(closure)
	if err != nil {
		return nil, err
	}
	return &CodeGenerators{
		Rune: func(c rune) (string, error) {
			var buf strings.Builder
			if err := tmplRune.Execute(&buf, string(c)); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
		Concat: func(a, b string) (string, error) {
			var buf strings.Builder
			if err := tmplConcat.Execute(&buf, struct {
				MatcherFuncA, MatcherFuncB string
			}{a, b}); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
		Or: func(a, b string) (string, error) {
			var buf strings.Builder
			if err := tmplOr.Execute(&buf, struct {
				MatcherFuncA, MatcherFuncB string
			}{a, b}); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
		Closure: func(a string, min int) (string, error) {
			var buf strings.Builder
			if err := tmplClosure.Execute(&buf, struct {
				MatcherFuncA string
				Min          int
			}{a, min}); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
	}, nil
}

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
