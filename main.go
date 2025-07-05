package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Expr interface {
}

type Ap struct {
	Left  Expr
	Right Expr
}

type Number struct {
	Value int64
}

type Symbol struct {
	Name string
}

type Value interface {
	Call(arg Value) Value
}

type Num int64

func (n Num) Call(arg Value) Value {
	panic(fmt.Sprintf("call: expected function, got number %d", n))
}

type Fun func(arg Value) Value

func (f Fun) Call(arg Value) Value {
	return f(arg)
}

// Assign "nil" after builtins is initialized
func builtins() map[string]Fun {
	builtins := map[string]Fun{}

	// 9
	builtins["mul"] = Fun(func(x0 Value) Value {
		return Fun(func(x1 Value) Value {
			if n0, ok := x0.(Num); ok {
				if n1, ok := x1.(Num); ok {
					return n0 * n1
				}
			}
			panic(fmt.Sprintf("mul: expected two numbers, got %T and %T", x0, x1))
		})
	})

	// 11
	builtins["eq"] = Fun(func(x0 Value) Value {
		return Fun(func(x1 Value) Value {
			if n0, ok := x0.(Num); ok {
				if n1, ok := x1.(Num); ok {
					if n0 == n1 {
						return builtins["t"]
					}
					return builtins["f"]
				}
			}
			panic(fmt.Sprintf("eq: expected two numbers, got %T and %T", x0, x1))
		})
	})

	// 18
	builtins["s"] = Fun(func(x0 Value) Value {
		return Fun(func(x1 Value) Value {
			return Fun(func(x2 Value) Value {
				return x0.Call(x2).Call(x1.Call(x2))
			})
		})
	})

	// 19
	builtins["c"] = Fun(func(x0 Value) Value {
		return Fun(func(x1 Value) Value {
			return Fun(func(x2 Value) Value {
				return x0.Call(x2).Call(x1)
			})
		})
	})

	// 20
	builtins["b"] = Fun(func(x0 Value) Value {
		return Fun(func(x1 Value) Value {
			return Fun(func(x2 Value) Value {
				return x0.Call(x1.Call(x2))
			})
		})
	})

	// 21
	builtins["t"] = Fun(func(x0 Value) Value {
		return Fun(func(x1 Value) Value {
			return x0
		})
	})
	builtins["cons"] = Fun(func(x0 Value) Value {
		return Fun(func(x1 Value) Value {
			return Fun(func(x2 Value) Value {
				return x2.Call(x0).Call(x1)
			})
		})
	})
	builtins["nil"] = Fun(func(x0 Value) Value {
		return builtins["t"]
	})
	return builtins
}

func parseProgram(path string) (map[string]Expr, error) {
	byts, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(byts), "\n")

	symbols := map[string]Expr{}
	for _, line := range lines {
		pieces := strings.Split(line, " = ")
		if len(pieces) != 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		symbols[pieces[0]], _ = parseExpr(strings.Split(pieces[1], " "))
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
			return &Number{Value: num}, rest
		}
		return &Symbol{Name: terms[0]}, rest
	}
}

func eval(expr Expr, symbols map[string]Fun) (Value, error) {
	switch e := expr.(type) {
	case *Ap:
		left, err := eval(e.Left, symbols)
		if err != nil {
			return nil, fmt.Errorf("error evaluating left expression: %w", err)
		}
		right, err := eval(e.Right, symbols)
		if err != nil {
			return nil, fmt.Errorf("error evaluating right expression: %w", err)
		}
		if fun, ok := left.(Fun); ok {
			return fun(right), nil
		}
		return nil, fmt.Errorf("left expression is not a function: %T", left)
	case *Number:
		return Num(e.Value), nil
	case *Symbol:
		if fun, ok := symbols[e.Name]; ok {
			return fun, nil
		}
		return nil, fmt.Errorf("undefined symbol: %s", e.Name)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", e)
	}
}

func doit() error {
	program, err := parseProgram("./galaxy.txt")
	if err != nil {
		return fmt.Errorf("failed to parse program: %w", err)
	}
	fmt.Printf("%v", program)
	return nil
}

func main() {
	err := doit()
	if err != nil {
		fmt.Print(err.Error())
	}
}
