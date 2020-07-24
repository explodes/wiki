package markov

import (
	"golang.org/x/net/html"
	"strings"
	"unicode"
)

type Stream struct {
	stream <-chan token
}

type token struct {
	start bool
	end   bool
	value string
}

func Split(s string) Stream {
	ch := make(chan token)
	go func() {
		defer close(ch)
		start := true
		end := false
		out := make([]rune, 0, 12)
		for _, r := range s {
			switch {
			case unicode.IsSpace(r) || r == '.':
				end = r == '.'
				if len(out) != 0 {
					ch <- token{start: start, end: end, value: string(out)}
				}
				start = end == true
				end = false
				out = make([]rune, 0, 12)
			case unicode.IsLetter(r) || unicode.IsPunct(r):
				out = append(out, unicode.ToLower(r))
			}
		}
		if len(out) != 0 {
			ch <- token{start: start, end: end, value: string(out)}
		}
	}()
	return Stream{ch}
}

func Html(s string) Stream {
	ch := make(chan token)
	go func() {
		defer close(ch)
		domDocTest := html.NewTokenizer(strings.NewReader(s))
		previousStartTokenTest := domDocTest.Token()
	loopDomTest:
		for {
			tt := domDocTest.Next()
			switch {
			case tt == html.ErrorToken:
				break loopDomTest // End of the document,  done
			case tt == html.StartTagToken:
				previousStartTokenTest = domDocTest.Token()
			case tt == html.TextToken:
				if previousStartTokenTest.Data == "script" {
					continue
				}
				out := strings.ToLower(strings.TrimSpace(html.UnescapeString(string(domDocTest.Text()))))
				if len(out) > 0 {
					for s := range Split(out).stream {
						ch <- s
					}
				}
			}
		}
	}()
	return Stream{ch}
}

func (s Stream) ToMarkov(m *Markov) {
	last := ""
	for token := range s.stream {
		switch {
		case token.start && token.end:
			m.Start(token.value)
			m.End(token.value)
			last = ""
		case token.start:
			m.Start(token.value)
			last = token.value
		case token.end:
			m.Add(last, token.value)
			m.End(token.value)
		default:
			if len(last) > 0 {
				m.Add(last, token.value)
			}
		}
	}
}

func (s Stream) Flatten() []string {
	out := make([]string, 0, 32)
	for word := range s.stream {
		out = append(out, word.value)
	}
	return out
}
