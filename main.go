package main

import (
	"fmt"
	"strconv"
	"unicode"
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
	kind  TokenType
	pos   int
	extra any
}

func (t Token) String() string {
	s := fmt.Sprintf("{ %s @ %d", t.kind, t.pos)
	if t.extra != nil {
		s = fmt.Sprintf("%s (%d)", s, t.extra)
	}
	return fmt.Sprintf("%s }", s)
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
	length := len(str)
	out := &txr.parseTokens

	for pos < length {
		start := pos
		char := str[pos]
		pos += 1

		switch char {
		case byte(' '), byte('\t'), byte('\r'), byte('\n'):
			break

		case byte('('):
			*out = append(*out, Token{ParOpen, start, nil})

		case byte(')'):
			*out = append(*out, Token{ParClose, start, nil})

		case byte('+'):
			*out = append(*out, Token{Op, start, Add})

		case byte('-'):
			*out = append(*out, Token{Op, start, Sub})

		case byte('*'):
			*out = append(*out, Token{Op, start, Mul})

		case byte('/'):
			*out = append(*out, Token{Op, start, FDiv})

		case byte('%'):
			*out = append(*out, Token{Op, start, FMod})

		default:
			if unicode.IsDigit(rune(char)) {
				for pos < length {
					char = str[pos]
					if unicode.IsDigit(rune(char)) {
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
				*out = append(*out, Token{Number, start, val})
			}
		}
	}

	return false
}

func NewTxr() Txr {
	return Txr{error: ""}
}

func main() {
	txr := NewTxr()
	txr.Parse("Hello World ()() 123 456")
	fmt.Println(txr.parseTokens)
}
