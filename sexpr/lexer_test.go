package sexpr

import (
	"fmt"
	"testing"
	"reflect"
)

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
				{ itemSymbol, "+" },
				{ itemNumber, "1" },
				{ itemNumber, "2" },
				{ itemRparen, ")" },
				{ itemEOF, ""},
			},
		},
		{
			"o+",
			[]item {
				{ itemSymbol, "o+" },
				{ itemEOF, ""},
			},
		},
		{
			"o- o+",
			[]item {
				{ itemSymbol, "o-" },
				{ itemSymbol, "o+" },
				{ itemEOF, ""},
			},
		},
		{
			"1st",
			[]item {
				{ itemSymbol, "1st" },
				{ itemEOF, ""},
			},
		},
		{
			"lambda(x)",
			[]item {
				{ itemSymbol, "lambda" },
				{ itemLparen, "(" },
				{ itemSymbol, "x" },
				{ itemRparen, ")" },
				{ itemEOF, "" },
			},
		},
		{
			"([cond (null? x)])",
			[]item {
				{ itemLparen, "(" },
				{ itemLparen, "[" },
				{ itemSymbol, "cond" },
				{ itemLparen, "(" },
				{ itemSymbol, "null?" },
				{ itemSymbol, "x" },
				{ itemRparen, ")" },
				{ itemRparen, "]" },
				{ itemRparen, ")" },
				{ itemEOF, "" },
			},
		},
	}
	for _, test := range tests {
		_, ch := lex("test", mkRuneChannel(test.input))
		var got []item
		for it := range ch {
			got = append(got, it)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("Lexed '%s' got '%v', wanted '%v'",
				test.input, got, test.want,
			)
		}
	}
}

func TestLexComments (t *testing.T) {
	var tests = []struct {
		input string
		want []itemType
	}{
		{ "", []itemType{ itemEOF } },
		{
			`; Hello
World`,
			[]itemType{ itemComment, itemSymbol, itemEOF },
		},
		{
			"1 ; A number",
			[]itemType{ itemNumber, itemComment, itemEOF },
		},
	}

	for _, test := range tests {
		_, ch := lex("test", mkRuneChannel(test.input))
		var got []itemType
		for it := range ch {
			got = append(got, it.typ)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("Lexed %q got %v, wanted %v",
				test.input, got, test.want,
			)
		}
	}
}

func ExampleLexer() {
	sexpr := " 'abc (3.14159)"
	_, ch := lex("test", mkRuneChannel(sexpr))
	for tok := range ch {
		fmt.Println(tok)
	}
	// Output:
	// QUOTE
	// SYMBOL(abc)
	// LPAREN
	// NUMBER(3.14159)
	// RPAREN
	// EOF
}


func ExampleLexing2 () {
	sexpr := "+"
		_, ch := lex("test", mkRuneChannel(sexpr))
	for tok := range ch {
		fmt.Println(tok)
	}
	// Output:
	// SYMBOL(+)
	// EOF
}

func ExampleBadLexing () {
	sexpr := "(add '0 1)"
		_, ch := lex("test", mkRuneChannel(sexpr))
	for tok := range ch {
		fmt.Println(tok)
	}
	// Output:
	// LPAREN
	// SYMBOL(add)
	// QUOTE
	// NUMBER(0)
	// NUMBER(1)
	// RPAREN
	// EOF
}

func ExampleSymbolLexing () {
	sexpr := "(0add 0 1)"
		_, ch := lex("test", mkRuneChannel(sexpr))
	for tok := range ch {
		fmt.Println(tok)
	}
	// Output:
	// LPAREN
	// SYMBOL(0add)
	// NUMBER(0)
	// NUMBER(1)
	// RPAREN
	// EOF
}
