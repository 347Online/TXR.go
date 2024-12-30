package main

import (
	"fmt"
)

//go:generate stringer -type=TokenType
type TokenType int

const (
	Eof TokenType = iota
	Op
	ParOpen
	ParClose
	Number
	Ident
)

//go:generate stringer -type=OpType
type OpType int

const (
	Mul  OpType = 0x01
	FDiv OpType = 0x02
	FMod OpType = 0x03
	IDiv OpType = 0x04
	Add  OpType = 0x10
	Sub  OpType = 0x11
	MaxP OpType = 0x20
)

type Token struct {
	TokenType
	int
}

func (t Token) String() string {
	return fmt.Sprintf("{ %s @ %d }", t.TokenType, t.int)
}

type Txr struct {
	parseTokens []Token
	error       string
}

func (txr *Txr) Throw(msg string, pos int) bool {
	txr.error = fmt.Sprintf("%s at position %d", msg, pos)
	return true
}

func (txr *Txr) Parse(str string) bool {
	pos := 0
	for pos < len(str) {
		start := pos
		out := &txr.parseTokens
		char := str[pos]
		pos += 1

		switch char {
		case byte(' '), byte('\t'), byte('\r'), byte('\n'):
			break

		case byte('('):
			*out = append(*out, Token{ParOpen, start})
			break

		case byte(')'):
			*out = append(*out, Token{ParClose, start})
			break

		default:
			break
		}
	}

	return false
}

func NewTxr() Txr {
	return Txr{error: ""}
}

func main() {
	txr := NewTxr()
	txr.Parse("Hello World ()()")
	fmt.Println(txr.parseTokens)
}
