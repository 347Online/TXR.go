// Code generated by "stringer -type=OpType"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[OpMul-1]
	_ = x[OpFDiv-2]
	_ = x[OpFMod-3]
	_ = x[OpIDiv-4]
	_ = x[OpAdd-16]
	_ = x[OpSub-17]
	_ = x[OpMaxP-32]
}

const (
	_OpType_name_0 = "OpMulOpFDivOpFModOpIDiv"
	_OpType_name_1 = "OpAddOpSub"
	_OpType_name_2 = "OpMaxP"
)

var (
	_OpType_index_0 = [...]uint8{0, 5, 11, 17, 23}
	_OpType_index_1 = [...]uint8{0, 5, 10}
)

func (i OpType) String() string {
	switch {
	case 1 <= i && i <= 4:
		i -= 1
		return _OpType_name_0[_OpType_index_0[i]:_OpType_index_0[i+1]]
	case 16 <= i && i <= 17:
		i -= 16
		return _OpType_name_1[_OpType_index_1[i]:_OpType_index_1[i+1]]
	case i == 32:
		return _OpType_name_2
	default:
		return "OpType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
