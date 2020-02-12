package util

type comparable interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64
}

func Iff[T comparable](condition bool, v1 T, v2 T) T {
	if condition {
		return v1
	} else {
		return v2
	}
}
