package test

type Provider[T any] interface {
	Get() T
}

type StringProvider string

func (p StringProvider) Get() string {
	return string(p)
}
