package gitdiff

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/txie/verdiff/internal/verdiff"
)

// CLIDiffer uses the git CLI for diff computation, suitable for large repos.
type CLIDiffer struct{}

func NewCLIDiffer() *CLIDiffer {
	return &CLIDiffer{}
}

var numstatRe = regexp.MustCompile(`^(\d+|-)\t(\d+|-)\t(.+)$`)
var renameRe = regexp.MustCompile(`^(.+)\{(.+) => (.+)\}(.*)$`)

func (c *CLIDiffer) Diff(ctx context.Context, repoPath, versionA, versionB string, topN int, pathFilter string) (verdiff.DiffResult, error) {
	args := []string{"diff", "--numstat", "-M", versionA + ".." + versionB}
	if pathFilter != "" {
		args = append(args, "--", pathFilter)
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return verdiff.DiffResult{}, fmt.Errorf("git diff --numstat: %w", err)
	}

	files, err := parseNumstat(string(out))
	if err != nil {
		return verdiff.DiffResult{}, err
	}

	repoName := filepath.Base(repoPath)
	result := verdiff.DiffResult{
		RepoName: repoName,
		VersionA: versionA,
		VersionB: versionB,
		Files:    files,
		Tree:     BuildDirTree(files),
		Hotspots: topHotspots(files, topN),
	}

	return result, nil
}

func parseNumstat(output string) ([]verdiff.FileDiff, error) {
	var files []verdiff.FileDiff
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		m := numstatRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		fd := verdiff.FileDiff{}

		if m[1] == "-" {
			fd.IsBinary = true
		} else {
			fd.LinesAdded, _ = strconv.Atoi(m[1])
			fd.LinesDeleted, _ = strconv.Atoi(m[2])
		}

		path := m[3]
		if rm := renameRe.FindStringSubmatch(path); rm != nil {
			fd.OldPath = rm[1] + rm[2] + rm[4]
			fd.Path = rm[1] + rm[3] + rm[4]
			fd.ChangeType = verdiff.ChangeRenamed
		} else {
			fd.Path = path
			if fd.LinesDeleted > 0 && fd.LinesAdded == 0 {
				fd.ChangeType = verdiff.ChangeDeleted
			} else if fd.LinesAdded > 0 && fd.LinesDeleted == 0 {
				fd.ChangeType = verdiff.ChangeAdded
			} else {
				fd.ChangeType = verdiff.ChangeModified
			}
		}

		files = append(files, fd)
	}
	return files, scanner.Err()
}
