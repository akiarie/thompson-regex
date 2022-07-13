package compiler

import (
	"fmt"
	"log"
	"strings"
	"text/template"
)

var matchernames []string = []string{}

func newmatcher() string {
	name := fmt.Sprintf("match%d", len(matchernames))
	matchernames = append(matchernames, name)
	return name
}

var (
	nnodeTmpl,
	cnodeTmpl *template.Template
)

func init() {
	nnode := `func {{.Func}}(input string, pos int) (bool, int) {
	if input[pos] == '{{.Char}}' {
		return true, 1
	}
	return false, 0
}`
	tmpl, err := template.New("nnode").Parse(nnode)
	if err != nil {
		log.Fatalln(err)
	}
	nnodeTmpl = tmpl

	cnode := `func {{.Func}}(input string, pos int) (bool, int) {
	if ok, n := {{.MatchA}}(input, pos); ok {
		return true, n
	} else if ok, n := {{.MatchB}}(input, pos); ok {
		return true, n
	}
	return false, 0
}`
	tmpl, err = template.New("cnode").Parse(cnode)
	if err != nil {
		log.Fatalln(err)
	}
	cnodeTmpl = tmpl
}

// A matchcode contains the code for matching a parsed expression
type matchcode struct {
	name, code string
}

// nnode returns a matchcode for the given character
func nnode(c rune) matchcode {
	data := struct {
		Func, Char string
	}{newmatcher(), string(c)}
	var buf strings.Builder
	if err := nnodeTmpl.Execute(&buf, &data); err != nil {
		panic(err)
	}
	return matchcode{data.Func, buf.String()}
}

// cnode returns a matchcode which permits matching of either of the given
// matchcodes
func cnode(a, b matchcode) matchcode {
	data := struct {
		Func, MatchA, MatchB string
	}{newmatcher(), a.name, b.name}
	var buf strings.Builder
	if err := cnodeTmpl.Execute(&buf, &data); err != nil {
		panic(err)
	}
	return matchcode{data.Func, buf.String()}
}

// Compile returns a string containing the Go source for a command-line program
// that takes a string as its input and outputs the matches of the given regex.
// The regex must be in reverse Polish notation.
func Compile(regex string) (string, error) {
	stack := []matchcode{}
	pop := func() matchcode {
		mc := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return mc
	}
	push := func(mc matchcode) {
		stack = append(stack, mc)
	}
	for _, c := range regex {
		switch c {
		case '+':
		case '*':
			continue
		case '|':
			if len(stack) < 2 {
				return "", fmt.Errorf("cannot use %q with less than 2 elements", c)
			}
			push(cnode(pop(), pop()))
			continue
		default:
			push(nnode(c))
		}
	}
	return "", nil
}
