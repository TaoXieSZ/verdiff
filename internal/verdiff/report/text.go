package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/txie/verdiff/internal/verdiff"
)

// WriteText outputs a plain-text summary of the analysis result.
func WriteText(w io.Writer, result *verdiff.AnalysisResult) error {
	d := result.Diff
	fmt.Fprintf(w, "Version Diff: %s  %s..%s\n", d.RepoName, d.VersionA, d.VersionB)
	fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 60))

	// Overview
	totalAdded, totalDeleted := 0, 0
	for _, f := range d.Files {
		totalAdded += f.LinesAdded
		totalDeleted += f.LinesDeleted
	}
	fmt.Fprintf(w, "Files changed: %d\n", len(d.Files))
	fmt.Fprintf(w, "Lines added:   +%d\n", totalAdded)
	fmt.Fprintf(w, "Lines deleted: -%d\n", totalDeleted)
	fmt.Fprintf(w, "\n")

	// Hotspots
	if len(d.Hotspots) > 0 {
		fmt.Fprintf(w, "Top Changed Files:\n")
		for i, h := range d.Hotspots {
			fmt.Fprintf(w, "  %2d. +%-5d -%5d  %s\n", i+1, h.LinesAdded, h.LinesDeleted, h.Path)
		}
		fmt.Fprintf(w, "\n")
	}

	// Version changes
	if len(result.VersionChanges) > 0 {
		fmt.Fprintf(w, "Dependency Changes:\n")
		for _, vc := range result.VersionChanges {
			switch vc.Direction {
			case verdiff.DirAdded:
				fmt.Fprintf(w, "  + %-40s %s (new)\n", vc.Name, vc.NewVersion)
			case verdiff.DirRemoved:
				fmt.Fprintf(w, "  - %-40s %s (removed)\n", vc.Name, vc.OldVersion)
			default:
				marker := ""
				if vc.BumpType == verdiff.BumpMajor {
					marker = " ⚠ MAJOR"
				}
				fmt.Fprintf(w, "  ~ %-40s %s → %s%s\n", vc.Name, vc.OldVersion, vc.NewVersion, marker)
			}
		}
		fmt.Fprintf(w, "\n")
	}

	// Findings
	breakingFindings := filterByCategory(result.Findings, "breaking-change")
	if len(breakingFindings) > 0 {
		fmt.Fprintf(w, "Breaking Change Candidates (%d):\n", len(breakingFindings))
		for _, f := range breakingFindings {
			fmt.Fprintf(w, "  [%s] %s\n", strings.ToUpper(string(f.Severity)), f.Title)
			if f.Description != "" {
				fmt.Fprintf(w, "         %s\n", f.Description)
			}
		}
		fmt.Fprintf(w, "\n")
	}

	otherFindings := filterNotCategory(result.Findings, "breaking-change")
	if len(otherFindings) > 0 {
		fmt.Fprintf(w, "Other Findings (%d):\n", len(otherFindings))
		for _, f := range otherFindings {
			fmt.Fprintf(w, "  [%s] %s\n", strings.ToUpper(string(f.Severity)), f.Title)
		}
	}

	return nil
}

func filterByCategory(findings []verdiff.Finding, cat string) []verdiff.Finding {
	var out []verdiff.Finding
	for _, f := range findings {
		if f.Category == cat {
			out = append(out, f)
		}
	}
	return out
}

func filterNotCategory(findings []verdiff.Finding, cat string) []verdiff.Finding {
	var out []verdiff.Finding
	for _, f := range findings {
		if f.Category != cat {
			out = append(out, f)
		}
	}
	return out
}
