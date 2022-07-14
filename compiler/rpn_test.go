package compiler

import "testing"

func TestRPNConvert(t *testing.T) {
	cases := map[string]string{
		//"a⋅(bc|)*⋅d": "a⋅bc|*⋅d",
		//"a⋅b|c⋅d":
		//"(a⋅b)|(c⋅d)":
		//"a⋅n⋅d⋅r⋅e⋅w|j⋅a⋅c⋅k⋅s⋅o⋅n":
		//"(a⋅n⋅d⋅r⋅e⋅w)|(j⋅a⋅c⋅k⋅s⋅o⋅n)":
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
