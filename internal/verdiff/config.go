package verdiff

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the .verdiff.yaml configuration file.
type Config struct {
	Analyzers       []CustomAnalyzerDef `yaml:"analyzers"`
	VersionPatterns []VersionPatternDef `yaml:"version_patterns"`
	BreakingRules   []BreakingRuleDef   `yaml:"breaking_change_rules"`
	Ignore          []IgnoreRule        `yaml:"ignore"`
	Report          ReportConfig        `yaml:"report"`
}

// CustomAnalyzerDef defines a user-specified analyzer via file glob + regex.
type CustomAnalyzerDef struct {
	Name     string `yaml:"name"`
	FileGlob string `yaml:"file_glob"`
	Pattern  string `yaml:"pattern"`
	Category string `yaml:"category"`
	Severity string `yaml:"severity"`
}

// VersionPatternDef defines a custom version extraction pattern.
type VersionPatternDef struct {
	Name     string `yaml:"name"`
	FileGlob string `yaml:"file_glob"`
	Pattern  string `yaml:"pattern"` // regex with named capture group "version"
}

// BreakingRuleDef defines a custom breaking change detection rule.
type BreakingRuleDef struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	FileGlob string `yaml:"file_glob"`
	Pattern  string `yaml:"pattern"`
}

// IgnoreRule suppresses findings by file path pattern or rule ID.
type IgnoreRule struct {
	Path   string `yaml:"path,omitempty"`
	RuleID string `yaml:"rule_id,omitempty"`
}

// ReportConfig holds report rendering options.
type ReportConfig struct {
	DefaultTheme  string `yaml:"default_theme"` // "light" or "dark"
	DiffLineLimit int    `yaml:"diff_line_limit"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Report: ReportConfig{
			DefaultTheme:  "auto",
			DiffLineLimit: 500,
		},
	}
}

// LoadConfig reads a .verdiff.yaml file and merges with defaults.
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Report.DiffLineLimit == 0 {
		cfg.Report.DiffLineLimit = 500
	}
	return cfg, nil
}
