package compiler

import (
	"fmt"
	"log"
	"strings"
	"text/template"
)

const (
	funcSeparator string = "\n\n"
)

var matchernames []string = []string{}

func newmatcher() string {
	name := fmt.Sprintf("match%d", len(matchernames))
	matchernames = append(matchernames, name)
	return name
}

type tmpset struct {
	runematcher,
	concatmatcher,
	ormatcher,
	closurematcher *template.Template
}

var templates tmpset

func init() {
	var err error
	templates.runematcher, err = template.New("c").Parse(`runematcher('{{ .Rune }}')`)
	if err != nil {
		log.Fatalln(err)
	}

	templates.ormatcher, err = template.New("cnode").Parse(`ormatcher({{ .MatchA }}, {{ .MatchB }})`)
	if err != nil {
		log.Fatalln(err)
	}

	templates.concatmatcher, err = template.New("concat").Parse(`concatmatcher({{ .MatchA }}, {{ .MatchB }})`)
	if err != nil {
		log.Fatalln(err)
	}

	templates.closurematcher, err = template.New("closure").Parse(`closurematcher({{ .MatchA }}, {{ .Min }})`)
	if err != nil {
		log.Fatalln(err)
	}
}

// A exprmatcher represents the matcher for a given expression.
type exprmatcher interface {
	// genfunc returns the name of a function to match the expression of the
	// exprmatcher.
	genfunc() string
}

// A runematcher matches a single character.
type runematcher rune

func (c runematcher) genfunc() string {
	data := struct{ Rune string }{string(c)}
	var buf strings.Builder
	if err := templates.runematcher.Execute(&buf, &data); err != nil {
		panic(err)
	}
	return buf.String()
}

// An binopmatcher matches based on the provided matchers and binary operation.
type binopmatcher struct {
	a, b exprmatcher
	op   rune
}

func (m *binopmatcher) genfunc() string {
	data := struct{ MatchA, MatchB string }{m.a.genfunc(), m.b.genfunc()}
	var buf strings.Builder
	tmplmap := map[rune]*template.Template{
		'|': templates.ormatcher,
		'⋅': templates.concatmatcher,
	}
	if err := tmplmap[m.op].Execute(&buf, &data); err != nil {
		panic(err)
	}
	return buf.String()
}

// An closermatcher matches based on the provided matcher and unary operation.
type closermatcher struct {
	a  exprmatcher
	op rune
}

func (m *closermatcher) genfunc() string {
	data := struct {
		MatchA string
		Min    int
	}{
		MatchA: m.a.genfunc(),
		Min:    map[rune]int{'*': 0, '+': 1}[m.op],
	}
	var buf strings.Builder
	if err := templates.closurematcher.Execute(&buf, &data); err != nil {
		panic(err)
	}
	return buf.String()
}

// Compile returns a string containing the Go source for a command-line program
// that takes a string as its input and outputs the matches of the given regex.
// The regex must be in reverse Polish notation.
func Compile(regex string) (string, error) {
	stack := []exprmatcher{}
	pop := func() exprmatcher {
		mc := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return mc
	}
	push := func(m exprmatcher) {
		stack = append(stack, m)
	}
	for _, c := range regex {
		switch c {
		case '+', '*':
			push(&closermatcher{pop(), c})
			continue
		case '|', '⋅':
			if len(stack) < 2 {
				return "", fmt.Errorf("cannot use %q with less than 2 elements", c)
			}
			// reversal for the sequential matcher case
			a, b := pop(), pop()
			push(&binopmatcher{b, a, c})
			continue
		default:
			push(runematcher(c))
		}
	}

	if len(stack) != 1 {
		log.Fatalln("stack length not 1")
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
type matcher func([]rune, int) (bool, int)

// runematcher returns a matcher for the given rune.
func runematcher(c rune) matcher {
	return func(input []rune, pos int) (bool, int) {
		return input[pos] == c, 1
	}
}

// ormatcher returns a single matcher for strings matching any of the
// given matchers
func ormatcher(matchers ...matcher) matcher {
	return func(input []rune, pos int) (bool, int) {
		for _, m := range matchers {
			if ok, n := m(input, pos); ok {
				return true, n
			}		
		}
		return false, 0
	}
}

// concatmatcher returns a single matcher for strings matching the
// concatenation of the given matchers
func concatmatcher(matchers ...matcher) matcher {
	return func(input []rune, pos int) (bool, int) {
		p := 0
		for _, m := range matchers {
			if ok, n := m(input, pos+p); ok {
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
	return func(input []rune, pos int) (bool, int) {
		// match min occurrences first
		if ok, n := kmatcher(m, min)(input, pos); ok {
			// then match 1 or more additional 
			if ok, subn := closurematcher(m, 1)(input, pos+n); ok {
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
	input := os.Args[1]
	if ok, _ := {{ . }}([]rune(input), 0); !ok {
		log.Fatalln("unmatching")
	}
	fmt.Println("matching")
}
`)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, stack[0].genfunc()); err != nil {
		panic(err)
	}
	return buf.String(), nil
}
