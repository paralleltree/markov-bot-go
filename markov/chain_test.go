package markov_test

import (
	"slices"
	"testing"

	"github.com/paralleltree/markov-bot-go/markov"
)

func TestChain_Generate_WithEmptyModel_Returns_Empty(t *testing.T) {
	// arrange
	stateSize := 3
	chain := markov.NewChain(stateSize)

	// act
	result := chain.Generate()

	// assert
	wantResultLength := 0
	if wantResultLength != len(result) {
		t.Fatalf("unexpected result length: want %d, but got %d", wantResultLength, len(result))
	}
}

func TestChain_Generate_WithSingleSentence_Returns_Same_Sentence(t *testing.T) {
	// arrange
	stateSize := 3
	chain := markov.NewChain(stateSize)

	chain.AddSource([]string{"A", "B", "C"})

	// act
	gotResult := chain.Generate()

	// assert
	wantResult := []string{"A", "B", "C"}
	if !slices.Equal(wantResult, gotResult) {
		t.Fatalf("unexpected result: want %v, but got %v", wantResult, gotResult)
	}
}
