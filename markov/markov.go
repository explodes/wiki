package markov

import (
	"fmt"
	"io"
	"math/rand"
	"sort"
	"time"
)

const (
	endSentinel = ""
)

type Markov struct {
	counts    map[string]*counts
	startSet  map[string]struct{}
	startList []string
	sizeHint  int
	rng       *rand.Rand
}

func New(sizeHint int) *Markov {
	return &Markov{
		counts:    make(map[string]*counts, sizeHint),
		startSet:  make(map[string]struct{}, sizeHint),
		startList: make([]string, 0, sizeHint),
		sizeHint:  sizeHint,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *Markov) Start(a string) {
	if _, ok := m.startSet[a]; !ok {
		m.startSet[a] = struct{}{}
		m.startList = append(m.startList, a)
	}
}

func (m *Markov) Add(a, b string) {
	counts, ok := m.counts[a]
	if !ok {
		counts = newCounts(m.sizeHint)
		m.counts[a] = counts
	}
	counts.Increment(b)
}

func (m *Markov) End(b string) {
	m.Add(b, endSentinel)
}

func (m *Markov) Generate() []string {
	if len(m.startList) == 0 {
		return nil
	}

	out := make([]string, 0, 32)

	word := m.startList[m.rng.Intn(len(m.startList))]
	for word != endSentinel {
		out = append(out, word)
		word = m.nextWord(word)
	}

	return out
}

func (m *Markov) nextWord(root string) string {
	counts, ok := m.counts[root]
	if !ok {
		return endSentinel
	}
	return counts.NextWord(m.rng)
}

func (m *Markov) DumpStats(w io.Writer) (int, error) {
	written := 0
	for word, counts := range m.counts {
		invsum := 1.0 / float32(counts.sum)
		for next, count := range counts.valuesSet {
			s := fmt.Sprintf("%s->%s (%d/%d) (%f)\n", word, next, count, counts.sum, float32(count)*invsum)
			n, err := w.Write([]byte(s))
			written += n
			if err != nil {
				return written, err
			}
		}
	}
	return written, nil
}

type counts struct {
	valuesSet     map[string]int
	valuesList    []string
	sum           int
	dirty         bool
	summedWeights []int
}

func newCounts(sizeHint int) *counts {
	return &counts{
		valuesSet:     make(map[string]int, sizeHint),
		valuesList:    make([]string, 0, sizeHint),
		sum:           0,
		dirty:         false,
		summedWeights: nil,
	}
}

func (c *counts) Increment(b string) {
	count, ok := c.valuesSet[b]
	if !ok {
		c.valuesList = append(c.valuesList, b)
	}
	c.valuesSet[b] = count + 1
	c.sum++
	c.dirty = true
}

func (c *counts) computeWeights() {
	if !c.dirty {
		return
	}
	sum := 0
	summedWeights := make([]int, len(c.valuesList))
	for index, word := range c.valuesList {
		sum += c.valuesSet[word]
		summedWeights[index] = sum
	}
	c.summedWeights = summedWeights
	c.dirty = false
}

func (c *counts) NextWord(rng *rand.Rand) string {
	if len(c.valuesList) == 0 {
		return endSentinel
	}
	if c.dirty {
		c.computeWeights()
	}
	weightIndex := rng.Intn(c.sum)
	index := sort.Search(len(c.summedWeights), func(index int) bool {
		return c.summedWeights[index] >= weightIndex
	})
	return c.valuesList[index]
}
