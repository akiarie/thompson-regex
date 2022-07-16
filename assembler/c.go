package assembler

import (
	"strings"
	"text/template"
)

func C(root MatcherGenerator) (string, error) {
	tmpl, err := template.New("program").Parse(`#include <stdio.h>
#include <stdbool.h>

/* This implementation is hideous because I'm shoehorning C into the structures I
 * developed for Go source, which rely on closures for simple output code. */

/* matches and increments the ptr */
bool matchc(char c, char **input) {
	if (*input[0] == '\0' || *input[0] != c) {
		return false;
	}
	(*input)++;
	return true;
}

bool match(char **input) {
	bool matched = true;

	{{ . }}

	return matched;
}

int main(int argc, char **argv) {
   if (argc != 2) {
	   printf("must supply input string\n");
	   return 1;
   }
   char* input = argv[1];
   while (input[0] != '\0') {
	   char* start = input;
	   if (match(&input)) {
		   printf("%.*s\n", (int)(input-start), start);
		   continue;
	   } 
	   input++;
   }
}
	`)
	if err != nil {
		return "", err
	}

	gens, err := codeGens(`if (matched) {
if (!matchc('{{ . }}', input)) {
			matched = false;
		}
	}
`,
		`// BEGIN OR
	if (matched) {
		{{ .MatcherFuncA }}
		if (!matched) {
			matched = true;
			{{ .MatcherFuncB }}
		}
	} // END OR`,
		`// BEGIN CONCAT
	if (matched) {
		{{ .MatcherFuncA }}
		{{ .MatcherFuncB }}
	} // END CONCAT `,
		`// BEGIN CLOSURE
	if (matched) {
		for (int i = 0; i < 0; i++) {
			{{ .MatcherFuncA }}
		}
		while (matched) {
			{{ .MatcherFuncA }}
		}
		matched = true;
	} // END CLOSURE`,
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
