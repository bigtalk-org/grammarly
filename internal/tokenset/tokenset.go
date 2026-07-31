// Package tokenset extracts a lexer's token enum and keyword table from its
// Go source, e.g. lexer/token.go's TokenType const block and keywords map.
package tokenset

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
)

// Set is one file's token vocabulary.
type Set struct {
	Tokens   []string          // enum identifiers, in declaration order
	Keywords map[string]string // keyword string -> token identifier, e.g. "const" -> "CONST"
}

// Load extracts enumName's const block and keywordsVar's map from path.
func Load(path, enumName, keywordsVar string) (*Set, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadSource(path, src, enumName, keywordsVar)
}

// LoadSource is Load with the content already in hand, for tests.
func LoadSource(filename string, src []byte, enumName, keywordsVar string) (*Set, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}

	set := &Set{Keywords: map[string]string{}}
	var foundEnum, foundKeywords bool

	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		switch gd.Tok {
		case token.CONST:
			if !foundEnum && declaresType(gd, enumName) {
				for _, spec := range gd.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, name := range vs.Names {
						if name.Name != "_" {
							set.Tokens = append(set.Tokens, name.Name)
						}
					}
				}
				foundEnum = true
			}
		case token.VAR:
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || len(vs.Names) != 1 || vs.Names[0].Name != keywordsVar {
					continue
				}
				if len(vs.Values) != 1 {
					continue
				}
				cl, ok := vs.Values[0].(*ast.CompositeLit)
				if !ok {
					continue
				}
				for _, elt := range cl.Elts {
					kv, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					keyLit, ok := kv.Key.(*ast.BasicLit)
					if !ok || keyLit.Kind != token.STRING {
						continue
					}
					keyword, err := strconv.Unquote(keyLit.Value)
					if err != nil {
						continue
					}
					valIdent, ok := kv.Value.(*ast.Ident)
					if !ok {
						continue
					}
					set.Keywords[keyword] = valIdent.Name
				}
				foundKeywords = true
			}
		}
	}

	if !foundEnum {
		return nil, fmt.Errorf("%s: no const block declaring type %s found", filename, enumName)
	}
	if !foundKeywords {
		return nil, fmt.Errorf("%s: no var %s = map[string]%s{...} found", filename, keywordsVar, enumName)
	}
	return set, nil
}

// declaresType checks gd's opening `NAME Type = iota` spec. The rest of the
// block shares that type under Go's iota-repetition rule, so one match is enough.
func declaresType(gd *ast.GenDecl, enumName string) bool {
	if len(gd.Specs) == 0 {
		return false
	}
	vs, ok := gd.Specs[0].(*ast.ValueSpec)
	if !ok || vs.Type == nil {
		return false
	}
	ident, ok := vs.Type.(*ast.Ident)
	return ok && ident.Name == enumName
}
