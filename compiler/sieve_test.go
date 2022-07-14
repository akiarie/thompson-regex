package compiler

import "testing"

func TestSieveConvert(t *testing.T) {
	cases := map[string]string{
		"a(b|c)*d":           "a⋅(bc|)*⋅d",
		"ab|cd":              "a⋅b|c⋅d",
		"(ab)|(cd)":          "(a⋅b)|(c⋅d)",
		"andrew|jackson":     "a⋅n⋅d⋅r⋅e⋅w|j⋅a⋅c⋅k⋅s⋅o⋅n",
		"(andrew)|(jackson)": "(a⋅n⋅d⋅r⋅e⋅w)|(j⋅a⋅c⋅k⋅s⋅o⋅n)",
	}
	for r, rpn := range cases {
		out, err := Sieve(r)
		if err != nil {
			t.Fatal(err)
		}
		if rpn != out {
			t.Fatalf("expected %q got %q", rpn, out)
		}
	}
}
