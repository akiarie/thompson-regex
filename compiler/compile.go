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
	nnode,
	cnode,
	seqnode,
	closure,
	posClosure *template.Template
}

var templates tmpset

func init() {
	var err error
	templates.nnode, err = template.New("nnode").Parse(`var {{.Func}} matcher = cmatcher('{{.Rune}}')`)
	if err != nil {
		log.Fatalln(err)
	}

	templates.cnode, err = template.New("cnode").Parse(`func {{.Func}}(input []rune, pos int) (bool, int) {
	if ok, n := {{.MatchA}}(input, pos); ok {
		return true, n
	} else if ok, n := {{.MatchB}}(input, pos); ok {
		return true, n
	}
	return false, 0
}`)
	if err != nil {
		log.Fatalln(err)
	}

	templates.seqnode, err = template.New("seqnode").Parse(`func {{.Func}}(input []rune, pos int) (bool, int) {
	if ok, n := {{.MatchA}}(input, pos); ok {
		if ok, m := {{.MatchB}}(input[n:], pos); ok {
			return true, n + m
		}
	}
	return false, 0
}`)
	if err != nil {
		log.Fatalln(err)
	}

	templates.closure, err = template.New("closure").Parse(`func {{.Func}}(input []rune, pos int) (bool, int) {
	if ok, n := {{.MatchA}}(input, pos); ok {
		if pos+n >= len(input) {
			return false, n
		}
		if ok, subn := {{.Func}}(input, pos+n); ok {
			return true, n + subn
		}
	} 
	return true, 0
}`)
	if err != nil {
		log.Fatalln(err)
	}

	templates.posClosure, err = template.New("posclosure").Parse(`func {{.Func}}(input string, pos int) (bool, int) {
	if ok, n := {{.MatchA}}(input, pos); ok {
		if pos+n >= len(input) {
			return false, n
		}
		if ok, subn := {{.Func}}(input, pos+n); ok {
			return true, n + subn
		}
		return true, n
	} 
	return false, 0
}`)
	if err != nil {
		log.Fatalln(err)
	}
}

// A exprmatcher represents the matcher for a given expression.
type exprmatcher interface {
	// genfunc returns the code for a function to match the expression, using the
	// provided name
	genfunc(name string) string
}

// A cmatcher matches a single character.
type cmatcher rune

func (c cmatcher) genfunc(name string) string {
	data := struct {
		Func string
		Rune string
	}{name, string(c)}
	var buf strings.Builder
	if err := templates.nnode.Execute(&buf, &data); err != nil {
		panic(err)
	}
	return buf.String()
}

// An binopmatcher matches based on the provided matchers and binary operation.
type binopmatcher struct {
	a, b exprmatcher
	op   rune
}

func (m *binopmatcher) genfunc(name string) string {
	data := struct {
		Func, MatchA, MatchB string
	}{name, newmatcher(), newmatcher()}
	var buf strings.Builder
	tmplmap := map[rune]*template.Template{
		'|': templates.cnode,
		'⋅': templates.seqnode,
	}
	if err := tmplmap[m.op].Execute(&buf, &data); err != nil {
		panic(err)
	}
	return strings.Join([]string{
		m.a.genfunc(data.MatchA),
		m.b.genfunc(data.MatchB),
		buf.String(),
	}, funcSeparator)
}

// An unopmatcher matches based on the provided matcher and unary operation.
type unopmatcher struct {
	a  exprmatcher
	op rune
}

func (m *unopmatcher) genfunc(name string) string {
	data := struct {
		Func, MatchA string
	}{name, newmatcher()}
	var buf strings.Builder
	tmplmap := map[rune]*template.Template{
		'*': templates.closure,
		'+': templates.posClosure,
	}
	if err := tmplmap[m.op].Execute(&buf, &data); err != nil {
		panic(err)
	}
	return strings.Join([]string{
		m.a.genfunc(data.MatchA),
		buf.String(),
	}, funcSeparator)
}

// A closurematcher matches
type closurematcher rune

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
			push(&unopmatcher{pop(), c})
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
			push(cmatcher(c))
		}
	}

	type matchcode struct {
		Name, Code string
	}
	matchcodes := make([]matchcode, len(stack))
	for i := range stack {
		name := newmatcher()
		matchcodes[i] = matchcode{
			Name: name,
			Code: stack[i].genfunc(name),
		}
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

// cmatcher returns a matcher for the given rune.
func cmatcher(c rune) matcher {
	return func(input []rune, pos int) (bool, int) {
		return input[pos] == c, 1
	}
}

{{ range . }}
{{ .Code }}

{{ end }}
func main() {
	if len(os.Args) != 2 {
		log.Fatalln("must supply input string")
	}
	input := os.Args[1]

	matchers := []func([]rune, int) (bool, int){
		{{  range . -}}
		{{ .Name }}, 
		{{ end -}}
	}

	pos := 0
	for _, matcher := range matchers {
		if pos >= len(input) {
			log.Fatalln("unmatching: input string too short")
		}
		ok, n := matcher([]rune(input), pos)
		if !ok {
			log.Fatalln("unmatching")
		}
		pos += n
	}
	fmt.Println("matching")
}
`)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, matchcodes); err != nil {
		panic(err)
	}
	return buf.String(), nil
}
