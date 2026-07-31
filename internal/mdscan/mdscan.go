// Package mdscan finds lexer-token references in a literate grammar doc,
// using its own convention that ALLCAPS names are tokens (nonterminals are
// CamelCase, so they never match).
package mdscan

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strings"
)

var tokenRef = regexp.MustCompile(`\b[A-Z][A-Z0-9_]+\b`)

// Mention is one occurrence of an ALLCAPS token-shaped word.
type Mention struct {
	Token   string
	Line    int // 1-based
	InFence bool
}

// Result is one document's scan.
type Result struct {
	Mentions []Mention
	Raw      string
}

// HasWord reports whether word appears anywhere as a whole word
// (case-sensitive) — for reserved keywords like "region" that get
// discussed in prose by their lowercase spelling, not their token name.
func (r *Result) HasWord(word string) bool {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)
	return re.MatchString(r.Raw)
}

// Fenced returns tokens named inside fenced code blocks — actual grammar
// terminals, not just prose mentions.
func (r *Result) Fenced() map[string][]int {
	out := map[string][]int{}
	for _, m := range r.Mentions {
		if m.InFence {
			out[m.Token] = append(out[m.Token], m.Line)
		}
	}
	return out
}

// All returns every token name mentioned anywhere, fenced or prose.
func (r *Result) All() map[string]bool {
	out := map[string]bool{}
	for _, m := range r.Mentions {
		out[m.Token] = true
	}
	return out
}

// Scan reads path and extracts token mentions.
func Scan(path string) (*Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ScanReader(f)
}

// ScanReader is Scan with the content already in hand, for tests.
func ScanReader(r io.Reader) (*Result, error) {
	res := &Result{}
	var raw strings.Builder
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	inFence := false
	line := 0
	for sc.Scan() {
		line++
		text := sc.Text()
		raw.WriteString(text)
		raw.WriteByte('\n')
		if strings.HasPrefix(strings.TrimSpace(text), "```") {
			inFence = !inFence
			continue
		}
		for _, tok := range tokenRef.FindAllString(text, -1) {
			res.Mentions = append(res.Mentions, Mention{Token: tok, Line: line, InFence: inFence})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	res.Raw = raw.String()
	return res, nil
}
