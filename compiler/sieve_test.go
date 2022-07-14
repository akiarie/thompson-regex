package compiler

/*
func TestRPNConvert(t *testing.T) {
	cases := map[string]string{
		"a(b|c)*d": "a⋅bc|*⋅d",
		"ab|cd":    "ab⋅cd|",
		//"(ab)|(cd)": "abcd|",
		//"andrew|jackson":     "andrewj|ackson",
		//"(andrew)|(jackson)": "andrew⋅jackson⋅|",
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
*/
