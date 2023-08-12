package feel

// for FEEL syntax refer to https://learn-dmn-in-15-minutes.com/learn/the-feel-language.html
// for BNF forms and handbook refer to https://kiegroup.github.io/dmn-feel-handbook

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

type UnexpectedToken struct {
	token   ScannerToken
	callers []string
	expects []string
}

func NewUnexpectedToken(token ScannerToken, callers []string, expects []string) *UnexpectedToken {
	return &UnexpectedToken{token: token, callers: callers, expects: expects}
}

func (self UnexpectedToken) Error() string {
	return fmt.Sprintf(
		"unexpected %s %s, at %d %d, expect %s\ncallers:\n%s\n",
		self.token.Kind, self.token.Value,
		self.token.Pos.Row, self.token.Pos.Column,
		strings.Join(self.expects, ", "),
		strings.Join(self.callers, "\n"),
	)
}

func hasDupName(names []string) (bool, string) {
	nameSet := make(map[string]bool)
	for _, name := range names {
		if _, ok := nameSet[name]; ok {
			return true, name
		}
		nameSet[name] = true
	}
	return false, ""
}

func ParseString(input string) (Node, error) {
	parser := NewParser(NewScanner(input))
	return parser.Parse()
}

type Parser struct {
	scanner *Scanner
}

func NewParser(scanner *Scanner) *Parser {
	return &Parser{
		scanner: scanner,
	}
}

func (self Parser) Unexpected(expects ...string) *UnexpectedToken {
	// extract caller stack dump
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	var callers []string
	if n > 0 {
		pc = pc[:n]
		frames := runtime.CallersFrames(pc)
		for {
			frame, more := frames.Next()
			callers = append(callers, fmt.Sprintf("%s:%d", frame.Function, frame.Line))
			if !more {
				break
			}
		}
	}
	return NewUnexpectedToken(self.CurrentToken(), callers, expects)
}

func (self Parser) CurrentToken() ScannerToken {
	return self.scanner.Current()
}

func (self *Parser) Parse() (Node, error) {
	self.scanner.Next()
	var exps []Node

	for !self.CurrentToken().Expect(TokenEOF) {
		if self.CurrentToken().Expect(";") {
			self.scanner.Next()
		} else {
			exp, err := self.parseUnaryTest()
			if err != nil {
				return nil, err
			}
			exps = append(exps, exp)
		}
	}

	if len(exps) == 1 {
		return exps[0], nil
	} else {
		return &ExprList{
			Elements: exps,
		}, nil
	}
}

func (self Parser) startTextRange() TextRange {
	return TextRange{Start: self.CurrentToken().Pos}
}

func (self *Parser) parseUnaryTestElement() (Node, error) {
	if self.CurrentToken().Expect(">", ">=", "<", "<=", "!=", "=") {
		textRange := self.startTextRange()
		op := self.CurrentToken().Kind
		self.scanner.Next()
		right, err := self.expression()
		if err != nil {
			return nil, err
		}
		textRange.End = self.CurrentToken().Pos
		exp := &Binop{
			Left:      &Var{Name: "?"},
			Op:        op,
			Right:     right,
			textRange: textRange,
		}
		return exp, nil
	} else {
		return self.expression()
	}
}

func (self *Parser) parseUnaryTest() (Node, error) {
	textRange := self.startTextRange()
	exp, err := self.parseUnaryTestElement()
	if err != nil {
		return nil, err
	}

	if self.CurrentToken().Expect(",") {
		elements := []Node{exp}
		for self.CurrentToken().Expect(",") {
			self.scanner.Next()

			uexp, err := self.parseUnaryTestElement()
			if err != nil {
				return nil, err
			}
			elements = append(elements, uexp)
		}
		textRange.End = self.CurrentToken().Pos
		return &MultiTests{Elements: elements, textRange: textRange}, nil
	} else {
		return exp, nil
	}
}

func (self *Parser) expression() (Node, error) {
	return self.inOp()
}

type astFunc func() (Node, error)

func (self *Parser) binop(ops []string, subfunc astFunc) (Node, error) {
	left, err := subfunc()
	if err != nil {
		return nil, err
	}

	for self.CurrentToken().Expect(ops...) {
		op := self.CurrentToken().Kind
		self.scanner.Next()

		right, err := subfunc()
		if err != nil {
			return nil, err
		}
		textRange := TextRange{Start: left.TextRange().Start}
		textRange.End = self.CurrentToken().Pos
		left = &Binop{Op: op, Left: left, Right: right, textRange: textRange}
	}
	return left, nil
}

func (self *Parser) binopKeywords(ops []string, subfunc astFunc) (Node, error) {
	left, err := subfunc()
	if err != nil {
		return nil, err
	}

	for self.CurrentToken().ExpectKeywords(ops...) {
		op := self.CurrentToken().Value
		self.scanner.Next()

		right, err := subfunc()
		if err != nil {
			return nil, err
		}
		textRange := TextRange{Start: left.TextRange().Start}
		textRange.End = self.CurrentToken().Pos

		left = &Binop{Op: op, Left: left, Right: right, textRange: textRange}
	}
	return left, nil
}

// pase chains
func (self *Parser) inOp() (Node, error) {
	return self.binopKeywords(
		[]string{"in"},
		self.logicOrOp,
	)
}

func (self *Parser) logicOrOp() (Node, error) {
	return self.binopKeywords(
		[]string{"or"},
		self.logicAndOp,
	)
}

func (self *Parser) logicAndOp() (Node, error) {
	return self.binopKeywords(
		[]string{"and"},
		self.compareOp,
	)
}

func (self *Parser) compareOp() (Node, error) {
	return self.binop(
		[]string{">", ">=", "<", "<=", "!=", "="},
		self.addOrSubOp,
	)
}

func (self *Parser) addOrSubOp() (Node, error) {
	return self.binop(
		[]string{"+", "-"},
		self.mulOrDivOp,
	)
}

func (self *Parser) mulOrDivOp() (Node, error) {
	return self.binop(
		[]string{"*", "/", "%"},
		self.parseFuncallOrIndexOrDot,
	)
}

func (self *Parser) parseFuncallOrIndexOrDot() (Node, error) {
	exp, err := self.singleElement()
	if err != nil {
		return nil, err
	}
	for {
		switch self.CurrentToken().Kind {
		case "(":
			nexp, err := self.parseFuncallRest(exp)
			if err != nil {
				return nil, err
			}
			exp = nexp
		case "[":
			nexp, err := self.parseIndexRest(exp)
			if err != nil {
				return nil, err
			}
			exp = nexp
		case ".":
			nexp, err := self.parseDotRest(exp)
			if err != nil {
				return nil, err
			}
			exp = nexp
		default:
			return exp, nil
		}
	}
}

var funcallTrailing = regexp.MustCompile(`\s*\($`)

// func (self *Parser) parseFuncall() (Node, error) {
// 	funcallWithRbracket := self.CurrentToken().Value
// 	funcName := funcallTrailing.ReplaceAllString(funcallWithRbracket, "")
// 	textRange := TextRange{Start: Node.TextRange().Start, End: self.CurrentToken().Pos}
// 	return self.parseFuncallRest(&Var{Name: funcName, textRange: })

// }

func (self *Parser) parseFunccallArg() (funcallArg, error) {
	arg, err := self.expression()
	if err != nil {
		return funcallArg{}, err
	}

	if self.CurrentToken().Expect(":") { // kwargs
		if varArg, ok := arg.(*Var); ok {
			self.scanner.Next()
			argValue, err := self.expression()
			if err != nil {
				return funcallArg{}, err
			}
			return funcallArg{argName: varArg.Name, arg: argValue}, nil
		} else {
			return funcallArg{}, self.Unexpected("var")
		}
	} else {
		return funcallArg{argName: "", arg: arg}, nil
	}
}

func (self *Parser) parseFuncallRest(funExpr Node) (Node, error) {
	self.scanner.Next()
	// parse function arguments
	var args []funcallArg = nil
	keywordArgs := false
	for !self.CurrentToken().Expect(")") {
		arg, err := self.parseFunccallArg()
		if err != nil {
			return nil, err
		}
		if !keywordArgs && arg.argName != "" {
			keywordArgs = true
		}
		if len(args) > 0 {
			if arg.argName != "" && args[0].argName == "" {
				return nil, self.Unexpected("non var")
			}
			if arg.argName == "" && args[0].argName != "" {
				return nil, self.Unexpected("var")
			}
		}
		args = append(args, arg)
		if self.CurrentToken().Expect(",") {
			self.scanner.Next()
		} else if !self.CurrentToken().Expect(")") {
			return nil, self.Unexpected(",", ")")
		}
	}

	if self.CurrentToken().Expect(")") {
		self.scanner.Next()
	}

	textRange := TextRange{Start: funExpr.TextRange().Start, End: self.CurrentToken().Pos}
	return &FunCall{
		FunRef:      funExpr,
		Args:        args,
		keywordArgs: keywordArgs,
		textRange:   textRange,
	}, nil
}

func (self *Parser) parseIndexRest(exp Node) (Node, error) {
	self.scanner.Next()

	// parse index arguments
	at, err := self.expression()
	if err != nil {
		return nil, err
	}
	if !self.CurrentToken().Expect("]") {
		return nil, self.Unexpected("]")
	}

	self.scanner.Next()
	textRange := TextRange{Start: exp.TextRange().Start, End: self.CurrentToken().Pos}

	return &Binop{Left: exp, Op: "[]", Right: at, textRange: textRange}, nil
}

func (self *Parser) parseDotRest(exp Node) (Node, error) {
	self.scanner.Next()
	// parse index arguments
	attr, err := self.parseName()
	if err != nil {
		return nil, err
	}
	textRange := TextRange{Start: exp.TextRange().Start, End: self.CurrentToken().Pos}
	return &DotOp{Left: exp, Attr: attr, textRange: textRange}, nil
}

func (self *Parser) singleElement() (Node, error) {
	curr := self.CurrentToken()
	switch curr.Kind {
	case TokenName:
		return self.parseVar()
	// case TokenFuncall:
	// 	return self.parseFuncall()
	case TokenNumber:
		return self.parseNumberNode()
	case TokenString:
		return self.parseStringNode()
	case TokenTemporal:
		return self.parseTemporalNode()
	case "(":
		return self.parseBracketOrRange()
	case "[":
		return self.parseRangeOrArray()
	case "{":
		return self.parseMapNode()
	case "?":
		return &Var{Name: "?"}, nil
	case TokenKeyword:
		switch curr.Value {
		case "true":
			return self.parseBool()
		case "false":
			return self.parseBool()
		case "null":
			return self.parseNull()
		case "if":
			return self.parseIfExpression()
		case "for":
			return self.parseForExpr()
		case "function":
			return self.parseFunDef()
		case "some":
			return self.parseSomeOrEvery()
		case "every":
			return self.parseSomeOrEvery()
		default:
			//return nil, self.Unexpected("keywords")
			// unexpected keywords can be part of names
			return self.parseVar()
		}
	default:
		return nil, self.Unexpected("name", "number", "string", "(", "[", "keyword")
	}
}

func (self *Parser) parseVar() (Node, error) {
	textRange := self.startTextRange()
	name, err := self.parseName()
	if err != nil {
		return nil, err
	}
	textRange.End = self.CurrentToken().Pos
	return &Var{Name: name, textRange: textRange}, nil
}

func (self *Parser) parseBool() (Node, error) {
	textRange := self.startTextRange()
	v := self.CurrentToken().Value
	self.scanner.Next()
	textRange.End = self.CurrentToken().Pos
	switch v {
	case "true":
		return &BoolNode{Value: true, textRange: textRange}, nil
	case "false":
		return &BoolNode{Value: false, textRange: textRange}, nil
	default:
		return nil, self.Unexpected("true", "false")
	}
}

func (self *Parser) parseNull() (Node, error) {
	textRange := self.startTextRange()
	self.scanner.Next()
	textRange.End = self.CurrentToken().Pos
	return &NullNode{textRange: textRange}, nil
}

func containsKeywords(keywords []string, kw string) bool {
	for _, stopKw := range keywords {
		if stopKw == kw {
			return true
		}
	}
	return false
}

func (self *Parser) parseName(stopKeywords ...string) (string, error) {
	names := make([]string, 0)

	for self.CurrentToken().Expect(TokenName, TokenKeyword) {
		if self.CurrentToken().Kind == "name" {
			names = append(names, self.CurrentToken().Value)
			self.scanner.Next()
		} else if self.CurrentToken().Kind == TokenKeyword {
			// keyworlds
			//if self.CurrentToken()
			kwVal := self.CurrentToken().Value
			if len(names) > 0 && containsKeywords(stopKeywords, kwVal) {
				break
			} else {
				names = append(names, kwVal)
				self.scanner.Next()
			}
		} else {
			break
		}
	}
	if len(names) <= 0 {
		return "", self.Unexpected(TokenName)
	}
	return strings.Join(names, " "), nil
}

func (self *Parser) parseBracketOrRange() (Node, error) {
	textRange := self.startTextRange()
	self.scanner.Next()
	c, err := self.expression()
	if err != nil {
		return nil, err
	}
	if self.CurrentToken().Kind == ".." {
		self.scanner.Next()
		d, err := self.expression()
		if err != nil {
			return nil, err
		}

		if self.CurrentToken().Kind == ")" {
			self.scanner.Next()
			textRange.End = self.CurrentToken().Pos
			return &RangeNode{StartOpen: true, Start: c, EndOpen: true, End: d, textRange: textRange}, nil
		} else if self.CurrentToken().Kind == "]" {
			self.scanner.Next()
			textRange.End = self.CurrentToken().Pos
			return &RangeNode{StartOpen: true, Start: c, EndOpen: false, End: d, textRange: textRange}, nil
		}
		return nil, self.Unexpected(")", "]")
	} else if self.CurrentToken().Expect(")") {
		self.scanner.Next()
	} else {
		return nil, self.Unexpected(")")
	}
	return c, nil
}

func (self *Parser) parseRangeOrArray() (Node, error) {
	rng := self.startTextRange()
	prefixKind := self.CurrentToken().Kind // prefixKind is '['
	self.scanner.Next()
	if self.CurrentToken().Expect("]") {
		self.scanner.Next()
		// empty array
		return &ArrayNode{}, nil
	}
	c, err := self.expression()
	if err != nil {
		return nil, err
	}

	if self.CurrentToken().Expect(",", "]") {
		return self.parseArrayGivenFirst(prefixKind, c)
	}

	if !self.CurrentToken().Expect("..") {
		return nil, self.Unexpected("..")
	}
	self.scanner.Next()
	d, err := self.expression()
	if err != nil {
		return nil, err
	}

	startOpen := prefixKind == "("
	if self.CurrentToken().Kind == ")" {
		self.scanner.Next()
		rng.End = self.CurrentToken().Pos
		return &RangeNode{StartOpen: startOpen, Start: c, EndOpen: true, End: d, textRange: rng}, nil
	} else if self.CurrentToken().Kind == "]" {
		self.scanner.Next()
		rng.End = self.CurrentToken().Pos
		return &RangeNode{StartOpen: startOpen, Start: c, EndOpen: false, End: d, textRange: rng}, nil
	}
	return nil, self.Unexpected(")", "]")
}

func (self *Parser) parseArrayGivenFirst(prefixKind string, firstElem Node) (Node, error) {
	rng := self.startTextRange()
	elements := []Node{firstElem}
	for self.CurrentToken().Expect(",") {
		self.scanner.Next()
		elem, err := self.expression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
	}
	if !self.CurrentToken().Expect("]") {
		return nil, self.Unexpected("]")
	}
	self.scanner.Next()
	rng.End = self.CurrentToken().Pos
	return &ArrayNode{Elements: elements, textRange: rng}, nil
}

func (self *Parser) parseNumberNode() (Node, error) {
	rng := self.startTextRange()
	v := self.CurrentToken().Value
	self.scanner.Next()
	rng.End = self.CurrentToken().Pos
	return &NumberNode{Value: v, textRange: rng}, nil
}

func (self *Parser) parseStringNode() (Node, error) {
	rng := self.startTextRange()
	v := self.CurrentToken().Value
	self.scanner.Next()
	rng.End = self.CurrentToken().Pos
	return &StringNode{Value: v, textRange: rng}, nil
}

func (self *Parser) parseMapKey() (string, error) {
	switch self.CurrentToken().Kind {
	case TokenName:
		return self.parseName()
	case TokenString:
		node, err := self.parseStringNode()
		if err != nil {
			return "", err
		}
		return node.(*StringNode).Content(), nil
	default:
		return "", self.Unexpected(TokenName, TokenString)
	}
}

func (self *Parser) parseTemporalNode() (Node, error) {
	rng := self.startTextRange()
	v := self.CurrentToken().Value
	self.scanner.Next()
	rng.End = self.CurrentToken().Pos
	return &TemporalNode{Value: v, textRange: rng}, nil
}

func (self *Parser) parseMapNode() (Node, error) {
	rng := self.startTextRange()
	self.scanner.Next()
	var mapValues []mapItem

	for !self.CurrentToken().Expect("}") {
		key, err := self.parseMapKey()
		if err != nil {
			return nil, err
		}

		if !self.CurrentToken().Expect(":") {
			return nil, self.Unexpected(":")
		}
		self.scanner.Next()

		exp, err := self.expression()
		if err != nil {
			return nil, err
		}

		mapValues = append(mapValues, mapItem{Name: key, Value: exp})

		if self.CurrentToken().Expect(",") {
			self.scanner.Next()
		} else if !self.CurrentToken().Expect("}") {
			return nil, self.Unexpected(",", "}")
		}
	}
	if self.CurrentToken().Expect("}") {
		self.scanner.Next()
	}
	rng.End = self.CurrentToken().Pos
	return &MapNode{Values: mapValues, textRange: rng}, nil
}

func (self *Parser) parseIfExpression() (Node, error) {
	rng := self.startTextRange()
	self.scanner.Next()
	cond, err := self.expression()
	if err != nil {
		return nil, err
	}
	if !self.CurrentToken().ExpectKeywords("then") {
		return nil, self.Unexpected("then")
	}
	self.scanner.Next()

	then_branch, err := self.expression()
	if err != nil {
		return nil, err
	}
	if !self.CurrentToken().ExpectKeywords("else") {
		return nil, self.Unexpected("else")
	}
	self.scanner.Next()

	else_branch, err := self.expression()
	if err != nil {
		return nil, err
	}

	rng.End = self.CurrentToken().Pos
	return &IfExpr{Cond: cond, ThenBranch: then_branch, ElseBranch: else_branch, textRange: rng}, nil

}

func (self *Parser) parseForExpr() (Node, error) {
	rng := self.startTextRange()
	self.scanner.Next()
	varName, err := self.parseName("in", "for")

	if !self.CurrentToken().ExpectKeywords("in") {
		return nil, self.Unexpected("in")
	}
	self.scanner.Next()

	listExpr, err := self.expression()
	if err != nil {
		return nil, err
	}
	//fmt.Printf("list expr %s\n", listExpr.Repr())

	if self.CurrentToken().Expect(",") {
		returnExpr, err := self.parseForExpr()
		if err != nil {
			return nil, err
		}
		return &ForExpr{
			Varname:    varName,
			ListExpr:   listExpr,
			ReturnExpr: returnExpr,
		}, nil
	}

	if !self.CurrentToken().ExpectKeywords("return") {
		return nil, self.Unexpected("return")
	}
	self.scanner.Next()
	//fmt.Printf("return\n")

	returnExpr, err := self.expression()
	if err != nil {
		return nil, err
	}
	rng.End = self.CurrentToken().Pos
	return &ForExpr{
		Varname:    varName,
		ListExpr:   listExpr,
		ReturnExpr: returnExpr,
		textRange:  rng,
	}, nil
}

func (self *Parser) parseSomeOrEvery() (Node, error) {
	rng := self.startTextRange()
	cmd := self.CurrentToken().Value
	self.scanner.Next()
	// parse variable name
	varName, err := self.parseName("in")
	if err != nil {
		return nil, err
	}

	if !self.CurrentToken().ExpectKeywords("in") {
		return nil, self.Unexpected("in")
	}
	self.scanner.Next()

	listExpr, err := self.expression()
	if err != nil {
		return nil, err
	}

	if !self.CurrentToken().ExpectKeywords("satisfies") {
		return nil, self.Unexpected("satisfies")
	}
	self.scanner.Next()

	filterExpr, err := self.expression()
	if err != nil {
		return nil, err
	}
	rng.End = self.CurrentToken().Pos
	if cmd == "some" {
		return &SomeExpr{
			Varname:    varName,
			ListExpr:   listExpr,
			FilterExpr: filterExpr,
			textRange:  rng,
		}, nil
	} else {
		return &EveryExpr{
			Varname:    varName,
			ListExpr:   listExpr,
			FilterExpr: filterExpr,
			textRange:  rng,
		}, nil
	}

}

func (self *Parser) parseFunDef() (Node, error) {
	rng := self.startTextRange()
	self.scanner.Next()
	if !self.CurrentToken().Expect("(") {
		return nil, self.Unexpected("(")
	}
	self.scanner.Next()

	// parse var list
	var args []string
	for !self.CurrentToken().Expect(")") {
		argName, err := self.parseName()
		if err != nil {
			return nil, err
		}

		args = append(args, argName)

		if self.CurrentToken().Expect(",") {
			self.scanner.Next()
		} else if !self.CurrentToken().Expect(")") {
			return nil, self.Unexpected(")", ",")
		}
	}
	if isdup, name := hasDupName(args); isdup {
		return nil, errors.New(fmt.Sprintf("function arg name '%s' duplicates", name))
	}

	if self.CurrentToken().Expect(")") {
		self.scanner.Next()
	}

	exp, err := self.expression()
	if err != nil {
		return nil, err
	}
	rng.End = self.CurrentToken().Pos
	return &FunDef{
		Args:      args,
		Body:      exp,
		textRange: rng,
	}, nil
}
