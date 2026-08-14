//go:build !js || !wasm

package main

import (
	"fmt"
	"os"
)

func main() {
	show = true
	var err error
	environment, err = InitEnvironment()
	if err != nil {
		fmt.Println(err)
		return
	}
	input := getExpression()
	if input == "" {
		return
	}
	_, err = Process(input)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getExpression() string {
	input := ""
	var err error
	if len(os.Args) <= 1{
		fmt.Println("usage: calculator [option] <code>\n",
			"        -r/--repl to enter repl\n",
			"        -f/--file to read code from file\n",
			"        -e/--eval to read from arguments")
		return ""
	}
	switch os.Args[1] {
	case "--file", "-f":
		filepath := os.Args[2]
		input, err = ReadFile(filepath)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		show = false
	case "--eval", "-e":
		input = os.Args[2]
	case "--repl", "-r":
		input = "(display \"type 'exit to exit\") (define (foldr func init l) (if (null? l) init (func (car l) (foldr func init (cdr l))))) (let loop () (let ((x (readline))) (if (eq? x 'exit) (display \"good bye!\")(loop))))"
	default:
		fmt.Println("usage: calculator [option] <code>\n",
			"        -r/--repl to enter repl\n",
			"        -f/--file to read code from file\n",
			"        -e/--eval to read from arguments")
		return ""
	}
	return input
}
