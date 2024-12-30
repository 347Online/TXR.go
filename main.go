package main

import "fmt"

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

type Token struct{}

type Txr struct {
	parseTokens []Token
	error       string
}

func (txr *Txr) Throw(msg string, pos int) bool {
	txr.error = fmt.Sprintf("%s at position %d", msg, pos)
	return true
}

func NewTxr() Txr {
	return Txr{error: ""}
}

func main() {
	txr := NewTxr()
	txr.Throw("Test", 0)

	fmt.Println(txr.error)
	fmt.Println(Eof)
	fmt.Println(FDiv)
}
