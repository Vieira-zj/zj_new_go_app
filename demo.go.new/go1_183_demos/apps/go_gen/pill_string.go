// Code generated by "stringer -type=AutoPill -output=pill_string.go"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AutoPlacebo-0]
	_ = x[AutoAspirin-1]
	_ = x[AutoIbuprofen-2]
	_ = x[AutoParacetamol-3]
}

const _AutoPill_name = "AutoPlaceboAutoAspirinAutoIbuprofenAutoParacetamol"

var _AutoPill_index = [...]uint8{0, 11, 22, 35, 50}

func (i AutoPill) String() string {
	if i < 0 || i >= AutoPill(len(_AutoPill_index)-1) {
		return "AutoPill(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AutoPill_name[_AutoPill_index[i]:_AutoPill_index[i+1]]
}