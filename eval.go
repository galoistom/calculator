package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

var keywords = map[string]bool{
	"quote": true, "quasiquote": true, "unquote": true, "unquote-splicing": true,
	"set!": true, "define": true, "let": true, "if": true, "and": true, "or": true,
	"lambda": true, "begin": true, "defmacro": true, "cond": true,
	"map": true, "for-each": true, "filter": true, "fold": true, "eval": true, "apply": true,
}

func isKeyword(name Symbol) bool {
	return keywords[name.content]
}

func make_frame(vars []Expr, vals []Expr) (frame, error) {
	res := make(frame)
	for i := range vars {
		if j, ok := vars[i].(Symbol); ok {
			res[j.content] = vals[i]
		} else {
			fmt.Println(vars[i].Print())
			return nil, errors.New("frame should only contain symbols")
		}
	}
	return res, nil
}

func (env Environment) extend_environment(vars []Expr, vals []Expr) (Environment, error) {
	if len(vars) == len(vals) {
		res, err := make_frame(vars, vals)
		if err != nil {
			return Environment{}, err
		}
		env.env = append([]frame{res}, env.env...)
		return env, nil
	}
	return Environment{},
		fmt.Errorf("extend_environment: number of variables (%d) does not match number of values (%d)", len(vars), len(vals))
}

func (Number) exprNode() {}
func (n Number) Print() string {
	return Reduce(n.value).Print()
}
func (x stringNumber) getValue() (Number, error) {
	num := x.Value
	if strings.Contains(num, ".") {
		f, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return Number{}, fmt.Errorf("failed to convert to float: %v", err)
		}
		return Number{Reduce(Real(f))}, nil
	}
	if strings.Contains(num, "/") {
		parts := strings.Split(num, "/")
		if len(parts) != 2 {
			return Number{}, fmt.Errorf("invalid quotient: %s", num)
		}
		p, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return Number{}, err
		}

		q, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return Number{}, err
		}
		return Number{Reduce(Rational{p: int(p), q: int(q)})}, nil
	}
	i, err := strconv.Atoi(num)
	if err != nil {
		return Number{}, err
	}
	return Number{value: Integer(i)}, nil
}

func (String) exprNode() {}
func (s String) Print() string {
	return fmt.Sprintf("%s", s.content)
}

func (Symbol) exprNode() {}
func (s Symbol) Print() string {
	return fmt.Sprintf("'%s", s.content)
}

func (List) exprNode() {}
func (l List) Print() string {
	var b strings.Builder
	b.WriteString("(listof")
	for _, i := range l.args {
		fmt.Fprintf(&b, " %s", i.Print())
	}
	b.WriteString(")")
	return b.String()
}

func (Action) exprNode() {}
func (f Action) Print() string {
	return f.name
}

func (splice) exprNode() {}
func (splice) Print() string {
	return "(splice)"
}

func (Procedure) exprNode() {}
func (Procedure) Print() string {
	return "(Procedure)"
}

func (Hash) exprNode() {}
func (h Hash) Print() string {
	var b strings.Builder
	b.WriteString("(hash:")
	for i, j := range h.hash {
		fmt.Fprintf(&b, " (%s: %s)", i, j.Print())
	}
	b.WriteString(")")
	return b.String()
}
func (h *Hash) set(vari Expr, val Expr) {
	h.hash[vari.Print()] = val
}
func (h Hash) ref(vari Expr) (Expr, bool) {
	if e, ok := h.hash[vari.Print()]; ok {
		return e, true
	}
	return nil, false
}

func (Thunk) exprNode() {}
func (t Thunk) Print() string {
	return fmt.Sprintf("(thunk: %v (%s))", t.thunk, t.exp.Print())
}

func (Macro) exprNode() {}
func (m Macro) Print() string {
	return fmt.Sprintf("(Macro %s)", m.name)
}

func Print(e Expr) string {
	if e == nil {
		return "Expr:nil"
	}
	return e.Print()
}

func InitEnvironment() (*Environment, error) {
	env := Environment{[]frame{{
		"null":  nil,
		"true":  Symbol{"true"},
		"false": Symbol{"false"},
		"Pi":    Number{Real(math.Pi)},
		"E":     Number{Real(math.E)}}}}
	vals := []Expr{}
	vars := []Expr{}
	var inputScanner = bufio.NewScanner(os.Stdin)
	var primitiveAction = map[string]Action{
		"make-hash": {name: "make-hash",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 0 {
					return &Hash{make(map[string]Expr)}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of make-hash, need 0, get %d", len(args))
			}},
		"hash-ref": {name: "hash-ref",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					if a, ok := args[0].(*Hash).ref(args[1]); ok {
						return a, nil
					}
					return List{}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of hash-ref, need 2, get %d", len(args))
			}},
		"hash-ref!": {name: "hash-ref!",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 3 {
					h := args[0].(*Hash)
					if a, ok := h.ref(args[1]); ok {
						return a, nil
					}
					h.set(args[1], args[2])
					return Symbol{"ok"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of hash-ref!, need 3, get %d", len(args))
			}},
		"hash-set!": {name: "hash-set!",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 3 {
					args[0].(*Hash).set(args[1], args[2])
					return Symbol{"ok"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of hash-set!, need 3, get %d", len(args))
			}},
		"car": {name: "car",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					return args[0].(List).args[0], nil
				}
				return nil, fmt.Errorf("wrong number of arguments of car, need 1, get %d", len(args))
			}},
		"cdr": {name: "cdr",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					return List{args[0].(List).args[1:]}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of cdr, need 1, get %d", len(args))
			}},
		"display": {name: "display",
			f: func(args []Expr) (Expr, error) {
				for _, i := range args {
					if i != nil {
						fmt.Fprint(output, Print(i))
					}
				}
				return nil, nil
			}},
		"displayln": {name: "displayln",
			f: func(args []Expr) (Expr, error) {
				for _, i := range args {
					if i != nil {
						fmt.Fprintln(output, Print(i))
					}
				}
				return nil, nil
			}},
		"format": {name: "format",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					return String{Print(args[0])}, nil
				}
				return nil, nil
			}},
		"exit": {name: "exit",
			f: func(args []Expr) (Expr, error) {
				fmt.Println("Exited")
				os.Exit(0)
				return nil, nil
			}},
		"load": {name: "load",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					file := args[0].(String).content
					code, err := ReadFile(file)
					if err != nil {
						return nil, err
					}
					fmt.Fprintf(output, "file %s loaded successfully\n", file)
					return Process(code)
				}
				return nil, fmt.Errorf("wrong number of arguments of load, need 1, get %d", len(args))
			}},
		"error": {name: "error",
			f: func(args []Expr) (Expr, error) {
				b := strings.Builder{}
				b.WriteString("error:")
				for _, a := range args {
					fmt.Fprintf(&b, " %s", Print(a))
				}
				return nil, errors.New(b.String())
			}},
		"newline": {name: "newline",
			f: func(args []Expr) (Expr, error) {
				fmt.Fprintln(output)
				return nil, nil
			}},
		"pair?": {name: "pair?",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					if _, ok := args[0].(List); ok {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of pair?, need 1, get %d", len(args))
			}},
		"number?": {name: "number?",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					if _, ok := args[0].(Number); ok {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of number?, need 1, get %d", len(args))
			}},
		"string?": {name: "string?",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					if _, ok := args[0].(String); ok {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of string?, need 1, get %d", len(args))
			}},
		"symbol?": {name: "symbol?",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					if _, ok := args[0].(Symbol); ok {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of pair?, need 1, get %d", len(args))
			}},
		"eq?": {name: "eq?",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					if args[0] == args[1] {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of eq?, need 2, get %d", len(args))
			}},
		"list": {name: "list",
			f: func(args []Expr) (Expr, error) {
				return List{args}, nil
			}},
		"length": {name: "length",
			f: func(args []Expr) (Expr, error) {
				if list, ok := args[0].(List); ok {
					return Number{Integer(len(list.args))}, nil
				}
				return nil, fmt.Errorf("list-len expects a list, got %T", args[0])
			}},
		"list-tail": {name: "list-tail",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					list := args[0].(List)
					num := args[1]
					index, err := Int(num.(Number).value)
					if err != nil {
						return nil, err
					}
					return List{list.args[index:]}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of list-tail, need 2, get %d", len(args))
			}},
		"list-set!": {name: "list-set!",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 3 {
					list := args[0].(List)
					num := args[1]
					index, err := Int(num.(Number).value)
					if err != nil {
						return nil, err
					}
					res := args[2]
					list.args[int(index)] = res
					return list, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of list-set!, need 3, get %d", len(args))
			}},
		"list-ref": {name: "list-ref",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					list := args[0].(List)
					num := args[1]
					index, err := Int(num.(Number).value)
					if err != nil {
						return nil, err
					}
					return list.args[int(index)], nil
				}
				return nil, fmt.Errorf("wrong number of arguments of list-ref, need 2, get %d", len(args))
			}},
		"null?": {name: "null?",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					if args[0] == nil {
						return Symbol{"true"}, nil
					}
					if l, ok := args[0].(List); ok {
						if len(l.args) == 0 {
							return Symbol{"true"}, nil
						}
						return Symbol{"false"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of null?, need 1, get %d", len(args))
			}},
		"cons": {name: "cons",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					y := []Expr{args[1]}
					if lst, ok := args[1].(List); ok {
						y = lst.args
					}
					return List{append([]Expr{args[0]}, y...)}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of cons, need 2, get %d", len(args))
			}},
		"readline": {name: "readline",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 0 {
					fmt.Fprint(output, "input:")
					if inputScanner.Scan() {
						line := inputScanner.Text()
						return Process(line)
					}
					if err := inputScanner.Err(); err != nil {
						fmt.Fprintln(output, "failed to read", err)
					}
				}
				return nil, fmt.Errorf("readline expects no arguments, got %d", len(args))
			}},
		"readline-raw": {name: "readline-raw",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 0 {
					fmt.Fprint(output, "input:")
					if inputScanner.Scan() {
						line := inputScanner.Text()
						l := Lexer{input: line}
						p := Parser{tokens: l.GetToken(), position: 0}
						exp, err := p.Parse()
						if err != nil {
							return nil, err
						}
						return exp, nil
					}
					if err := inputScanner.Err(); err != nil {
						fmt.Fprintln(output, "failed to read", err)
					}
				}
				return nil, fmt.Errorf("readline expects no arguments, got %d", len(args))
			}},
		"not": {name: "not",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					t, err := evalTrue(args[0])
					if err != nil {
						return nil, err
					}
					if t {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("readline expects no arguments, got %d", len(args))
			}},
		"+": {name: "+",
			f: func(args []Expr) (Expr, error) {
				switch args[0].(type) {
				case Number:
					num := Value(Integer(0))
					for _, k := range args {
						num = Add(k.(Number).value, num)
					}
					return Number{num}, nil
				case String:
					var b strings.Builder
					for _, k := range args {
						fmt.Fprintf(&b, "%s", k.(String).content)
					}
					return String{b.String()}, nil
				}
				return nil, fmt.Errorf("cannot apply + to %T, expected numbers or strings", args[0])
			}},
		"-": {name: "-",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					return Number{Minu(a, b)}, nil
				}
				if len(args) == 1 {
					a := args[0].(Number).value
					return Number{Minu(Integer(0), a)}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of -, need 1 or 2, get %d", len(args))
			}},
		"*": {name: "*",
			f: func(args []Expr) (Expr, error) {
				num := Value(Integer(1))
				for _, k := range args {
					num = Times(k.(Number).value, num)
				}
				return Number{num}, nil
			}},
		"/": {name: "/",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					res, err := Div(a, b)
					return Number{res}, err
				}
				return nil, fmt.Errorf("wrong number of arguments of /, need 2, get %d", len(args))
			}},
		"=": {name: "=",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					if IsZero(Minu(a, b)) {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of =, need 2, get %d", len(args))
			}},
		"<": {name: "<",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					if a.Type() == ComplexType || b.Type() == ComplexType {
						return nil, errors.New("complex numbers cannot be compared with <")
					}
					res := Add(Real(0), Minu(a, b))
					if res.(Real) < 0 {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of <, need 2, get %d", len(args))
			}},
		"<=": {name: "<=",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					if a.Type() == ComplexType || b.Type() == ComplexType {
						return nil, errors.New("complex numbers cannot be compared with <")
					}
					res := Add(Real(0), Minu(a, b))
					if res.(Real) <= 0 {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of <, need 2, get %d", len(args))
			}},
		">": {name: ">",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					if a.Type() == ComplexType || b.Type() == ComplexType {
						return nil, errors.New("complex numbers cannot be compared with >")
					}
					res := Add(Real(0), Minu(a, b))
					if res.(Real) > 0 {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of >, need 2, get %d", len(args))
			}},
		">=": {name: ">=",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					if a.Type() == ComplexType || b.Type() == ComplexType {
						return nil, errors.New("complex numbers cannot be compared with >")
					}
					res := Add(Real(0), Minu(a, b))
					if res.(Real) >= 0 {
						return Symbol{"true"}, nil
					}
					return Symbol{"false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of >, need 2, get %d", len(args))
			}},
		"pow": {name: "pow",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					x, ok := b.(Integer)
					if !ok {
						return nil, errors.New("pow with a non-integer exponent is not implemented yet")
					}
					return Number{Power(a, int(x))}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of pow, need 2, get %d", len(args))
			}},
		"c": {name: "c",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					return Number{Add(a, Times(Complex{a: 0, b: 1}, b))}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of complex, need 2, get %d", len(args))
			}},
		"abs": {name: "abs",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					a := args[0].(Number).value
					return Number{Abs(a)}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of abs, need 1, get %d", len(args))
			}},
		"sin": {name: "sin",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					a := args[0].(Number).value
					if a.Type() <= ComplexType {
						t := Add(a, Real(0)).(Real)
						return Number{Reduce(Real(math.Sin(float64(t))))}, nil
					}
					return nil, errors.New("complex sin not implemented")
				}
				return nil, fmt.Errorf("wrong number of arguments of sin, need 1, get %d", len(args))
			}},
		"arcsin": {name: "arcsin",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					a := args[0].(Number).value
					if a.Type() <= ComplexType {
						t := Add(a, Real(0)).(Real)
						return Number{Reduce(Real(math.Asin(float64(t))))}, nil
					}
					return nil, errors.New("complex sin not implemented")
				}
				return nil, fmt.Errorf("wrong number of arguments of arcsin, need 1, get %d", len(args))
			}},
		"cos": {name: "cos",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					a := args[0].(Number).value
					if a.Type() <= ComplexType {
						t := Add(a, Real(0)).(Real)
						return Number{Reduce(Real(math.Cos(float64(t))))}, nil
					}
					return nil, errors.New("complex cos not implemented")
				}
				return nil, fmt.Errorf("wrong number of arguments of cos, need 1, get %d", len(args))
			}},
		"arccos": {name: "arccos",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					a := args[0].(Number).value
					if a.Type() <= ComplexType {
						t := Add(a, Real(0)).(Real)
						return Number{Reduce(Real(math.Acos(float64(t))))}, nil
					}
					return nil, errors.New("complex cos not implemented")
				}
				return nil, fmt.Errorf("wrong number of arguments of arccos, need 1, get %d", len(args))
			}},
		"ln": {name: "ln",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					a := args[0].(Number).value
					if a.Type() <= RealType {
						t := Add(a, Real(0)).(Real)
						return Number{Reduce(Real(math.Log(float64(t))))}, nil
					}
					return nil, errors.New("complex cos not implemented")
				}
				return nil, fmt.Errorf("wrong number of arguments of ln, need 1, get %d", len(args))
			}},
		"exp": {name: "exp",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					a := args[0].(Number).value
					if a.Type() <= RealType {
						t := Add(a, Real(0)).(Real)
						return Number{Reduce(Real(math.Pow(math.E, float64(t))))}, nil
					} else {
						za := a.(Complex).a
						zb := a.(Complex).b
						p := math.Pow(math.E, za) * math.Cos(zb)
						q := math.Pow(math.E, za) * math.Sin(zb)
						return Number{Reduce(Complex{p, q})}, nil
					}
				}
				return nil, fmt.Errorf("wrong number of arguments of exp, need 1, get %d", len(args))
			}},
		"int": {name: "int",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					a := args[0].(Number).value
					res, err := Int(a)
					return Number{value: res}, err
				}
				return nil, fmt.Errorf("wrong number of arguments of int, need 1, get %d", len(args))
			}},
	}
	for _, a := range composedAccessors() {
		primitiveAction[a.name] = a
	}

	for i := range primitiveAction {
		vars = append(vars, Symbol{i})
		vals = append(vals, List{[]Expr{Symbol{"primitive"}, primitiveAction[i]}})
	}
	new_env, err := env.extend_environment(vars, vals)
	if err != nil {
		return nil, err
	}
	return &new_env, nil
}

func makeAccessor(ops string) Action {
	name := "c" + ops + "r"
	return Action{
		name: name,
		f: func(args []Expr) (Expr, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("wrong number of arguments of %s, need 1, get %d", name, len(args))
			}
			var cur Expr = args[0]
			for i := len(ops) - 1; i >= 0; i-- {
				l, ok := cur.(List)
				if !ok {
					return nil, fmt.Errorf("%s: not a list", name)
				}
				if ops[i] == 'a' {
					if len(l.args) == 0 {
						return nil, fmt.Errorf("%s: empty list", name)
					}
					cur = l.args[0]
				} else {
					cur = List{l.args[1:]}
				}
			}
			return cur, nil
		},
	}
}

func composedAccessors() []Action {
	res := []Action{}
	for depth := 2; depth <= 4; depth++ {
		for mask := 0; mask < (1 << depth); mask++ {
			ops := make([]byte, depth)
			for i := 0; i < depth; i++ {
				if mask&(1<<i) != 0 {
					ops[i] = 'a'
				} else {
					ops[i] = 'd'
				}
			}
			res = append(res, makeAccessor(string(ops)))
		}
	}
	return res
}

func lookUpVariable(exp Symbol, env *Environment) (Expr, error) {
	for _, frame := range env.env {
		res, ok := frame[exp.content]
		if ok {
			return res, nil
		}
	}
	return nil, fmt.Errorf("unbound variable: %s", Print(exp))
}

func evalAssignment(exp []Expr, env *Environment) error {
	if len(exp) != 2 {
		return fmt.Errorf("wrong number of arguments of set!, need 2, get %d", len(exp))
	}
	target, ok := exp[0].(Symbol)
	if !ok {
		return errors.New("set! requires a variable name")
	}
	for _, j := range (*env).env {
		_, a := j[target.content]
		if a {
			res, err := Eval(exp[1], env)
			if err != nil {
				return err
			}
			j[target.content] = res
			return nil
		}
	}
	return fmt.Errorf("unbound variable: %s", Print(target))
}

func definitionVariable(exp []Expr) (Symbol, error) {
	switch x := exp[0].(type) {
	case Symbol:
		return x, nil
	case List:
		switch y := x.args[0].(type) {
		case Symbol:
			return y, nil
		default:
			return Symbol{}, errors.New("define: the name of a procedure definition must be a symbol")
		}
	default:
		return Symbol{}, errors.New("define: invalid definition")
	}
}

func makeLambda(par []Expr, body []Expr) List {
	return List{append([]Expr{Symbol{"lambda"}, List{par}}, body...)}
}

func definitionValue(exp []Expr) Expr {
	switch x := exp[0].(type) {
	case Symbol:
		return exp[1]
	case List:
		return makeLambda(x.args[1:], exp[1:])
	}
	return nil
}

func evalDefinition(exp []Expr, env *Environment) error {
	res, err := definitionVariable(exp)
	if err != nil {
		return err
	}
	val, err := Eval(definitionValue(exp), env)
	if err != nil {
		return err
	}
	(*env).env[0][res.content] = val
	return nil
}

func sequenceToExp(exp []Expr) Expr {
	switch len(exp) {
	case 0:
		return nil
	case 1:
		return exp[0]
	default:
		return List{append([]Expr{Symbol{"begin"}}, exp...)}
	}
}

func splitVarVal(exp []Expr) ([]Expr, []Expr, error) {
	variables := []Expr{}
	values := []Expr{}
	for _, i := range exp {
		switch j := i.(type) {
		case List:
			if len(j.args) != 2 {
				return nil, nil, errors.New("let: each binding must be a (name value) pair")
			}
			values = append(values, j.args[1])
			switch k := j.args[0].(type) {
			case Symbol:
				variables = append(variables, k)
				continue
			}
		default:
			return nil, nil, errors.New("let: binding must be a list of (name value)")
		}
	}
	return variables, values, nil
}

func evalLet(exp []Expr, env *Environment) (Expr, error) {
	if len(exp) <= 1 {
		return nil, errors.New("let: expected (let <bindings> <body>...)")
	}
	switch x := exp[0].(type) {
	case Symbol:
		ass := exp[1]
		switch assignments := ass.(type) {
		case List:
			variables, values, err := splitVarVal(assignments.args)
			if err != nil {
				return nil, err
			}
			function := List{append([]Expr{Symbol{"define"},
				List{append([]Expr{x}, variables...)}}, exp[2:]...)}
			procedure := sequenceToExp([]Expr{function, List{append([]Expr{x}, values...)}})
			new_env, err := env.extend_environment([]Expr{}, []Expr{})
			return Eval(procedure, &new_env)
		}
		return nil, errors.New("let: expected a list of bindings")
	case List:
		variables, values, err := splitVarVal(x.args)
		if err != nil {
			return nil, err
		}
		lambda := List{append(
			[]Expr{List{append(
				[]Expr{Symbol{"lambda"}, List{variables}},
				exp[1:]...)}},
			values...)}
		return Eval(lambda, env)
	default:
		return nil, errors.New("let: bindings must be a list")
	}
}

func letStripStar(exps []Expr) (Expr, error) {
	ass := exps[0].(List)
	if len(ass.args) == 0 {
		return List{append([]Expr{Symbol{"begin"}}, exps[1:]...)}, nil
	}
	rest, err := letStripStar(append([]Expr{
		List{ass.args[1:]}},
		exps[1:]...))
	if err != nil {
		return nil, err
	}
	return List{[]Expr{
		Symbol{"let"},
		List{[]Expr{List{ass.args[0].(List).args}}},
		rest}}, nil
}

func evalLetStar(exps []Expr, env *Environment) (Expr, error) {
	exp, err := letStripStar(exps)
	if err != nil {
		return nil, err
	}
	return Eval(exp, env)
}

func evalTrue(exp Expr) (bool, error) {
	switch x := exp.(type) {
	case Number:
		return !IsZero(x.value), nil
	case Symbol:
		return !(x.content == "false"), nil
	case List:
		return !(len(x.args) == 0), nil
	}
	return false, fmt.Errorf("cannot use %s as a boolean", Print(exp))
}

func evalIf(exp []Expr, env *Environment) (Expr, error) {
	if len(exp) != 3 {
		return nil, fmt.Errorf("wrong number of arguments of if, need 3, get %d", len(exp))
	}
	condition, err := Eval(exp[0], env)
	if err != nil {
		return nil, err
	}
	c, err := evalTrue(condition)
	if err != nil {
		return nil, err
	}
	if c {
		return Eval(exp[1], env)
	} else {
		return Eval(exp[2], env)
	}
}

func evalOr(exp []Expr, env *Environment) (Expr, error) {
	if len(exp) == 0 {
		return nil, errors.New("or expects at least one argument")
	}
	for _, e := range exp {
		res, err := Eval(e, env)
		if err != nil {
			return nil, err
		}
		r, err := evalTrue(res)
		if err != nil {
			return nil, err
		}
		if r {
			return Symbol{"true"}, nil
		}
	}
	return Symbol{"false"}, nil
}

func evalAnd(exp []Expr, env *Environment) (Expr, error) {
	if len(exp) == 0 {
		return nil, errors.New("and expects at least one argument")
	}
	for _, e := range exp {
		res, err := Eval(e, env)
		if err != nil {
			return nil, err
		}
		r, err := evalTrue(res)
		if err != nil {
			return nil, err
		}
		if !r {
			return Symbol{"false"}, nil
		}
	}
	return Symbol{"true"}, nil
}

func makeProcedure(par Expr, body []Expr, env *Environment) (Expr, error) {
	if p, ok := par.(List); ok {
		return List{[]Expr{
			Symbol{"procedure"},
			Procedure{body, p.args, env},
		}}, nil
	}
	return nil, errors.New("makeProcedure: expected a list of parameters")
}

func EvalSequence(exprs []Expr, env *Environment) (Expr, error) {
	if len(exprs) == 0 {
		return nil, errors.New("begin body should not be empty")
	}
	var res Expr
	var err error
	for _, e := range exprs {
		res, err = Eval(e, env)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func evalCond(exps []Expr, env *Environment) (Expr, error) {
	exp, err := condToIf(exps)
	if err != nil {
		return nil, err
	}
	return Eval(exp, env)
}

func condToIf(exps []Expr) (Expr, error) {
	if len(exps) == 0 {
		return Symbol{"false"}, nil
	}
	switch j := exps[0].(type) {
	case List:
		if len(j.args) == 3 {
			if s, ok := j.args[1].(Symbol); ok && s.content == "=>" {
				res, err := condToIf(exps[1:])
				if err != nil {
					return nil, err
				}
				return List{[]Expr{Symbol{"let"},
					List{[]Expr{
						List{[]Expr{Symbol{"value"}, s}}}},
					List{[]Expr{
						Symbol{"if"}, Symbol{"value"},
						List{[]Expr{j.args[2], Symbol{"value"}}},
						res,
					}}}}, nil
			}
		}
		if k, ok := j.args[0].(Symbol); ok && k.content == "else" {
			if 1 != len(exps) {
				return nil, errors.New("cond: the else clause must be the last one")
			}
			return sequenceToExp(j.args[1:]), nil
		}
		res, err := condToIf(exps[1:])
		if err != nil {
			return nil, err
		}
		return List{[]Expr{Symbol{"if"}, j.args[0],
			sequenceToExp(j.args[1:]), res}}, nil
	default:
		return nil, errors.New("cond: each clause must be a list")
	}
}

func primitiveApply(proc List, args []Expr) (Expr, error) {
	if p, ok := proc.args[1].(Action); ok {
		return p.f(args)
	}
	return nil, errors.New("malformed primitive procedure")
}

func apply(proc Expr, args []Expr, env *Environment) (Expr, error) {
	if pro, ok := proc.(List); ok {
		if pr, ok := pro.args[0].(Symbol); ok {
			switch pr.content {
			case "primitive":
				a, err := listOfArgValues(args, env)
				if err != nil {
					return nil, err
				}
				return primitiveApply(pro, a)
			case "procedure":
				if p, ok := pro.args[1].(Procedure); ok {
					a := listOfDelayedArgs(args, env)
					new_env, err := p.env.extend_environment(p.parameters, a)
					if err != nil {
						return nil, err
					}
					return EvalSequence(p.body, &new_env)
				}
			case "macro":
				macro := pro.args[1].(Macro)
				newEnv, err := macro.defEnv.extend_environment(macro.para, args)
				if err != nil {
					return nil, err
				}
				expansion, err := EvalSequence(macro.body, &newEnv)
				return Eval(expansion, env)
			}
		}
	}
	return nil, fmt.Errorf("invalid procedure: %s", Print(proc))
}

func evalApply(exps []Expr, env *Environment) (Expr, error) {
	procedure, err := actualValue(exps[0], env)
	if err != nil {
		return nil, err
	}
	return apply(procedure, exps[1:], env)
}

func delayIt(exps Expr, env *Environment) *Thunk {
	return &Thunk{true, exps, env}
}

func listOfDelayedArgs(exps []Expr, env *Environment) []Expr {
	res := make([]Expr, len(exps))
	for k, i := range exps {
		res[k] = delayIt(i, env)
	}
	return res
}

func listOfArgValues(exps []Expr, env *Environment) ([]Expr, error) {
	res := make([]Expr, len(exps))
	var err error
	for k, i := range exps {
		res[k], err = actualValue(i, env)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func forceIt(obj *Thunk) (Expr, error) {
	if obj.thunk {
		result, err := actualValue(obj.exp, obj.env)
		if err != nil {
			return nil, err
		}
		obj.thunk = false
		obj.exp = result
		obj.env = nil
		return result, nil
	}
	return obj.exp, nil
}

func actualValue(exp Expr, env *Environment) (Expr, error) {
	res, err := Eval(exp, env)
	if err != nil {
		return nil, err
	}
	if l, ok := res.(*Thunk); ok {
		return forceIt(l)
	}
	return res, nil
}

func evalMap(args []Expr, env *Environment) (Expr, error) {
	proc, err := actualValue(args[0], env)
	if err != nil {
		return nil, err
	}
	xs, err := actualValue(args[1], env)
	if err != nil {
		return nil, err
	}
	l := xs.(List).args
	res := make([]Expr, len(l))
	for i, k := range l {
		r, err := apply(proc, []Expr{&Thunk{false, k, nil}}, env)
		if err != nil {
			return nil, err
		}
		for {
			t, ok := r.(*Thunk)
			if !ok {
				break
			}
			r, err = forceIt(t)
			if err != nil {
				return nil, err
			}
		}
		res[i] = r
	}
	return List{res}, nil
}

func evalForEach(args []Expr, env *Environment) (Expr, error) {
	proc, err := actualValue(args[0], env)
	if err != nil {
		return nil, err
	}
	x := [][]Expr{}
	for _, xst := range args[1:] {
		xs, err := actualValue(xst, env)
		if err != nil {
			return nil, err
		}
		l := xs.(List).args
		x = append(x, l)
	}
	l := len(x[0])
	le := len(x)
	res := make([]Expr, le)
	for k := range l {
		for i := range le {
			res[i] = &Thunk{false, x[i][k], nil}
		}
		if _, err := apply(proc, res, env); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func evalFilter(args []Expr, env *Environment) (Expr, error) {
	proc, err := actualValue(args[0], env)
	if err != nil {
		return nil, err
	}
	xs, err := actualValue(args[1], env)
	if err != nil {
		return nil, err
	}
	l := xs.(List).args
	res := []Expr{}
	for _, k := range l {
		r, err := apply(proc, []Expr{&Thunk{false, k, nil}}, env)
		if err != nil {
			return nil, err
		}
		for {
			t, ok := r.(*Thunk)
			if !ok {
				break
			}
			r, err = forceIt(t)
			if err != nil {
				return nil, err
			}
		}
		b, err := evalTrue(r)
		if err != nil {
			return nil, err
		}
		if b {
			res = append(res, k)
		}
	}
	return List{res}, nil
}

func evalFold(args []Expr, env *Environment) (Expr, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("wrong number of arguments of fold, need 3, get %d", len(args))
	}
	proc, err := actualValue(args[0], env)
	if err != nil {
		return nil, err
	}
	init, err := actualValue(args[1], env)
	if err != nil {
		return nil, err
	}
	plist, err := actualValue(args[2], env)
	if err != nil {
		return nil, err
	}
	list, ok := plist.(List)
	if !ok {
		return nil, fmt.Errorf("fold expects a list as its third argument, got %T", plist)
	}
	for _, k := range list.args {
		init, err = apply(proc, []Expr{&Thunk{false, init, nil}, &Thunk{false, k, nil}}, env)
		if err != nil {
			return nil, err
		}
		for {
			t, ok := init.(*Thunk)
			if !ok {
				break
			}
			init, err = forceIt(t)
			if err != nil {
				return nil, err
			}
		}
	}
	return init, nil
}

func evalFuncApply(args []Expr, env *Environment) (Expr, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("wrong number of arguments of apply, need at least 2, get %d", len(args))
	}
	proc, err := actualValue(args[0], env)
	if err != nil {
		return nil, err
	}
	argList, err := actualValue(args[len(args)-1], env)
	if err != nil {
		return nil, err
	}
	lst, ok := argList.(List)
	if !ok {
		return nil, fmt.Errorf("apply expects a list as its last argument, got %T", argList)
	}
	callArgs := make([]Expr, 0, len(args)-2+len(lst.args))
	for _, e := range args[1 : len(args)-1] {
		v, err := actualValue(e, env)
		if err != nil {
			return nil, err
		}
		callArgs = append(callArgs, &Thunk{false, v, nil})
	}
	for _, e := range lst.args {
		callArgs = append(callArgs, &Thunk{false, e, nil})
	}
	return apply(proc, callArgs, env)
}

func evalQuasi(e Expr, depth int, env *Environment) (Expr, error) {
	switch t := e.(type) {
	case List:
		if sym, ok := t.args[0].(Symbol); ok {
			switch sym.content {
			case "quasiquote":
				if len(t.args) < 2 {
					return nil, errors.New("quasiquote requires an argument")
				}
				inner, err := evalQuasi(t.args[1], depth+1, env)
				if err != nil {
					return nil, err
				}
				return List{[]Expr{Symbol{"quasiquote"}, inner}}, nil
			case "unquote":
				if len(t.args) < 2 {
					return nil, errors.New("unquote requires an argument")
				}
				if depth == 1 {
					return actualValue(t.args[1], env)
				}
				inner, err := evalQuasi(t.args[1], depth-1, env)
				if err != nil {
					return nil, err
				}
				return List{[]Expr{Symbol{"unquote"}, inner}}, nil
			case "unquote-splicing":
				if len(t.args) < 2 {
					return nil, errors.New("unquote-splicing requires an argument")
				}
				if depth == 1 {
					val, err := actualValue(t.args[1], env)
					if err != nil {
						return nil, err
					}
					lv, ok := val.(List)
					if !ok {
						return nil, fmt.Errorf("unquote-splicing requires a list, got %s", Print(val))
					}
					return &splice{lv.args}, nil
				}
				inner, err := evalQuasi(t.args[1], depth-1, env)
				if err != nil {
					return nil, err
				}
				return List{[]Expr{Symbol{"unquote-splicing"}, inner}}, nil
			}
		}
		out := []Expr{}
		for _, el := range t.args {
			r, err := evalQuasi(el, depth, env)
			if err != nil {
				return nil, err
			}
			if s, ok := r.(*splice); ok {
				out = append(out, s.args...)
			} else {
				out = append(out, r)
			}
		}
		return List{out}, nil
	default:
		return e, nil
	}
}

func evalDefMacro(args []Expr, env *Environment) (Expr, error) {
	macro_Name := args[0].(Symbol).content
	para := args[1].(List)
	body := args[2:]
	macro := Macro{macro_Name, para.args, body, env}
	env.env[0][macro_Name] = List{[]Expr{Symbol{"macro"}, macro}}
	return macro, nil
}

func Eval(exp Expr, env *Environment) (Expr, error) {
	if exp == nil {
		return nil, nil
	}
	switch x := exp.(type) {
	case Number, String:
		return x, nil
	case *Thunk:
		return x, nil
	case Symbol:
		if v, err := lookUpVariable(x, env); err == nil {
			return v, nil
		}
		if isKeyword(x) {
			return x, nil
		}
		return nil, fmt.Errorf("unbound variable: %s", Print(x))
	case List:
		switch y := x.args[0].(type) {
		case Symbol:
			if _, err := lookUpVariable(y, env); err == nil {
				return evalApply(x.args, env)
			}
			switch y.content {
			case "quote":
				return x.args[1], nil
			case "quasiquote":
				res, err := evalQuasi(x.args[1], 1, env)
				if err != nil {
					return nil, err
				}
				if _, ok := res.(*splice); ok {
					return nil, errors.New("unquote-splicing is not allowed at top level of a quasiquote")
				}
				return res, nil
			case "set!":
				return Symbol{"ok"}, evalAssignment(x.args[1:], env)
			case "define":
				return Symbol{"ok"}, evalDefinition(x.args[1:], env)
			case "let":
				return evalLet(x.args[1:], env)
			case "let*":
				return evalLetStar(x.args[1:], env)
			case "if":
				return evalIf(x.args[1:], env)
			case "and":
				return evalAnd(x.args[1:], env)
			case "or":
				return evalOr(x.args[1:], env)
			case "lambda":
				return makeProcedure(x.args[1], x.args[2:], env)
			case "begin":
				return EvalSequence(x.args[1:], env)
			case "defmacro":
				return evalDefMacro(x.args[1:], env)
			case "map":
				return evalMap(x.args[1:], env)
			case "for-each":
				return evalForEach(x.args[1:], env)
			case "filter":
				return evalFilter(x.args[1:], env)
			case "fold":
				return evalFold(x.args[1:], env)
			case "cond":
				return evalCond(x.args[1:], env)
			case "eval":
				return Eval(x.args[1], env)
			case "apply":
				return evalFuncApply(x.args[1:], env)
			default:
				return evalApply(x.args, env)
			}
		case List:
			return evalApply(x.args, env)
		default:
			return nil, fmt.Errorf("not a procedure: %v", y)
		}
	default:
		panic(fmt.Sprintf("not implemented yet: %T", exp))
	}
}
