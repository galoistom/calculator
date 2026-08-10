package main

import (
	"fmt"
	"os"
)

var environment *Environment

func Process(code string) (Expr, error) {
	l := Lexer{input: code}
	p := Parser{tokens: l.GetToken(), position: 0}
	exp, err := p.ParseSequence()
	if err != nil {
		return nil, err
	}
	res, err := EvalSequence(exp, environment)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: calculator <filename>")
		return
	}
	input := ""
	switch os.Args[1] {
	case "--help", "-h":
		fmt.Println("usage: calculator <filename>")
		return
	case "--file", "-f":
		filepath := os.Args[2]
		file, err := os.ReadFile(filepath)
		if err != nil {
			fmt.Println(err)
			return
		}
		input = string(file)
	case "--eval", "-e":
		input = os.Args[2]
	}
	var err error
	environment, err = InitEnvironment()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = Process(input)
	if err != nil {
		fmt.Println(err)
		return
	}
}
