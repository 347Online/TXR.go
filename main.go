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

//go:generate stringer -type=Unary
type Unary int

const (
	UnNegate Unary = 1
)

//go:generate stringer -type=BuildFlag
type BuildFlag int

const (
	NoOps BuildFlag = 1
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

type Txr struct {
	tokens []Token
	error  string
}

func (txr *Txr) Throw(msg string, pos int) bool {
	txr.error = fmt.Sprintf("%s at position %d", msg, pos)
	fmt.Println(txr.error)
	return true
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

func NewTxr() Txr {
	return Txr{error: ""}
}

func main() {
	txr := NewTxr()
	txr.Parse("Hello World ()() 123 + 456 ")
	fmt.Println(txr.tokens)
}
