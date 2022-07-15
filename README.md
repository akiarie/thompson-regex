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

should produce output like the below (after running `go fmt`):

```Golang
package main

import (
	"fmt"
	"log"
	"os"
)

/* trimmed some stuff up here */

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("must supply input string")
	}
	input := []rune(os.Args[1])

	match := concat(
		concat(
			char('a'),
			closure(
				or(
					char('b'),
					char('c'),
				),
				0,
			),
		),
		char('d'),
	)

	matches := []string{}
	for i := 0; i < len(input); {
		if ok, n := match(input[i:]); ok {
			matches = append(matches, string(input[i:i+n]))
			i += n
			continue
		}
		i++
	}

	fmt.Printf("matches: %q\n", matches)
}

```

## Purpose.

Ken Thompson's [famous paper](https://dl.acm.org/doi/10.1145/363347.363387) on implementing regular
expressions is surprisingly simple and yet extremely sophisticated at the same time.
The "heart" of the method he proposed was the use of a stack in order to compile the expression to
code that parses it, which in this case was IBM 7094 object code.

I find this idea beautiful, and wanted a chance to play around with it for myself, as well as
demonstrate to anyone interested the way the algorithm works. Since the aim is demonstration rather
than a concrete (or usable) implementation, it makes sense to compile to a high-level language,
which is why I have started by targeting Go.

## Note on terminology.

I have used the term _assembler_ to refer to the portion of the code that produces Go source (and
hopefully other sources in future), which may be slightly confusing given its usual association with
the program that produces machine code out of assembly langauge. The aim is to draw an analogy:
Thompson's goal was efficiency, so his implementation produced object code; mine is illustration and
education, so it makes sense to target a high level language with the human mind being the real
intended machine.

## Contributing.

Feel free to create issues to share feedback or make suggestions. The way the codebase is designed
it is quite easy to add other output formats, provided they are text-based: see
[here](assembler/).
