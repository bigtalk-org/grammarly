package tokenset

import (
	"reflect"
	"testing"
)

const sampleSrc = `package lexer

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	CONST
	LET
)

var keywords = map[string]TokenType{
	"const": CONST, "let": LET,
}
`

func TestLoadSource(t *testing.T) {
	set, err := LoadSource("token.go", []byte(sampleSrc), "TokenType", "keywords")
	if err != nil {
		t.Fatalf("LoadSource: %v", err)
	}

	wantTokens := []string{"ILLEGAL", "EOF", "CONST", "LET"}
	if !reflect.DeepEqual(set.Tokens, wantTokens) {
		t.Errorf("Tokens = %v, want %v", set.Tokens, wantTokens)
	}

	wantKeywords := map[string]string{"const": "CONST", "let": "LET"}
	if !reflect.DeepEqual(set.Keywords, wantKeywords) {
		t.Errorf("Keywords = %v, want %v", set.Keywords, wantKeywords)
	}
}

func TestLoadSourceMissingEnum(t *testing.T) {
	_, err := LoadSource("x.go", []byte(sampleSrc), "NoSuchType", "keywords")
	if err == nil {
		t.Fatal("expected an error for a nonexistent enum type, got nil")
	}
}

func TestLoadSourceMissingKeywordsVar(t *testing.T) {
	_, err := LoadSource("x.go", []byte(sampleSrc), "TokenType", "noSuchVar")
	if err == nil {
		t.Fatal("expected an error for a nonexistent keywords var, got nil")
	}
}

func TestLoadSourceBlankIdentifierSkipped(t *testing.T) {
	src := `package lexer

type TokenType int

const (
	ILLEGAL TokenType = iota
	_
	CONST
)

var keywords = map[string]TokenType{
	"const": CONST,
}
`
	set, err := LoadSource("token.go", []byte(src), "TokenType", "keywords")
	if err != nil {
		t.Fatalf("LoadSource: %v", err)
	}
	want := []string{"ILLEGAL", "CONST"}
	if !reflect.DeepEqual(set.Tokens, want) {
		t.Errorf("Tokens = %v, want %v (blank identifier should be skipped)", set.Tokens, want)
	}
}
