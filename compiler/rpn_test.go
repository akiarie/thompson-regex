package compiler

import "testing"

func TestRPNConvert(t *testing.T) {
	cases := map[string]string{
		"a^(b|c)*^d":                "abc|*d^^",
		"a^b|c^d":                   "ab^cd^|",
		"(a^b)|(c^d)":               "ab^cd^|",
		"a^n^d^r^e^w|j^a^c^k^s^o^n": "andrew^^^^^jackson^^^^^^|",
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
