package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/txie/verdiff/internal/verdiff"
	"github.com/txie/verdiff/internal/verdiff/breaking"
	"github.com/txie/verdiff/internal/verdiff/deps"
	"github.com/txie/verdiff/internal/verdiff/gitdiff"
	"github.com/txie/verdiff/internal/verdiff/report"
)

func main() {
	var (
		repoPath   string
		output     string
		configPath string
		format     string
		useGitCLI  bool
		topN       int
		pathFilter string
	)

	flag.StringVar(&repoPath, "repo", ".", "path to the git repository")
	flag.StringVar(&output, "output", "", "output file path (default: auto-generated)")
	flag.StringVar(&configPath, "config", ".verdiff.yaml", "path to configuration file")
	flag.StringVar(&format, "format", "html", "output format: html or text")
	flag.BoolVar(&useGitCLI, "use-git-cli", false, "use git CLI instead of go-git")
	flag.IntVar(&topN, "top", 10, "number of hotspot files to highlight")
	flag.StringVar(&pathFilter, "path", "", "limit diff to a specific directory (e.g. chef/, internal/)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: verdiff [flags] <version-a> <version-b>\n\n")
		fmt.Fprintf(os.Stderr, "Analyze differences between two git versions and generate a structured report.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  verdiff v0.30.1 v0.37.0\n")
		fmt.Fprintf(os.Stderr, "  verdiff --repo /path/to/repo main release/v2\n")
		fmt.Fprintf(os.Stderr, "  verdiff --format text HEAD~10 HEAD\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	versionA := flag.Arg(0)
	versionB := flag.Arg(1)

	if err := run(repoPath, versionA, versionB, output, configPath, format, useGitCLI, topN, pathFilter); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(repoPath, versionA, versionB, output, configPath, format string, useGitCLI bool, topN int, pathFilter string) error {
	ctx := context.Background()
	start := time.Now()

	// Load config
	cfg, err := verdiff.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config %s: %v\n", configPath, err)
		cfg = verdiff.DefaultConfig()
	}

	// Resolve absolute repo path
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("resolve repo path: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Analyzing %s..%s in %s\n", versionA, versionB, absRepo)

	// Run git diff engine
	var differ gitdiff.Differ
	if useGitCLI {
		differ = gitdiff.NewCLIDiffer()
	} else {
		differ = gitdiff.NewGoGitDiffer()
	}

	fmt.Fprintf(os.Stderr, "  Computing diff...\n")
	diff, err := differ.Diff(ctx, absRepo, versionA, versionB, topN, pathFilter)
	if err != nil {
		return fmt.Errorf("git diff: %w", err)
	}

	// Register and run analyzers
	reg := verdiff.NewRegistry()
	reg.Register(deps.NewTracker())
	reg.Register(breaking.NewDetector())

	fmt.Fprintf(os.Stderr, "  Running analyzers...\n")
	result, err := reg.RunAll(ctx, diff, cfg)
	if err != nil {
		return fmt.Errorf("analysis: %w", err)
	}

	// Generate output
	elapsed := time.Since(start)
	fmt.Fprintf(os.Stderr, "  Analysis complete in %s\n", elapsed.Round(time.Millisecond))

	switch format {
	case "text":
		return report.WriteText(os.Stdout, result)
	case "html":
		if output == "" {
			repoName := filepath.Base(absRepo)
			output = fmt.Sprintf("verdiff-%s-%s-%s.html", repoName, sanitize(versionA), sanitize(versionB))
		}
		if err := report.WriteHTML(output, result, cfg); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "  Report written to %s\n", output)
		return nil
	default:
		return fmt.Errorf("unknown format: %s (use html or text)", format)
	}
}

func sanitize(s string) string {
	r := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '.' {
			r = append(r, c)
		} else {
			r = append(r, '_')
		}
	}
	return string(r)
}
