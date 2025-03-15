package feel

import (
	"fmt"
	"strings"
)

type Scope map[string]any

type Interpreter struct {
	ScopeStack []Scope
}

type Node interface {
	Repr() string
	Eval(*Interpreter) (any, error)
	TextRange() TextRange
}

type HasAttrs interface {
	GetAttr(name string) (any, bool)
}

type TextRange struct {
	Start ScanPosition
	End   ScanPosition
}

// binary operator
type Binop struct {
	Op    string
	Left  Node
	Right Node

	textRange TextRange
}

func (op Binop) TextRange() TextRange {
	return op.textRange
}
func (op Binop) Repr() string {
	return fmt.Sprintf("(%s %s %s)", op.Op, op.Left.Repr(), op.Right.Repr())
}

// function call
type DotOp struct {
	Left Node
	Attr string

	textRange TextRange
}

func (op DotOp) TextRange() TextRange {
	return op.textRange
}

func (op DotOp) Repr() string {
	return fmt.Sprintf("(. %s %s)", op.Left.Repr(), op.Attr)
}

// function call
type funcallArg struct {
	argName string
	arg     Node
}

type FunCall struct {
	FunRef      Node
	Args        []funcallArg
	keywordArgs bool

	textRange TextRange
}

func (fc FunCall) TextRange() TextRange {
	return fc.textRange
}
func (fc FunCall) Repr() string {
	argReprs := make([]string, 0)
	if fc.keywordArgs {
		for _, arg := range fc.Args {
			s := fmt.Sprintf("(%s %s)", arg.argName, arg.arg.Repr())
			argReprs = append(argReprs, s)
		}
	} else {
		for _, arg := range fc.Args {
			argReprs = append(argReprs, arg.arg.Repr())
		}
	}
	return fmt.Sprintf("(call %s [%s])", fc.FunRef.Repr(), strings.Join(argReprs, ", "))
}

// function definition
type FunDef struct {
	Args []string
	Body Node

	textRange TextRange
}

func (fdef FunDef) TextRange() TextRange {
	return fdef.textRange
}

func (fdef FunDef) Repr() string {
	return fmt.Sprintf("(function [%s] %s)", strings.Join(fdef.Args, ", "), fdef.Body.Repr())
}

// variable
type Var struct {
	Name      string
	textRange TextRange
}

func (v Var) TextRange() TextRange {
	return v.textRange
}

func (v Var) Repr() string {
	if strings.Contains(v.Name, " ") {
		return fmt.Sprintf("`%s`", v.Name)
	}
	return v.Name
}

// number
type NumberNode struct {
	Value string

	textRange TextRange
}

func (node NumberNode) TextRange() TextRange {
	return node.textRange
}

func (node NumberNode) Repr() string {
	return node.Value
}

// bool
type BoolNode struct {
	Value bool

	textRange TextRange
}

func (node BoolNode) TextRange() TextRange {
	return node.textRange
}
func (node BoolNode) Repr() string {
	if node.Value {
		return "true"
	} else {
		return "false"
	}
}

// null
type NullNode struct {
	textRange TextRange
}

func (node NullNode) Repr() string {
	return "null"
}

func (node NullNode) TextRange() TextRange {
	return node.textRange
}

// string
type StringNode struct {
	Value string

	textRange TextRange
}

func (node StringNode) Repr() string {
	return node.Value
}
func (node StringNode) TextRange() TextRange {
	return node.textRange
}
func (node StringNode) Content() string {
	// trim leading and trailing quotes
	s := node.Value[1 : len(node.Value)-1]

	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	return s
}

// Map

type mapItem struct {
	Name  string
	Value Node
}

type MapNode struct {
	Values []mapItem

	textRange TextRange
}

func (node MapNode) TextRange() TextRange {
	return node.textRange
}
func (node MapNode) Repr() string {
	var ss []string
	for _, item := range node.Values {
		s := fmt.Sprintf("(\"%s\" %s)", item.Name, item.Value.Repr())
		ss = append(ss, s)
	}
	return fmt.Sprintf("(map %s)", strings.Join(ss, " "))
}

// temporal
type TemporalNode struct {
	Value     string
	textRange TextRange
}

func (node TemporalNode) TextRange() TextRange {
	return node.textRange
}
func (node TemporalNode) Repr() string {
	return node.Value
}

func (node TemporalNode) Content() string {
	return node.Value[2 : len(node.Value)-1]
}

// range
type RangeNode struct {
	StartOpen bool
	Start     Node

	EndOpen bool
	End     Node

	textRange TextRange
}

func (node RangeNode) TextRange() TextRange {
	return node.textRange
}
func (node RangeNode) Repr() string {
	startQuote := "["
	if node.StartOpen {
		startQuote = "("
	}
	endQuote := "]"
	if node.EndOpen {
		endQuote = ")"
	}
	return fmt.Sprintf("%s%s..%s%s", startQuote, node.Start.Repr(), node.End.Repr(), endQuote)
}

// if expression
type IfExpr struct {
	Cond       Node
	ThenBranch Node
	ElseBranch Node

	textRange TextRange
}

func (node IfExpr) TextRange() TextRange {
	return node.textRange
}
func (node IfExpr) Repr() string {
	return fmt.Sprintf("(if %s %s %s)", node.Cond.Repr(), node.ThenBranch.Repr(), node.ElseBranch.Repr())
}

// array
type ArrayNode struct {
	Elements []Node

	textRange TextRange
}

func (node ArrayNode) TextRange() TextRange {
	return node.textRange
}
func (node ArrayNode) Repr() string {
	s := make([]string, 0)
	for _, elem := range node.Elements {
		s = append(s, elem.Repr())
	}
	return fmt.Sprintf("[%s]", strings.Join(s, ", "))
}

// ExpressList
type ExprList struct {
	Elements []Node

	textRange TextRange
}

func (node ExprList) TextRange() TextRange {
	return node.textRange
}
func (node ExprList) Repr() string {
	s := make([]string, 0)
	for _, elem := range node.Elements {
		s = append(s, elem.Repr())
	}
	return fmt.Sprintf("(explist %s)", strings.Join(s, " "))
}

// MultiTests
type MultiTests struct {
	Elements  []Node
	textRange TextRange
}

func (node MultiTests) TextRange() TextRange {
	return node.textRange
}
func (node MultiTests) Repr() string {
	s := make([]string, 0)
	for _, elem := range node.Elements {
		s = append(s, elem.Repr())
	}
	return fmt.Sprintf("(multitests %s)", strings.Join(s, " "))
}

// For expression
type ForExpr struct {
	Varname    string
	ListExpr   Node
	ReturnExpr Node
	textRange  TextRange
}

func (node ForExpr) TextRange() TextRange {
	return node.textRange
}
func (node ForExpr) Repr() string {
	return fmt.Sprintf("(for %s %s %s)", node.Varname, node.ListExpr.Repr(), node.ReturnExpr.Repr())
}

// Some expression
type SomeExpr struct {
	Varname    string
	ListExpr   Node
	FilterExpr Node
	textRange  TextRange
}

func (node SomeExpr) TextRange() TextRange {
	return node.textRange
}
func (node SomeExpr) Repr() string {
	return fmt.Sprintf("(some \"%s\" %s %s)", node.Varname, node.ListExpr.Repr(), node.FilterExpr.Repr())
}

// Every expression
type EveryExpr struct {
	Varname    string
	ListExpr   Node
	FilterExpr Node

	textRange TextRange
}

func (node EveryExpr) TextRange() TextRange {
	return node.textRange
}
func (node EveryExpr) Repr() string {
	return fmt.Sprintf("(every \"%s\" %s %s)", node.Varname, node.ListExpr.Repr(), node.FilterExpr.Repr())
}
