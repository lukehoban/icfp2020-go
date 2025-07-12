package main

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// clearCache recursively clears the v field from Ap expressions for testing
func clearCache(expr Expr) Expr {
	switch e := expr.(type) {
	case *Ap:
		return &Ap{
			Left:  clearCache(e.Left),
			Right: clearCache(e.Right),
			v:     nil,
		}
	case Number, Symbol:
		return e
	default:
		return expr
	}
}

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
		{"ap ap cons ap ap cons 1 2 ap ap cons 3 4", &Ap{Left: &Ap{Left: Symbol("cons"), Right: &Ap{Left: &Ap{Left: Symbol("cons"), Right: Number(1)}, Right: Number(2)}}, Right: &Ap{Left: &Ap{Left: Symbol("cons"), Right: Number(3)}, Right: Number(4)}}},
	} {
		expr, _ := parseExpr(strings.Split(testCase.input, " "))
		v := eval(expr, map[Symbol]Expr{})
		// Clear cache for comparison
		cleanV := clearCache(v)
		assert.Equal(t, testCase.expected, cleanV)
	}
}

func TestGalaxy(t *testing.T) {
	symbols, err := parseProgram("galaxy.txt")
	assert.NoError(t, err)

	expr, _ := parseExpr([]string{"ap", "ap", "galaxy", "nil", "ap", "ap", "cons", "0", "0"})
	v := eval(expr, symbols)
	assert.Equal(t, "ap ap cons 0 ap ap cons ap ap cons 0 ap ap cons ap ap cons 0 nil ap ap cons 0 ap ap cons nil nil ap ap cons ap ap cons ap ap cons ap ap cons -1 -3 ap ap cons ap ap cons 0 -3 ap ap cons ap ap cons 1 -3 ap ap cons ap ap cons 2 -2 ap ap cons ap ap cons -2 -1 ap ap cons ap ap cons -1 -1 ap ap cons ap ap cons 0 -1 ap ap cons ap ap cons 3 -1 ap ap cons ap ap cons -3 0 ap ap cons ap ap cons -1 0 ap ap cons ap ap cons 1 0 ap ap cons ap ap cons 3 0 ap ap cons ap ap cons -3 1 ap ap cons ap ap cons 0 1 ap ap cons ap ap cons 1 1 ap ap cons ap ap cons 2 1 ap ap cons ap ap cons -2 2 ap ap cons ap ap cons -1 3 ap ap cons ap ap cons 0 3 ap ap cons ap ap cons 1 3 nil ap ap cons ap ap cons ap ap cons -7 -3 ap ap cons ap ap cons -8 -2 nil ap ap cons nil nil nil", printExpr(v))
	raw := toValue(v)

	// The galaxy function returns a complex structure representing game state
	// [0 [0 [0] 0 []] [[coordinates...] [other_coordinates...] []]]
	expected := []interface{}{
		int64(0),
		[]interface{}{
			int64(0),
			[]interface{}{int64(0)},
			int64(0),
			[]interface{}(nil),
		},
		[]interface{}{
			[]interface{}{
				struct{ Left, Right interface{} }{Left: int64(-1), Right: int64(-3)},
				struct{ Left, Right interface{} }{Left: int64(0), Right: int64(-3)},
				struct{ Left, Right interface{} }{Left: int64(1), Right: int64(-3)},
				struct{ Left, Right interface{} }{Left: int64(2), Right: int64(-2)},
				struct{ Left, Right interface{} }{Left: int64(-2), Right: int64(-1)},
				struct{ Left, Right interface{} }{Left: int64(-1), Right: int64(-1)},
				struct{ Left, Right interface{} }{Left: int64(0), Right: int64(-1)},
				struct{ Left, Right interface{} }{Left: int64(3), Right: int64(-1)},
				struct{ Left, Right interface{} }{Left: int64(-3), Right: int64(0)},
				struct{ Left, Right interface{} }{Left: int64(-1), Right: int64(0)},
				struct{ Left, Right interface{} }{Left: int64(1), Right: int64(0)},
				struct{ Left, Right interface{} }{Left: int64(3), Right: int64(0)},
				struct{ Left, Right interface{} }{Left: int64(-3), Right: int64(1)},
				struct{ Left, Right interface{} }{Left: int64(0), Right: int64(1)},
				struct{ Left, Right interface{} }{Left: int64(1), Right: int64(1)},
				struct{ Left, Right interface{} }{Left: int64(2), Right: int64(1)},
				struct{ Left, Right interface{} }{Left: int64(-2), Right: int64(2)},
				struct{ Left, Right interface{} }{Left: int64(-1), Right: int64(3)},
				struct{ Left, Right interface{} }{Left: int64(0), Right: int64(3)},
				struct{ Left, Right interface{} }{Left: int64(1), Right: int64(3)},
			},
			[]interface{}{
				struct{ Left, Right interface{} }{Left: int64(-7), Right: int64(-3)},
				struct{ Left, Right interface{} }{Left: int64(-8), Right: int64(-2)},
			},
			[]interface{}(nil),
		},
	}

	assert.Equal(t, expected, raw)
}
