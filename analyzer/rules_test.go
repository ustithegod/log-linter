package analyzer

import (
	"go/parser"
	"strings"
	"testing"
)

func TestSplitIdentTokens(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "camel", in: "authToken", want: []string{"auth", "Token"}},
		{name: "snake", in: "user_token", want: []string{"user", "token"}},
		{name: "mixed", in: "userToken_id", want: []string{"user", "Token", "id"}},
		{name: "caps", in: "APIKey", want: []string{"APIKey"}},
		{name: "single", in: "password", want: []string{"password"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitIdentTokens(tt.in)
			if !equalSlices(got, tt.want) {
				t.Fatalf("splitIdentTokens(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}

func TestContainsKeywordWordBoundary(t *testing.T) {
	tests := []struct {
		name         string
		original     string
		keyword      string
		wordBoundary bool
		caseInsensitive bool
		want         bool
	}{
		{name: "camel hit insensitive", original: "authToken", keyword: "token", wordBoundary: true, caseInsensitive: true, want: true},
		{name: "snake hit insensitive", original: "user_token", keyword: "token", wordBoundary: true, caseInsensitive: true, want: true},
		{name: "camel auth hit insensitive", original: "authToken", keyword: "auth", wordBoundary: true, caseInsensitive: true, want: true},
		{name: "no boundary miss insensitive", original: "tokentest", keyword: "token", wordBoundary: true, caseInsensitive: true, want: false},
		{name: "substring hit insensitive", original: "tokentest", keyword: "token", wordBoundary: false, caseInsensitive: true, want: true},
		{name: "case sensitive miss", original: "authToken", keyword: "token", wordBoundary: true, caseInsensitive: false, want: false},
		{name: "case sensitive hit with underscore", original: "auth_token", keyword: "token", wordBoundary: true, caseInsensitive: false, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := tt.original
			normalized := orig
			kw := tt.keyword
			if tt.caseInsensitive {
				normalized = strings.ToLower(orig)
				kw = strings.ToLower(tt.keyword)
			}
			if got := containsKeyword(orig, normalized, kw, tt.wordBoundary, tt.caseInsensitive); got != tt.want {
				t.Fatalf("containsKeyword(%q, %q, %q, %v, %v) = %v, want %v", orig, normalized, kw, tt.wordBoundary, tt.caseInsensitive, got, tt.want)
			}
		})
	}
}

func TestMatchSensitive(t *testing.T) {
	cfg := SensitiveConfig{
		CaseInsensitive: true,
		WordBoundary:    true,
		Keywords:        []string{"token", " password ", ""},
	}

	if kw, ok := matchSensitive("authToken", cfg); !ok || kw != "token" {
		t.Fatalf("matchSensitive(authToken) = %q, %v; want token, true", kw, ok)
	}
	if kw, ok := matchSensitive("passwordHash", cfg); !ok || strings.TrimSpace(kw) != "password" {
		t.Fatalf("matchSensitive(passwordHash) = %q, %v; want password, true", kw, ok)
	}
	if kw, ok := matchSensitive("sessionID", cfg); ok {
		t.Fatalf("matchSensitive(sessionID) = %q, %v; want false", kw, ok)
	}

	cfgCaseSensitive := SensitiveConfig{
		CaseInsensitive: false,
		WordBoundary:    true,
		Keywords:        []string{"token"},
	}

	if _, ok := matchSensitive("authToken", cfgCaseSensitive); ok {
		t.Fatalf("matchSensitive(authToken) should not match when case_insensitive=false")
	}
	if _, ok := matchSensitive("auth_token", cfgCaseSensitive); !ok {
		t.Fatalf("matchSensitive(auth_token) should match when case_insensitive=false and token boundary matches")
	}
}

func TestConcatLiteralsOnly(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
		ok   bool
	}{
		{name: "simple", expr: `"a" + "b"`, want: "ab", ok: true},
		{name: "with ident", expr: `"a" + x + "b"`, want: "ab", ok: true},
		{name: "paren", expr: `("a" + "b") + "c"`, want: "abc", ok: true},
		{name: "no literal", expr: `x + y`, want: "", ok: false},
		{name: "non add", expr: `"a" - "b"`, want: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpr(tt.expr)
			if err != nil {
				t.Fatalf("ParseExpr(%q) error: %v", tt.expr, err)
			}
			got, ok := concatLiteralsOnly(expr)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("concatLiteralsOnly(%q) = %q, %v; want %q, %v", tt.expr, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
