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

// A matcher represents the compiled code for matching a particular expression.
type matcher interface {
	match(input []rune) (bool, int)
}

// A char is a matcher for the given rune.
type char rune

func (c char) match(input []rune) (bool, int) {
	if len(input) > 0 {
		return input[0] == rune(c), 1
	}
	return false, 0
}

// an or is a matcher for strings matching any of the given matchers
type or []matcher

func (matchers or) match(input []rune) (bool, int) {
	for _, m := range matchers {
		if ok, n := m.match(input); ok {
			return true, n
		}		
	}
	return false, 0
}

// concat is a matcher for strings matching the concatenation of the
// given matchers
type concat []matcher

func (matchers concat) match(input []rune) (bool, int) {
	p := 0
	for _, m := range matchers {
		if ok, n := m.match(input[p:]); ok {
			p += n
			continue
		}
		return false, p
	}
	return true, p
}

// kmatcher returns a single matcher for strings matching exactly k occurrences
// of the given matcher
func kmatcher(m matcher, k int) matcher {
	matchers := make([]matcher, k)
	for i := 0; i < k; i++ {
		matchers[i] = m
	}
	return concat(matchers)
}

// closure is single matcher for strings matching the closure of the
// given matcher
type closure struct {
	m matcher
	min int
}

func (cl closure) match(input []rune) (bool, int) {
	// match min occurrences first
	if ok, n := kmatcher(cl.m, cl.min).match(input); ok {
		// then match 1 or more additional 
		nextcl := closure{cl.m, 1}
		if ok, subn := nextcl.match(input[n:]); ok {
			n += subn
		}
		return true, n
	}
	return false, 0
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("must supply input string")
	}
	input := []rune(os.Args[1])

	expmatcher := {{ . }}

	matches := []string{}
	for i := 0; i < len(input); {
		if ok, n := expmatcher.match(input[i:]); ok {
			matches = append(matches, string(input[i:i+n]))
			i += n
			continue
		}
		i++
	}

	fmt.Printf("%q\n", matches)
}
`)
	if err != nil {
		return "", err
	}

	gens, err := codeGens(
		"char('{{ . }}')",
		`or{
	{{ .MatcherFuncA }},
	{{ .MatcherFuncB }},
}`,
		`concat{
	{{ .MatcherFuncA }},
	{{ .MatcherFuncB }},
}`,
		`closure{
	{{ .MatcherFuncA }},
	{{ .Min }},
}`,
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
