package lib

type IteratorFunc[T any] func() ([]T, bool, error)
