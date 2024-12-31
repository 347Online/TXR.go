package main

import (
	"fmt"
	"strconv"
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

	switch t.kind {
	case Op:
		s = fmt.Sprintf("%s, %s", s, t.extra.(OpType))
	case Number:
		s = fmt.Sprintf("%s (%d)", s, t.extra)
	case Ident:
		s = fmt.Sprintf("%s `%s`", s, t.extra)
	}

	return fmt.Sprintf("%s }", s)
}

type Txr struct {
	parseTokens []Token
	error       string
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
				*out = append(*out, Token{Number, start, val})
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
					*out = append(*out, Token{Op, start, FMod})
				case "div":
					*out = append(*out, Token{Op, start, IDiv})
				default:
					*out = append(*out, Token{Ident, start, name})
				}
			} else {
				*out = []Token{}
				msg := fmt.Sprintf("Unexpected character `%c`", char)
				return txr.Throw(msg, start)
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
	txr.Parse("Hello World ()() 123 + 456 .")
	fmt.Println(txr.parseTokens)
}
