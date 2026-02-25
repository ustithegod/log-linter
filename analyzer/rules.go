package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

func checkForUppercase(pass *analysis.Pass, e ast.Expr) {
	s, ok := getStringLiteral(e)
	if !ok || s == "" {
		return
	}

	if !startsWithUpper(s) {
		return
	}

	pass.Report(analysis.Diagnostic{
		Pos:     e.Pos(),
		End:     e.End(),
		Message: fmt.Sprintf("message \"%s\" starts with the capital letter", s),

		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "convert the first letter to lowercase",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     e.Pos(),
						End:     e.End(),
						NewText: lowercaseFirst(s),
					},
				},
			},
		},
	})
}

func checkForEnglishOnly(pass *analysis.Pass, e ast.Expr) {
	s, ok := getStringLiteral(e)
	if !ok || s == "" {
		return
	}

	var hasNonEnglish bool

	for _, r := range s {
		if unicode.IsLetter(r) && !isEnglishLetter(r) {
			hasNonEnglish = true
		}
	}

	if !hasNonEnglish {
		return
	}

	clean := cleanNonEnglish(s)
	pass.Report(analysis.Diagnostic{
		Pos:     e.Pos(),
		End:     e.End(),
		Message: fmt.Sprintf("message \"%s\" contains non-english letters", s),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "remove non-english letters",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     e.Pos(),
						End:     e.End(),
						NewText: []byte(strconv.Quote(clean)),
					},
				},
			},
		},
	})
}

func checkForSpecials(pass *analysis.Pass, e ast.Expr) {
	s, ok := getStringLiteral(e)
	if !ok || s == "" {
		return
	}

	var hasSpecial bool

	for _, r := range s {
		if isAllowedSpecial(r) {
			continue
		}
		hasSpecial = true
		break
	}

	if !hasSpecial {
		return
	}

	clean := cleanAllowed(s)
	pass.Report(analysis.Diagnostic{
		Pos:     e.Pos(),
		End:     e.End(),
		Message: fmt.Sprintf("message \"%s\" contains special symbols", s),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "remove special symbols",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     e.Pos(),
						End:     e.End(),
						NewText: []byte(strconv.Quote(clean)),
					},
				},
			},
		},
	})
}

func checkForSensitive(pass *analysis.Pass, e ast.Expr, cfg Config) {
	if len(cfg.Sensitive.Keywords) == 0 {
		return
	}

	names := collectNamesFromExpr(e)
	if len(names) == 0 {
		return
	}

	for _, name := range names {
		if kw, ok := matchSensitive(name, cfg.Sensitive); ok {
			diag := analysis.Diagnostic{
				Pos:     e.Pos(),
				End:     e.End(),
				Message: fmt.Sprintf("message contains sensitive keyword %q from identifier %q", kw, name),
			}

			if lit, ok := concatLiteralsOnly(e); ok {
				diag.SuggestedFixes = []analysis.SuggestedFix{
					{
						Message: "remove concatenation and keep only literal parts",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     e.Pos(),
								End:     e.End(),
								NewText: []byte(strconv.Quote(lit)),
							},
						},
					},
				}
			}

			pass.Report(diag)
			return
		}
	}
}

func getStringLiteral(node ast.Expr) (string, bool) {
	lit, ok := node.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	unq, err := strconv.Unquote(lit.Value)
	if err != nil || unq == "" {
		return "", false
	}

	return unq, true
}

func cleanAllowed(s string) string {
	var b strings.Builder
	for _, r := range s {
		if isAllowedSpecial(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func cleanNonEnglish(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) && !isEnglishLetter(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func startsWithUpper(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(r)
}

func lowercaseFirst(s string) []byte {
	r, size := utf8.DecodeRuneInString(s)
	new := unicode.ToLower(r)

	return []byte(strconv.Quote(string(new) + s[size:]))
}

func isEnglishLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isAllowedSpecial(r rune) bool {
	return unicode.IsLetter(r) || (r >= '0' && r <= '9') || r == ' '
}

func collectNamesFromExpr(e ast.Expr) []string {
	var names []string
	var visit func(ast.Expr)

	visit = func(expr ast.Expr) {
		switch v := expr.(type) {
		case *ast.BasicLit:
			return
		case *ast.Ident:
			names = append(names, v.Name)
		case *ast.SelectorExpr:
			names = append(names, v.Sel.Name)
			visit(v.X)
		case *ast.BinaryExpr:
			if v.Op != token.ADD {
				return
			}
			visit(v.X)
			visit(v.Y)
		case *ast.CallExpr:
			visit(v.Fun)
			for _, arg := range v.Args {
				visit(arg)
			}
		case *ast.ParenExpr:
			visit(v.X)
		case *ast.UnaryExpr:
			visit(v.X)
		case *ast.IndexExpr:
			visit(v.X)
			visit(v.Index)
		case *ast.SliceExpr:
			visit(v.X)
			if v.Low != nil {
				visit(v.Low)
			}
			if v.High != nil {
				visit(v.High)
			}
			if v.Max != nil {
				visit(v.Max)
			}
		case *ast.TypeAssertExpr:
			visit(v.X)
		}
	}

	visit(e)

	return names
}

func concatLiteralsOnly(e ast.Expr) (string, bool) {
	var b strings.Builder
	var hasLiteral bool
	var ok = true

	var visit func(ast.Expr)
	visit = func(expr ast.Expr) {
		if !ok || expr == nil {
			return
		}
		switch v := expr.(type) {
		case *ast.BasicLit:
			if v.Kind != token.STRING {
				return
			}
			unq, err := strconv.Unquote(v.Value)
			if err != nil {
				return
			}
			hasLiteral = true
			b.WriteString(unq)
		case *ast.BinaryExpr:
			if v.Op != token.ADD {
				ok = false
				return
			}
			visit(v.X)
			visit(v.Y)
		case *ast.ParenExpr:
			visit(v.X)
		default:
			// allow other nodes in the concat chain, but only literals are kept
			return
		}
	}

	visit(e)

	if !ok || !hasLiteral {
		return "", false
	}

	return b.String(), true
}

func matchSensitive(name string, cfg SensitiveConfig) (string, bool) {
	original := name
	normalized := name
	if cfg.CaseInsensitive {
		normalized = strings.ToLower(name)
	}

	for _, kw := range cfg.Keywords {
		keyword := strings.TrimSpace(kw)
		if keyword == "" {
			continue
		}
		if cfg.CaseInsensitive {
			keyword = strings.ToLower(keyword)
		}
		if containsKeyword(original, normalized, keyword, cfg.WordBoundary, cfg.CaseInsensitive) {
			return kw, true
		}
	}
	return "", false
}

func containsKeyword(original, normalized, keyword string, wordBoundary, caseInsensitive bool) bool {
	if !wordBoundary {
		return strings.Contains(normalized, keyword)
	}

	tokens := splitIdentTokens(original)
	for _, t := range tokens {
		if t == "" {
			continue
		}
		if caseInsensitive {
			if strings.EqualFold(t, keyword) {
				return true
			}
			continue
		}
		if t == keyword {
			return true
		}
	}
	return false
}

func splitIdentTokens(name string) []string {
	if name == "" {
		return nil
	}

	var tokens []string
	var buf []rune

	flush := func() {
		if len(buf) == 0 {
			return
		}
		tokens = append(tokens, string(buf))
		buf = buf[:0]
	}

	var prev rune
	for i, r := range name {
		if r == '_' {
			flush()
			prev = 0
			continue
		}
		if !isIdentRune(r) {
			flush()
			prev = 0
			continue
		}

		if i > 0 && shouldSplitOnCase(prev, r) {
			flush()
		}
		buf = append(buf, r)
		prev = r
	}
	flush()

	return tokens
}

func isIdentRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func shouldSplitOnCase(prev, curr rune) bool {
	if prev == 0 {
		return false
	}
	if unicode.IsLower(prev) && unicode.IsUpper(curr) {
		return true
	}
	return false
}
