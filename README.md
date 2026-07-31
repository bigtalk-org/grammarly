# grammarly

Checks a literate grammar doc (e.g. BigTalk's `docs/spec/grammar.md`) against
the lexer's actual token enum (e.g. `lexer/token.go`) for drift.

It doesn't generate one from the other. The grammar doc's value is its
hand-written rationale, which no enum can produce; the enum's layout
shouldn't be dictated by markdown either. This just flags where they've
grown apart.

## What it reports

Using the doc's own convention that ALLCAPS names are tokens:

- **DRIFT** — a token used in a fenced EBNF production that doesn't exist
  in the Go enum. The one that matters: the grammar documents syntax the
  lexer can't produce.
- **UNDOCUMENTED (tokens)** — an enum token whose ALLCAPS name never
  appears in the doc. Expect noise (`EOF`, `ILLEGAL`, ...); treat as a
  prompt to check, not a verdict.
- **UNDOCUMENTED (keywords)** — a reserved keyword whose lowercase spelling
  (e.g. `region`) never appears anywhere in the doc. The trustworthy one.

## Usage

```bash
go install github.com/bigtalk-org/grammarly/cmd/grammarly@latest

cd /path/to/bigtalk
grammarly -grammar docs/spec/grammar.md -token lexer/token.go
```

| Flag        | Default                | Meaning                                  |
|-------------|-------------------------|-------------------------------------------|
| `-grammar`  | `docs/spec/grammar.md` | grammar markdown file                     |
| `-token`    | `lexer/token.go`       | Go file declaring the token enum          |
| `-enum`     | `TokenType`            | token enum type name                      |
| `-keywords` | `keywords`             | reserved-keyword map variable name        |
| `-strict`   | `false`                | also fail on UNDOCUMENTED findings        |

Exit code is non-zero on DRIFT; UNDOCUMENTED only fails under `-strict`.
