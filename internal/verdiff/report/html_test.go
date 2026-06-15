package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/txie/verdiff/internal/verdiff"
)

func TestWriteHTML(t *testing.T) {
	result := &verdiff.AnalysisResult{
		Diff: verdiff.DiffResult{
			RepoName: "test-repo",
			VersionA: "v1.0.0",
			VersionB: "v2.0.0",
			Files: []verdiff.FileDiff{
				{Path: "main.go", ChangeType: verdiff.ChangeModified, LinesAdded: 10, LinesDeleted: 5},
				{Path: "new.go", ChangeType: verdiff.ChangeAdded, LinesAdded: 50},
				{Path: "old.go", ChangeType: verdiff.ChangeDeleted, LinesDeleted: 30},
			},
			Hotspots: []verdiff.FileDiff{
				{Path: "new.go", ChangeType: verdiff.ChangeAdded, LinesAdded: 50},
			},
			Tree: &verdiff.DirNode{Name: ".", Path: ".", FileCount: 3, LinesAdded: 60, LinesDeleted: 35},
		},
		VersionChanges: []verdiff.VersionChange{
			{Name: "github.com/foo/bar", Source: "go.mod", OldVersion: "v1.0.0", NewVersion: "v2.0.0", BumpType: verdiff.BumpMajor, Direction: verdiff.DirUpgrade},
		},
		Findings: []verdiff.Finding{
			{Title: "Removed func", Category: "breaking-change", Severity: verdiff.SeverityWarning, FilePaths: []string{"main.go"}},
			{Title: "New feature", Category: "info", Severity: verdiff.SeverityInfo},
		},
	}

	dir := t.TempDir()
	outPath := filepath.Join(dir, "report.html")

	err := WriteHTML(outPath, result, verdiff.DefaultConfig())
	if err != nil {
		t.Fatalf("WriteHTML failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	html := string(data)

	checks := []string{
		"test-repo",
		"v1.0.0",
		"v2.0.0",
		"Hotspots",
		"Dependency Changes",
		"Breaking Change Candidates",
		"All Files",
		"main.go",
		"new.go",
		"old.go",
		"github.com/foo/bar",
		"toggleTheme",
		"filterFiles",
		"<!DOCTYPE html>",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("HTML report missing expected content: %q", check)
		}
	}
}

func TestWriteHTML_EmptySections(t *testing.T) {
	result := &verdiff.AnalysisResult{
		Diff: verdiff.DiffResult{
			RepoName: "empty-repo",
			VersionA: "v1",
			VersionB: "v2",
			Tree:     &verdiff.DirNode{Name: "."},
		},
	}

	dir := t.TempDir()
	outPath := filepath.Join(dir, "empty.html")

	err := WriteHTML(outPath, result, verdiff.DefaultConfig())
	if err != nil {
		t.Fatalf("WriteHTML failed: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	html := string(data)

	if strings.Contains(html, "Breaking Change Candidates") {
		t.Error("empty report should not show Breaking Change section")
	}
}
