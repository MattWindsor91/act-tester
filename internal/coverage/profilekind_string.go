// Code generated by "stringer -type ProfileKind"; DO NOT EDIT.

package coverage

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Known-0]
	_ = x[Standalone-1]
}

const _ProfileKind_name = "KnownStandalone"

var _ProfileKind_index = [...]uint8{0, 5, 15}

func (i ProfileKind) String() string {
	if i >= ProfileKind(len(_ProfileKind_index)-1) {
		return "ProfileKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ProfileKind_name[_ProfileKind_index[i]:_ProfileKind_index[i+1]]
}