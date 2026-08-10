package main

import (
	"testing"
)

func TestParse(t *testing.T) {
	input := "(pow (abs (- 3 '(c_1 0 1))) set! 2)"
	l := Lexer{input: input}
	p := Parser{tokens: l.GetToken(), position: 0}
	exp, err := p.Parse()
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(exp)
}

func TestEval(t *testing.T) {
	input := "1\n"
	l := Lexer{input: input}
	p := Parser{tokens: l.GetToken(), position: 0}
	exp, err := p.Parse()
	if err != nil {
		t.Log(err)
		return
	}
	env, err := InitEnvironment()
	if err != nil {
		t.Log(err)
		return
	}
	res, err := Eval(exp, env)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(res)
}
