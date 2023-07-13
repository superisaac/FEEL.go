package feel

import (
	"fmt"
	"strings"
)

type Scope map[string]interface{}

type Interpreter struct {
	ScopeStack []Scope
}

type AST interface {
	Repr() string
	Eval(*Interpreter) (interface{}, error)
}

// binary operator
type Binop struct {
	Op    string
	Left  AST
	Right AST
}

func (self Binop) Repr() string {
	return fmt.Sprintf("(%s %s %s)", self.Op, self.Left.Repr(), self.Right.Repr())
}

// function call
type DotOp struct {
	Left AST
	Attr string
}

func (self DotOp) Repr() string {
	return fmt.Sprintf("(. %s %s)", self.Left.Repr(), self.Attr)
}

// function call
type funcallArg struct {
	argName string
	arg     AST
}

type FunCall struct {
	FunRef      AST
	Args        []funcallArg
	keywordArgs bool
}

func (self FunCall) Repr() string {
	argReprs := make([]string, 0)
	if self.keywordArgs {
		for _, arg := range self.Args {
			s := fmt.Sprintf("(%s %s)", arg.argName, arg.arg.Repr())
			argReprs = append(argReprs, s)
		}
	} else {
		for _, arg := range self.Args {
			argReprs = append(argReprs, arg.arg.Repr())
		}
	}
	return fmt.Sprintf("(call %s [%s])", self.FunRef.Repr(), strings.Join(argReprs, ", "))
}

// function definition
type FunDef struct {
	Args []string
	Body AST
}

func (self FunDef) Repr() string {
	return fmt.Sprintf("(function [%s] %s)", strings.Join(self.Args, ", "), self.Body.Repr())
}

// variable
type Var struct {
	Name string
}

func (self Var) Repr() string {
	if strings.Contains(self.Name, " ") {
		return fmt.Sprintf("`%s`", self.Name)
	}
	return self.Name
}

// number
type NumberNode struct {
	Value string
}

func (self NumberNode) Repr() string {
	return self.Value
}

// bool
type BoolNode struct {
	Value bool
}

func (self BoolNode) Repr() string {
	if self.Value {
		return "true"
	} else {
		return "false"
	}
}

// null
type NullNode struct {
}

func (self NullNode) Repr() string {
	return "null"
}

// string
type StringNode struct {
	Value string
}

func (self StringNode) Repr() string {
	return self.Value
}
func (self StringNode) Content() string {
	// trim leading and trailing quotes
	s := self.Value[1 : len(self.Value)-1]

	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	return s
}

// Map

type mapItem struct {
	Name  string
	Value AST
}

type MapNode struct {
	Values []mapItem
}

func (self MapNode) Repr() string {
	var ss []string
	for _, item := range self.Values {
		s := fmt.Sprintf("(\"%s\" %s)", item.Name, item.Value.Repr())
		ss = append(ss, s)
	}
	return fmt.Sprintf("(map %s)", strings.Join(ss, " "))
}

// temporal
type TemporalVal struct {
	Value string
}

func (self TemporalVal) Repr() string {
	return self.Value
}

// range
type RangeNode struct {
	StartOpen bool
	Start     AST

	EndOpen bool
	End     AST
}

func (self RangeNode) Repr() string {
	startQuote := "["
	if self.StartOpen {
		startQuote = "("
	}
	endQuote := "]"
	if self.EndOpen {
		endQuote = ")"
	}
	return fmt.Sprintf("%s%s..%s%s", startQuote, self.Start.Repr(), self.End.Repr(), endQuote)
}

// if expression
type IfExpr struct {
	Cond       AST
	ThenBranch AST
	ElseBranch AST
}

func (self IfExpr) Repr() string {
	return fmt.Sprintf("(if %s then %s else %s)", self.Cond.Repr(), self.ThenBranch.Repr(), self.ElseBranch.Repr())
}

// array
type ArrayNode struct {
	Elements []AST
}

func (self ArrayNode) Repr() string {
	s := make([]string, 0)
	for _, elem := range self.Elements {
		s = append(s, elem.Repr())
	}
	return fmt.Sprintf("[%s]", strings.Join(s, ", "))
}

// ExpressList
type ExprList struct {
	Elements []AST
}

func (self ExprList) Repr() string {
	s := make([]string, 0)
	for _, elem := range self.Elements {
		s = append(s, elem.Repr())
	}
	return fmt.Sprintf("(explist %s)", strings.Join(s, " "))
}

// MultiTests
type MultiTests struct {
	Elements []AST
}

func (self MultiTests) Repr() string {
	s := make([]string, 0)
	for _, elem := range self.Elements {
		s = append(s, elem.Repr())
	}
	return fmt.Sprintf("(multitests %s)", strings.Join(s, " "))
}

// For expression
type ForExpr struct {
	Varname    string
	ListExpr   AST
	ReturnExpr AST
}

func (self ForExpr) Repr() string {
	return fmt.Sprintf("(for %s %s %s)", self.Varname, self.ListExpr.Repr(), self.ReturnExpr.Repr())
}

// Some expression
type SomeExpr struct {
	Varname    string
	ListExpr   AST
	FilterExpr AST
}

func (self SomeExpr) Repr() string {
	return fmt.Sprintf("(some \"%s\" %s %s)", self.Varname, self.ListExpr.Repr(), self.FilterExpr.Repr())
}

// Every expression
type EveryExpr struct {
	Varname    string
	ListExpr   AST
	FilterExpr AST
}

func (self EveryExpr) Repr() string {
	return fmt.Sprintf("(every \"%s\" %s %s)", self.Varname, self.ListExpr.Repr(), self.FilterExpr.Repr())
}
