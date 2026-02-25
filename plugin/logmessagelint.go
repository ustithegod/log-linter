package plugin

import (
	"github.com/ustithegod/log-linter/analyzer"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("logmessagelint", New)
}

type modulePlugin struct {
	cfg analyzer.Config
}

// New is the module plugin entrypoint for golangci-lint.
// settings comes from linters.settings.custom.<linter>.settings in golangci-lint config.
func New(settings any) (register.LinterPlugin, error) {
	cfg := analyzer.DefaultConfig()
	if m, ok := settings.(map[string]any); ok {
		cfg = analyzer.ApplyConfig(cfg, m)
	}
	return &modulePlugin{cfg: cfg}, nil
}

func (p *modulePlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{analyzer.NewWithConfig(p.cfg)}, nil
}

func (p *modulePlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
