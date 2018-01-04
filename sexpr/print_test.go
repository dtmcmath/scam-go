package sexpr

import (
	"testing"
)

func TestPrint(t *testing.T) {
	var tests = []struct{
		input string
		want  []string
	} {
		{ "'(1 2)", []string{ "(1 2)" } },
		{ "(+ 1 2 3)", []string{ "6" } },
		{ "(eq? 2 (+ 1 1))", []string { "#t" } },
		{
			"(+ 10 20)(cons 'atom 'molecule)",
			[]string{
				"30",
				"(atom . molecule)",
			},
		},
	}

	for _, test := range tests {
		_, sexprs := Parse("test", mkRuneChannel(test.input))
		idx := 0
		for sx := range sexprs {
			val := Evaluate(sx)
			if got := Sprint(val) ; got != test.want[idx] {
				t.Errorf("Print %q[%d]=%v, want %v",
					test.input, idx, got, test.want[idx],
				)
			}
			idx += 1
		}
		if idx != len(test.want) {
			t.Errorf("Parse %q gave %d exprs, want %d",
				test.input, idx, len(test.want),
			)
		}
	}
}
