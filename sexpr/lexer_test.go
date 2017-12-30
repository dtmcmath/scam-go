package sexpr

import (
	"fmt"
	"testing"
)

// func deepEqual(a []interface{}, b []interface{}) bool {
func deepEqual(a []item, b []item) bool {
	for len(a) == len(b) {
		if len(a) == 0 {
			return true
		} else if a[0] != b[0] {
			return false
		}
		// else
		a = a[1:]
		b = b[1:]
	}
	return false
}

func TestSimpleLexer(t *testing.T) {
	var tests = []struct {
		input string
		want []item
	}{
		{
			"",
			[]item { {itemEOF, ""} },
		},
		{
			"3.14159",
			[]item { {itemNumber, "3.14159"}, {itemEOF, ""} },
		},
		{
			"()",
			[]item {
				{ itemLparen, "(" },
				{ itemRparen, ")" },
				{ itemEOF, "" },
			},
		},
		{
			"(+ 1 2)",
			[]item {
				{ itemLparen, "(" },
				{ itemPlus, "+" },
				{ itemNumber, "1" },
				{ itemNumber, "2" },
				{ itemRparen, ")" },
				{ itemEOF, ""},
			},
		},
	}
	for _, test := range tests {
		_, ch := lex("test", test.input)
		var got []item
		for it := range ch {
			got = append(got, it)
		}
		if !deepEqual(got, test.want) {
			t.Errorf("Lexed '%s' got '%v', wanted '%v'",
				test.input, got, test.want,
			)
		}
	}
}
			
func ExampleLexer() {
	sexpr := " 'abc (3.14159)"
	_, ch := lex("test", sexpr)
	for tok := range ch {
		fmt.Println(tok)
	}
	// Output:
	// QSYMBOL(abc)
	// LPAREN
	// NUMBER(3.14159)
	// RPAREN
	// EOF
}


func ExampleLexing2 () {
	// The answer here is a little controversial.  The lexer thinks a
	// bare "+" looks more like a number than the keyword.  It's
	// neither, so I really can't complain that the lexer thinks it's
	// a number.  The Parser will reject such an expression as a bad
	// number.
	sexpr := "+"
		_, ch := lex("test", sexpr)
	for tok := range ch {
		fmt.Println(tok)
	}
	// Output:
	// NUMBER(+)
	// EOF
}

func ExampleBadLexing () {
	sexpr := "(add '0 1)"
		_, ch := lex("test", sexpr)
	for tok := range ch {
		fmt.Println(tok)
	}
	// Output:
	// LPAREN
	// SYMBOL(add)
	// ERROR(quoted symbols must start with a letter, not "'0 1)")
}

func ExampleBadLexing2 () {
	sexpr := "(0add 0 1)"
		_, ch := lex("test", sexpr)
	for tok := range ch {
		fmt.Println(tok)
	}
	// Output:
	// LPAREN
	// ERROR(bad number syntax: "0a")
}
