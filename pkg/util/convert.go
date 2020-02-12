package util

func BoolToUint8(v bool) uint8 {
	if v {
		return 1
	} else {
		return 0
	}
}
