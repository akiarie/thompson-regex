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

type CodeGenerators struct {
	Rune       func(rune) (string, error)
	Or, Concat func(string, string) (string, error)
	Closure    func(string, int) (string, error)
}

func codeGens() (*CodeGenerators, error) {
	tmplRune, err := template.New("c").Parse(`runematcher('{{ . }}')`)
	if err != nil {
		return nil, err
	}

	tmplOr, err := template.New("cnode").Parse(`ormatcher(
	{{ .MatchA }},
	{{ .MatchB }},
)`)
	if err != nil {
		return nil, err
	}

	tmplConcat, err := template.New("concat").Parse(`concatmatcher(
	{{ .MatchA }},
	{{ .MatchB }},
)`)
	if err != nil {
		return nil, err
	}

	tmplClosure, err := template.New("closure").Parse(`closurematcher(
	{{ .MatchA }},
	{{ .Min }},
)`)
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
				MatchA, MatchB string
			}{a, b}); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
		Or: func(a, b string) (string, error) {
			var buf strings.Builder
			if err := tmplOr.Execute(&buf, struct {
				MatchA, MatchB string
			}{a, b}); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
		Closure: func(a string, min int) (string, error) {
			var buf strings.Builder
			if err := tmplClosure.Execute(&buf, struct {
				MatchA string
				Min    int
			}{a, min}); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
	}, nil
}

func Go(root MatcherGenerator) (string, error) {
	gens, err := codeGens()
	if err != nil {
		return "", err
	}

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

// runematcher returns a matcher for the given rune.
func runematcher(c rune) matcher {
	return func(input []rune) (bool, int) {
		if len(input) > 0 {
			return input[0] == c, 1
		}
		return false, 0
	}
}

// ormatcher returns a single matcher for strings matching any of the
// given matchers
func ormatcher(matchers ...matcher) matcher {
	return func(input []rune) (bool, int) {
		for _, m := range matchers {
			if ok, n := m(input); ok {
				return true, n
			}		
		}
		return false, 0
	}
}

// concatmatcher returns a single matcher for strings matching the
// concatenation of the given matchers
func concatmatcher(matchers ...matcher) matcher {
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
	return concatmatcher(matchers...)
}

// closurematcher returns a single matcher for strings matching the closure of
// the given matcher
func closurematcher(m matcher, min int) matcher {
	return func(input []rune) (bool, int) {
		// match min occurrences first
		if ok, n := kmatcher(m, min)(input); ok {
			// then match 1 or more additional 
			if ok, subn := closurematcher(m, 1)(input[n:]); ok {
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

	matcher := {{ . }}

	matches := []string{}
	for i := 0; i < len(input); {
		if ok, n := matcher(input[i:]); ok {
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
