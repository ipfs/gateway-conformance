package test

type Provider[T any] interface {
	Get() T
}

type StringProvider string

func (p StringProvider) Get() string {
	return string(p)
}

type FuncProvider[T any] func() T

func (p FuncProvider[T]) Get() T {
	return p()
}
