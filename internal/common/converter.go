package common

func PointerVal[T any](t *T) T {
	if t == nil {
		var zero T
		return zero
	}

	return *t
}

func Pointer[T any](t T) *T {
	return &t
}
