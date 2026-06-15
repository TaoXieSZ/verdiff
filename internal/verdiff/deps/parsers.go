package deps

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/txie/verdiff/internal/verdiff"
)

// GoModParser extracts dependency versions from go.mod patches.
type GoModParser struct{}

var goModRequireRe = regexp.MustCompile(`^\s*(\S+)\s+(v\S+)`)

func (p *GoModParser) ExtractChanges(path, patch string) []verdiff.VersionChange {
	added, removed := parsePatchLines(patch)
	oldDeps := parseGoModLines(removed)
	newDeps := parseGoModLines(added)
	return diffVersionMaps(oldDeps, newDeps, path)
}

func parseGoModLines(lines []string) map[string]string {
	m := make(map[string]string)
	for _, line := range lines {
		match := goModRequireRe.FindStringSubmatch(line)
		if match != nil {
			mod := match[1]
			ver := match[2]
			if !strings.HasPrefix(mod, "//") {
				m[mod] = ver
			}
		}
	}
	return m
}

// PackageJSONParser extracts dependency versions from package.json patches.
type PackageJSONParser struct{}

func (p *PackageJSONParser) ExtractChanges(path, patch string) []verdiff.VersionChange {
	added, removed := parsePatchLines(patch)
	oldDeps := parseJSONDepLines(removed)
	newDeps := parseJSONDepLines(added)
	return diffVersionMaps(oldDeps, newDeps, path)
}

var jsonDepRe = regexp.MustCompile(`"([^"]+)"\s*:\s*"([^"]+)"`)

func parseJSONDepLines(lines []string) map[string]string {
	m := make(map[string]string)
	for _, line := range lines {
		match := jsonDepRe.FindStringSubmatch(line)
		if match != nil {
			name := match[1]
			ver := match[2]
			if !strings.HasPrefix(name, "@types/") || true {
				m[name] = strings.TrimPrefix(ver, "^")
			}
		}
	}
	return m
}

// GemfileLockParser extracts gem versions from Gemfile.lock patches.
type GemfileLockParser struct{}

var gemLockRe = regexp.MustCompile(`^\s{4}(\S+)\s+\((\S+)\)`)

func (p *GemfileLockParser) ExtractChanges(path, patch string) []verdiff.VersionChange {
	added, removed := parsePatchLines(patch)
	oldGems := parseGemLines(removed)
	newGems := parseGemLines(added)
	return diffVersionMaps(oldGems, newGems, path)
}

func parseGemLines(lines []string) map[string]string {
	m := make(map[string]string)
	for _, line := range lines {
		match := gemLockRe.FindStringSubmatch(line)
		if match != nil {
			m[match[1]] = match[2]
		}
	}
	return m
}

// PolicyfileParser extracts cookbook versions from Policyfile.rb or Policyfile.lock.json.
type PolicyfileParser struct{}

var policyRbCookbookRe = regexp.MustCompile(`cookbook\s+['"](\S+)['"]\s*,\s*['"]([~>=<!\s]*[\d.]+)['"]`)

func (p *PolicyfileParser) ExtractChanges(path, patch string) []verdiff.VersionChange {
	if strings.HasSuffix(path, ".lock.json") {
		return p.extractFromLockJSON(path, patch)
	}
	return p.extractFromRb(path, patch)
}

func (p *PolicyfileParser) extractFromRb(path, patch string) []verdiff.VersionChange {
	added, removed := parsePatchLines(patch)
	oldCB := parsePolicyRbLines(removed)
	newCB := parsePolicyRbLines(added)
	return diffVersionMaps(oldCB, newCB, path)
}

func parsePolicyRbLines(lines []string) map[string]string {
	m := make(map[string]string)
	for _, line := range lines {
		match := policyRbCookbookRe.FindStringSubmatch(line)
		if match != nil {
			m[match[1]] = match[2]
		}
	}
	return m
}

func (p *PolicyfileParser) extractFromLockJSON(path, patch string) []verdiff.VersionChange {
	added, removed := parsePatchLines(patch)
	oldCB := parsePolicyLockLines(removed)
	newCB := parsePolicyLockLines(added)
	return diffVersionMaps(oldCB, newCB, path)
}

func parsePolicyLockLines(lines []string) map[string]string {
	// Policyfile.lock.json has "cookbook_locks" with entries like:
	// "name": { "version": "x.y.z", ... }
	// Since we're parsing patch lines, we look for version fields
	m := make(map[string]string)
	content := strings.Join(lines, "\n")
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err == nil {
		if locks, ok := data["cookbook_locks"].(map[string]interface{}); ok {
			for name, v := range locks {
				if entry, ok := v.(map[string]interface{}); ok {
					if ver, ok := entry["version"].(string); ok {
						m[name] = ver
					}
				}
			}
		}
	}
	// Fallback: line-by-line regex for partial patches
	if len(m) == 0 {
		for _, line := range lines {
			match := jsonDepRe.FindStringSubmatch(line)
			if match != nil {
				m[match[1]] = match[2]
			}
		}
	}
	return m
}

// RequirementsParser extracts Python dependency versions.
type RequirementsParser struct{}

var requirementsRe = regexp.MustCompile(`^([a-zA-Z0-9_-]+)\s*[=<>!~]+\s*(\S+)`)

func (p *RequirementsParser) ExtractChanges(path, patch string) []verdiff.VersionChange {
	added, removed := parsePatchLines(patch)
	oldPkgs := parseRequirementsLines(removed)
	newPkgs := parseRequirementsLines(added)
	return diffVersionMaps(oldPkgs, newPkgs, path)
}

func parseRequirementsLines(lines []string) map[string]string {
	m := make(map[string]string)
	for _, line := range lines {
		match := requirementsRe.FindStringSubmatch(line)
		if match != nil {
			m[match[1]] = match[2]
		}
	}
	return m
}
