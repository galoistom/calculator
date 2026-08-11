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
	if res != nil {
		fmt.Println(res.Print())
	}

	return res, nil
}

func ReadFile(path string) (string, error){
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(file),nil
}

func main() {
	input := ""
	var err error
	switch os.Args[1] {
	case "--help", "-h":
		fmt.Println("usage: calculator [option] <code>\n      -f/--file to read code from file\n      -e/--eval to read from arguments")
		return
	case "--file", "-f":
		filepath := os.Args[2]
		input,err=ReadFile(filepath)
		if err!=nil{fmt.Println(err);return}
	case "--eval", "-e":
		input = os.Args[2]
	}
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
