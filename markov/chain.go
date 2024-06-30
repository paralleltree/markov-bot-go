package markov

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

const (
	BOS = "__BOS__"
	EOS = "__EOS__"
)

type Chain struct {
	StateSize int        `json:"state_size"`
	RootNode  *chainNode `json:"root_node"`
}

func NewChain(stateSize int) *Chain {
	return &Chain{
		StateSize: stateSize,
		RootNode:  newChainNode(),
	}
}

type chainNode struct {
	Children    map[string]*chainNode `json:"children"`
	Occurrences int                   `json:"occurences"`
}

func newChainNode() *chainNode {
	return &chainNode{Children: map[string]*chainNode{}}
}

// Adds single source to this chain.
func (c *Chain) AddSource(source []string) {
	// fill BOS/EOS
	run := makeRun(c.StateSize, source)

	// skip empty source
	if len(run) == c.StateSize+1 {
		return
	}

	for i := 0; i < len(run)-c.StateSize; i++ {
		tailNode := c.findOrAddTailNode(run[i : i+c.StateSize])
		// find or add leaf node
		nextElement := run[i+c.StateSize]
		followNode, ok := tailNode.Children[nextElement]
		if !ok {
			followNode = newChainNode()
			tailNode.Children[nextElement] = followNode
		}
		// increment occurrence at leaf node
		followNode.Occurrences += 1
	}
}

// Generates the sequence from this chain.
func (c *Chain) Generate() []string {
	buf := make([]string, c.StateSize)
	for i := 0; i < c.StateSize; i++ {
		buf[i] = BOS
	}
	for i := 0; ; i++ {
		tailNode := c.findOrAddTailNode(buf[i : i+c.StateSize])
		items, cumsum := tailNode.accumulateOccurrences()
		if len(items) == 0 {
			// no words found in this Chain
			return nil
		}
		r := rand.Intn(cumsum[len(cumsum)-1])
		elected := items[sort.SearchInts(cumsum, r)]
		if elected == EOS {
			break
		}
		buf = append(buf, elected)
	}
	return buf[c.StateSize:]
}

// Find or add tail node that is last of state sequence
func (c *Chain) findOrAddTailNode(state []string) *chainNode {
	tailNode := c.RootNode
	for i := 0; i < len(state); i++ {
		nextNode, ok := tailNode.Children[state[i]]
		if !ok {
			nextNode = newChainNode()
			tailNode.Children[state[i]] = nextNode
		}
		tailNode = nextNode
	}
	return tailNode
}

// Returns cumulative sum and its value
func (n *chainNode) accumulateOccurrences() ([]string, []int) {
	sum := 0
	accum := make([]int, 0, len(n.Children))
	values := make([]string, 0, len(n.Children))
	for k, v := range n.Children {
		sum += v.Occurrences
		values = append(values, k)
		accum = append(accum, sum)
	}
	return values, accum
}

func (c *Chain) Dump() ([]byte, error) {
	return json.Marshal(c)
}

func LoadChain(s []byte) (*Chain, error) {
	c := &Chain{}
	if err := json.Unmarshal(s, c); err != nil {
		return nil, fmt.Errorf("unmarshal chain: %w", err)
	}
	return c, nil
}

// Makes string slice that contains [BOS * stateSize, source, EOS]
func makeRun(stateSize int, source []string) []string {
	run := make([]string, 0, len(source)+stateSize+1)
	for i := 0; i < stateSize; i++ {
		run = append(run, BOS)
	}
	for _, v := range source {
		if v == "" {
			continue
		}
		run = append(run, v)
	}
	run = append(run, EOS)
	return run
}
