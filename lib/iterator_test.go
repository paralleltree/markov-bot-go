package lib_test

import (
	"testing"

	"github.com/paralleltree/markov-bot-go/lib"
)

func TestBuildIterator_IteratesOverMultipleChunks(t *testing.T) {
	itemsPerPage := 10
	breakAtTheLastChunkFunc := func(wantItemsCount int) func() ([]int, bool, error) {
		head := 0
		return func() ([]int, bool, error) {
			buf := make([]int, 0, itemsPerPage)
			for i := 0; i < itemsPerPage; i++ {
				if wantItemsCount <= head+len(buf) {
					break
				}
				buf = append(buf, head)
			}
			head += len(buf)
			return buf, head < wantItemsCount, nil
		}
	}

	breakAfterTheLastChunkFunc := func(wantItemsCount int) func() ([]int, bool, error) {
		head := 0
		return func() ([]int, bool, error) {
			if head == wantItemsCount {
				return nil, false, nil
			}
			buf := make([]int, 0, itemsPerPage)
			for i := 0; i < itemsPerPage; i++ {
				if wantItemsCount <= head+len(buf) {
					break
				}
				buf = append(buf, head)
			}
			head += len(buf)
			return buf, true, nil
		}
	}

	cases := []struct {
		name           string
		wantItemsCount int
		buildChunkFunc func(int) func() ([]int, bool, error)
	}{
		{
			name:           "chunkFunc returns false as hasNext at the last chunk",
			wantItemsCount: 15,
			buildChunkFunc: breakAfterTheLastChunkFunc,
		},
		{
			name:           "chunkFunc returns false as hasNext after the last chunk",
			wantItemsCount: 15,
			buildChunkFunc: breakAtTheLastChunkFunc,
		},
		{
			name:           "chunkFunc returns false as hasNext at the last chunk and the number of items is multiple of itemsPerPage",
			wantItemsCount: 20,
			buildChunkFunc: breakAfterTheLastChunkFunc,
		},
		{
			name:           "chunkFunc returns false as hasNext after the last chunk and the number of items is multiple of itemsPerPage",
			wantItemsCount: 20,
			buildChunkFunc: breakAtTheLastChunkFunc,
		},
		{
			name:           "when chunkFunc returns empty result first but hasNext is true, iterator should make next call to chunkFunc",
			wantItemsCount: 1,
			buildChunkFunc: func(i int) func() ([]int, bool, error) {
				invokeCount := 0
				return func() ([]int, bool, error) {
					defer func() { invokeCount++ }()
					switch invokeCount {
					case 0:
						return nil, true, nil
					case 1:
						return []int{0}, false, nil
					default:
						panic("unexpected invoke!")
					}
				}
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			chunkFunc := tt.buildChunkFunc(tt.wantItemsCount)
			iter := lib.BuildIterator(chunkFunc)

			gotCount := 0
			for {
				_, hasNext, err := iter()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !hasNext {
					break
				}

				gotCount++
			}

			if tt.wantItemsCount != gotCount {
				t.Errorf("unexpected count: want %d items, but got %d items", tt.wantItemsCount, gotCount)
			}
		})
	}
}
