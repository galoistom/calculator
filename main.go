package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: calculator <filename>")
		return
	}
	input:=""
	switch os.Args[1]{
	case "--help", "-h" :
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
	l := Lexer{input: input}
	p := Parser{tokens: l.GetToken(), position: 0}
	exp, err := p.ParseSequence()
	if err != nil {
		fmt.Println(err)
		return
	}
	env, err := InitEnvironment()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = EvalSequence(exp, env)
	if err != nil {
		fmt.Println(err)
		return
	}
}
