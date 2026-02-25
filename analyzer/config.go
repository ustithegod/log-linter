package analyzer

type Config struct {
	Rules     RulesConfig     `yaml:"rules"`
	Sensitive SensitiveConfig `yaml:"sensitive"`
}

type RulesConfig struct {
	LowercaseStart bool `yaml:"lowercase_start"`
	EnglishOnly    bool `yaml:"english_only"`
	NoSpecials     bool `yaml:"no_specials"`
	Sensitive      bool `yaml:"sensitive"`
}

type SensitiveConfig struct {
	CaseInsensitive bool     `yaml:"case_insensitive"`
	WordBoundary    bool     `yaml:"word_boundary"`
	Keywords        []string `yaml:"keywords"`
}

func DefaultConfig() Config {
	return Config{
		Rules: RulesConfig{
			LowercaseStart: true,
			EnglishOnly:    true,
			NoSpecials:     true,
			Sensitive:      true,
		},
		Sensitive: SensitiveConfig{
			CaseInsensitive: true,
			WordBoundary:    true,
			Keywords: []string{
				"password",
				"passwd",
				"pass",
				"token",
				"api_key",
				"apikey",
				"secret",
				"private_key",
			},
		},
	}
}

// парсинг конфига из .golangci.yml
func ApplyConfig(cfg Config, settings map[string]any) Config {
	if settings == nil {
		return cfg
	}

	if rules, ok := settings["rules"].(map[string]any); ok {
		if v, ok := parseBool(rules["lowercase_start"]); ok {
			cfg.Rules.LowercaseStart = v
		}
		if v, ok := parseBool(rules["english_only"]); ok {
			cfg.Rules.EnglishOnly = v
		}
		if v, ok := parseBool(rules["no_specials"]); ok {
			cfg.Rules.NoSpecials = v
		}
		if v, ok := parseBool(rules["sensitive"]); ok {
			cfg.Rules.Sensitive = v
		}
	}

	if sens, ok := settings["sensitive"].(map[string]any); ok {
		if v, ok := parseBool(sens["case_insensitive"]); ok {
			cfg.Sensitive.CaseInsensitive = v
		}
		if v, ok := parseBool(sens["word_boundary"]); ok {
			cfg.Sensitive.WordBoundary = v
		}
		if kws, ok := parseStringSlice(sens["keywords"]); ok {
			cfg.Sensitive.Keywords = kws
		}
	}

	return cfg
}

func parseBool(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}

func parseStringSlice(v any) ([]string, bool) {
	switch t := v.(type) {
	case []string:
		return t, true
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			s, ok := item.(string)
			if !ok {
				return nil, false
			}
			out = append(out, s)
		}
		return out, true
	default:
		return nil, false
	}
}
