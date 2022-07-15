# A Thompson-style Regex-to-Golang Compiler.

This repo contains a command-line app that takes in a simple regular expression (letters and digits
with no spaces and the operators "|", "&ast;" and "+" together with parenthesization) and outputs
the source to a Go program which parses inputs for matches to the same expression.

It is intended for educational purposes and not for use in any production system.

## Usage.

We will use the example from Thompson's paper for demonstrative purposes. Executing

```bash
./thompson-regex 'a(b|c)*d'
```

should produce output like

```Golang
package main

import (
	"fmt"
	"log"
	"os"
)

// A matcher is a function representing a particular expression, returning true
// if the given rune slice matches the expression with the length of the match,
// or false otherwise with an undefined integer.
type matcher func([]rune) (bool, int)

/* trimmed helper functions out */

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("must supply input string")
	}
	input := []rune(os.Args[1])

	matcher := concatmatcher(
		concatmatcher(
			runematcher('a'),
			closurematcher(
				ormatcher(
					runematcher('b'),
					runematcher('c'),
				),
				0,
			),
		),
		runematcher('d'),
	)

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
```

## Contributing.

Feel free to create issues to share feedback or make suggestions. The way the codebase is designed
it is quite easy to add other output formats, provided they are text-based, see
[here](assembler/).
