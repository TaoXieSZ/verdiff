package gitdiff

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/txie/verdiff/internal/verdiff"
)

func testRepoPath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", "test-repo")
}

func TestGoGitDiffer(t *testing.T) {
	repoPath := testRepoPath(t)
	d := NewGoGitDiffer()
	result, err := d.Diff(context.Background(), repoPath, "v1.0.0", "v2.0.0", 10, "")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("expected files in diff, got 0")
	}
	if result.RepoName != "test-repo" {
		t.Errorf("RepoName = %q, want %q", result.RepoName, "test-repo")
	}
	if result.VersionA != "v1.0.0" || result.VersionB != "v2.0.0" {
		t.Errorf("versions = %s..%s, want v1.0.0..v2.0.0", result.VersionA, result.VersionB)
	}

	found := map[string]bool{}
	for _, f := range result.Files {
		found[f.Path] = true
	}
	if !found["added.go"] {
		t.Error("expected added.go in diff")
	}
	if !found["file.txt"] {
		t.Error("expected file.txt in diff")
	}
}

func TestGoGitDiffer_PathFilter(t *testing.T) {
	repoPath := testRepoPath(t)
	d := NewGoGitDiffer()
	result, err := d.Diff(context.Background(), repoPath, "v1.0.0", "v2.0.0", 10, "pkg/")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file with path filter 'pkg/', got %d", len(result.Files))
	}
	if result.Files[0].Path != "pkg/utils.go" {
		t.Errorf("file path = %q, want %q", result.Files[0].Path, "pkg/utils.go")
	}
}

func TestGoGitDiffer_InvalidRef(t *testing.T) {
	repoPath := testRepoPath(t)
	d := NewGoGitDiffer()
	_, err := d.Diff(context.Background(), repoPath, "nonexistent", "v2.0.0", 10, "")
	if err == nil {
		t.Fatal("expected error for invalid ref, got nil")
	}
}

func TestCLIDiffer(t *testing.T) {
	repoPath := testRepoPath(t)
	d := NewCLIDiffer()
	result, err := d.Diff(context.Background(), repoPath, "v1.0.0", "v2.0.0", 10, "")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if len(result.Files) == 0 {
		t.Fatal("expected files in diff, got 0")
	}
}

func TestCLIDiffer_PathFilter(t *testing.T) {
	repoPath := testRepoPath(t)
	d := NewCLIDiffer()
	result, err := d.Diff(context.Background(), repoPath, "v1.0.0", "v2.0.0", 10, "pkg/")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
}

func TestBuildDirTree(t *testing.T) {
	files := []verdiff.FileDiff{
		{Path: "a/b/c.go", LinesAdded: 10, LinesDeleted: 5},
		{Path: "a/b/d.go", LinesAdded: 20, LinesDeleted: 0},
		{Path: "a/e.go", LinesAdded: 3, LinesDeleted: 1},
		{Path: "f.go", LinesAdded: 1, LinesDeleted: 1},
	}
	tree := BuildDirTree(files)

	if tree.FileCount != 4 {
		t.Errorf("root FileCount = %d, want 4", tree.FileCount)
	}
	if tree.LinesAdded != 34 {
		t.Errorf("root LinesAdded = %d, want 34", tree.LinesAdded)
	}

	// "a" should be the first child (most changes)
	if len(tree.Children) == 0 || tree.Children[0].Name != "a" {
		t.Error("expected 'a' as first child of root")
	}
}

func TestTopHotspots(t *testing.T) {
	files := []verdiff.FileDiff{
		{Path: "small.go", LinesAdded: 1},
		{Path: "big.go", LinesAdded: 100, LinesDeleted: 50},
		{Path: "medium.go", LinesAdded: 20},
	}
	hot := topHotspots(files, 2)
	if len(hot) != 2 {
		t.Fatalf("expected 2 hotspots, got %d", len(hot))
	}
	if hot[0].Path != "big.go" {
		t.Errorf("top hotspot = %q, want big.go", hot[0].Path)
	}
}
