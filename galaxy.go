package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Expr interface {
	isExpr()
}

type Ap struct {
	Left  Expr
	Right Expr

	// Cached computed value
	v Expr
}

func (a *Ap) isExpr() {}

type Number int64

func (n Number) isExpr() {}

type Symbol string

func (s Symbol) isExpr() {}

func parseProgram(path string) (map[Symbol]Expr, error) {
	byts, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(byts), "\n")

	symbols := map[Symbol]Expr{}
	for _, line := range lines {
		pieces := strings.Split(line, " = ")
		if len(pieces) != 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		symbols[Symbol(pieces[0])], _ = parseExpr(strings.Split(pieces[1], " "))
	}

	return symbols, nil
}

func parseExpr(terms []string) (Expr, []string) {
	if len(terms) == 0 {
		return nil, terms
	}

	token := terms[0]
	rest := terms[1:]

	switch token {
	case "ap":
		left, rest := parseExpr(rest)
		right, rest := parseExpr(rest)
		return &Ap{Left: left, Right: right}, rest
	default:
		if num, err := strconv.ParseInt(token, 10, 64); err == nil {
			return Number(num), rest
		}
		return Symbol(token), rest
	}
}

func printExpr(expr Expr) string {
	switch e := expr.(type) {
	case Number:
		return fmt.Sprintf("%d", e)
	case Symbol:
		return string(e)
	case *Ap:
		return fmt.Sprintf("ap %s %s", printExpr(e.Left), printExpr(e.Right))
	default:
		return fmt.Sprintf("unknown(%T)", expr)
	}
}

func eval(expr Expr, symbols map[Symbol]Expr) Expr {
	if a, ok := expr.(*Ap); ok && a.v != nil {
		return a.v
	}
	initialExpr := expr
	for {
		result := tryEval(expr, symbols)
		if result == expr {
			if a, ok := initialExpr.(*Ap); ok && a.v == nil {
				a.v = expr
			}
			return result
		}
		expr = result
	}
}

const t = Symbol("t")
const f = Symbol("f")
const cons = Symbol("cons")

func tryEval(expr Expr, symbols map[Symbol]Expr) Expr {
	if a, ok := expr.(*Ap); ok && a.v != nil {
		return a.v
	}
	switch e := expr.(type) {
	case Symbol:
		if val, ok := symbols[e]; ok {
			return val
		}
	case *Ap:
		fun := eval(e.Left, symbols)
		x := e.Right
		switch fun := fun.(type) {
		case Symbol:
			switch fun {
			case "neg":
				return -eval(x, symbols).(Number)
			case "i":
				return x
			case "nil":
				return t
			case "isnil":
				return &Ap{x, &Ap{t, &Ap{t, f, nil}, nil}, nil}
			case "car":
				return &Ap{x, t, nil}
			case "cdr":
				return &Ap{x, f, nil}
			}
		case *Ap:
			fun2 := eval(fun.Left, symbols)
			y := fun.Right
			switch fun2 := fun2.(type) {
			case Symbol:
				switch fun2 {
				case "t":
					return y
				case "f":
					return x
				case "add":
					return eval(x, symbols).(Number) + eval(y, symbols).(Number)
				case "mul":
					return eval(x, symbols).(Number) * eval(y, symbols).(Number)
				case "div":
					return eval(y, symbols).(Number) / eval(x, symbols).(Number)
				case "lt":
					if eval(y, symbols).(Number) < eval(x, symbols).(Number) {
						return t
					}
					return f
				case "eq":
					vx := eval(x, symbols)
					vy := eval(y, symbols)
					if vx.(Number) == vy.(Number) {
						return t
					}
					return f
				case "cons":
					res := &Ap{Left: &Ap{Left: cons, Right: eval(y, symbols)}, Right: eval(x, symbols)}
					res.v = res
					return res
				}
			case *Ap:
				fun3 := eval(fun2.Left, symbols)
				z := fun2.Right
				switch fun3 := fun3.(type) {
				case Symbol:
					switch fun3 {
					case "s":
						return &Ap{Left: &Ap{Left: z, Right: x}, Right: &Ap{Left: y, Right: x}}
					case "c":
						return &Ap{Left: &Ap{Left: z, Right: x}, Right: y}
					case "b":
						return &Ap{Left: z, Right: &Ap{Left: y, Right: x}}
					case "cons":
						return &Ap{Left: &Ap{Left: x, Right: z}, Right: y}
					}
				}
			}
		}
	}
	return expr
}

func toValue(expr Expr) interface{} {
	switch e := expr.(type) {
	case Number:
		return int64(e)
	case Symbol:
		if e == "nil" {
			return []interface{}(nil)
		}
		panic(fmt.Sprintf("unexpected symbol: %s", e))
	case *Ap:
		switch e2 := e.Left.(type) {
		case *Ap:
			switch e3 := e2.Left.(type) {
			case Symbol:
				switch e3 {
				case "cons":
					right := toValue(e.Right)
					left := toValue(e2.Right)
					if rightarr, ok := right.([]interface{}); ok {
						return append([]interface{}{left}, rightarr...)
					} else {
						return struct{ Left, Right interface{} }{Left: left, Right: right}
					}
				default:
					panic(fmt.Sprintf("unexpected Ap.Left: %s", printExpr(e2.Left)))
				}
			default:
				panic(fmt.Sprintf("unexpected Ap.Left: %s", printExpr(e2.Left)))
			}
		default:
			panic(fmt.Sprintf("unexpected Ap.Left: %s", printExpr(e.Left)))
		}
	default:
		panic(fmt.Sprintf("unexpected expr type: %T", expr))
	}
}
