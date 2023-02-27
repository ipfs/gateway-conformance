package test

func H[T any](hint string, v T) WithHint[T] {
	return WithHint[T]{v, hint}
}
