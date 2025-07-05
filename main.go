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

	// Cached computed value
	v Expr
}

type Number struct {
	Value int64
}

type Symbol struct {
	Name string
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

func printExpr(expr Expr) string {
	switch e := expr.(type) {
	case *Number:
		return fmt.Sprintf("%d", e.Value)
	case *Symbol:
		return e.Name
	case *Ap:
		return fmt.Sprintf("ap %s %s", printExpr(e.Left), printExpr(e.Right))
	default:
		return fmt.Sprintf("unknown(%T)", expr)
	}
}

func eval(expr Expr, symbols map[string]Expr) (Expr, error) {
	originalExpr := expr
	for {
		if a, ok := expr.(*Ap); ok && a.v != nil {
			return a.v, nil
		}
		newExpr, err := evalInner2(expr, symbols)
		if err != nil {
			return nil, fmt.Errorf("error evaluating expression: %w", err)
		}
		if expr == newExpr {
			if a, ok := originalExpr.(*Ap); ok {
				if a.v != nil {
					panic("shouldn't be here, should have already returned cached value")
				}
				fmt.Printf("Caching value for %s: %v\n", printExpr(originalExpr), printExpr(newExpr))
				a.v = expr
			}
			return expr, nil
		}
		expr = newExpr
	}
}

func evalInner2(expr Expr, symbols map[string]Expr) (Expr, error) {
	fmt.Printf("Evaluating: %T %v %q\n", expr, expr, printExpr(expr))
	// Walk down the left spine, pushing arguments to app onto a stack and then applying when we reach
	// a function.
	args := []Expr{}
loop:
	for {
		fmt.Printf("Walking down: %T %v %q %v\n", expr, expr, printExpr(expr), args)
		switch e := expr.(type) {
		case *Number:
			if len(args) > 0 {
				return nil, fmt.Errorf("unexpected number %v with args %v", e.Value, args)
			}
			return e, nil
		case *Symbol:
			if val, ok := symbols[e.Name]; ok {
				expr = val
				continue
			}
			switch e.Name {
			case "add":
				if len(args) < 2 {
					return nil, fmt.Errorf("symbol 'add' requires 2 arguments, got %d: %v", len(args), args)
				}
				var err error
				expr, err = twoArgCall(args[0], args[1], symbols, func(left, right *Number) Expr {
					return &Number{Value: left.Value + right.Value}
				})
				if err != nil {
					return nil, fmt.Errorf("error evaluating 'add': %w", err)
				}
				args = args[2:]
				break loop
			case "mul":
				if len(args) < 2 {
					return nil, fmt.Errorf("symbol 'mul' requires 2 arguments, got %d: %v", len(args), args)
				}
				var err error
				expr, err = twoArgCall(args[0], args[1], symbols, func(left, right *Number) Expr {
					return &Number{Value: left.Value * right.Value}
				})
				if err != nil {
					return nil, fmt.Errorf("error evaluating 'mul': %w", err)
				}
				args = args[2:]
				break loop
			case "eq":
				if len(args) < 2 {
					return nil, fmt.Errorf("symbol 'eq' requires 2 arguments, got %d: %v", len(args), args)
				}
				var err error
				expr, err = twoArgCall(args[0], args[1], symbols, func(left, right *Number) Expr {
					if left.Value == right.Value {
						return &Symbol{Name: "t"}
					}
					return &Symbol{Name: "f"}
				})
				if err != nil {
					return nil, fmt.Errorf("error evaluating 'eq': %w", err)
				}
				args = args[2:]
				break loop
			case "s":
				// ap ap ap s x0 x1 x2   =   ap ap x0 x2 ap x1 x2
				if len(args) < 3 {
					return nil, fmt.Errorf("symbol 's' requires 3 arguments, got %d: %v", len(args), args)
				}
				expr = &Ap{Left: &Ap{Left: args[0], Right: args[2]}, Right: &Ap{Left: args[1], Right: args[2]}}
				args = args[3:]
				break loop
			case "c":
				// ap ap ap c x0 x1 x2   =   ap ap x0 x2 x1
				if len(args) < 3 {
					return nil, fmt.Errorf("symbol 'c' requires 3 arguments, got %d: %v", len(args), args)
				}
				expr = &Ap{Left: &Ap{Left: args[0], Right: args[2]}, Right: args[1]}
				args = args[3:]
				break loop
			case "b":
				// ap ap ap b x0 x1 x2   =   ap x0 ap x1 x2
				if len(args) < 3 {
					return nil, fmt.Errorf("symbol 'b' requires 3 arguments, got %d: %v", len(args), args)
				}
				expr = &Ap{Left: args[0], Right: &Ap{Left: args[1], Right: args[2]}}
				args = args[3:]
				break loop
			case "t":
				// ap ap t x0 x1   =   x0
				if len(args) < 2 {
					return nil, fmt.Errorf("symbol 't' requires 2 arguments, got %d: %v", len(args), args)
				}
				expr = args[0]
				args = args[2:]
				break loop
			case "f":
				// ap ap f x0 x1   =   x1
				if len(args) < 2 {
					return nil, fmt.Errorf("symbol 'f' requires 2 arguments, got %d: %v", len(args), args)
				}
				expr = args[1]
				args = args[2:]
				break loop
			default:
				return nil, fmt.Errorf("not yet implemented symbol: %s", e.Name)
			}
		case *Ap:
			// Note: don't eval the right side yet, just push it onto the stack
			args = append([]Expr{e.Right}, args...)
			expr = e.Left
		default:
			return nil, fmt.Errorf("unknown expression type: %T", expr)
		}
	}
	// Re-apply any remaining arguments
	for _, arg := range args {
		expr = &Ap{Left: expr, Right: arg}
	}
	return expr, nil
}

func twoArgCall(x0, x1 Expr, symbols map[string]Expr, f func(left, right *Number) Expr) (Expr, error) {
	v0, err := eval(x0, symbols)
	if err != nil {
		return nil, fmt.Errorf("error evaluating first argument of 'mul': %w", err)
	}
	v1, err := eval(x1, symbols)
	if err != nil {
		return nil, fmt.Errorf("error evaluating second argument of 'mul': %w", err)
	}
	if n1, ok := v0.(*Number); ok {
		if n2, ok := v1.(*Number); ok {
			return f(n1, n2), nil
		} else {
			return nil, fmt.Errorf("expected number for second argument of 'mul', got: %v", v1)
		}
	} else {
		return nil, fmt.Errorf("expected number for first argument of 'mul', got: %v", v0)
	}
}

func evalInner(expr Expr, symbols map[string]Expr) (Expr, error) {
	fmt.Printf("Evaluating: %T %v\n", expr, expr)
	switch e := expr.(type) {
	case *Number:
		return e, nil
	case *Symbol:
		if val, ok := symbols[e.Name]; ok {
			return val, nil
		}
		return e, nil
	case *Ap:
		if e.v != nil {
			return e.v, nil
		}
		left, err := eval(e.Left, symbols)
		if err != nil {
			return nil, fmt.Errorf("error evaluating left side of application: %w", err)
		}
		switch l := left.(type) {
		case *Number:
			return nil, fmt.Errorf("left side of application is a number, expected a function: %v", l)
		case *Symbol:
			return nil, fmt.Errorf("not yet implemeneted handling of 1 arg symbol: %s", l.Name)
		case *Ap:
			switch l2 := l.Left.(type) {
			case *Number:
				return nil, fmt.Errorf("not yet implemeneted handling of application with number: %v", l2)
			case *Symbol:
				return nil, fmt.Errorf("not yet implemeneted handling of 2 arg symbol: %s", l2.Name)
			case *Ap:
				switch l3 := l2.Left.(type) {
				case *Number:
					return nil, fmt.Errorf("not yet implemeneted handling of application with number: %v", l3)
				case *Symbol:
					x0 := l2.Right
					x1 := l.Right
					x2 := e.Right
					switch l3.Name {
					case "s":
						return &Ap{&Ap{x0, x2, nil}, &Ap{x1, x2, nil}, nil}, nil
					case "k":
						return nil, fmt.Errorf("not yet implemeneted handling of symbol: %s", l3.Name)
					case "i":
						return nil, fmt.Errorf("not yet implemeneted handling of symbol: %s", l3.Name)
					default:
						return nil, fmt.Errorf("unknown function: %s", l3.Name)
					}
				case *Ap:
					return nil, fmt.Errorf("unexpected 4 args application: %v", l3)
				default:
					return nil, fmt.Errorf("unknown expression type: %T", l3)
				}
			default:
				return nil, fmt.Errorf("unknown expression type: %T", l2)
			}
		default:
			return nil, fmt.Errorf("unknown expression type: %T", l)
		}

	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
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
