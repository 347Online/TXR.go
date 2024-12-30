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
}
