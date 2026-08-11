package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

// Core expression types
type Expr interface {
	exprNode()
	Print() string
}

type Environment struct {
	env []frame
}

type frame map[string]Expr

type stringNumber struct {
	Value string
}

type Number struct {
	value Value
}

type Symbol struct {
	content string
}

type List struct {
	args []Expr
}

type Action struct {
	name string
	f    func([]Expr) (Expr, error)
}

type String struct {
	content string
}

// Lexer and Parser types
type TokenType int

const (
	LPAREN TokenType = iota
	RPAREN
	NUMBER
	IDENT
	QUOTE
	STRING
	EOF
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	input    string
	position int
}

type Parser struct {
	tokens   []Token
	position int
}

type Type int

const (
	IntType Type = iota
	RationalType
	RealType
	ComplexType
)

// Implementation of Integers
type Integer int

func (Integer) Type() Type {
	return IntType
}

func (a Integer) add(b Integer) Integer {
	return a + b
}

func (a Integer) minu(b Integer) Integer {
	return a - b
}

func (a Integer) times(b Integer) Integer {
	return a * b
}

func (a Integer) divUp(b Integer) (Integer, error) {
	return (a + b - 1).divDown(b)
}

func (a Integer) divDown(b Integer) (Integer, error) {
	if b == 0 {
		return a, errors.New("illegal action: divisor is 0")
	}
	return a / b, nil
}

func (a Integer) mod(b Integer) Integer {
	return a % b
}

func (a Integer) Abs() Integer {
	if a < 0 {
		return -a
	}
	return a
}

func (a Integer) Print() string {
	return strconv.Itoa(int(a))
}

// Implementation of rationals
type Rational struct {
	p int
	q int
}

func (Rational) Type() Type {
	return RationalType
}

func (a Rational) simp() Rational {
	m, n := a.p, a.q
	for n != 0 {
		m, n = n, m%n
	}
	if a.q < 0 {
		m = -m
	}
	return Rational{
		p: a.p / m,
		q: a.q / m,
	}
}

func (a Rational) neg() Rational {
	return Rational{
		p: -a.p,
		q: a.q,
	}
}

func (a Rational) inv() (Rational, error) {
	if a.p == 0 {
		return a, errors.New("illegal action: the divisor is 0")
	}
	return Rational{
		p: a.q,
		q: a.p,
	}, nil
}

func (a Rational) add(b Rational) Rational {
	return Rational{
		p: a.p*b.q + a.q*b.p,
		q: a.q * b.q,
	}.simp()
}

func (a Rational) times(b Rational) Rational {
	return Rational{
		p: a.p * b.p,
		q: a.q * b.q,
	}.simp()
}

func (a Rational) minu(b Rational) Rational {
	return a.add(b.neg())
}

func (a Rational) div(b Rational) (Rational, error) {
	c, err := b.inv()
	if err != nil {
		return a, err
	}
	return a.times(c), nil
}

func (a Rational) Int() Integer {
	return Integer(a.p / a.q)
}

func (a Rational) Abs() Rational {
	if (a.p * a.q) < 0 {
		return Rational{p: -a.p, q: a.q}
	}
	return a
}

func (x Rational) Print() string {
	if x.p%x.q == 0 {
		return fmt.Sprintf("%d", x.p/x.q)
	}
	return fmt.Sprintf("%d/%d", x.p, x.q)
}

// Implementation of Reals
type Real float64

func (Real) Type() Type {
	return RealType
}

func (a Real) add(b Real) Real {
	return a + b
}

func (a Real) minu(b Real) Real {
	return a - b
}

func (a Real) times(b Real) Real {
	return a * b
}

func (a Real) div(b Real) (Real, error) {
	if b == 0 {
		return a, errors.New("illegal action: divisor is 0")
	}
	return a / b, nil
}

func (a Real) Int() Integer {
	return Integer(a)
}

func (a Real) Abs() Real {
	if a < 0 {
		return -a
	}
	return a
}

func (x Real) Print() string {
	if math.Trunc(float64(x)) == float64(x) {
		return fmt.Sprintf("%d", int(x))
	}
	return fmt.Sprintf("%f", x)
}

// Implementation of Complex
type Complex struct {
	a float64
	b float64
}

func (Complex) Type() Type {
	return ComplexType
}

func (x Complex) neg() Complex {
	return Complex{
		a: -x.a,
		b: -x.b,
	}
}

func (x Complex) inv() (Complex, error) {
	if x.a == 0 && x.b == 0 {
		return x, nil
	}
	frac := x.a*x.a + x.b*x.b
	return Complex{
		a: x.a / frac,
		b: -x.b / frac,
	}, nil
}

func (x Complex) add(y Complex) Complex {
	return Complex{
		a: x.a + y.a,
		b: x.b + y.b,
	}
}

func (x Complex) times(y Complex) Complex {
	return Complex{
		a: x.a*y.a - x.b*y.b,
		b: x.b*y.a + x.a*y.b,
	}
}

func (x Complex) minu(y Complex) Complex {
	return x.add(y.neg())
}

func (x Complex) div(y Complex) (Complex, error) {
	z, err := y.inv()
	if err != nil {
		return z, err
	}
	return x.times(z), nil
}

func (x Complex) Abs() Real {
	return Real(math.Sqrt(x.a*x.a + x.b*x.b))
}

func (x Complex) Print() string {
	if x.a == 0 {
		return Real(x.b).Print()
	} else if x.b == 0 {
		return Real(x.a).Print()
	} else if x.b < 0 {
		return fmt.Sprintf("%f%fi", x.a, x.b)
	}
	return fmt.Sprintf("%f+%fi", x.a, x.b)
}

func (String) exprNode() {}

func (s String) Print() string {
	return fmt.Sprintf("%s", s.content)
}

// Implement types
type Value interface {
	Type() Type
	Print() string
}

func nextType(t Type) Type {
	switch t {
	case IntType:
		return RationalType
	case RationalType:
		return RealType
	case RealType:
		return ComplexType
	default:
		return t
	}
}

func upOnce(v Value) Value {
	switch x := v.(type) {
	case Integer:
		return Rational{p: int(x), q: 1}
	case Rational:
		return Real(x.p) / Real(x.q)
	case Real:
		return Complex{a: float64(x), b: 0}
	default:
		return v
	}
}

func upTo(v Value, t Type) Value {
	for v.Type() != t {
		v = upOnce(v)
	}
	return v
}

func levelUp(a Value, b Value) (Value, Value) {
	ta, tb := a.Type(), b.Type()
	if ta == tb {
		return a, b
	}
	if ta < tb {
		return upTo(a, tb), b
	}
	return a, upTo(b, ta)
}

// Implement basic calculations
func Add(a Value, b Value) Value {
	a, b = levelUp(a, b)
	switch x := a.(type) {
	case Integer:
		return x.add(b.(Integer))
	case Rational:
		return x.add(b.(Rational))
	case Real:
		return x.add(b.(Real))
	case Complex:
		return x.add(b.(Complex))
	default:
		panic("unsupport types")
	}
}

func Minu(a Value, b Value) Value {
	a, b = levelUp(a, b)
	switch x := a.(type) {
	case Integer:
		return x.minu(b.(Integer))
	case Rational:
		return x.minu(b.(Rational))
	case Real:
		return x.minu(b.(Real))
	case Complex:
		return x.minu(b.(Complex))
	default:
		panic("unsupport types")
	}
}

func Times(a Value, b Value) Value {
	a, b = levelUp(a, b)
	switch x := a.(type) {
	case Integer:
		return x.times(b.(Integer))
	case Rational:
		return x.times(b.(Rational))
	case Real:
		return x.times(b.(Real))
	case Complex:
		return x.times(b.(Complex))
	default:
		panic("unsupport types")
	}
}

func Div(a Value, b Value) (Value, error) {
	a, b = levelUp(a, b)
	switch x := a.(type) {
	case Integer:
		return Div(upOnce(a), upOnce(b))
	case Rational:
		res, err := x.div(b.(Rational))
		if err != nil {
			return res, err
		}
		return res, nil
	case Real:
		res, err := x.div(b.(Real))
		if err != nil {
			return res, err
		}
		return res, nil
	case Complex:
		res, err := x.div(b.(Complex))
		if err != nil {
			return res, err
		}
		return res, nil
	default:
		panic("unsupport types")
	}
}

func Int(a Value) (Integer, error) {
	switch x := a.(type) {
	case Integer:
		return x, nil
	case Rational:
		return x.Int(), nil
	case Real:
		return x.Int(), nil
	default:
		panic("not implemented yet")
	}
}

func Abs(a Value) Value {
	switch x := a.(type) {
	case Integer:
		return x.Abs()
	case Rational:
		return x.Abs()
	case Real:
		return x.Abs()
	case Complex:
		return x.Abs()
	default:
		panic("not implemented yet")
	}
}

func oneOf(a Value) Value {
	switch a.(type) {
	case Integer:
		return Integer(1)
	case Rational:
		return Rational{p: 1, q: 1}
	case Real:
		return Real(1)
	case Complex:
		return Complex{a: 1, b: 0}
	default:
		panic("not implemented yet")
	}
}

func Power(a Value, n int) Value {
	res := oneOf(a)
	for n > 0 {
		if n&1 == 1 {
			res = Times(res, a)
		}
		a = Times(a, a)
		n >>= 1
	}
	return res
}

func IsZero(a Value) bool {
	switch x := a.(type) {
	case Integer:
		return x == Integer(0)
	case Rational:
		return x.p == 0
	case Real:
		return x == Real(0)
	case Complex:
		return x == Complex{a: 0, b: 0}
	default:
		panic("not implemented yet")
	}
}
