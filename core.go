package main

import (
	"fmt"
	"os"
)

var environment *Environment
var show bool

func process(code string) (Expr, error) {
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
	if l, ok := res.(*Thunk); ok {
		res, err = forceIt(l)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func Process(code string) (Expr, error) {
	res, err := process(code)
	if err != nil {
		return nil, err
	}
	fmt.Println(Print(res))
	return res, nil
}

func ReadFile(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(file), nil
}
