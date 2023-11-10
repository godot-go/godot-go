package util

func Iff[T comparable](condition bool, v1 T, v2 T) T {
	if condition {
		return v1
	} else {
		return v2
	}
}
