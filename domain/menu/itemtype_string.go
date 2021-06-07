// Code generated by "stringer -type=ItemType"; DO NOT EDIT.

package menu

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[UndefinedItem-0]
	_ = x[Plate-1]
	_ = x[Snack-2]
	_ = x[Side-3]
	_ = x[AddOn-4]
	_ = x[Condiment-5]
}

const _ItemType_name = "UndefinedItemPlateSnackSideAddOnCondiment"

var _ItemType_index = [...]uint8{0, 13, 18, 23, 27, 32, 41}

func (i ItemType) String() string {
	if i < 0 || i >= ItemType(len(_ItemType_index)-1) {
		return "ItemType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ItemType_name[_ItemType_index[i]:_ItemType_index[i+1]]
}
