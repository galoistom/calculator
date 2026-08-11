package main

import (
	"strings"
	"testing"
)

func parseExpr(t *testing.T, code string) Expr {
	t.Helper()
	l := Lexer{input: code}
	p := Parser{tokens: l.GetToken(), position: 0}
	got, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse(%q) error: %v", code, err)
	}
	return got
}

func parseSequence(t *testing.T, code string) []Expr {
	t.Helper()
	l := Lexer{input: code}
	p := Parser{tokens: l.GetToken(), position: 0}
	got, err := p.ParseSequence()
	if err != nil {
		t.Fatalf("ParseSequence(%q) error: %v", code, err)
	}
	return got
}

func TestParseNumber(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"0", "0"},
		{"42", "42"},
		{"1.5", "1.500000"},
		{"7/4", "7/4"},
	}
	for _, c := range cases {
		got := parseExpr(t, c.in)
		if got.Print() != c.want {
			t.Errorf("Parse(%q).Print() = %q, want %q", c.in, got.Print(), c.want)
		}
		if _, ok := got.(Number); !ok {
			t.Errorf("Parse(%q) type = %T, want Number", c.in, got)
		}
	}
}

func TestParseNumberTypes(t *testing.T) {
	got := parseSequence(t, "1 2.5 3/8")
	if len(got) != 3 {
		t.Fatalf("ParseSequence length = %d, want 3", len(got))
	}
	n0, ok := got[0].(Number)
	if !ok {
		t.Fatalf("got[0] type = %T, want Number", got[0])
	}
	if _, ok := n0.value.(Integer); !ok {
		t.Errorf("1 parsed as %T, want Integer", n0.value)
	}
	n1 := got[1].(Number)
	if _, ok := n1.value.(Real); !ok {
		t.Errorf("2.5 parsed as %T, want Real", n1.value)
	}
	n2 := got[2].(Number)
	if _, ok := n2.value.(Rational); !ok {
		t.Errorf("3/8 parsed as %T, want Rational", n2.value)
	}
}

func TestParseSymbol(t *testing.T) {
	got := parseExpr(t, "foo")
	s, ok := got.(Symbol)
	if !ok || s.content != "foo" {
		t.Errorf("Parse(foo) = %v, want Symbol(foo)", got)
	}
}

func TestParseString(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{`"hello"`, `hello`},
		{`"a b c"`, `a b c`},
		{`""`, ``},
		{`"say \"hi\""`, `say "hi"`},
	}
	for _, c := range cases {
		got := parseExpr(t, c.in)
		if _, ok := got.(String); !ok {
			t.Errorf("Parse(%q) type = %T, want String", c.in, got)
			continue
		}
		if got.Print() != c.want {
			t.Errorf("Parse(%q).Print() = %q, want %q", c.in, got.Print(), c.want)
		}
	}
}

func TestParseStringInList(t *testing.T) {
	got := parseExpr(t, `("hello" world)`)
	lst, ok := got.(List)
	if !ok || len(lst.args) != 2 {
		t.Fatalf("Parse = %v, want a 2-element List", got)
	}
	if _, ok := lst.args[0].(String); !ok {
		t.Errorf("args[0] type = %T, want String", lst.args[0])
	}
	if _, ok := lst.args[1].(Symbol); !ok {
		t.Errorf("args[1] type = %T, want Symbol", lst.args[1])
	}
}

func TestParseUnterminatedString(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("Parse(%q) expected a panic for unterminated string", `"abc`)
		}
	}()
	l := Lexer{input: `"abc`}
	p := Parser{tokens: l.GetToken(), position: 0}
	_, _ = p.Parse()
}

func TestParseList(t *testing.T) {
	got := parseExpr(t, "(+ 1 2)")
	lst, ok := got.(List)
	if !ok {
		t.Fatalf("Parse((+ 1 2)) type = %T, want List", got)
	}
	if len(lst.args) != 3 {
		t.Fatalf("List has %d elements, want 3", len(lst.args))
	}
	if s, ok := lst.args[0].(Symbol); !ok || s.content != "+" {
		t.Errorf("args[0] = %v, want Symbol(+)", lst.args[0])
	}
	if _, ok := lst.args[1].(Number); !ok {
		t.Errorf("args[1] type = %T, want Number", lst.args[1])
	}
}

func TestParseQuote(t *testing.T) {
	got := parseExpr(t, "'(1 2)")
	lst, ok := got.(List)
	if !ok || len(lst.args) != 2 {
		t.Fatalf("Parse('(1 2)) = %v, want a 2-element List", got)
	}
	if s, ok := lst.args[0].(Symbol); !ok || s.content != "quote" {
		t.Errorf("args[0] = %v, want Symbol(quote)", lst.args[0])
	}
	inner, ok := lst.args[1].(List)
	if !ok || len(inner.args) != 2 {
		t.Errorf("quoted body = %v, want List(1 2)", lst.args[1])
	}
}

func TestParseSequence(t *testing.T) {
	got := parseSequence(t, "1 2 3")
	if len(got) != 3 {
		t.Errorf("ParseSequence length = %d, want 3", len(got))
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"( + 1", "( is not closed"},
		{")", "invalid sentences"},
	}
	for _, c := range cases {
		l := Lexer{input: c.in}
		p := Parser{tokens: l.GetToken(), position: 0}
		_, err := p.Parse()
		if err == nil {
			t.Errorf("Parse(%q) expected error containing %q, got none", c.in, c.want)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("Parse(%q) error = %q, want it to contain %q", c.in, err, c.want)
		}
	}
}

func evalCode(t *testing.T, code string) (string, error) {
	t.Helper()
	exps := parseSequence(t, code)
	env, err := InitEnvironment()
	if err != nil {
		t.Fatalf("InitEnvironment error: %v", err)
	}
	res, err := EvalSequence(exps, env)
	if err != nil {
		return "", err
	}
	if res == nil {
		return "", nil
	}
	return res.Print(), nil
}

func mustEval(t *testing.T, code, want string) {
	t.Helper()
	got, err := evalCode(t, code)
	if err != nil {
		t.Fatalf("eval(%q) unexpected error: %v", code, err)
	}
	if got != want {
		t.Errorf("eval(%q) = %q, want %q", code, got, want)
	}
}

func mustError(t *testing.T, code, wantErr string) {
	t.Helper()
	_, err := evalCode(t, code)
	if err == nil {
		t.Fatalf("eval(%q) expected error containing %q, got none", code, wantErr)
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Errorf("eval(%q) error = %q, want it to contain %q", code, err, wantErr)
	}
}

func TestEvalArithmetic(t *testing.T) {
	cases := []struct {
		code, want string
	}{
		{"(+ 1 2)", "3"},
		{"(- 5 3)", "2"},
		{"(* 6 7)", "42"},
		{"(/ 10 4)", "5/2"},
		{"(- 5)", "-5"},
		{"(+ 1 1/2)", "3/2"},
		{"(+ 1 0.5)", "1.500000"},
		{"(+ 1/2 1/4)", "3/4"},
		{"(pow 2 10)", "1024"},
		{"(abs (- 5))", "5"},
		{"(int 7/2)", "3"},
		{"(int 3.7)", "3"},
		{"(c 3 4)", "3.000000+4.000000i"},
		{"(abs (c 3 4))", "5"},
	}
	for _, c := range cases {
		mustEval(t, c.code, c.want)
	}
}

func TestEvalListAndQuote(t *testing.T) {
	mustEval(t, "(list 1 2 3)", "(listof 1 2 3)")
	mustEval(t, "(car (quote (1 2 3)))", "1")
	mustEval(t, "(cdr (quote (1 2 3)))", "(listof 2 3)")
}

func TestEvalConditionals(t *testing.T) {
	mustEval(t, "(if (= 1 1) 10 20)", "10")
	mustEval(t, "(if (= 1 2) 10 20)", "20")
	mustEval(t, "(cond ((= 1 2) 10) ((= 2 2) 20) (else 30))", "20")
	mustEval(t, "(and 1 2)", "'true")
	mustEval(t, "(and 0 2)", "'false")
	mustEval(t, "(or 0 0)", "'false")
	mustEval(t, "(or 0 3)", "'true")
}

func TestEvalLambda(t *testing.T) {
	mustEval(t, "((lambda (x) (+ x 1)) 41)", "42")
	mustEval(t, "((lambda (a b) (+ a b)) 2 3)", "5")
	mustEval(t, "(begin 1 2 3)", "3")
	mustEval(t, "(let ((a 3) (b 4)) (+ a b))", "7")
}

func TestEvalIdentityReturnsThunk(t *testing.T) {
	exps := parseSequence(t, "((lambda (x) x) 5)")
	env, err := InitEnvironment()
	if err != nil {
		t.Fatalf("InitEnvironment error: %v", err)
	}
	res, err := EvalSequence(exps, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lst, ok := res.(*List)
	if !ok {
		t.Fatalf("identity applied to 5 returned %T, want an unevaluated thunk", res)
	}
	if s, ok := lst.args[0].(Symbol); !ok || s.content != "thunk" {
		t.Errorf("thunk head = %v, want Symbol(thunk)", lst.args[0])
	}
}

func TestEvalDefineAndRecursion(t *testing.T) {
	mustEval(t, "(define x 5) x", "5")
	mustEval(t, "(define x 5) (define y 3) (+ x y)", "8")
	mustEval(t, "(define (fact n) (if (= n 0) 1 (* n (fact (- n 1))))) (fact 10)", "3628800")
	mustEval(t, "(define (make-adder n) (lambda (x) (+ x n))) ((make-adder 3) 4)", "7")
}

func TestEvalErrors(t *testing.T) {
	mustError(t, "foo", "unable to find variable")
	mustError(t, "(+ 1 2 3)", "args legth incorrect to call +")
	mustError(t, "(/ 1 0)", "divisor is 0")
	mustError(t, "(if 1 2)", "wrong number of args to call if")
	mustError(t, "(let ((a)) a)", "Wrong numbers of arguments in let")
}

func TestLazyArguments(t *testing.T) {
	mustEval(t, "((lambda (x) 5) (/ 1 0))", "5")
	mustError(t, "((lambda (x) (+ x 1)) (/ 1 0))", "divisor is 0")
}

func TestLetBindingsAreValuesNotReevaluated(t *testing.T) {
	mustEval(t, "(let ((x (quote exit))) (eq? x (quote exit)))", "'true")
	mustEval(t, "(let ((x (quote (1 2)))) (car x))", "1")
	mustEval(t, "(let ((a 3) (b 4)) (+ a b))", "7")
}

func evalWithCounter(t *testing.T, code string, count *int) (string, error) {
	t.Helper()
	exps := parseSequence(t, code)
	env, err := InitEnvironment()
	if err != nil {
		t.Fatalf("InitEnvironment error: %v", err)
	}
	action := Action{
		name: "count",
		f: func(args []Expr) (Expr, error) {
			*count++
			return Number{value: Integer(1)}, nil
		},
	}
	newEnv, err := env.extend_environment(
		[]Expr{Symbol{content: "count"}},
		[]Expr{List{args: []Expr{Symbol{content: "primitive"}, action}}},
	)
	if err != nil {
		t.Fatalf("extend_environment error: %v", err)
	}
	res, err := EvalSequence(exps, &newEnv)
	if err != nil {
		return "", err
	}
	if res == nil {
		return "", nil
	}
	return res.Print(), nil
}

func TestLazyMemoization(t *testing.T) {
	count := 0
	got, err := evalWithCounter(t, "((lambda (x) (+ x x)) (count))", &count)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "2" {
		t.Errorf("eval = %q, want 2", got)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (thunk should be forced exactly once)", count)
	}

	count = 0
	got, err = evalWithCounter(t, "((lambda (x) 5) (count))", &count)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "5" {
		t.Errorf("eval = %q, want 5", got)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0 (unused argument must not be forced)", count)
	}
}
