package compiler

import "testing"

func TestRPNConvert(t *testing.T) {
	cases := map[string]string{
		"a(b|c)*d":           "abc|⋅*d⋅",
		"ab|cd":              "abc|⋅d⋅",
		"(andrew)|(jackson)": "ab⋅cd⋅|",
	}
	for r, rpn := range cases {
		out, err := RPNConvert(r)
		if err != nil {
			t.Fatal(err)
		}
		if rpn != out {
			t.Fatalf("Expected %q got %q", rpn, out)
		}
	}
}
