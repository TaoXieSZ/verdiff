package breaking

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/txie/verdiff/internal/verdiff"
)

// Rule detects breaking changes in a file diff.
type Rule interface {
	Apply(f verdiff.FileDiff) []verdiff.Finding
}

// regexRule matches removed lines against a regex and reports them as candidates.
type regexRule struct {
	id       string
	name     string
	fileGlob string
	pattern  *regexp.Regexp
	FileGlob string // exported for defaultRules
}

func (r *regexRule) Apply(f verdiff.FileDiff) []verdiff.Finding {
	_, removed := parsePatchLines(f.Patch)
	var findings []verdiff.Finding
	for _, line := range removed {
		if r.pattern.MatchString(line) {
			match := r.pattern.FindString(line)
			findings = append(findings, verdiff.Finding{
				Title:       fmt.Sprintf("[Candidate] %s: %s", r.name, truncate(match, 80)),
				Description: fmt.Sprintf("Removed in %s", f.Path),
				Severity:    verdiff.SeverityWarning,
				Category:    "breaking-change",
				FilePaths:   []string{f.Path},
				OldValue:    strings.TrimSpace(line),
			})
		}
	}
	return findings
}

func defaultRules() []struct {
	FileGlob string
	Rule
} {
	return []struct {
		FileGlob string
		Rule
	}{
		{
			FileGlob: "*.go",
			Rule: &regexRule{
				id:       "go-exported-func",
				name:     "Go exported function removed",
				fileGlob: "*.go",
				pattern:  regexp.MustCompile(`^func\s+[A-Z]\w*\s*\(`),
			},
		},
		{
			FileGlob: "*.go",
			Rule: &regexRule{
				id:       "go-exported-type",
				name:     "Go exported type removed",
				fileGlob: "*.go",
				pattern:  regexp.MustCompile(`^type\s+[A-Z]\w*\s+(struct|interface)\s*\{`),
			},
		},
		{
			FileGlob: "*.yaml",
			Rule:     &configKeyRule{name: "YAML config key removed"},
		},
		{
			FileGlob: "*.yml",
			Rule:     &configKeyRule{name: "YAML config key removed"},
		},
		{
			FileGlob: "*.json",
			Rule:     &configKeyRule{name: "JSON config key removed"},
		},
		{
			FileGlob: "*.go",
			Rule: &regexRule{
				id:       "env-var-removed",
				name:     "Environment variable reference removed",
				fileGlob: "*.go",
				pattern:  regexp.MustCompile(`os\.Getenv\("([^"]+)"\)`),
			},
		},
		{
			FileGlob: "*.go",
			Rule: &regexRule{
				id:       "flag-removed",
				name:     "CLI flag removed",
				fileGlob: "*.go",
				pattern:  regexp.MustCompile(`flag\.(String|Int|Bool|Float64|Duration)(Var)?\(`),
			},
		},
	}
}

// configKeyRule detects removed top-level config keys.
type configKeyRule struct {
	name string
}

var configKeyRe = regexp.MustCompile(`^(\s*)([a-zA-Z_][a-zA-Z0-9_-]*)\s*:`)

func (r *configKeyRule) Apply(f verdiff.FileDiff) []verdiff.Finding {
	added, removed := parsePatchLines(f.Patch)
	removedKeys := extractConfigKeys(removed)
	addedKeys := extractConfigKeys(added)

	var findings []verdiff.Finding
	for key := range removedKeys {
		if _, stillExists := addedKeys[key]; !stillExists {
			findings = append(findings, verdiff.Finding{
				Title:       fmt.Sprintf("[Candidate] %s: %s", r.name, key),
				Description: fmt.Sprintf("Key '%s' removed from %s", key, f.Path),
				Severity:    verdiff.SeverityWarning,
				Category:    "breaking-change",
				FilePaths:   []string{f.Path},
				OldValue:    key,
			})
		}
	}
	return findings
}

func extractConfigKeys(lines []string) map[string]bool {
	keys := make(map[string]bool)
	for _, line := range lines {
		m := configKeyRe.FindStringSubmatch(line)
		if m != nil && m[1] == "" { // top-level only (no indentation)
			keys[m[2]] = true
		}
	}
	return keys
}

func parsePatchLines(patch string) (added, removed []string) {
	for _, line := range strings.Split(patch, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added = append(added, strings.TrimPrefix(line, "+"))
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removed = append(removed, strings.TrimPrefix(line, "-"))
		}
	}
	return
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
