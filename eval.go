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
					fmt.Println(i.Print())
				}
				return nil, nil
			}},
		"load": {name: "load",
			f: func(args []Expr) (Expr, error) {
				if len(args)==1{
					file:=args[0].(String).content
					code, err:= ReadFile(file)
					if err!=nil{return nil,err}
					fmt.Printf("file %s loaded successfully", file)
					return Process(code)
				}
				return nil, fmt.Errorf("wrong number of arguments of load, need 1, get %d", len(args))
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
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					return Number{value: Add(a, b)}, nil
				}
				return nil, errors.New("args legth incorrect to call + ")
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
				if len(args) == 2 {
					a := args[0].(Number).value
					b := args[1].(Number).value
					return Number{value: Times(a, b)}, nil
				}
				return nil, errors.New("args legth incorrect to call * ")
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

func lookUpVariable(exp Symbol, env Environment) (Expr, error) {
	for _, frame := range env.env {
		res, ok := frame[exp.content]
		if ok {
			return res, nil
		}
	}
	return nil, fmt.Errorf("unable to find variable %s's value", exp.content)
}

func evalAssignment(vari Symbol, exp []Expr, env *Environment) error {
	if len(exp) != 2 {
		return errors.New("in correct numbers of argumens for set!")
	}
	for _, j := range (*env).env {
		_, a := j[vari.content]
		if a {
			res, err := Eval(exp[1], env)
			if err != nil {
				return err
			}
			j[vari.content] = res
			return nil
		}
	}
	return fmt.Errorf("unbounded variable: %s", vari.content)
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
			}
		}
	}
	fmt.Println(proc)
	return nil, errors.New("invalid procedure")
}

func evalApply(exps []Expr, env *Environment) (Expr, error) {
	procedure, err := Eval(exps[0], env)
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

func evalMap(args []Expr, env *Environment) (Expr,error){
	proc, err:= Eval(args[0],env)
	if err!=nil{return nil,err}
	xs,err:=Eval(args[1],env)
	if err!=nil{return nil,err}
	l := xs.(List).args
	res := make([]Expr, len(l))
	for i,k := range l{
		r, err:= apply(proc, []Expr{k}, env)
		if err!=nil{return nil,err}
		res[i] = r
	}
	return List{args:res},nil
}

func evalFilter(args []Expr, env *Environment) (Expr, error){
	proc, err:= Eval(args[0],env)
	if err!=nil{return nil,err}
	xs,err:=Eval(args[1],env)
	if err!=nil{return nil,err}
	l := xs.(List).args
	res := []Expr{}
	for _,k := range l{
		r, err:= apply(proc, []Expr{k}, env)
		if err!=nil{return nil,err}
		b,err:=evalTrue(r)
		if err!=nil{return nil,err}
		if b{res= append(res,k)}
	}
	return List{args:res},nil
}

func evalListRef(args []Expr, env *Environment) (Expr, error){
	if len(args)!=2{
		return nil, errors.New("wrong number of arguments to call list-ref")
	}
	list := args[0].(List)
	num, err:= Eval(args[1],env)
	if err!=nil{return nil,err}
	index,err := Int(num.(Number).value)
	if err!=nil{return nil,err}
	return list.args[int(index)],nil
}

func evalListSet(args []Expr, env *Environment) (Expr, error){
	if len(args)!=3{
		return nil, errors.New("wrong number of arguments to call list-set")
	}
	list := args[0].(List)
	num, err:= Eval(args[1],env)
	if err!=nil{return nil,err}
	index,err := Int(num.(Number).value)
	if err!=nil{return nil,err}
	res ,err:= Eval(args[2], env)
	if err!=nil{return nil,err}
	list.args[int(index)]= res
	return list,nil
}

func evalListTail(args []Expr, env *Environment) (Expr, error){
	if len(args)!=2{
		return nil, errors.New("wrong number of arguments to call list-tail")
	}
	list := args[0].(List)
	num, err:= Eval(args[1],env)
	if err!=nil{return nil,err}
	index,err := Int(num.(Number).value)
	if err!=nil{return nil,err}
	return List{args:list.args[index:]},nil
}

func Eval(exp Expr, env *Environment) (Expr, error) {
	if exp == nil {
		return nil, nil
	}
	switch x := exp.(type) {
	case Number, String:
		return x, nil
	case Symbol:
		return lookUpVariable(x, *env)
	case List:
		switch y := x.args[0].(type) {
		case Symbol:
			switch y.content {
			case "quote":
				return x.args[1], nil
			case "set!":
				return Symbol{content: "ok"}, evalAssignment(y, x.args[1:], env)
			case "define":
				return Symbol{content: "ok"}, evalDefinition(x.args[1:], env)
			case "let":
				return evalLet(x.args[1:], env)
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
			case "map":
				return evalMap(x.args[1:],env)
			case "filter":
				return evalFilter(x.args[1:],env)
			case "list-ref":
				return evalListRef(x.args[1:], env)
			case "list-set!":
				return evalListSet(x.args[1:], env)
			case "list-tail":
				return evalListTail(x.args[1:], env)
			case "cond":
				return evalCond(x.args[1:], env)
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
