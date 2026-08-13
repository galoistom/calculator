//go:build js && wasm

package main

import (
	"bytes"
	"strings"
	"syscall/js"
)

var calcEval js.Func
var stdoutBuf bytes.Buffer

func getExpression() string {
	if expr := js.Global().Get("expression"); !expr.IsUndefined() {
		return expr.String()
	}
	return ""
}

func main() {
	output = &stdoutBuf
	var err error
	environment, err = InitEnvironment()
	if err != nil {
		panic(err)
	}
	if input := getExpression(); input != "" {
		if _, err := Process(input); err != nil {
			println("error:", err.Error())
		}
	}
	calcEval = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return "error: expected a code string"
		}
		stdoutBuf.Reset()
		res, err := process(args[0].String())
		printed := strings.TrimRight(stdoutBuf.String(), "\n")
		stdoutBuf.Reset()
		if err != nil {
			if printed != "" {
				return printed + "\nerror: " + err.Error()
			}
			return "error: " + err.Error()
		}
		var result string
		if res != nil {
			result = Print(res)
		}
		switch {
		case printed != "" && result != "":
			return printed + "\n" + result
		case printed != "":
			return printed
		default:
			return result
		}
	})
	js.Global().Set("calcEval", calcEval)
	<-make(chan struct{})
}
