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
	successWords              = 4
	timeLimit                 = 5 * time.Second

	prefixHttp  = "http://"
	prefixHttps = "https://"
	prefixFile  = "file://"

	debugMagic = "debug"
)

type app struct {
	r        *bufio.Reader
	debugOut io.Writer
}

func main() {
	newApp().begin()
}

func newApp() *app {
	return &app{
		r:        bufio.NewReader(os.Stdin),
		debugOut: ioutil.Discard,
	}
}

func (app *app) begin() {
	for {
		app.loop()
	}
}

func (app *app) loop() {
	url := app.prompt("Where do you want to learn from today?")
	app.note("Ok, learning from %s...", url)
	app.doPage(url)
}

func (app *app) doPage(url string) {

	var m *markov.Markov
	var err error
	if strings.HasPrefix(url, prefixHttp) || strings.HasPrefix(url, prefixHttps) {
		m, err = app.htmlMarkov(url)
	} else if strings.HasPrefix(url, prefixFile) {
		m, err = app.fileMarkov(url[len(prefixFile):])
	} else {
		app.note("I don'app know how to handle %s", url)
		return
	}
	if err != nil {
		app.note("Oops, something went wrong: %s", err)
		return
	}

	success := 0
	start := time.Now()
	for success < generateNumberOfSuccesses {
		now := time.Now()
		if now.Sub(start) > timeLimit {
			app.note("Sorry, ran out of time trying to generate new sentences.")
			break
		}
		generated := m.Generate()
		if len(generated) < successWords {
			continue
		}
		success++
		app.note("> %s", strings.Join(generated, " "))
	}

	mustWrite(m.DumpStats(app.debugOut))
}

func (app *app) htmlMarkov(url string) (*markov.Markov, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer mustClose(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := markov.New(2048)
	markov.Html(string(body)).ToMarkov(m)
	return m, nil
}

func (app *app) fileMarkov(path string) (*markov.Markov, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer mustClose(f)
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	m := markov.New(2048)
	markov.Split(string(contents)).ToMarkov(m)
	return m, nil
}

func (app *app) prompt(msg string) string {
	fmt.Printf("%s: ", msg)
	s, err := app.r.ReadString('\n')
	if err != nil {
		panic(err)
	}
	s = s[:len(s)-1]

	if s == debugMagic {
		app.debugOut = os.Stderr
		app.note("Debug mode enabled.")
		return app.prompt(msg)
	}

	return s
}

func (app *app) note(format string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(format, args...))
}

func (app *app) debug(format string, args ...interface{}) {
	mustWrite(fmt.Fprintln(app.debugOut, fmt.Sprintf(format, args...)))
}

func mustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		panic(err)
	}
}

func mustWrite(n int, err error) {
	if err != nil {
		panic(err)
	}
}
