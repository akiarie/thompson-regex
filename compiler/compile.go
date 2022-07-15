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
	seqnode,
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

	templates.seqnode, err = template.New("seqnode").Parse(`func {{.Func}}(input string, pos int) (bool, int) {
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

	templates.closure, err = template.New("closure").Parse(`func {{.Func}}(input string, pos int) (bool, int) {
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

type matchcode struct {
	Name, Code string
}

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

func binopnode(a, b matchcode, op rune) matchcode {
	data := struct {
		Func, MatchA, MatchB string
	}{
		Func: newmatcher(),
		// note the swap, the topmost element on the stack is second in sequence
		MatchA: b.Name,
		MatchB: a.Name,
	}
	var buf strings.Builder
	tmplmap := map[rune]*template.Template{
		'|': templates.cnode,
		'⋅': templates.seqnode,
	}
	if err := tmplmap[op].Execute(&buf, &data); err != nil {
		panic(err)
	}
	matchers := []string{a.Code, b.Code, buf.String()}
	return matchcode{data.Func, strings.Join(matchers, methodSeparator)}
}

func unopnode(a matchcode, op rune) matchcode {
	data := struct {
		Func, MatchA string
	}{newmatcher(), a.Name}
	var buf strings.Builder
	tmplmap := map[rune]*template.Template{
		'*': templates.closure,
		'+': templates.posClosure,
	}
	if err := tmplmap[op].Execute(&buf, &data); err != nil {
		panic(err)
	}
	matchers := []string{a.Code, buf.String()}
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
		case '+', '*':
			push(unopnode(pop(), c))
			continue
		case '|', '⋅':
			if len(stack) < 2 {
				return "", fmt.Errorf("cannot use %q with less than 2 elements", c)
			}
			push(binopnode(pop(), pop(), c))
			continue
		default:
			push(nnode(c))
		}
	}

	tmpl, err := template.New("program").Parse(`package main

import (
	"fmt"
	"log"
	"os"
)

{{ range . }}
{{ .Code }}

{{ end }}
func main() {
	if len(os.Args) != 2 {
		log.Fatalln("must supply input string")
	}
	input := os.Args[1]

	matchers := []func(string, int) (bool, int){
		{{  range . -}}
		{{ .Name }}, 
		{{ end -}}
	}

	pos := 0
	for _, matcher := range matchers {
		if pos >= len(input) {
			log.Fatalln("unmatching: input string too short")
		}
		ok, n := matcher(input, pos)
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
	if err := tmpl.Execute(&buf, stack); err != nil {
		panic(err)
	}
	return buf.String(), nil
}
