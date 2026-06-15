package breaking

import (
	"context"
	"path/filepath"
	"regexp"

	"github.com/txie/verdiff/internal/verdiff"
)

// Detector is an analyzer that identifies candidate breaking changes.
type Detector struct{}

func NewDetector() *Detector { return &Detector{} }

func (d *Detector) Name() string { return "breaking-change-detector" }

func (d *Detector) Analyze(ctx context.Context, input *verdiff.AnalysisInput) (*verdiff.AnalysisOutput, error) {
	out := &verdiff.AnalysisOutput{}

	// Built-in rules
	builtinRules := defaultRules()
	for _, f := range input.Diff.Files {
		if f.Patch == "" || f.ChangeType == verdiff.ChangeAdded {
			continue
		}
		for _, rule := range builtinRules {
			matched, _ := filepath.Match(rule.FileGlob, f.Path)
			if !matched {
				// Also try matching just the base name
				matched, _ = filepath.Match(rule.FileGlob, filepath.Base(f.Path))
			}
			if !matched {
				continue
			}
			findings := rule.Apply(f)
			out.Findings = append(out.Findings, findings...)
		}
	}

	// Custom rules from config
	for _, cr := range input.Config.BreakingRules {
		re, err := regexp.Compile(cr.Pattern)
		if err != nil {
			continue
		}
		rule := &regexRule{
			id:       cr.ID,
			name:     cr.Name,
			fileGlob: cr.FileGlob,
			pattern:  re,
		}
		for _, f := range input.Diff.Files {
			if f.Patch == "" {
				continue
			}
			matched, _ := filepath.Match(rule.fileGlob, f.Path)
			if !matched {
				matched, _ = filepath.Match(rule.fileGlob, filepath.Base(f.Path))
			}
			if !matched {
				continue
			}
			findings := rule.Apply(f)
			out.Findings = append(out.Findings, findings...)
		}
	}

	// Apply ignore rules
	out.Findings = applyIgnoreRules(out.Findings, input.Config.Ignore)

	// Check prior version changes for major bumps → breaking candidates
	for _, vc := range input.PriorVersions {
		if vc.BumpType == verdiff.BumpMajor {
			out.Findings = append(out.Findings, verdiff.Finding{
				Title:       "Major dependency bump: " + vc.Name,
				Description: vc.OldVersion + " → " + vc.NewVersion + " (may contain breaking changes)",
				Severity:    verdiff.SeverityWarning,
				Category:    "breaking-change",
				FilePaths:   []string{vc.Source},
				OldValue:    vc.OldVersion,
				NewValue:    vc.NewVersion,
			})
		}
	}

	return out, nil
}

func applyIgnoreRules(findings []verdiff.Finding, ignores []verdiff.IgnoreRule) []verdiff.Finding {
	if len(ignores) == 0 {
		return findings
	}
	var filtered []verdiff.Finding
	for _, f := range findings {
		ignored := false
		for _, ig := range ignores {
			if ig.Path != "" {
				for _, fp := range f.FilePaths {
					if matched, _ := filepath.Match(ig.Path, fp); matched {
						ignored = true
						break
					}
				}
			}
		}
		if !ignored {
			filtered = append(filtered, f)
		}
	}
	return filtered
}
