# A Thompson-style Regex-to-Golang Compiler.

This application takes in a simple regular expression (letters and digits with no spaces and the
operators _|_, _*_ and _+_ together with parenthesization) and outputs the source to a Go program
which parses inputs for matches to the same expression.

It is (obviously) for educational purposes, intended as an illustration of the basic logic involved
in Thompson's construction algorithm.
