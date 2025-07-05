package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPowerOf2(t *testing.T) {
	symbols := builtins()

	pwr2 := "ap ap s ap ap c ap eq 0 1 ap ap b ap mul 2 ap ap b pwr2 ap add -1"
	e, _ := parseExpr(strings.Split(pwr2, " "))
	pwr2Val, err := eval(e, symbols)
	assert.NoError(t, err)

	symbols["pwr2"] = pwr2Val.(Fun)

	expr, _ := parseExpr([]string{"ap", "pwr2", "0"})
	v, err := eval(expr, symbols)
	assert.NoError(t, err)
	assert.Equal(t, Num(1), v)
}
