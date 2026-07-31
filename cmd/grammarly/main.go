// Command grammarly checks a literate grammar doc against the lexer's
// token enum and keyword table for drift. It doesn't generate one from the
// other — it just reports where they've grown apart.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/bigtalk-org/grammarly/internal/mdscan"
	"github.com/bigtalk-org/grammarly/internal/tokenset"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("grammarly", flag.ContinueOnError)
	grammarPath := fs.String("grammar", "docs/spec/grammar.md", "path to the literate grammar markdown file")
	tokenPath := fs.String("token", "lexer/token.go", "path to the Go file declaring the token enum")
	enumName := fs.String("enum", "TokenType", "name of the token-kind enum type")
	keywordsVar := fs.String("keywords", "keywords", "name of the package-level reserved-keyword map variable")
	strict := fs.Bool("strict", false, "exit non-zero on undocumented tokens/keywords too, not just hard drift")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	ts, err := tokenset.Load(*tokenPath, *enumName, *keywordsVar)
	if err != nil {
		fmt.Fprintln(os.Stderr, "grammarly:", err)
		return 2
	}
	md, err := mdscan.Scan(*grammarPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "grammarly:", err)
		return 2
	}

	codeTokens := map[string]bool{}
	for _, t := range ts.Tokens {
		codeTokens[t] = true
	}
	fenced := md.Fenced()
	allMentions := md.All()

	var drift []string
	for tok, lines := range fenced {
		if !codeTokens[tok] {
			for _, ln := range lines {
				drift = append(drift, fmt.Sprintf("%s:%d: %s", *grammarPath, ln, tok))
			}
		}
	}
	sort.Strings(drift)

	var undocumentedTokens []string
	for _, tok := range ts.Tokens {
		if !allMentions[tok] {
			undocumentedTokens = append(undocumentedTokens, tok)
		}
	}
	sort.Strings(undocumentedTokens)

	var undocumentedKeywords []string
	for kw, tok := range ts.Keywords {
		if !allMentions[tok] && !md.HasWord(kw) {
			undocumentedKeywords = append(undocumentedKeywords, fmt.Sprintf("%q -> %s", kw, tok))
		}
	}
	sort.Strings(undocumentedKeywords)

	exit := 0

	if len(drift) > 0 {
		fmt.Println("DRIFT — grammar references tokens that don't exist in", *tokenPath+":")
		for _, d := range drift {
			fmt.Println("  " + d)
		}
		exit = 1
	}

	if len(undocumentedTokens) > 0 {
		fmt.Println("UNDOCUMENTED — tokens exist in", *tokenPath, "but", *grammarPath, "never mentions them:")
		for _, t := range undocumentedTokens {
			fmt.Println("  " + t)
		}
		if *strict {
			exit = 1
		}
	}

	if len(undocumentedKeywords) > 0 {
		fmt.Println("UNDOCUMENTED — reserved keywords whose token is never mentioned in", *grammarPath+":")
		for _, k := range undocumentedKeywords {
			fmt.Println("  " + k)
		}
		if *strict {
			exit = 1
		}
	}

	if exit == 0 && len(undocumentedTokens) == 0 && len(undocumentedKeywords) == 0 {
		fmt.Println("grammarly: no drift found —", len(ts.Tokens), "tokens,", len(ts.Keywords), "keywords, all accounted for")
	}

	return exit
}
