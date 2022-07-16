package assembler

import (
	"strings"
	"text/template"
)

func Python3(root MatcherGenerator) (string, error) {
	tmpl, err := template.New("program").Parse(`import sys
from typing import Union


class Matcher():
    def match(self, inputstr: str) -> Union[bool, int]:
        raise Exception("not implemented")

    def __or__(self, a):
        return Or(self, a)

    def __add__(self, a):
        return Concat(self, a)

    def __pow__(self, min: int):
        return Closure(self, min)


class Char(Matcher):
    def __init__(self, c: str):
        self.c = c

    def match(self, inputstr: str) -> Union[bool, int]:
        if len(inputstr) == 0:
            return False, 0
        return inputstr[0] == self.c, 1


class Or(Matcher):
    def __init__(self, a: Matcher, b: Matcher):
        self.am, self.bm = a.match, b.match

    def match(self, inputstr: str) -> Union[bool, int]:
        ismatch, n = self.am(inputstr)
        if ismatch:
            return True, n
        return self.bm(inputstr)


class Concat(Matcher):
    def __init__(self, a: Matcher, b: Matcher):
        self.am, self.bm = a.match, b.match

    def match(self, inputstr: str) -> Union[bool, int]:
        ismatch, n = self.am(inputstr)
        if not ismatch:
            return False, 0
        ismatch, m = self.bm(inputstr[n:])
        return ismatch, n+m


class Closure(Matcher):
    def __init__(self, a: Matcher, min: int):
        self.am = a.match
        self.min = min

    def match(self, inputstr: str) -> Union[bool, int]:
        n = 0
        matches = 0
        while matches < self.min: 
            ismatch, m = self.am(inputstr[n:])
            if not ismatch:
               return False, n
            n += m
            matches += 1
        while True:
            ismatch, m = self.am(inputstr[n:])
            if not ismatch:
               return True, n
            n += m


if len(sys.argv) != 2:
    print("must supply input string")
    sys.exit()

inputstr = sys.argv[1]

exprmatcher = {{ . }}

matches = []

i = 0
while i < len(inputstr):
    ismatch, n = exprmatcher.match(inputstr[i:])
    if ismatch:
        matches += [inputstr[i:i+n]]
        i += n
    else:
        i += 1

print(matches)
`)
	if err != nil {
		return "", err
	}

	gens, err := codeGens(
		"Char('{{ . }}')",
		`({{ .MatcherFuncA }} | {{ .MatcherFuncB }})`,
		`({{ .MatcherFuncA }} + {{ .MatcherFuncB }})`,
		"{{ .MatcherFuncA }} ** {{ .Min }}",
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
