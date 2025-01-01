package main

import (
	"fmt"
	"strconv"
)

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

//go:generate stringer -type=NodeType
type NodeType int

const (
	NodeNumber NodeType = iota + 1
	NodeIdent
	NodeUnOp
	NodeBinOp
)

type Node struct {
	kind NodeType
	pos  int
	op   any
	lhs  any
	rhs  any
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

func (n Node) String() string {
	return fmt.Sprintf("<%s @ %d | lhs: %v, rhs: %v>", n.kind, n.pos, n.lhs, n.rhs)
}

type Txr struct {
	tokens    []Token
	error     string
	buildPos  int
	buildLen  int
	buildNode Node
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

func IsAsciiDigit(c byte) bool {
	return c >= byte('0') && c <= byte('9')
}

func IsAsciiAlphabetic(c byte) bool {
	return (c >= byte('a') && c <= byte('z')) || (c >= byte('A') && c <= byte('Z'))
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
				val, err := strconv.Atoi(numstr)
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

func RemoveIndex[T any](s []T, index int) []T {
	return append(s[:index], s[:index+1]...)
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
			if (int(tk.extra.(OpType)) >> 4) != pri {
				continue
			}
			nodes[i] = Node{NodeBinOp, tk.pos, nil, nodes[i], nodes[i+1]}
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
		txr.buildNode = Node{NodeNumber, tk.pos, nil, tk.extra, nil}
	case TokIdent:
		txr.buildNode = Node{NodeIdent, tk.pos, nil, tk.extra, nil}
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
			txr.buildNode = Node{NodeUnOp, tk.pos, UnNegate, txr.buildNode, nil}
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

func NewTxr() Txr {
	return Txr{
		tokens:   []Token{},
		error:    "",
		buildPos: 0,
		buildLen: 0,
	}
}

func main() {
	txr := NewTxr()
	txr.Parse("(10 + 2) * 4")
	fmt.Println(txr.tokens)
	txr.Build()
	fmt.Println(txr.buildNode)
}
