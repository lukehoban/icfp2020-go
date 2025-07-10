package main

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPowerOf2(t *testing.T) {

	pwr2 := "ap ap s ap ap c ap eq 0 1 ap ap b ap mul 2 ap ap b pwr2 ap add -1"
	e, _ := parseExpr(strings.Split(pwr2, " "))
	symbols := map[Symbol]Expr{
		"pwr2": e,
	}

	for _, testCase := range []struct {
		input    int64
		expected int64
	}{
		// ap pwr2 0   =   ap ap ap s ap ap c ap eq 0 1 ap ap b ap mul 2 ap ap b pwr2 ap add -1 0
		// ap pwr2 0   =   ap ap ap ap c ap eq 0 1 0 ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 0
		// ap pwr2 0   =   ap ap ap ap eq 0 0 1 ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 0
		// ap pwr2 0   =   ap ap t 1 ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 0
		// ap pwr2 0   =   1
		{0, 1},
		// ap pwr2 1   =   ap ap ap s ap ap c ap eq 0 1 ap ap b ap mul 2 ap ap b pwr2 ap add -1 1
		// ap pwr2 1   =   ap ap ap ap c ap eq 0 1 1 ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 1
		// ap pwr2 1   =   ap ap ap ap eq 0 1 1 ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 1
		// ap pwr2 1   =   ap ap f 1 ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 1
		// ap pwr2 1   =   ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 1
		// ap pwr2 1   =   ap ap mul 2 ap ap ap b pwr2 ap add -1 1
		// ap pwr2 1   =   ap ap mul 2 ap pwr2 ap ap add -1 1
		// ap pwr2 1   =   ap ap mul 2 ap ap ap s ap ap c ap eq 0 1 ap ap b ap mul 2 ap ap b pwr2 ap add -1 ap ap add -1 1
		// ap pwr2 1   =   ap ap mul 2 ap ap ap ap c ap eq 0 1 ap ap add -1 1 ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 ap ap add -1 1
		// ap pwr2 1   =   ap ap mul 2 ap ap ap ap eq 0 ap ap add -1 1 1 ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 ap ap add -1 1
		// ap pwr2 1   =   ap ap mul 2 ap ap ap ap eq 0 0 1 ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 ap ap add -1 1
		// ap pwr2 1   =   ap ap mul 2 ap ap t 1 ap ap ap b ap mul 2 ap ap b pwr2 ap add -1 ap ap add -1 1
		// ap pwr2 1   =   ap ap mul 2 1
		// ap pwr2 1   =   2
		{1, 2},
		{3, 8},
		{4, 16},
		{5, 32},
		{8, 256},
	} {
		expr, _ := parseExpr([]string{"ap", "pwr2", strconv.FormatInt(testCase.input, 10)})
		v := eval(expr, symbols)
		assert.Equal(t, Number(testCase.expected), v)
	}
}

func TestCons(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected Expr
	}{
		{"ap ap cons 1 nil", &Ap{Left: &Ap{Left: Symbol("cons"), Right: Number(1)}, Right: Symbol("nil")}},
		{"ap ap cons 1 ap ap cons 2 nil", &Ap{Left: &Ap{Left: Symbol("cons"), Right: Number(1)}, Right: &Ap{Left: &Ap{Left: Symbol("cons"), Right: Number(2)}, Right: Symbol("nil")}}},
		{"ap ap cons ap ap cons 1 2 ap ap cons 3 4", &Ap{Left: &Ap{Left: Symbol("cons"), Right: &Ap{Left: Number(1), Right: Number(2)}}, Right: &Ap{Left: Symbol("cons"), Right: &Ap{Left: Number(3), Right: Number(4)}}}},
	} {
		expr, _ := parseExpr(strings.Split(testCase.input, " "))
		v := eval(expr, map[Symbol]Expr{})
		assert.Equal(t, testCase.expected, v)
	}
}

func TestGalaxy(t *testing.T) {
	symbols, err := parseProgram("galaxy.txt")
	assert.NoError(t, err)

	expr, _ := parseExpr([]string{"ap", "ap", "galaxy", "nil", "ap", "ap", "cons", "0", "0"})
	v := eval(expr, symbols)
	assert.Equal(t, "ap ap cons 0 ap ap cons ap ap cons 0 ap ap cons ap ap cons 0 nil ap ap cons 0 ap ap cons nil nil ap ap cons ap ap cons ap ap cons ap ap cons -1 -3 ap ap cons ap ap cons 0 -3 ap ap cons ap ap cons 1 -3 ap ap cons ap ap cons 2 -2 ap ap cons ap ap cons -2 -1 ap ap cons ap ap cons -1 -1 ap ap cons ap ap cons 0 -1 ap ap cons ap ap cons 3 -1 ap ap cons ap ap cons -3 0 ap ap cons ap ap cons -1 0 ap ap cons ap ap cons 1 0 ap ap cons ap ap cons 3 0 ap ap cons ap ap cons -3 1 ap ap cons ap ap cons 0 1 ap ap cons ap ap cons 1 1 ap ap cons ap ap cons 2 1 ap ap cons ap ap cons -2 2 ap ap cons ap ap cons -1 3 ap ap cons ap ap cons 0 3 ap ap cons ap ap cons 1 3 nil ap ap cons ap ap cons ap ap cons -7 -3 ap ap cons ap ap cons -8 -2 nil ap ap cons nil nil nil", printExpr(v))
	raw := toValue(v)
	assert.Equal(t, []interface{}{}, raw)
	// ap ap cons 0 ap ap cons ap ap cons 0 ap ap cons ap ap cons 0 nil ap ap cons 0 ap ap cons nil nil ap ap cons ap ap cons ap ap cons ap ap cons -1 -3 ap ap cons ap ap cons 0 -3 ap ap cons ap ap cons 1 -3 ap ap cons ap ap cons 2 -2 ap ap cons ap ap cons -2 -1 ap ap cons ap ap cons -1 -1 ap ap cons ap ap cons 0 -1 ap ap cons ap ap cons 3 -1 ap ap cons ap ap cons -3 0 ap ap cons ap ap cons -1 0 ap ap cons ap ap cons 1 0 ap ap cons ap ap cons 3 0 ap ap cons ap ap cons -3 1 ap ap cons ap ap cons 0 1 ap ap cons ap ap cons 1 1 ap ap cons ap ap cons 2 1 ap ap cons ap ap cons -2 2 ap ap cons ap ap cons -1 3 ap ap cons ap ap cons 0 3 ap ap cons ap ap cons 1 3 nil ap ap cons ap ap cons ap ap cons -7 -3 ap ap cons ap ap cons -8 -2 nil ap ap cons nil nil nil
	//

}

// Tests for parsing functions
func TestParseExpr(t *testing.T) {
	// Test parsing numbers
	expr, rest := parseExpr([]string{"42"})
	assert.Equal(t, Number(42), expr)
	assert.Empty(t, rest)

	// Test parsing negative numbers
	expr, rest = parseExpr([]string{"-42"})
	assert.Equal(t, Number(-42), expr)
	assert.Empty(t, rest)

	// Test parsing symbols
	expr, rest = parseExpr([]string{"symbol"})
	assert.Equal(t, Symbol("symbol"), expr)
	assert.Empty(t, rest)

	// Test parsing application
	expr, rest = parseExpr([]string{"ap", "f", "x"})
	expected := &Ap{Left: Symbol("f"), Right: Symbol("x")}
	assert.Equal(t, expected, expr)
	assert.Empty(t, rest)

	// Test nested applications
	expr, rest = parseExpr([]string{"ap", "ap", "f", "x", "y"})
	expected = &Ap{Left: &Ap{Left: Symbol("f"), Right: Symbol("x")}, Right: Symbol("y")}
	assert.Equal(t, expected, expr)
	assert.Empty(t, rest)

	// Test parsing with remaining tokens
	expr, rest = parseExpr([]string{"42", "extra", "tokens"})
	assert.Equal(t, Number(42), expr)
	assert.Equal(t, []string{"extra", "tokens"}, rest)

	// Test empty input
	expr, rest = parseExpr([]string{})
	assert.Nil(t, expr)
	assert.Empty(t, rest)
}

// Tests for built-in functions
func TestBuiltinFunctions(t *testing.T) {
	symbols := map[Symbol]Expr{}

	// Test neg function
	expr, _ := parseExpr([]string{"ap", "neg", "42"})
	result := eval(expr, symbols)
	assert.Equal(t, Number(-42), result)

	expr, _ = parseExpr([]string{"ap", "neg", "-10"})
	result = eval(expr, symbols)
	assert.Equal(t, Number(10), result)

	// Test i (identity) function
	expr, _ = parseExpr([]string{"ap", "i", "42"})
	result = eval(expr, symbols)
	assert.Equal(t, Number(42), result)

	expr, _ = parseExpr([]string{"ap", "i", "symbol"})
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("symbol"), result)

	// Test nil function - nil applied to anything returns t
	expr, _ = parseExpr([]string{"ap", "nil", "anything"})
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("t"), result)

	// Test basic cons creation - cons creates a data structure
	expr, _ = parseExpr([]string{"ap", "ap", "cons", "1", "2"})
	result = eval(expr, symbols)
	// cons creates a structure that can be accessed with car/cdr
	assert.NotNil(t, result)
	
	// Test that cons result is an Ap structure
	consResult, ok := result.(*Ap)
	assert.True(t, ok)
	assert.NotNil(t, consResult)
}

// Tests for binary operations
func TestBinaryOperations(t *testing.T) {
	symbols := map[Symbol]Expr{}

	// Test add operation
	expr, _ := parseExpr([]string{"ap", "ap", "add", "3", "4"})
	result := eval(expr, symbols)
	assert.Equal(t, Number(7), result)

	expr, _ = parseExpr([]string{"ap", "ap", "add", "-5", "3"})
	result = eval(expr, symbols)
	assert.Equal(t, Number(-2), result)

	// Test mul operation
	expr, _ = parseExpr([]string{"ap", "ap", "mul", "3", "4"})
	result = eval(expr, symbols)
	assert.Equal(t, Number(12), result)

	expr, _ = parseExpr([]string{"ap", "ap", "mul", "-2", "5"})
	result = eval(expr, symbols)
	assert.Equal(t, Number(-10), result)

	// Test div operation
	expr, _ = parseExpr([]string{"ap", "ap", "div", "10", "2"})
	result = eval(expr, symbols)
	assert.Equal(t, Number(5), result)

	expr, _ = parseExpr([]string{"ap", "ap", "div", "7", "3"})
	result = eval(expr, symbols)
	assert.Equal(t, Number(2), result) // integer division

	// Test lt (less than) operation
	expr, _ = parseExpr([]string{"ap", "ap", "lt", "3", "5"})
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("t"), result)

	expr, _ = parseExpr([]string{"ap", "ap", "lt", "5", "3"})
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("f"), result)

	expr, _ = parseExpr([]string{"ap", "ap", "lt", "3", "3"})
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("f"), result)

	// Test eq (equal) operation
	expr, _ = parseExpr([]string{"ap", "ap", "eq", "3", "3"})
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("t"), result)

	expr, _ = parseExpr([]string{"ap", "ap", "eq", "3", "4"})
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("f"), result)

	expr, _ = parseExpr([]string{"ap", "ap", "eq", "-5", "-5"})
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("t"), result)
}

// Tests for combinator operations
func TestCombinatorOperations(t *testing.T) {
	symbols := map[Symbol]Expr{}

	// Test s combinator: ap ap ap s f g x = ap ap f x ap g x
	// Test with known identity: s i i x = x
	iSym := Symbol("i")
	expr := &Ap{Left: &Ap{Left: &Ap{Left: Symbol("s"), Right: iSym}, Right: iSym}, Right: Number(42)}
	result := eval(expr, symbols)
	assert.Equal(t, Number(42), result)

	// Test c combinator: ap ap ap c f g x = ap ap f x g
	// c with add should flip arguments: c add 3 5 should be add 5 3
	cExpr, _ := parseExpr([]string{"ap", "ap", "ap", "c", "add", "3", "5"})
	result = eval(cExpr, symbols)
	assert.Equal(t, Number(8), result) // should be same as add 5 3

	// Test b combinator: ap ap ap b f g x = ap f ap g x
	// For simple test, let's use b i add 5 which should be i (add 5) = add 5
	expr = &Ap{
		Left: &Ap{
			Left: &Ap{Left: Symbol("b"), Right: Symbol("i")}, 
			Right: Symbol("add"),
		}, 
		Right: Number(5),
	}
	result = eval(expr, symbols)
	// The result should be a partial application of add with 5
	assert.NotNil(t, result)
	
	// Apply the result to another number to complete the addition
	finalExpr := &Ap{Left: result, Right: Number(3)}
	finalResult := eval(finalExpr, symbols)
	assert.Equal(t, Number(8), finalResult) // i(add 5) 3 = add 5 3 = 8

	// Test cons operation with three arguments - cons creates a selector function
	consExpr1, _ := parseExpr([]string{"ap", "ap", "ap", "cons", "1", "2", "t"})
	result = eval(consExpr1, symbols)
	assert.Equal(t, Number(1), result) // cons 1 2 t = 1

	consExpr2, _ := parseExpr([]string{"ap", "ap", "ap", "cons", "1", "2", "f"})
	result = eval(consExpr2, symbols)
	assert.Equal(t, Number(2), result) // cons 1 2 f = 2
}

// Tests for printExpr function
func TestPrintExpr(t *testing.T) {
	// Test printing numbers
	assert.Equal(t, "42", printExpr(Number(42)))
	assert.Equal(t, "-10", printExpr(Number(-10)))
	assert.Equal(t, "0", printExpr(Number(0)))

	// Test printing symbols
	assert.Equal(t, "symbol", printExpr(Symbol("symbol")))
	assert.Equal(t, "add", printExpr(Symbol("add")))
	assert.Equal(t, "t", printExpr(Symbol("t")))

	// Test printing applications
	ap := &Ap{Left: Symbol("f"), Right: Symbol("x")}
	assert.Equal(t, "ap f x", printExpr(ap))

	// Test printing nested applications
	nested := &Ap{Left: &Ap{Left: Symbol("f"), Right: Symbol("x")}, Right: Symbol("y")}
	assert.Equal(t, "ap ap f x y", printExpr(nested))

	// Test printing application with numbers
	apNum := &Ap{Left: Symbol("add"), Right: Number(42)}
	assert.Equal(t, "ap add 42", printExpr(apNum))

	// Test complex expression
	complex := &Ap{
		Left: &Ap{Left: Symbol("add"), Right: Number(3)},
		Right: &Ap{Left: Symbol("mul"), Right: Number(4)},
	}
	assert.Equal(t, "ap ap add 3 ap mul 4", printExpr(complex))
}

// Tests for toValue function
func TestToValue(t *testing.T) {
	// Test converting numbers
	assert.Equal(t, int64(42), toValue(Number(42)))
	assert.Equal(t, int64(-10), toValue(Number(-10)))
	assert.Equal(t, int64(0), toValue(Number(0)))

	// Test converting nil symbol
	assert.Equal(t, []interface{}(nil), toValue(Symbol("nil")))

	// Test converting cons structures to arrays
	// Create: cons 1 nil
	consOne := &Ap{
		Left: &Ap{Left: Symbol("cons"), Right: Number(1)},
		Right: Symbol("nil"),
	}
	expected := []interface{}{int64(1)}
	assert.Equal(t, expected, toValue(consOne))

	// Create: cons 1 (cons 2 nil)
	consTwo := &Ap{
		Left: &Ap{Left: Symbol("cons"), Right: Number(2)},
		Right: Symbol("nil"),
	}
	consList := &Ap{
		Left: &Ap{Left: Symbol("cons"), Right: Number(1)},
		Right: consTwo,
	}
	expected = []interface{}{int64(1), int64(2)}
	assert.Equal(t, expected, toValue(consList))

	// Test converting cons pair (not a list)
	consPair := &Ap{
		Left: &Ap{Left: Symbol("cons"), Right: Number(1)},
		Right: Number(2),
	}
	expectedPair := struct{ Left, Right interface{} }{Left: int64(1), Right: int64(2)}
	assert.Equal(t, expectedPair, toValue(consPair))
}

// Tests for edge cases and complex scenarios
func TestEdgeCases(t *testing.T) {
	symbols := map[Symbol]Expr{}

	// Test evaluation with undefined symbols
	expr, _ := parseExpr([]string{"undefined_symbol"})
	result := eval(expr, symbols)
	assert.Equal(t, Symbol("undefined_symbol"), result) // should return the symbol unchanged

	// Test deeply nested applications
	deep := &Ap{
		Left: &Ap{
			Left: &Ap{Left: Symbol("add"), Right: Number(1)},
			Right: Number(2),
		},
		Right: Number(3),
	}
	result = eval(deep, symbols)
	assert.Equal(t, Number(3), result) // ((add 1) 2) should be add 1 2 = 3, then applied to 3

	// Test cons with complex nested structure
	expr, _ = parseExpr([]string{"ap", "ap", "cons", "ap", "ap", "add", "1", "2", "ap", "ap", "mul", "3", "4"})
	result = eval(expr, symbols)
	// This creates a cons pair of (3, 12)
	expectedStruct := struct{ Left, Right interface{} }{Left: int64(3), Right: int64(12)}
	assert.Equal(t, expectedStruct, toValue(result))

	// Test boolean operations with t and f
	expr, _ = parseExpr([]string{"ap", "ap", "t", "42", "0"})
	result = eval(expr, symbols)
	assert.Equal(t, Number(42), result) // t x y = x

	expr, _ = parseExpr([]string{"ap", "ap", "f", "42", "0"})
	result = eval(expr, symbols)
	assert.Equal(t, Number(0), result) // f x y = y

	// Test identity with different types
	expr, _ = parseExpr([]string{"ap", "i", "t"})
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("t"), result)

	expr, _ = parseExpr([]string{"ap", "i", "f"})
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("f"), result)
}

// Tests for symbol resolution
func TestSymbolResolution(t *testing.T) {
	// Create some symbols
	symbols := map[Symbol]Expr{
		"x": Number(42),
		"y": Symbol("add"),
		"z": &Ap{Left: Symbol("add"), Right: Number(1)},
	}

	// Test resolving simple symbol
	expr := Symbol("x")
	result := eval(expr, symbols)
	assert.Equal(t, Number(42), result)

	// Test resolving symbol to another symbol
	expr = Symbol("y")
	result = eval(expr, symbols)
	assert.Equal(t, Symbol("add"), result)

	// Test resolving symbol to expression
	expr = Symbol("z")
	result = eval(expr, symbols)
	expectedAp := &Ap{Left: Symbol("add"), Right: Number(1)}
	assert.Equal(t, expectedAp, result)

	// Test using resolved symbols in calculations
	parsedExpr, _ := parseExpr([]string{"ap", "ap", "y", "x", "10"})
	result = eval(parsedExpr, symbols)
	assert.Equal(t, Number(52), result) // y is add, x is 42, so add 42 10 = 52
}

// Tests for parseProgram function
func TestParseProgram(t *testing.T) {
	// Test parsing a simple program file
	symbols, err := parseProgram("/tmp/test_program.txt")
	assert.NoError(t, err)
	assert.NotNil(t, symbols)

	// Check that symbols were parsed correctly
	assert.Equal(t, Number(42), symbols["x"])
	assert.Equal(t, Symbol("add"), symbols["y"])
	
	expectedZ := &Ap{Left: Symbol("add"), Right: Number(1)}
	assert.Equal(t, expectedZ, symbols["z"])

	expectedComplex := &Ap{Left: &Ap{Left: Symbol("add"), Right: Symbol("x")}, Right: Number(10)}
	assert.Equal(t, expectedComplex, symbols["complex"])

	// Test error case with non-existent file
	_, err = parseProgram("/nonexistent/file.txt")
	assert.Error(t, err)
}
