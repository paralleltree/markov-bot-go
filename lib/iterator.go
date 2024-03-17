package lib

// Type IteratorFunc[T any] is a function type that returns an item, a boolean indicating whether the item is valid and can try to move on to next one, and an error.
// Caller should not process the item if the second returned value is false.'
// ex. `item, found, err := iter()` // where `found` is false, `item` is not valid.
type IteratorFunc[T any] func() (T, bool, error)

// Type ChunkIteratorFunc[T any] is a function type that returns a chunk of items, a boolean indicating whether there are more items, and an error.
// Caller should call this function repeatedly while the second returned value is true.
type ChunkIteratorFunc[T any] func() ([]T, bool, error)

func BuildIterator[T any](chunkIterator ChunkIteratorFunc[T]) IteratorFunc[T] {
	var buffer []T
	var empty T
	hasNext := true
	current := 0
	return func() (T, bool, error) {
		for {
			// reached the end of the buffer
			if len(buffer) <= current && hasNext {
				newBuf, newHasNext, err := chunkIterator()
				if err != nil {
					return empty, false, err
				}
				buffer = newBuf
				hasNext = newHasNext
				current = 0
			} else {
				break
			}
		}

		if len(buffer) <= current && !hasNext {
			return empty, false, nil
		}

		current++
		return buffer[current-1], current-1 < len(buffer), nil
	}
}
