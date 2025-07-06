package main

import "fmt"

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
