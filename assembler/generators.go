package assembler

import (
	"strings"
	"text/template"
)

// The Assemblers are the functions that construct output source.
var Assemblers = map[string]func(MatcherGenerator) (string, error){
	"golang": Go,
	"go":     Go,

	"C": C,
	"c": C,

	"python3": Python3,
}

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
