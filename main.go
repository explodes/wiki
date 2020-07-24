package main

import (
	"bufio"
	"fmt"
	"github.com/explodes/wiki/markov"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	generateNumberOfSuccesses = 15
	successWords              = 7
	timeLimit                 = 3 * time.Second
)

func main() {
	term := newTerminal()

	doPage(term, "https://blog.bruce-hill.com/a-faster-weighted-random-choice")

	for {
		loop(term)
	}
}

func loop(term *terminal) {
	url := term.prompt("Where do you want to learn from today?")
	term.note("Ok, learning from %s...", url)
	doPage(term, url)
}

func doPage(term *terminal, url string) {

	resp, err := http.Get(url)
	if err != nil {
		term.note("Oops, something went wrong: %s", err)
		return
	}
	defer mustClose(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		term.note("Oops, something went wrong: %s", err)
		return
	}

	m := markov.New(2048)
	markov.Html(string(body)).ToMarkov(m)

	success := 0
	start := time.Now()
	for success < generateNumberOfSuccesses {
		now := time.Now()
		if now.Sub(start) > timeLimit {
			term.note("Sorry, ran out of time trying to generate new sentences.")
			break
		}
		generated := m.Generate()
		if len(generated) < successWords {
			continue
		}
		success++
		term.note("> %s", strings.Join(generated, " "))
	}

	m.DumpStats(os.Stdout)
}

func mustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		panic(err)
	}
}

type terminal struct {
	r *bufio.Reader
}

func newTerminal() *terminal {
	return &terminal{
		r: bufio.NewReader(os.Stdin),
	}
}

func (p *terminal) prompt(msg string) string {
	fmt.Printf("%s: ", msg)
	s, err := p.r.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return s[:len(s)-1]
}

func (p *terminal) note(format string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(format, args...))
}
