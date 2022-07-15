package compiler

import "testing"

func TestRPNConvert(t *testing.T) {
	cases := map[string]string{
		"a⋅(b|c)*⋅d":   "abc|*⋅d⋅",
		"a⋅(a⋅b|c)*⋅d": "aab⋅c|*⋅d⋅",
		"a⋅b|c⋅d":      "ab⋅cd⋅|",
		"(a⋅b)|(c⋅d)":  "ab⋅cd⋅|",
		"a⋅n⋅d⋅r⋅e⋅w|j⋅a⋅c⋅k⋅s⋅o⋅n":     "an⋅d⋅r⋅e⋅w⋅ja⋅c⋅k⋅s⋅o⋅n⋅|",
		"(a⋅n⋅d⋅r⋅e⋅w)|(j⋅a⋅c⋅k⋅s⋅o⋅n)": "an⋅d⋅r⋅e⋅w⋅ja⋅c⋅k⋅s⋅o⋅n⋅|",
	}
	for r, rpn := range cases {
		out, err := RPNConvert(r)
		if err != nil {
			t.Fatal(err)
		}
		if rpn != out {
			t.Fatalf("expected %q got %q", rpn, out)
		}
	}
}
