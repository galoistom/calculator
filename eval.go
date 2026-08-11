package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func make_frame(vars []Expr, vals []Expr) (frame, error) {
	res := make(frame)
	for i := range vars {
		if j, ok := vars[i].(Symbol); ok {
			res[j.content] = vals[i]
		} else {
			return nil, errors.New("frame should only contain symbol")
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
	return Environment{}, errors.New("numbers of values and variables dose not match")
}
func (env *Environment) exprNode() {}
func (env *Environment) Print() string {
	return fmt.Sprint(*env)
}

func (Number) exprNode() {}
func (n Number) Print() string {
	return n.value.Print()
}

func (x stringNumber) getValue() (Number, error) {
	num := x.Value
	if strings.Contains(num, ".") {
		f, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return Number{}, fmt.Errorf("failed to convert to float: %v", err)
		}
		return Number{value: Real(f)}, nil
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
		return Number{value: Rational{p: int(p), q: int(q)}}, nil
	}
	i, err := strconv.Atoi(num)
	if err != nil {
		return Number{}, err
	}
	return Number{value: Integer(i)}, nil
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
func (s splice) Print() string {
	return "(splice)"
}

func (Macro) exprNode() {}
func (m Macro) Print() string {
	return fmt.Sprintf("(Macro %s)", m.name)
}

func InitEnvironment() (*Environment, error) {
	env := Environment{[]frame{{
		"true":  Symbol{content: "true"},
		"false": Symbol{content: "false"}}}}
	vals := []Expr{}
	vars := []Expr{}
	var inputScanner = bufio.NewScanner(os.Stdin)

	var primitiveAction = map[string]Action{
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
					return List{args: args[0].(List).args[1:]}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of cdr, need 1, get %d", len(args))
			}},
		"display": {name: "display",
			f: func(args []Expr) (Expr, error) {
				for _, i := range args {
					if i != nil {
						fmt.Println(i.Print())
					}
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
					fmt.Printf("file %s loaded successfully\n", file)
					return Process(code)
				}
				return nil, fmt.Errorf("wrong number of arguments of load, need 1, get %d", len(args))
			}},
		"error": {name: "error",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					message := args[0].(String).content
					return nil, errors.New(message)
				}
				return nil, fmt.Errorf("wrong number of arguments of error, need 1, get %d", len(args))
			}},
		"newline": {name: "newline",
			f: func(args []Expr) (Expr, error) {
				fmt.Println()
				return nil, nil
			}},
		"pair?": {name: "pair?",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					if a, ok := args[0].(List); ok {
						if len(a.args) >= 2 {
							return Symbol{content: "true"}, nil
						}
						return Symbol{content: "false"}, nil
					}
					return Symbol{content: "false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of pair?, need 1, get %d", len(args))
			}},
		"eq?": {name: "eq?",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					if args[0] == args[1] {
						return Symbol{content: "true"}, nil
					}
					return Symbol{content: "false"}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of eq?, need 2, get %d", len(args))
			}},
		"list": {name: "list",
			f: func(args []Expr) (Expr, error) {
				return List{args: args}, nil
			}},
		"list-len": {name: "list-len",
			f: func(args []Expr) (Expr, error) {
				if list, ok := args[0].(List); ok {
					return Number{value: Integer(len(list.args))}, nil
				}
				return nil, errors.New("list-len only accept list")
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
					return List{args: list.args[index:]}, nil
				}
				return nil, errors.New("wrong number of arguments to call list-tail")
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
				return nil, errors.New("wrong number of arguments to call list-set")
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
				return nil, errors.New("wrong number of arguments to call list-ref")
			}},
		"null?": {name: "null?",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					if l, ok := args[0].(List); ok {
						if len(l.args) == 0 {
							return Symbol{content: "true"}, nil
						}
						return Symbol{content: "false"}, nil
					}
					return nil, errors.New("Wrong type to call null?")
				}
				return nil, fmt.Errorf("wrong number of arguments of eq?, need 2, get %d", len(args))
			}},
		"cons": {name: "cons",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					var x, y []Expr
					switch a := args[0].(type) {
					case List:
						x = a.args
					default:
						x = []Expr{a}
					}
					switch a := args[1].(type) {
					case List:
						y = a.args
					default:
						y = []Expr{a}
					}
					return List{args: append(x, y...)}, nil
				}
				return nil, fmt.Errorf("wrong number of arguments of eq?, need 2, get %d", len(args))
			}},
		"readline": {name: "readline",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 0 {
					fmt.Print("input:")
					if inputScanner.Scan() {
						line := inputScanner.Text()
						return Process(line)
					}
					if err := inputScanner.Err(); err != nil {
						fmt.Println("failed to read", err)
					}
				}
				return nil, errors.New("readline should not have any args")
			}},
		"+": {name: "+",
			f: func(args []Expr) (Expr, error) {
				switch args[0].(type) {
				case Number:
					num := Value(Integer(0))
					for _, k := range args {
						num = Add(k.(Number).value, num)
					}
					return Number{value: num}, nil
				case String:
					var b strings.Builder
					for _, k := range args {
						fmt.Fprintf(&b, "%s", k.(String).content)
					}
					return String{content: b.String()}, nil
				}
				return nil, errors.New("wrong type to call +")
			}},
		"-": {name: "-",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					return Number{value: Minu(a, b)}, nil
				}
				if len(args) == 1 {
					a := args[0].(Number).value
					return Number{value: Minu(Integer(0), a)}, nil
				}
				return nil, errors.New("args legth incorrect to call - ")
			}},
		"*": {name: "*",
			f: func(args []Expr) (Expr, error) {
				num := Value(Integer(1))
				for _, k := range args {
					num = Times(k.(Number).value, num)
				}
				return Number{value: num}, nil
			}},
		"/": {name: "/",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					res, err := Div(a, b)
					return Number{value: res}, err
				}
				return nil, errors.New("args legth incorrect to call / ")
			}},
		"=": {name: "=",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					if IsZero(Minu(a, b)) {
						return Symbol{content: "true"}, nil
					}
					return Symbol{content: "false"}, nil

				}
				return nil, errors.New("args legth incorrect to call = ")
			}},
		"pow": {name: "pow",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					x, ok := b.(Integer)
					if !ok {
						return Number{value: x}, errors.New("none integer power is not implemented yet")
					}
					return Number{value: Power(a, int(x))}, nil
				}
				return nil, errors.New("args legth incorrect to call pow ")
			}},
		"c": {name: "c",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					return Number{value: Add(a, Times(Complex{a: 0, b: 1}, b))}, nil
				}
				return nil, errors.New("args legth incorrect to call c ")
			}},
		"abs": {name: "abs",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					a := args[0].(Number).value
					return Number{value: Abs(a)}, nil
				}
				return nil, errors.New("args legth incorrect to call abs ")
			}},
		"int": {name: "int",
			f: func(args []Expr) (Expr, error) {
				if len(args) == 1 {
					a := args[0].(Number).value
					res, err := Int(a)
					return Number{value: res}, err
				}
				return nil, errors.New("args legth incorrect to call abs ")
			}},
	}
	for _, a := range composedAccessors() {
		primitiveAction[a.name] = a
	}

	for i := range primitiveAction {
		vars = append(vars, Symbol{content: i})
		vals = append(vals, List{args: []Expr{Symbol{content: "primitive"}, primitiveAction[i]}})
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
					cur = List{args: l.args[1:]}
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
	return nil, fmt.Errorf("unable to find variable %s's value", exp.content)
}

func evalAssignment(exp []Expr, env *Environment) error {
	if len(exp) != 2 {
		return errors.New("in correct numbers of argumens for set!")
	}
	target, ok := exp[0].(Symbol)
	if !ok {
		return errors.New("set! requires a variable")
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
	return fmt.Errorf("unbounded variable: %s", target.content)
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
			return Symbol{}, errors.New("definition should start with a literal")
		}
	default:
		return Symbol{}, errors.New("Wrong type of definition object")
	}
}

func makeLambda(par []Expr, body []Expr) List {
	return List{args: append([]Expr{Symbol{content: "lambda"}, List{args: par}}, body...)}
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
		return List{args: append([]Expr{Symbol{content: "begin"}}, exp...)}
	}
}

func splitVarVal(exp []Expr) ([]Expr, []Expr, error) {
	variables := []Expr{}
	values := []Expr{}
	for _, i := range exp {
		switch j := i.(type) {
		case List:
			if len(j.args) != 2 {
				return nil, nil, errors.New("Wrong numbers of arguments in let")
			}
			values = append(values, j.args[1])
			switch k := j.args[0].(type) {
			case Symbol:
				variables = append(variables, k)
				continue
			}
		default:
			return nil, nil, errors.New("assigment with wrong type")
		}
	}
	return variables, values, nil
}

func evalLet(exp []Expr, env *Environment) (Expr, error) {
	if len(exp) <= 1 {
		return nil, errors.New("let command syntax error, need at least 3 indexes")
	}
	switch x := exp[0].(type) {
	case Symbol:
		ass := exp[1]
		switch assignments := ass.(type) {
		case List:
			values, variables, err := splitVarVal(assignments.args)
			if err != nil {
				return nil, err
			}
			function := List{args: append([]Expr{Symbol{content: "define"},
				List{args: append([]Expr{x}, variables...)}}, exp[2:]...)}
			procedure := sequenceToExp([]Expr{function, List{args: append([]Expr{x}, values...)}})
			new_env, err := env.extend_environment([]Expr{}, []Expr{})
			return Eval(procedure, &new_env)
		}
		return nil, errors.New("syntax error in let")
	case List:
		variables, values, err := splitVarVal(x.args)
		if err != nil {
			return nil, err
		}
		lambda := List{args: append(
			[]Expr{List{args: append(
				[]Expr{Symbol{content: "lambda"},
					List{args: variables}},
				exp[1:]...)}},
			values...)}
		return Eval(lambda, env)
	default:
		return nil, errors.New("assignemnt syntax error")
	}
}

func letStripStar(exps []Expr) (Expr, error) {
	ass := exps[0].(List)
	if len(ass.args) == 0 {
		return List{args: append([]Expr{Symbol{content: "begin"}}, exps[1:]...)}, nil
	}
	rest, err := letStripStar(append([]Expr{
		List{args: ass.args[1:]}},
		exps[1:]...))
	if err != nil {
		return nil, err
	}
	return List{args: []Expr{
		Symbol{content: "let"},
		List{args: []Expr{List{args: ass.args[0].(List).args}}},
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
	return false, errors.New("not valid expr")
}

func evalIf(exp []Expr, env *Environment) (Expr, error) {
	if len(exp) != 3 {
		return nil, fmt.Errorf("wrong number of args to call if, need 3, get %d", len(exp))
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
		return nil, errors.New("No argument to or")
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
			return Symbol{content: "true"}, nil
		}
	}
	return Symbol{content: "false"}, nil
}

func evalAnd(exp []Expr, env *Environment) (Expr, error) {
	if len(exp) == 0 {
		return nil, errors.New("No argument to and")
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
			return Symbol{content: "false"}, nil
		}
	}
	return Symbol{content: "true"}, nil
}

func makeProcedure(par Expr, body []Expr, env *Environment) (Expr, error) {
	if p, ok := par.(List); ok {
		return List{args: []Expr{
			Symbol{content: "procedure"},
			p, List{args: body}, env}}, nil
	}
	return nil, errors.New("makeProcedure expression format error")
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
		return Symbol{content: "false"}, nil
	}
	switch j := exps[0].(type) {
	case List:
		if len(j.args) == 3 {
			if s, ok := j.args[1].(Symbol); ok && s.content == "=>" {
				res, err := condToIf(exps[1:])
				if err != nil {
					return nil, err
				}
				return List{args: []Expr{Symbol{content: "let"},
					List{args: []Expr{
						List{args: []Expr{Symbol{content: "value"}, s}}}},
					List{args: []Expr{
						Symbol{content: "if"}, Symbol{content: "value"},
						List{args: []Expr{j.args[2], Symbol{content: "value"}}},
						res,
					}}}}, nil
			}
		}
		if k, ok := j.args[0].(Symbol); ok && k.content == "else" {
			if k.content == "else" {
				if 1 != len(exps) {
					return nil, errors.New("ELSE case is not the last in CNOD")
				}
				return sequenceToExp(j.args[1:]), nil
			}

		}
		res, err := condToIf(exps[1:])
		if err != nil {
			return nil, err
		}
		return List{args: []Expr{Symbol{content: "if"}, j.args[0],
			sequenceToExp(j.args[1:]), res}}, nil
	default:
		return nil, errors.New("Wrong type for condition")
	}
}

func primitiveApply(proc List, args []Expr) (Expr, error) {
	if p, ok := proc.args[1].(Action); ok {
		return p.f(args)
	}
	return nil, errors.New("Wrong type for primitive procedure")
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
				if old_env, ok := pro.args[3].(*Environment); ok {
					procedure_body := pro.args[2].(List).args
					procedure_para := pro.args[1].(List).args
					a := listOfDelayedArgs(args, env)
					new_env, err := old_env.extend_environment(procedure_para, a)
					if err != nil {
						return nil, err
					}
					return EvalSequence(procedure_body, &new_env)
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
	fmt.Println(proc)
	return nil, errors.New("invalid procedure")
}

func evalApply(exps []Expr, env *Environment) (Expr, error) {
	procedure, err := actualValue(exps[0], env)
	if err != nil {
		return nil, err
	}
	return apply(procedure, exps[1:], env)
}

func delayIt(exps Expr, env *Environment) Expr {
	return &List{args: append([]Expr{Symbol{content: "thunk"}}, exps, env)}
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

func forceIt(obj *List) (Expr, error) {
	if x, ok := obj.args[0].(Symbol); ok {
		switch x.content {
		case "thunk":
			result, err := actualValue(obj.args[1], obj.args[2].(*Environment))
			if err != nil {
				return nil, err
			}
			obj.args = []Expr{Symbol{content: "evaluated-thunk"}, result, nil}
			return result, nil
		case "evaluated-thunk":
			return obj.args[1], nil
		}
	}
	return obj, nil
}

func actualValue(exp Expr, env *Environment) (Expr, error) {
	res, err := Eval(exp, env)
	if err != nil {
		return nil, err
	}
	if l, ok := res.(*List); ok {
		return forceIt(l)
	}
	return res, nil
}

func evalMap(args []Expr, env *Environment) (Expr, error) {
	proc := args[0]
	xs, err := Eval(args[1], env)
	if err != nil {
		return nil, err
	}
	l := xs.(List).args
	res := make([]Expr, len(l))
	for i, k := range l {
		r, err := apply(proc, []Expr{k}, env)
		if err != nil {
			return nil, err
		}
		res[i] = r
	}
	return List{args: res}, nil
}

func evalFilter(args []Expr, env *Environment) (Expr, error) {
	proc := args[0]
	xs, err := Eval(args[1], env)
	if err != nil {
		return nil, err
	}
	l := xs.(List).args
	res := []Expr{}
	for _, k := range l {
		r, err := apply(proc, []Expr{k}, env)
		if err != nil {
			return nil, err
		}
		b, err := evalTrue(r)
		if err != nil {
			return nil, err
		}
		if b {
			res = append(res, k)
		}
	}
	return List{args: res}, nil
}

func evalFold(args []Expr, env *Environment) (Expr, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments, need 3")
	}
	proc := args[0]
	init, err := Eval(args[1], env)
	if err != nil {
		return nil, err
	}
	plist, err := Eval(args[2], env)
	if list, ok := plist.(List); ok {
		for _, k := range list.args {
			action := List{args: []Expr{proc, init, k}}
			init, err = Eval(action, env)
			if err != nil {
				return nil, err
			}
		}
		return init, nil
	}
	return nil, errors.New("fold can only process list")
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
				return List{args: []Expr{Symbol{content: "quasiquote"}, inner}}, nil
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
				return List{args: []Expr{Symbol{content: "unquote"}, inner}}, nil
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
						return nil, fmt.Errorf("unquote-splicing requires a list, got %s", val.Print())
					}
					return &splice{args: lv.args}, nil
				}
				inner, err := evalQuasi(t.args[1], depth-1, env)
				if err != nil {
					return nil, err
				}
				return List{args: []Expr{Symbol{content: "unquote-splicing"}, inner}}, nil
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
		return List{args: out}, nil
	default:
		return e, nil
	}
}

func evalDefMacro(args []Expr, env *Environment) (Expr, error) {
	macro_Name := args[0].(Symbol)
	para := args[1].(List)
	body := args[2:]
	macro := Macro{name: macro_Name.content, para: para.args, body: body, defEnv: env}
	(*env).env[0][macro_Name.content] = List{args: []Expr{Symbol{content: "macro"}, macro}}
	return macro, nil
}

func Eval(exp Expr, env *Environment) (Expr, error) {
	if exp == nil {
		return nil, nil
	}
	switch x := exp.(type) {
	case Number, String:
		return x, nil
	case Symbol:
		return lookUpVariable(x, env)
	case List:
		switch y := x.args[0].(type) {
		case Symbol:
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
				return Symbol{content: "ok"}, evalAssignment(x.args[1:], env)
			case "define":
				return Symbol{content: "ok"}, evalDefinition(x.args[1:], env)
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
			case "filter":
				return evalFilter(x.args[1:], env)
			case "fold":
				return evalFold(x.args[1:], env)
			case "cond":
				return evalCond(x.args[1:], env)
			case "eval":
				return Eval(x.args[1], env)
			default:
				return evalApply(x.args, env)
			}
		case List:
			return evalApply(x.args, env)
		default:
			return nil, fmt.Errorf("not a procedure: %v", y)
		}
	default:
		panic("not implemented yet")
	}
}
