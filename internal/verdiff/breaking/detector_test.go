package breaking

import (
	"context"
	"testing"

	"github.com/txie/verdiff/internal/verdiff"
)

func TestGoExportedFuncRemoved(t *testing.T) {
	input := &verdiff.AnalysisInput{
		Diff: verdiff.DiffResult{
			Files: []verdiff.FileDiff{
				{
					Path:       "pkg/api/handler.go",
					ChangeType: verdiff.ChangeModified,
					Patch: `-func FetchData(ctx context.Context) error {
-	return nil
-}
+func fetchData(ctx context.Context) error {
+	return nil
+}
`,
				},
			},
		},
		Config: verdiff.DefaultConfig(),
	}

	detector := NewDetector()
	out, err := detector.Analyze(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}

	hasBreaking := false
	for _, f := range out.Findings {
		if f.Category == "breaking-change" {
			hasBreaking = true
			break
		}
	}
	if !hasBreaking {
		t.Error("expected breaking change finding for removed exported func")
	}
}

func TestConfigKeyRemoved(t *testing.T) {
	input := &verdiff.AnalysisInput{
		Diff: verdiff.DiffResult{
			Files: []verdiff.FileDiff{
				{
					Path:       "config.yaml",
					ChangeType: verdiff.ChangeModified,
					Patch: `-database_url: postgres://localhost/mydb
-cache_ttl: 300
+database_url: postgres://localhost/mydb
`,
				},
			},
		},
		Config: verdiff.DefaultConfig(),
	}

	detector := NewDetector()
	out, err := detector.Analyze(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, f := range out.Findings {
		if f.Category == "breaking-change" && f.OldValue == "cache_ttl" {
			found = true
		}
	}
	if !found {
		t.Error("expected breaking change for removed config key 'cache_ttl'")
	}
}

func TestEnvVarRemoved(t *testing.T) {
	input := &verdiff.AnalysisInput{
		Diff: verdiff.DiffResult{
			Files: []verdiff.FileDiff{
				{
					Path:       "main.go",
					ChangeType: verdiff.ChangeModified,
					Patch: `-	port := os.Getenv("API_PORT")
+	port := "8080"
`,
				},
			},
		},
		Config: verdiff.DefaultConfig(),
	}

	detector := NewDetector()
	out, err := detector.Analyze(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}

	hasEnvFinding := false
	for _, f := range out.Findings {
		if f.Category == "breaking-change" {
			hasEnvFinding = true
		}
	}
	if !hasEnvFinding {
		t.Error("expected finding for removed env var reference")
	}
}

func TestIgnoreRules(t *testing.T) {
	input := &verdiff.AnalysisInput{
		Diff: verdiff.DiffResult{
			Files: []verdiff.FileDiff{
				{
					Path:       "internal/legacy.go",
					ChangeType: verdiff.ChangeModified,
					Patch:      `-func OldAPI() {}`,
				},
			},
		},
		Config: verdiff.Config{
			Ignore: []verdiff.IgnoreRule{
				{Path: "internal/legacy.go"},
			},
		},
	}

	detector := NewDetector()
	out, err := detector.Analyze(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range out.Findings {
		if f.Category == "breaking-change" {
			for _, fp := range f.FilePaths {
				if fp == "internal/legacy.go" {
					t.Error("finding for internal/legacy.go should be suppressed by ignore rule")
				}
			}
		}
	}
}

func TestNoBreakingChanges(t *testing.T) {
	input := &verdiff.AnalysisInput{
		Diff: verdiff.DiffResult{
			Files: []verdiff.FileDiff{
				{
					Path:       "readme.md",
					ChangeType: verdiff.ChangeModified,
					Patch:      "+Updated docs\n-Old docs\n",
				},
			},
		},
		Config: verdiff.DefaultConfig(),
	}

	detector := NewDetector()
	out, err := detector.Analyze(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Findings) != 0 {
		t.Errorf("expected 0 findings for readme change, got %d", len(out.Findings))
	}
}

func TestMajorBumpFromPriorVersions(t *testing.T) {
	input := &verdiff.AnalysisInput{
		Diff:   verdiff.DiffResult{},
		Config: verdiff.DefaultConfig(),
		PriorVersions: []verdiff.VersionChange{
			{
				Name: "github.com/big/lib", Source: "go.mod",
				OldVersion: "v1.0.0", NewVersion: "v2.0.0",
				BumpType: verdiff.BumpMajor, Direction: verdiff.DirUpgrade,
			},
		},
	}

	detector := NewDetector()
	out, err := detector.Analyze(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Findings) == 0 {
		t.Error("expected finding for major dep bump from prior versions")
	}
}
