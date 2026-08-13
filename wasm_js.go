//go:build js && wasm

package main

import "syscall/js"

func getExpression() string {
	if expr := js.Global().Get("expression"); !expr.IsUndefined() {
		return expr.String()
	}
	return ""
}
