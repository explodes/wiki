package markov_test

import (
	"github.com/explodes/wiki/markov"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func addSentence(m *markov.Markov, words ...string) {
	m.Start(words[0])
	for index := range words[:len(words)-1] {
		m.Add(words[index], words[index+1])
	}
	m.End(words[len(words)-1])
}

func TestRegenerateSentences(t *testing.T) {
	cases := []struct {
		name     string
		sentence string
	}{
		{"shortest", "hello"},
		{"two", "hello world"},
		{"longer", "hello world my name is explodes"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			m := markov.New(1024)
			split := strings.Split(testCase.sentence, " ")
			addSentence(m, split...)
			assert.Equal(t, split, m.Generate())
		})
	}
}

func TestProbabilisticSentences(t *testing.T) {
	cases := []struct {
		name     string
		sentence string
	}{
		{"shortest", "hello"},
		{"two", "hello world"},
		{"longer", "hello world my name is explodes"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			m := markov.New(1024)
			split := strings.Split(testCase.sentence, " ")
			for i := 0; i < 1000; i++ {
				addSentence(m, split...)
			}
			assert.Equal(t, split, m.Generate())
		})
	}
}

func TestFormingNewSentences(t *testing.T) {
	sentences := []string{
		"We are spirit bound to this flesh",
		"We go round one foot nailed down",
		"But bound to reach out and beyond this flesh",
		"Become Pneuma",
		"We are will and wonder",
		"Bound to recall, remember",
		"We are born of one breath, one word",
		"We are all one spark, sun becoming",
		"Child, wake up",
		"Child, release the light",
		"Wake up now",
		"Child, wake up",
		"Child, release the light",
		"Wake up now, child",
		"(Spirit)",
		"(Spirit)",
		"(Spirit)",
		"(Spirit)",
		"Bound to this flesh",
		"This guise, this mask",
		"This dream",
		"Wake up…",
	}
	m := markov.New(1024)
	for _, sentence := range sentences {
		split := markov.Split(sentence).Flatten()
		addSentence(m, split...)
	}
	_, err := m.DumpStats(os.Stdout)
	assert.NoError(t, err)
	for i := 0; i < len(sentences); i++ {
		println(strings.Join(m.Generate(), " "))
	}
}
