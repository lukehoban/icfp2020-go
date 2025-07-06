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

func eval(expr Expr, symbols map[Symbol]Expr) (Expr, error) {
	originalExpr := expr
	for {
		if a, ok := expr.(*Ap); ok && a.v != nil {
			return a.v, nil
		}
		newExpr, err := evalInner(expr, symbols)
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

func evalInner(expr Expr, symbols map[Symbol]Expr) (Expr, error) {
	fmt.Printf("Evaluating: %T %v %q\n", expr, expr, printExpr(expr))
	// Walk down the left spine, pushing arguments to app onto a stack and then applying when we reach
	// a function.
	args := []Expr{}
loop:
	for {
		fmt.Printf("Walking down: %T %v %q %v\n", expr, expr, printExpr(expr), args)
		switch e := expr.(type) {
		case Number:
			if len(args) > 0 {
				return nil, fmt.Errorf("unexpected number %v with args %v", e, args)
			}
			return e, nil
		case Symbol:
			if val, ok := symbols[e]; ok {
				expr = val
				continue
			}
			switch e {
			case "add":
				e, err := call("add", args, 2, symbols, func(vals ...Number) Expr {
					return Number(vals[0] + vals[1])
				})
				if err != nil {
					return nil, fmt.Errorf("error evaluating 'add': %w", err)
				}
				expr = e
				args = args[2:]
				break loop
			case "mul":
				e, err := call("mul", args, 2, symbols, func(vals ...Number) Expr {
					return Number(vals[0] * vals[1])
				})
				if err != nil {
					return nil, fmt.Errorf("error evaluating 'mul': %w", err)
				}
				expr = e
				args = args[2:]
				break loop
			case "div":
				e, err := call("div", args, 2, symbols, func(vals ...Number) Expr {
					return Number(vals[0] / vals[1])
				})
				if err != nil {
					return nil, fmt.Errorf("error evaluating 'div': %w", err)
				}
				expr = e
				args = args[2:]
				break loop
			case "eq":
				e, err := call("eq", args, 2, symbols, func(vals ...Number) Expr {
					if vals[0] == vals[1] {
						return Symbol("t")
					}
					return Symbol("f")
				})
				if err != nil {
					return nil, fmt.Errorf("error evaluating 'eq': %w", err)
				}
				expr = e
				args = args[2:]
				break loop
			case "lt":
				e, err := call("lt", args, 2, symbols, func(vals ...Number) Expr {
					if vals[0] < vals[1] {
						return Symbol("t")
					}
					return Symbol("f")
				})
				if err != nil {
					return nil, fmt.Errorf("error evaluating 'lt': %w", err)
				}
				expr = e
				args = args[2:]
				break loop
			case "neg":
				e, err := call("neg", args, 1, symbols, func(vals ...Number) Expr {
					return Number(-vals[0])
				})
				if err != nil {
					return nil, fmt.Errorf("error evaluating 'neg': %w", err)
				}
				expr = e
				args = args[1:]
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
			case "i":
				// ap i x0  =   x0
				if len(args) < 1 {
					return nil, fmt.Errorf("symbol 'i' requires 1 argument, got %d: %v", len(args), args)
				}
				expr = args[0]
				args = args[1:]
				break loop
			case "cons", "vec":
				// ap ap ap cons x0 x1 x2   =   ap ap x2 x0 x1
				if len(args) < 2 {
					return nil, fmt.Errorf("symbol 'cons' requires 2 arguments, got %d: %v", len(args), args)
				} else if len(args) == 2 {
					left, err := eval(args[0], symbols)
					if err != nil {
						return nil, fmt.Errorf("error evaluating left argument of 'cons': %w", err)
					}
					right, err := eval(args[1], symbols)
					if err != nil {
						return nil, fmt.Errorf("error evaluating right argument of 'cons': %w", err)
					}
					cons := &Ap{Left: &Ap{Left: Symbol("cons"), Right: left}, Right: right}
					cons.v = cons
					expr = cons
					args = args[2:]
					break loop
				} else {
					expr = &Ap{Left: &Ap{Left: args[2], Right: args[0]}, Right: args[1]}
					args = args[3:]
					break loop
				}
			case "car":
				if len(args) < 1 {
					return nil, fmt.Errorf("symbol 'car' requires 1 argument, got %d: %v", len(args), args)
				}
				// ap car x0   =   ap x0 t
				expr = &Ap{Left: args[0], Right: Symbol("t")}
				args = args[1:]
				break loop
			case "cdr":
				if len(args) < 1 {
					return nil, fmt.Errorf("symbol 'cdr' requires 1 argument, got %d: %v", len(args), args)
				}
				// ap cdr x0   =   ap x0 f
				expr = &Ap{Left: args[0], Right: Symbol("f")}
				args = args[1:]
				break loop
			case "nil":
				// ap nil x0   =   t
				if len(args) < 1 {
					return e, nil
				}
				expr = Symbol("t")
				args = args[1:]
				break loop
			case "isnil":
				if len(args) < 1 {
					return nil, fmt.Errorf("symbol 'isnil' requires 1 argument, got %d: %v", len(args), args)
				}
				// TODO: Treat cons constuctions as their own type outside of raw functions?
				if s, ok := args[0].(Symbol); ok && s == Symbol("nil") {
					expr = Symbol("t")
				} else {
					expr = Symbol("f")
				}
				args = args[1:]
			default:
				return nil, fmt.Errorf("not yet implemented symbol: %s", e)
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

func call(name string, args []Expr, n int, symbols map[Symbol]Expr, f func(args ...Number) Expr) (Expr, error) {
	if len(args) < n {
		return nil, fmt.Errorf("expected at least %d arguments for '%s', got %d: %v", n, name, len(args), args)
	}
	vals := make([]Number, 0, n)
	for i := range n {
		v, err := eval(args[i], symbols)
		if err != nil {
			return nil, fmt.Errorf("error evaluating argument: %w", err)
		}
		if n, ok := v.(Number); ok {
			vals = append(vals, n)
		} else {
			return nil, fmt.Errorf("expected number argument, got: %v", v)
		}
	}
	return f(vals...), nil
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
