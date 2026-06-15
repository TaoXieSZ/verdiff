package deps

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/txie/verdiff/internal/verdiff"
)

// Tracker is an analyzer that detects dependency and version changes.
type Tracker struct{}

func NewTracker() *Tracker { return &Tracker{} }

func (t *Tracker) Name() string { return "dependency-tracker" }

func (t *Tracker) Analyze(ctx context.Context, input *verdiff.AnalysisInput) (*verdiff.AnalysisOutput, error) {
	out := &verdiff.AnalysisOutput{}

	for _, f := range input.Diff.Files {
		parser := matchParser(f.Path)
		if parser == nil {
			continue
		}
		if f.Patch == "" {
			continue
		}
		changes := parser.ExtractChanges(f.Path, f.Patch)
		out.VersionChanges = append(out.VersionChanges, changes...)
	}

	// Apply custom version patterns from config
	for _, vp := range input.Config.VersionPatterns {
		re, err := regexp.Compile(vp.Pattern)
		if err != nil {
			continue
		}
		for _, f := range input.Diff.Files {
			matched, _ := filepath.Match(vp.FileGlob, f.Path)
			if !matched {
				continue
			}
			changes := extractCustomVersions(f, re, vp.Name)
			out.VersionChanges = append(out.VersionChanges, changes...)
		}
	}

	// Flag major version bumps as findings
	for _, vc := range out.VersionChanges {
		if vc.BumpType == verdiff.BumpMajor {
			out.Findings = append(out.Findings, verdiff.Finding{
				Title:       fmt.Sprintf("Major version bump: %s", vc.Name),
				Description: fmt.Sprintf("%s → %s in %s", vc.OldVersion, vc.NewVersion, vc.Source),
				Severity:    verdiff.SeverityWarning,
				Category:    "dependency",
				FilePaths:   []string{vc.Source},
				OldValue:    vc.OldVersion,
				NewValue:    vc.NewVersion,
			})
		}
	}

	return out, nil
}

// Parser extracts version mappings from a specific file type.
type Parser interface {
	ExtractChanges(path, patch string) []verdiff.VersionChange
}

func matchParser(path string) Parser {
	base := filepath.Base(path)
	switch {
	case base == "go.mod":
		return &GoModParser{}
	case base == "package.json":
		return &PackageJSONParser{}
	case base == "Gemfile.lock":
		return &GemfileLockParser{}
	case base == "Policyfile.rb" || base == "Policyfile.lock.json":
		return &PolicyfileParser{}
	case base == "requirements.txt" || base == "pyproject.toml":
		return &RequirementsParser{}
	default:
		return nil
	}
}

// parsePatchLines splits a patch into added and removed lines.
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

// diffVersionMaps compares two name→version maps and produces VersionChanges.
func diffVersionMaps(old, new map[string]string, source string) []verdiff.VersionChange {
	var changes []verdiff.VersionChange
	seen := make(map[string]bool)

	for name, newVer := range new {
		seen[name] = true
		oldVer, existed := old[name]
		if !existed {
			changes = append(changes, verdiff.VersionChange{
				Name: name, Source: source,
				NewVersion: newVer, Direction: verdiff.DirAdded,
			})
		} else if oldVer != newVer {
			dir := verdiff.DirUpgrade
			if compareVersions(newVer, oldVer) < 0 {
				dir = verdiff.DirDowngrade
			}
			changes = append(changes, verdiff.VersionChange{
				Name: name, Source: source,
				OldVersion: oldVer, NewVersion: newVer,
				BumpType: classifyBump(oldVer, newVer), Direction: dir,
			})
		}
	}
	for name, oldVer := range old {
		if !seen[name] {
			changes = append(changes, verdiff.VersionChange{
				Name: name, Source: source,
				OldVersion: oldVer, Direction: verdiff.DirRemoved,
			})
		}
	}

	sort.Slice(changes, func(i, j int) bool { return changes[i].Name < changes[j].Name })
	return changes
}

// classifyBump determines the semver bump type between two version strings.
func classifyBump(oldVer, newVer string) verdiff.SemverBump {
	op := parseSemver(oldVer)
	np := parseSemver(newVer)
	if op == nil || np == nil {
		return verdiff.BumpUnknown
	}
	if op[0] != np[0] {
		return verdiff.BumpMajor
	}
	if op[1] != np[1] {
		return verdiff.BumpMinor
	}
	return verdiff.BumpPatch
}

func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return nil
	}
	result := make([]int, 3)
	for i := 0; i < len(parts) && i < 3; i++ {
		// Strip pre-release suffix
		num := strings.SplitN(parts[i], "-", 2)[0]
		n, err := strconv.Atoi(num)
		if err != nil {
			return nil
		}
		result[i] = n
	}
	return result
}

func compareVersions(a, b string) int {
	pa := parseSemver(a)
	pb := parseSemver(b)
	if pa == nil || pb == nil {
		return strings.Compare(a, b)
	}
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] - pb[i]
		}
	}
	return 0
}

func extractCustomVersions(f verdiff.FileDiff, re *regexp.Regexp, name string) []verdiff.VersionChange {
	added, removed := parsePatchLines(f.Patch)
	oldVersions := extractByRegex(removed, re)
	newVersions := extractByRegex(added, re)
	if len(oldVersions) == 0 && len(newVersions) == 0 {
		return nil
	}

	var changes []verdiff.VersionChange
	oldV := strings.Join(oldVersions, ", ")
	newV := strings.Join(newVersions, ", ")
	if oldV != newV {
		changes = append(changes, verdiff.VersionChange{
			Name:       name + " (" + f.Path + ")",
			Source:     f.Path,
			OldVersion: oldV,
			NewVersion: newV,
			Direction:  verdiff.DirUpgrade,
			BumpType:   classifyBump(oldV, newV),
		})
	}
	return changes
}

func extractByRegex(lines []string, re *regexp.Regexp) []string {
	var results []string
	vIdx := re.SubexpIndex("version")
	for _, line := range lines {
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if vIdx >= 0 && vIdx < len(m) {
			results = append(results, m[vIdx])
		} else if len(m) > 1 {
			results = append(results, m[1])
		}
	}
	return results
}
