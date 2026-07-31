package mdscan

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func TestScanReaderFencedVsProse(t *testing.T) {
	src := `# Grammar

Some prose mentioning ISO-EBNF and a CamelCase nonterminal TernaryExpr.

` + "```" + `
Expr ::= CONST | LET
` + "```" + `

More prose about STRING elsewhere, not fenced.
`
	res, err := ScanReader(strings.NewReader(src))
	if err != nil {
		t.Fatalf("ScanReader: %v", err)
	}

	fenced := res.Fenced()
	if _, ok := fenced["CONST"]; !ok {
		t.Errorf("expected CONST to be recorded as fenced, got %v", fenced)
	}
	if _, ok := fenced["LET"]; !ok {
		t.Errorf("expected LET to be recorded as fenced, got %v", fenced)
	}
	if _, ok := fenced["STRING"]; ok {
		t.Errorf("STRING appears only in prose, should not be fenced: %v", fenced)
	}

	all := res.All()
	wantAll := []string{"CONST", "EBNF", "ISO", "LET", "STRING"}
	if got := keys(all); !reflect.DeepEqual(got, wantAll) {
		t.Errorf("All() = %v, want %v", got, wantAll)
	}
}

func TestScanReaderSingleLetterNotAToken(t *testing.T) {
	src := "A curried stage is an ordinary value. I said so."
	res, err := ScanReader(strings.NewReader(src))
	if err != nil {
		t.Fatalf("ScanReader: %v", err)
	}
	if len(res.Mentions) != 0 {
		t.Errorf("expected no token mentions from single-letter capitals, got %v", res.Mentions)
	}
}

func TestHasWord(t *testing.T) {
	src := "`effect`/`effects`/`with`/`deny`/`allow` are reserved as lexer tokens."
	res, err := ScanReader(strings.NewReader(src))
	if err != nil {
		t.Fatalf("ScanReader: %v", err)
	}
	for _, w := range []string{"effect", "effects", "with", "deny", "allow"} {
		if !res.HasWord(w) {
			t.Errorf("HasWord(%q) = false, want true", w)
		}
	}
	if res.HasWord("region") {
		t.Error(`HasWord("region") = true, want false (not present in source)`)
	}
}

func TestHasWordWholeWordOnly(t *testing.T) {
	src := "the effective rate"
	res, err := ScanReader(strings.NewReader(src))
	if err != nil {
		t.Fatalf("ScanReader: %v", err)
	}
	if res.HasWord("effect") {
		t.Error(`HasWord("effect") = true, want false ("effective" should not match as a substring)`)
	}
}
