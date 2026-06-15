package gitdiff

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/txie/verdiff/internal/verdiff"
)

// Differ computes the diff between two versions of a git repository.
type Differ interface {
	Diff(ctx context.Context, repoPath, versionA, versionB string, topN int, pathFilter string) (verdiff.DiffResult, error)
}

// GoGitDiffer uses go-git for diff computation.
type GoGitDiffer struct{}

func NewGoGitDiffer() *GoGitDiffer {
	return &GoGitDiffer{}
}

func (g *GoGitDiffer) Diff(ctx context.Context, repoPath, versionA, versionB string, topN int, pathFilter string) (verdiff.DiffResult, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return verdiff.DiffResult{}, fmt.Errorf("open repo: %w", err)
	}

	commitA, err := resolveRef(repo, versionA)
	if err != nil {
		return verdiff.DiffResult{}, fmt.Errorf("resolve %s: %w", versionA, err)
	}
	commitB, err := resolveRef(repo, versionB)
	if err != nil {
		return verdiff.DiffResult{}, fmt.Errorf("resolve %s: %w", versionB, err)
	}

	treeA, err := commitA.Tree()
	if err != nil {
		return verdiff.DiffResult{}, fmt.Errorf("tree for %s: %w", versionA, err)
	}
	treeB, err := commitB.Tree()
	if err != nil {
		return verdiff.DiffResult{}, fmt.Errorf("tree for %s: %w", versionB, err)
	}

	changes, err := object.DiffTreeWithOptions(ctx, treeA, treeB, &object.DiffTreeOptions{
		DetectRenames: true,
	})
	if err != nil {
		return verdiff.DiffResult{}, fmt.Errorf("diff trees: %w", err)
	}

	files, err := changesToFileDiffs(changes)
	if err != nil {
		return verdiff.DiffResult{}, err
	}

	if pathFilter != "" {
		files = filterByPath(files, pathFilter)
	}

	repoName := filepath.Base(repoPath)
	result := verdiff.DiffResult{
		RepoName: repoName,
		VersionA: versionA,
		VersionB: versionB,
		Files:    files,
	}

	result.Tree = BuildDirTree(files)
	result.Hotspots = topHotspots(files, topN)

	return result, nil
}

// resolveRef resolves a version string (tag, branch, or commit hash) to a commit.
func resolveRef(repo *git.Repository, ref string) (*object.Commit, error) {
	// Try as a full reference (tag or branch)
	for _, prefix := range []string{"refs/tags/", "refs/heads/", ""} {
		h, err := repo.ResolveRevision(plumbing.Revision(prefix + ref))
		if err == nil {
			return repo.CommitObject(*h)
		}
	}
	return nil, fmt.Errorf("cannot resolve ref %q", ref)
}

func changesToFileDiffs(changes object.Changes) ([]verdiff.FileDiff, error) {
	var files []verdiff.FileDiff
	for _, c := range changes {
		fd, err := changeToFileDiff(c)
		if err != nil {
			return nil, err
		}
		files = append(files, fd)
	}
	return files, nil
}

func changeToFileDiff(c *object.Change) (verdiff.FileDiff, error) {
	action, err := c.Action()
	if err != nil {
		return verdiff.FileDiff{}, err
	}

	fd := verdiff.FileDiff{}

	switch action {
	case merkletrie.Insert:
		fd.Path = c.To.Name
		fd.ChangeType = verdiff.ChangeAdded
	case merkletrie.Delete:
		fd.Path = c.From.Name
		fd.ChangeType = verdiff.ChangeDeleted
	case merkletrie.Modify:
		fd.Path = c.To.Name
		if c.From.Name != c.To.Name {
			fd.ChangeType = verdiff.ChangeRenamed
			fd.OldPath = c.From.Name
		} else {
			fd.ChangeType = verdiff.ChangeModified
		}
	}

	patch, err := c.Patch()
	if err != nil {
		// Binary file or other issue — record what we can
		fd.IsBinary = true
		return fd, nil
	}

	fd.Patch = patch.String()
	for _, fp := range patch.FilePatches() {
		if fp.IsBinary() {
			fd.IsBinary = true
			continue
		}
		for _, chunk := range fp.Chunks() {
			lines := strings.Split(chunk.Content(), "\n")
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}
			switch chunk.Type() {
			case diff.Add:
				fd.LinesAdded += len(lines)
			case diff.Delete:
				fd.LinesDeleted += len(lines)
			}
		}
	}

	return fd, nil
}

func filterByPath(files []verdiff.FileDiff, prefix string) []verdiff.FileDiff {
	var filtered []verdiff.FileDiff
	for _, f := range files {
		if strings.HasPrefix(f.Path, prefix) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func topHotspots(files []verdiff.FileDiff, n int) []verdiff.FileDiff {
	sorted := make([]verdiff.FileDiff, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalChange() > sorted[j].TotalChange()
	})
	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}
