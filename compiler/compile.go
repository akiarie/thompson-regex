package compiler

import (
	"fmt"
	"log"
	"strings"
	"text/template"
)

const (
	methodSeparator string = "\n\n"
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
	closure,
	posClosure *template.Template
}

var templates tmpset

func init() {
	var err error
	templates.nnode, err = template.New("nnode").Parse(`func {{.Func}}(input string, pos int) (bool, int) {
	if input[pos] == '{{.Char}}' {
		return true, 1
	}
	return false, 0
}`)
	if err != nil {
		log.Fatalln(err)
	}

	templates.cnode, err = template.New("cnode").Parse(`func {{.Func}}(input string, pos int) (bool, int) {
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

	templates.closure, err = template.New("closure").Parse(`func {{.Func}}(input string, pos int) (bool, int) {
	if ok, n := {{.MatchA}}(input, pos); ok {
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
	if err := templates.nnode.Execute(&buf, &data); err != nil {
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
	if err := templates.cnode.Execute(&buf, &data); err != nil {
		panic(err)
	}
	matchers := []string{a.code, b.code, buf.String()}
	return matchcode{data.Func, strings.Join(matchers, methodSeparator)}
}

func closurenode(a matchcode, op rune) matchcode {
	data := struct {
		Func, MatchA string
	}{newmatcher(), a.name}
	var buf strings.Builder
	tmplmap := map[rune]*template.Template{
		'*': templates.closure,
		'+': templates.posClosure,
	}
	if err := tmplmap[op].Execute(&buf, &data); err != nil {
		panic(err)
	}
	matchers := []string{a.code, buf.String()}
	return matchcode{data.Func, strings.Join(matchers, methodSeparator)}
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
			push(closurenode(pop(), c))
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
	matchers := make([]string, len(stack))
	for i := range stack {
		matchers[i] = stack[i].code
	}
	return strings.Join(matchers, methodSeparator), nil
}