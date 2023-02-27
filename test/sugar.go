package test

import "strings"

func H[T any](hint string, v T) WithHint[T] {
	return WithHint[T]{v, hint}
}

func HeaderContains(hint string, v string) WithHint[Check[string]] {
	return H[Check[string]](hint, func(s string) bool {
		return strings.Contains(s, v)
	})
}