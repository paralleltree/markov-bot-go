package lib

// Type ChunkIteratorFunc[T any] is a function type that returns a chunk of items, a boolean indicating whether there are more items, and an error.
type ChunkIteratorFunc[T any] func() ([]T, bool, error)
