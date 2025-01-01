package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
)

func IsAsciiDigit(c byte) bool {
	return c >= byte('0') && c <= byte('9')
}

func IsAsciiAlphabetic(c byte) bool {
	return (c >= byte('a') && c <= byte('z')) || (c >= byte('A') && c <= byte('Z'))
}

func RemoveIndex[T any](s []T, index int) []T {
	return append(s[:index], s[:index+1]...)
}

type Stack[T any] struct {
	inner []T
}

func NewStack[T any]() Stack[T] {
	return Stack[T]{
		inner: []T{},
	}
}

func (s *Stack[T]) Size() int {
	return len(s.inner)
}

func (s *Stack[T]) Push(items ...T) {
	s.inner = append(s.inner, items...)
}

func (s *Stack[T]) Pop() T {
	index := len(s.inner) - 1
	item := s.inner[index]
	s.inner = s.inner[:index]
	return item
}

//go:generate stringer -type=TokenType
type TokenType int

const (
	TokEof TokenType = iota
	TokOp
	TokParOpen
	TokParClose
	TokNumber
	TokIdent
)

//go:generate stringer -type=OpType
type OpType int

const (
	OpMul  OpType = 0x01
	OpFDiv OpType = 0x02
	OpFMod OpType = 0x03
	OpIDiv OpType = 0x04
	OpAdd  OpType = 0x10
	OpSub  OpType = 0x11
	OpMaxP OpType = 0x20
)

type Token struct {
	kind  TokenType
	pos   int
	extra any
}

func (t Token) String() string {
	s := fmt.Sprintf("{ %s @ %d", t.kind, t.pos)

	switch t.kind {
	case TokOp:
		s = fmt.Sprintf("%s, %s", s, t.extra.(OpType))
	case TokNumber:
		s = fmt.Sprintf("%s (%d)", s, t.extra)
	case TokIdent:
		s = fmt.Sprintf("%s `%s`", s, t.extra)
	}

	return fmt.Sprintf("%s }", s)
}

//go:generate stringer -type=Unary
type Unary int

const (
	UnNegate Unary = 1
)

//go:generate stringer -type=BuildFlag
type BuildFlag int

const (
	FlagNoOps BuildFlag = 1
)

//go:generate stringer -type=NodeType
type NodeType int

const (
	NodeNumber NodeType = iota + 1
	NodeIdent
	NodeUnOp
	NodeBinOp
)

type Node struct {
	kind    NodeType
	pos     int
	content []any
}

func (n Node) String() string {
	return fmt.Sprintf("<%s | content: %v>", n.kind, n.content)
}

//go:generate stringer -type ActionType
type ActionType int

const (
	NUM ActionType = iota + 1
	IDENT
	UNARY
	BINARY
)

type Action struct {
	kind ActionType
	pos  int
	arg  any
}

func (a Action) String() string {
	return fmt.Sprintf("$%s %v;", a.kind, a.arg)
}

type Txr struct {
	tokens      []Token
	error       string
	buildPos    int
	buildLen    int
	buildNode   Node
	compileList []Action
}

func NewTxr() Txr {
	return Txr{
		tokens:   []Token{},
		error:    "",
		buildPos: 0,
		buildLen: 0,
	}
}

func (txr *Txr) Throw(msg string, pos any) bool {
	txr.error = fmt.Sprintf("%s at position %v", msg, pos)
	fmt.Println(txr.error)
	return true
}

func (txr *Txr) ThrowAt(msg string, tk Token) bool {
	if tk.kind == TokEof {
		return txr.Throw(msg, "<EOF>")
	}
	return txr.Throw(msg, tk.pos)
}

func (txr *Txr) Parse(str string) bool {
	pos := 0
	length := len(str)
	out := &txr.tokens

	for pos < length {
		start := pos
		char := str[pos]
		pos += 1

		switch char {
		case byte(' '), byte('\t'), byte('\r'), byte('\n'):
			break

		case byte('('):
			*out = append(*out, Token{TokParOpen, start, nil})

		case byte(')'):
			*out = append(*out, Token{TokParClose, start, nil})

		case byte('+'):
			*out = append(*out, Token{TokOp, start, OpAdd})

		case byte('-'):
			*out = append(*out, Token{TokOp, start, OpSub})

		case byte('*'):
			*out = append(*out, Token{TokOp, start, OpMul})

		case byte('/'):
			*out = append(*out, Token{TokOp, start, OpFDiv})

		case byte('%'):
			*out = append(*out, Token{TokOp, start, OpFMod})

		default:
			if IsAsciiDigit(char) {
				for pos < length {
					char = str[pos]
					if IsAsciiDigit(char) {
						pos += 1
					} else {
						break
					}
				}
				numstr := str[start:pos]
				val, err := strconv.ParseFloat(numstr, 64)
				if err != nil {
					panic(err)
				}
				*out = append(*out, Token{TokNumber, start, val})
			} else if char == byte('_') || IsAsciiAlphabetic(char) {
				for pos < length {
					char = str[pos]
					if char == byte('_') || IsAsciiAlphabetic(char) || IsAsciiDigit(char) {
						pos += 1
					} else {
						break
					}
				}
				switch name := str[start:pos]; name {
				case "mod":
					*out = append(*out, Token{TokOp, start, OpFMod})
				case "div":
					*out = append(*out, Token{TokOp, start, OpIDiv})
				default:
					*out = append(*out, Token{TokIdent, start, name})
				}
			} else {
				*out = []Token{}
				msg := fmt.Sprintf("Unexpected character `%c`", char)
				return txr.Throw(msg, start)
			}
		}
	}

	*out = append(*out, Token{TokEof, length, nil})
	return false
}

func (txr *Txr) BuildOps(first Token) bool {
	nodes := []Node{txr.buildNode}
	ops := []Token{first}
	var tk Token
	for {
		if txr.BuildExpr(int(FlagNoOps)) {
			return true
		}
		nodes = append(nodes, txr.buildNode)
		tk = txr.tokens[txr.buildPos]
		if tk.kind == TokOp {
			txr.buildPos += 1
			ops = append(ops, tk)
		} else {
			break
		}
	}
	n := len(ops)
	pmax := int(OpMaxP) >> 4
	pri := 0
	for pri < pmax {
		for i := 0; i < n; i += 1 {
			tk = ops[i]
			op := tk.extra.(OpType)
			if (int(op) >> 4) != pri {
				continue
			}
			nodes[i] = Node{NodeBinOp, tk.pos, []any{op, nodes[i], nodes[i+1]}}
			nodes = RemoveIndex(nodes, i+1)
			ops = RemoveIndex(ops, i)
			n -= 1
			i -= 1
		}
		pri += 1
	}
	txr.buildNode = nodes[0]
	return false
}

func (txr *Txr) BuildExpr(flags int) bool {
	tk := txr.tokens[txr.buildPos]
	txr.buildPos += 1

	switch tk.kind {
	case TokNumber:
		txr.buildNode = Node{NodeNumber, tk.pos, []any{tk.extra.(float64)}}
	case TokIdent:
		txr.buildNode = Node{NodeIdent, tk.pos, []any{tk.extra.(string)}}
	case TokParOpen:
		if txr.BuildExpr(0) {
			return true
		}
		tk = txr.tokens[txr.buildPos]
		txr.buildPos += 1
		if tk.kind != TokParClose {
			return txr.ThrowAt("Expected a `)`", tk)
		}
	case TokOp:
		switch tk.extra.(OpType) {
		case OpAdd:
			if txr.BuildExpr(int(FlagNoOps)) {
				return true
			}
		case OpSub:
			if txr.BuildExpr(int(FlagNoOps)) {
				return true
			}
			txr.buildNode = Node{NodeUnOp, tk.pos, []any{UnNegate, txr.buildNode}}
		default:
			return txr.ThrowAt("Unexpected token", tk)
		}
	default:
		return txr.ThrowAt("Unexpected token", tk)
	}
	if (flags & int(FlagNoOps)) == 0 {
		tk = txr.tokens[txr.buildPos]
		if tk.kind == TokOp {
			txr.buildPos += 1
			if txr.BuildOps(tk) {
				return true
			}
		}
	}

	return false
}

func (txr *Txr) Build() bool {
	txr.buildPos = 0
	txr.buildLen = len(txr.tokens)
	if txr.BuildExpr(0) {
		return true
	}
	if txr.buildPos < txr.buildLen-1 {
		return txr.ThrowAt("Trailing data", txr.tokens[txr.buildPos])
	}
	return false
}

func (txr *Txr) CompileExpr(node Node) bool {
	out := &txr.compileList
	switch node.kind {
	case NodeNumber:
		*out = append(*out, Action{NUM, node.pos, node.content[0].(float64)})
	case NodeIdent:
		*out = append(*out, Action{IDENT, node.pos, node.content[0].(string)})
	case NodeUnOp:
		if txr.CompileExpr(node.content[1].(Node)) {
			return true
		}
		*out = append(*out, Action{UNARY, node.pos, node.content})
	case NodeBinOp:
		if txr.CompileExpr(node.content[1].(Node)) {
			return true
		}
		if txr.CompileExpr(node.content[2].(Node)) {
			return true
		}
		*out = append(*out, Action{BINARY, node.pos, node.content[0].(OpType)})
	default:
		msg := fmt.Sprintf("Cannot compile node type %s", node.kind)
		return txr.ThrowAt(msg, Token{TokenType(node.kind), node.pos, nil})
	}

	return false
}

func (txr *Txr) Compile(source string) ([]Action, error) {
	fmt.Println(source)
	if txr.Parse(source) {
		return []Action{}, errors.New(txr.error)
	}
	fmt.Println(txr.tokens)

	if txr.Build() {
		return []Action{}, errors.New(txr.error)
	}
	fmt.Println(txr.buildNode)

	out := &txr.compileList
	*out = []Action{}
	if txr.CompileExpr(txr.buildNode) {
		return []Action{}, errors.New(txr.error)
	}

	n := len(*out)
	arr := make([]Action, n)
	for i := 0; i < n; i += 1 {
		arr[i] = (*out)[i]
	}
	*out = []Action{}

	fmt.Println(arr)

	return arr, nil
}

func (txr *Txr) ExecExit(msg string, action Action) bool {
	return txr.ThrowAt(msg, Token{TokenType(action.kind), action.pos, nil})
}

func (txr *Txr) Exec(actions []Action) any {
	length := len(actions)
	pos := 0
	stack := NewStack[any]()
	for pos < length {
		action := actions[pos]
		pos += 1
		switch action.kind {
		case NUM:
			stack.Push(action.arg.(float64))
		case UNARY:
			stack.Push(-action.arg.(float64))
		case BINARY:
			b := stack.Pop().(float64)
			a := stack.Pop().(float64)
			switch action.arg.(OpType) {
			case OpAdd:
				a += b
			case OpSub:
				a -= b
			case OpMul:
				a *= b
			case OpFDiv:
				a /= b
			case OpFMod:
				if b != 0 {
					a = math.Mod(a, b)
				} else {
					a = 0
				}
			case OpIDiv:
				if b != 0 {
					a = math.Trunc(a / b)
				} else {
					a = 0
				}
			default:
				msg := fmt.Sprintf("Can't apply operator %s", action.kind)
				return txr.ExecExit(msg, action)
			}
			stack.Push(a)
		case IDENT:
			panic("unimplemented")

		default:
			msg := fmt.Sprintf("Can't run action %s", action.kind)
			return txr.ExecExit(msg, action)
		}
	}
	r := stack.Pop()
	txr.error = ""
	return r
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: txr '<expr>'")
		os.Exit(1)
	}

	source := os.Args[1]
	txr := NewTxr()
	actions, err := txr.Compile(source)
	if err != nil {
		panic(err)
	}
	result := txr.Exec(actions)
	fmt.Println(result)
}
