package analyzer

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const docstr = `Linter for log messages.
Supports slog.Logger, zap.Logger, and package-level functions from slog. DOES NOT support zap.SugaredLogger.
Rules:
1. Log messages must start with a lowercase letter;
2. Log messages must be in English only;
3. Log messages must not contain special symbols or emoji;
4. Log messages must not contain potentially sensitive data (by keywords).
Config:
You can enable/disable rules and customize sensitive keywords via a config that will be passed from golangci-lint.
`

func New() *analysis.Analyzer {
	return NewWithConfig(DefaultConfig())
}

func NewWithConfig(cfg Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "logmessagelint",
		Doc:      docstr,
		Run:      func(pass *analysis.Pass) (any, error) { return runWithConfig(pass, cfg) },
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}
