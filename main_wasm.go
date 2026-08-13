//go:build js && wasm

package main

import (
	"syscall/js"
)

var calcEval js.Func

func main() {
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
		res, err := process(args[0].String())
		if err != nil {
			return "error: " + err.Error()
		}
		if res == nil {
			return ""
		}
		return Print(res)
	})
	js.Global().Set("calcEval", calcEval)
	<-make(chan struct{})
}
